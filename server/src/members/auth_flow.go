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

// allowedOrigins are the web origins that may be redirected to after sign-in
// and may call the API cross-origin.
var allowedOrigins = map[string]bool{
	"https://www.directactioneverywhere.com":     true,
	"https://directactioneverywhere.com":         true,
	"https://staging.directactioneverywhere.com": true,
	"http://localhost:3000":                      true,
	"http://localhost:4173":                      true,
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
	if !allowedOrigins[u.Scheme+"://"+u.Host] {
		return fallbackNextURL
	}
	return raw
}

// originAllowed reports whether a request's Origin header is acceptable: absent
// (non-browser client) or in the allowlist.
func originAllowed(origin string) bool {
	return origin == "" || allowedOrigins[origin]
}

// setCORS echoes an allowlisted Origin with credentials allowed. It returns
// false if the request carries an Origin that is not allowlisted.
func setCORS(w http.ResponseWriter, r *http.Request) bool {
	w.Header().Add("Vary", "Origin")
	origin := r.Header.Get("Origin")
	if origin == "" {
		return true
	}
	if !allowedOrigins[origin] {
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
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
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
	verifyGoogle            func(ctx context.Context, credential string) (email string, emailVerified bool, err error)
}

// routeAuth registers the website member auth and API endpoints.
func routeAuth(r *mux.Router, db *sqlx.DB) {
	a := &authServer{
		store:        memberStore{db},
		emailLimiter: newRateLimiter(),
		ipLimiter:    newRateLimiter(),
		now:          time.Now,
		sendEmail:    sendLoginEmail,
		verifyGoogle: verifyGoogleCredential,
	}
	r.HandleFunc("/auth/google/token", corsHandler(a.googleToken)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/auth/email", corsHandler(a.emailStart)).Methods(http.MethodPost, http.MethodOptions)
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

// googleClientID is the website's own OAuth client ("DxE website sign-in", in
// the ADB Google Cloud project), separate from the ADB admin login's client so
// tokens issued for one aren't accepted by the other. The website renders
// Google's button and POSTs the credential here.
const googleClientID = "975059814880-aacjqeiq14ng4288bom3put59m2u6nd6.apps.googleusercontent.com"

const devCredential = "dev"

var googleVerifier = func() *oidc.IDTokenVerifier {
	provider, err := oidc.NewProvider(context.Background(), "https://accounts.google.com")
	if err != nil {
		log.Println(
			"WARNING: failed to construct OIDC provider (no Internet connection?). Some features may not work. Error:",
			err,
		)
		return nil
	}
	return provider.Verifier(&oidc.Config{ClientID: googleClientID})
}()

// verifyGoogleCredential verifies a Google ID token and returns its email.
func verifyGoogleCredential(ctx context.Context, credential string) (email string, emailVerified bool, err error) {
	if googleVerifier == nil {
		return "", false, errors.New("google verifier unavailable")
	}
	idToken, err := googleVerifier.Verify(ctx, credential)
	if err != nil {
		return "", false, err
	}
	var claims struct {
		Email         string `json:"email"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return "", false, err
	}
	return claims.Email, claims.EmailVerified, nil
}

// googleToken signs in with a Google ID token ("credential") from Google
// Identity Services and responds with the same JSON as /api/me.
func (a *authServer) googleToken(w http.ResponseWriter, r *http.Request) {
	noStore(w)
	// Always a browser call, so an Origin is required (corsHandler already
	// rejected non-allowlisted ones).
	if r.Header.Get("Origin") == "" {
		http.Error(w, "origin required", http.StatusForbidden)
		return
	}
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
	if !config.IsProd && body.Credential == devCredential {
		// Mirror the existing members flow, which fakes Google in dev.
		rawEmail = "test-dev@directactioneverywhere.com"
	} else {
		email, verified, err := a.verifyGoogle(r.Context(), body.Credential)
		if err != nil {
			http.Error(w, "invalid or expired credential", http.StatusUnauthorized)
			return
		}
		if !verified {
			http.Error(w, "your Google email address is not verified", http.StatusUnauthorized)
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
	if !config.IsProd || config.SMTPHost == "" {
		log.Printf("member auth: sign-in link for %s: %s", to, link)
		return nil
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
	w.Header().Set("Referrer-Policy", "no-referrer")
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
	clearCookie(w, sessionCookieName, "/")
	w.WriteHeader(http.StatusNoContent)
}
