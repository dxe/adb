package members

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/mail"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc"
	"github.com/dxe/adb/config"
	"github.com/dxe/adb/mailer"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

// Website member auth: Google sign-in (Identity Services) or an emailed magic link, yielding a
// session cookie on members.directactioneverywhere.com that the website calls
// cross-origin. This is separate from the members.dxesf.org flow in auth.go.

const (
	fallbackNextURL = "https://www.directactioneverywhere.com/account"
	maxNextURLLen   = 512

	emailsPerAddressPerHour = 5
	emailsPerIPPerHour      = 30
)

// siteOrigins are the web origins that may be redirected to after sign-in and
// may call the API cross-origin. Localhost origins are added outside prod only.
var siteOrigins = map[string]bool{
	"https://www.directactioneverywhere.com":     true,
	"https://directactioneverywhere.com":         true,
	"https://staging.directactioneverywhere.com": true,
}

var devOrigins = map[string]bool{
	"http://localhost:3000": true,
	"http://localhost:4173": true,
}

func isAllowedOrigin(origin string) bool {
	return siteOrigins[origin] || (devLogin() && devOrigins[origin])
}

// devLogin enables the local-development conveniences: the fake Google
// credential, logging magic links instead of emailing them, and localhost
// origins. It needs an explicit MEMBERS_DEV_LOGIN=true (server/debug.env) as
// well as !IsProd, so a production task that lost PROD=true (IsProd defaults
// to false) doesn't hand out sessions or log sign-in links.
func devLogin() bool {
	return !config.IsProd && config.MembersDevLogin
}

// authBaseURL is this server's public origin for the website auth flow.
func authBaseURL() string {
	if config.IsProd {
		return "https://members.directactioneverywhere.com"
	}
	return "http://localhost:" + config.MembersPort
}

// validateNext returns raw if it is an absolute http(s) URL whose origin is in
// the allowlist, and fallbackNextURL otherwise.
func validateNext(raw string) string {
	if raw == "" || len(raw) > maxNextURLLen {
		return fallbackNextURL
	}
	for _, r := range raw {
		// Reject control characters, whitespace and backslashes, which
		// browsers treat inconsistently with url.Parse.
		if r <= ' ' || r == 0x7f || r == '\\' {
			return fallbackNextURL
		}
	}
	u, err := url.Parse(raw)
	if err != nil || u.User != nil || u.Opaque != "" || u.Host == "" {
		return fallbackNextURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fallbackNextURL
	}
	if !isAllowedOrigin(u.Scheme + "://" + u.Host) {
		return fallbackNextURL
	}
	return raw
}

// originAllowed reports whether a request's Origin header is acceptable: absent
// (non-browser client) or in the allowlist.
func originAllowed(origin string) bool {
	return origin == "" || isAllowedOrigin(origin)
}

// setCORS echoes an allowlisted Origin with credentials allowed. It returns
// false if the request carries an Origin that is not allowlisted.
func setCORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Add("Vary", "Origin")
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if !isAllowedOrigin(origin) {
		return false
	}
	w.Header().Set("Access-Control-Allow-Origin", origin)
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	return true
}

// corsHandler wraps h with CORS handling and OPTIONS preflight support, and
// rejects non-allowlisted origins on state-changing requests (CSRF).
func corsHandler(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ok := setCORS(w, r)
		if r.Method == http.MethodOptions {
			if !ok {
				w.WriteHeader(http.StatusForbidden)
				return
			}
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
			w.Header().Set("Access-Control-Max-Age", "600")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		if !ok && r.Method != http.MethodGet {
			http.Error(w, "forbidden origin", http.StatusForbidden)
			return
		}
		h(w, r)
	}
}

// clientIP returns the client address as seen by the ALB. The ALB appends the
// connecting client's address to X-Forwarded-For, so the last entry is the one
// a client cannot spoof.
func clientIP(r *http.Request) string {
	// Use the last entry of the last header line, so a client-sent X-Forwarded-For
	// (even a duplicated header) cannot win over the entry the ALB appended.
	if lines := r.Header.Values("X-Forwarded-For"); len(lines) > 0 {
		parts := strings.Split(lines[len(lines)-1], ",")
		if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
			return ip
		}
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

// rateLimiter is a simple in-memory sliding-window limiter (per process).
type rateLimiter struct {
	mu   sync.Mutex
	hits map[string][]time.Time
}

func newRateLimiter() *rateLimiter { return &rateLimiter{hits: map[string][]time.Time{}} }

// allow records a hit for key and reports whether it is within limit hits per window.
func (l *rateLimiter) allow(key string, limit int, window time.Duration, now time.Time) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	cutoff := now.Add(-window)
	kept := l.hits[key][:0]
	for _, t := range l.hits[key] {
		if t.After(cutoff) {
			kept = append(kept, t)
		}
	}
	if len(kept) >= limit {
		l.hits[key] = kept
		return false
	}
	l.hits[key] = append(kept, now)
	if len(l.hits) > 10000 { // Bound memory: drop idle keys.
		for k, ts := range l.hits {
			if len(ts) == 0 || !ts[len(ts)-1].After(cutoff) {
				delete(l.hits, k)
			}
		}
	}
	return true
}

// normalizeEmail lowercases and validates a bare email address.
func normalizeEmail(s string) (string, bool) {
	s = strings.TrimSpace(s)
	if s == "" || len(s) > 254 {
		return "", false
	}
	// ASCII only: non-ASCII (accents, zero-width characters) can collide with
	// other addresses under the database's accent-insensitive collation.
	for i := 0; i < len(s); i++ {
		if s[i] >= 0x80 {
			return "", false
		}
	}
	addr, err := mail.ParseAddress(s)
	if err != nil || addr.Address != s || addr.Name != "" {
		return "", false
	}
	return strings.ToLower(addr.Address), true
}

type authServer struct {
	store memberStore
	// emailLimiter limits per address, ipLimiter per client IP.
	emailLimiter, ipLimiter *rateLimiter
	now                     func() time.Time
	sendEmail               func(to, link string) error
	verifyGoogle            func(ctx context.Context, credential string) (email string, emailVerified bool, nonce string, err error)
}

// routeAuth registers the website member auth and API endpoints. legacyLogin
// serves GET /login for the members.dxesf.org flow when the request isn't for
// the website sign-in page (see isWebsiteLoginRequest).
func routeAuth(r *mux.Router, db *sqlx.DB, legacyLogin http.HandlerFunc) {
	a := &authServer{
		store:        memberStore{db},
		emailLimiter: newRateLimiter(),
		ipLimiter:    newRateLimiter(),
		now:          time.Now,
		sendEmail:    sendLoginEmail,
		verifyGoogle: verifyGoogleCredential,
	}
	r.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		if isWebsiteLoginRequest(r) {
			a.loginPage(w, r)
		} else {
			legacyLogin(w, r)
		}
	}).Methods(http.MethodGet, http.MethodHead)
	r.PathPrefix("/login/assets/").Handler(loginAssetsHandler()).Methods(http.MethodGet, http.MethodHead)
	routeAuthHandlers(r, a)
}

// routeAuthHandlers registers everything but /login.
func routeAuthHandlers(r *mux.Router, a *authServer) {
	// Same-origin only (the sign-in page): no CORS, so browser scripts on other
	// sites can't redeem tokens or trigger sign-in emails. Non-browser clients
	// can forge Origin; tokens are protected by the nonce and by only this
	// origin being authorized on the Google client.
	r.HandleFunc("/auth/google/token", a.googleToken).Methods(http.MethodPost)
	r.HandleFunc("/auth/email", a.emailStart).Methods(http.MethodPost)
	r.HandleFunc("/auth/email/confirm", a.emailConfirmPage).Methods(http.MethodGet)
	r.HandleFunc("/auth/email/confirm", a.emailConfirm).Methods(http.MethodPost)
	r.HandleFunc("/auth/logout", corsHandler(a.logout)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/me", corsHandler(a.me)).Methods(http.MethodGet, http.MethodOptions)
}

func clearCookie(w http.ResponseWriter, name, path string) {
	http.SetCookie(w, &http.Cookie{
		Name:     name,
		Value:    "",
		Path:     path,
		MaxAge:   -1,
		HttpOnly: true,
		Secure:   config.IsProd,
		SameSite: http.SameSiteLaxMode,
	})
}

func denyFraming(w http.ResponseWriter) {
	w.Header().Set("X-Frame-Options", "DENY")
	w.Header().Set("Content-Security-Policy", "frame-ancestors 'none'")
}

func noStore(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}

// startSession creates a session for email and redirects to next.
func (a *authServer) startSession(w http.ResponseWriter, r *http.Request, email, next string, status int) {
	if err := saveSession(w, r, email, a.now()); err != nil {
		log.Println("member auth: save session:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	noStore(w)
	http.Redirect(w, r, next, status)
}

// --- Google (Google Identity Services ID token) ---

const devCredential = "dev"

// googleVerifier checks ID tokens against the website's own OAuth client
// (config.WebsiteGoogleClientID, "DxE website sign-in"), separate from the ADB
// admin login's client so tokens issued for one aren't accepted by the other.
// The sign-in page renders Google's button and POSTs the credential here.
var googleVerifier = func() *oidc.IDTokenVerifier {
	if config.WebsiteGoogleClientID == "" {
		log.Println("WARNING: WEBSITE_GOOGLE_CLIENT_ID is not set; website Google sign-in is disabled.")
		return nil
	}
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Println(
			"WARNING: failed to construct OIDC provider (no Internet connection?). Some features may not work. Error:",
			err,
		)
		return nil
	}
	return provider.Verifier(&oidc.Config{ClientID: config.WebsiteGoogleClientID})
}()

// verifyGoogleCredential verifies a Google ID token and returns its email and
// nonce claim.
func verifyGoogleCredential(ctx context.Context, credential string) (email string, emailVerified bool, nonce string, err error) {
	if googleVerifier == nil {
		return "", false, "", errors.New("google verifier unavailable")
	}
	idToken, err := googleVerifier.Verify(ctx, credential)
	if err != nil {
		return "", false, "", err
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
		Nonce         string `json:"nonce"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", false, "", err
	}
	return claims.Email, claims.EmailVerified, claims.Nonce, nil
}

// googleToken signs in with a Google ID token ("credential") from Google
// Identity Services and responds with the same JSON as /api/me.
func (a *authServer) googleToken(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	if !sameOrigin(w, r) {
		return
	}
	// The nonce cookie is single-use: cleared whatever happens next.
	clearCookie(w, nonceCookieName(), "/")
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
		return
	}
	var body struct {
		Credential string `json:"credential"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 16384)).Decode(&body); err != nil || body.Credential == "" {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	var rawEmail string
	if devLogin() && body.Credential == devCredential {
		// Mirror the existing members flow, which fakes Google in dev.
		rawEmail = "test-dev@directactioneverywhere.com"
	} else {
		email, verified, nonce, err := a.verifyGoogle(r.Context(), body.Credential)
		if err != nil {
			http.Error(w, "invalid or expired credential", http.StatusUnauthorized)
			return
		}
		if !verified {
			http.Error(w, "your Google email address is not verified", http.StatusUnauthorized)
			return
		}
		// Binds the token to the browser that loaded the sign-in page.
		if !checkNonce(r, nonce) {
			http.Error(w, "sign-in expired; reload the page and try again", http.StatusUnauthorized)
			return
		}
		rawEmail = email
	}
	email, ok := normalizeEmail(rawEmail)
	if !ok {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}
	if err := saveSession(w, r, email, a.now()); err != nil {
		log.Println("member auth: save session:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	a.writeMe(w, r, email)
}

// --- Email magic link ---

func (a *authServer) emailStart(w http.ResponseWriter, r *http.Request) {
	if !sameOrigin(w, r) {
		return
	}
	if !strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
		http.Error(w, "expected application/json", http.StatusUnsupportedMediaType)
		return
	}
	var body struct {
		Email string `json:"email"`
		Next  string `json:"next"`
	}
	if err := json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}
	email, ok := normalizeEmail(body.Email)
	if !ok {
		http.Error(w, "invalid email address", http.StatusBadRequest)
		return
	}
	now := a.now()
	if !a.ipLimiter.allow(clientIP(r), emailsPerIPPerHour, time.Hour, now) {
		http.Error(w, "too many requests", http.StatusTooManyRequests)
		return
	}
	// Over the per-address limit: silently do nothing, same response.
	if a.emailLimiter.allow(email, emailsPerAddressPerHour, time.Hour, now) {
		next := validateNext(body.Next)
		raw, err := a.store.createLoginToken(r.Context(), email, next, now)
		if err != nil {
			log.Println("member auth: create login token:", err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		link := authBaseURL() + "/auth/email/confirm?token=" + url.QueryEscape(raw)
		go func() {
			if err := a.sendEmail(email, link); err != nil {
				log.Println("member auth: send login email:", err)
			}
		}()
		go func() {
			if err := a.store.deleteExpired(context.Background(), now.Add(-24*time.Hour)); err != nil {
				log.Println("member auth: cleanup:", err)
			}
		}()
	}
	w.WriteHeader(http.StatusNoContent)
}

func sendLoginEmail(to, link string) error {
	if devLogin() {
		log.Printf("member auth: sign-in link for %s: %s", to, link)
		return nil
	}
	if config.SMTPHost == "" {
		// Never log the link in prod: it is a login credential.
		return errors.New("SMTP is not configured")
	}
	body := fmt.Sprintf(
		`<p>Click the link below to sign in to Direct Action Everywhere. It expires in 15 minutes and can only be used once.</p><p><a href="%s">Sign in</a></p><p>If you didn't request this, you can ignore this email.</p>`,
		template.HTMLEscapeString(link))
	return mailer.Send(mailer.Message{
		FromName:    "Direct Action Everywhere",
		FromAddress: config.SurveyFromEmail,
		ToAddress:   to,
		Subject:     "Your DxE sign-in link",
		BodyHTML:    body,
	})
}

var confirmTmpl = template.Must(template.New("confirm").Parse(`<!doctype html>
<html lang="en"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta name="robots" content="noindex">
<title>Sign in to DxE</title>
<style>body{font:16px/1.5 system-ui,sans-serif;max-width:28rem;margin:4rem auto;padding:0 1rem}
button{font:inherit;padding:.6rem 1.2rem;cursor:pointer}</style></head><body>
{{if .Email}}<h1>Sign in</h1>
<p>Sign in as <strong>{{.Email}}</strong>?</p>
<form method="post" action="/auth/email/confirm">
<input type="hidden" name="token" value="{{.Token}}">
<button type="submit">Sign in</button></form>
{{else}}<h1>Link expired</h1>
<p>This sign-in link is invalid, expired, or has already been used. Please request a new one.</p>{{end}}
</body></html>`))

func (a *authServer) emailConfirmPage(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	email, err := a.store.peekLoginToken(r.Context(), token, a.now())
	data := struct{ Email, Token string }{}
	status := http.StatusOK
	if err == nil {
		data.Email, data.Token = email, token
	} else if err == errInvalidToken {
		status = http.StatusGone
	} else {
		log.Println("member auth: peek token:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	noStore(w)
	denyFraming(w)
	// Keeps the token in this URL from leaking to other sites. Not
	// "no-referrer": with that, browsers send `Origin: null` on the form's own
	// POST, which emailConfirm's origin check rejects.
	w.Header().Set("Referrer-Policy", "same-origin")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_ = confirmTmpl.Execute(w, data)
}

func (a *authServer) emailConfirm(w http.ResponseWriter, r *http.Request) {
	// The form is served from this host, so its own origin is acceptable too.
	if origin := r.Header.Get("Origin"); !originAllowed(origin) && origin != authBaseURL() {
		http.Error(w, "forbidden origin", http.StatusForbidden)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, 4096)
	email, next, err := a.store.consumeLoginToken(r.Context(), r.PostFormValue("token"), a.now())
	if err == errInvalidToken {
		noStore(w)
		denyFraming(w)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusGone)
		_ = confirmTmpl.Execute(w, struct{ Email, Token string }{})
		return
	}
	if err != nil {
		log.Println("member auth: consume token:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	a.startSession(w, r, email, validateNext(next), http.StatusSeeOther)
}

// --- Logout ---

func (a *authServer) logout(w http.ResponseWriter, r *http.Request) {
	clearCookie(w, sessionCookieName(), "/")
	w.WriteHeader(http.StatusNoContent)
}
