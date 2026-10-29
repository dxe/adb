package members

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"embed"
	"encoding/base64"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dxe/adb/config"
	"github.com/gorilla/securecookie"
)

// The website sign-in page, served on members.directactioneverywhere.com. It is
// the only origin on the website's Google OAuth client, and carries no
// third-party scripts besides Google's own, enforced by loginCSP.

//go:embed login_assets
var loginAssetsFS embed.FS

var loginAssets = func() fs.FS {
	sub, err := fs.Sub(loginAssetsFS, "login_assets")
	if err != nil {
		panic(err)
	}
	return sub
}()

var loginTmpl = template.Must(template.ParseFS(loginAssets, "login.html"))

// loginCSP is deliberately narrow: no inline script, and only Google's GIS
// script, button iframe, stylesheet and endpoints. style-src needs
// 'unsafe-inline' because the GIS client script injects a <style> element and
// sets style attributes on the button it renders (without them the button is
// unstyled); our own page uses neither, and with no script-src 'unsafe-inline'
// and img/font/connect limited to 'self', inline styles can't run or exfiltrate.
const loginCSP = "default-src 'none'; " +
	"script-src 'self' https://accounts.google.com/gsi/client; " +
	"frame-src https://accounts.google.com/gsi/; " +
	"connect-src 'self' https://accounts.google.com/gsi/; " +
	"style-src 'self' https://accounts.google.com/gsi/style 'unsafe-inline'; " +
	"img-src 'self' data:; " +
	"form-action 'self'; " +
	"base-uri 'none'; " +
	"frame-ancestors 'none'"

func loginHeaders(w http.ResponseWriter) {
	h := w.Header()
	h.Set("Content-Security-Policy", loginCSP)
	h.Set("X-Frame-Options", "DENY")
	h.Set("Referrer-Policy", "same-origin")
	h.Set("X-Content-Type-Options", "nosniff")
}

// isWebsiteLoginRequest reports whether GET /login should serve the website
// sign-in page rather than the legacy members.dxesf.org Google redirect, which
// shares this router and path. In prod they are told apart by Host. Locally
// there is a single host, so the website page is the default and the legacy
// flow is reached with ?legacy (or ?force, which its own links use).
func isWebsiteLoginRequest(r *http.Request) bool {
	if config.IsProd {
		return isWebsiteHost(r)
	}
	q := r.URL.Query()
	return !q.Has("legacy") && !q.Has("force")
}

// legacyLoginURL is where the members.dxesf.org pages send signed-out users.
func legacyLoginURL() string {
	if config.IsProd {
		return absURL("/login")
	}
	return absURL("/login?legacy=1")
}

// --- GIS nonce ---
//
// The page's nonce is N = base64url(SHA-256(S)) for a random secret S that
// only ever leaves the server inside a signed, HttpOnly cookie. Google embeds N
// in the ID token, so the token reveals nothing that redeems it: redeeming
// needs S's cookie, i.e. the browser the page was served to.

const nonceMaxAge = 10 * time.Minute

func nonceCookieName() string {
	if config.IsProd {
		return "__Host-dxe_gis_nonce"
	}
	return "dxe_gis_nonce"
}

// nonceCodec signs and timestamps S. securecookie binds the cookie name into
// the MAC, so other signed cookies (the session) cannot be replayed as a nonce.
var nonceCodec = func() *securecookie.SecureCookie {
	return securecookie.New([]byte(config.CookieSecret), nil).MaxAge(int(nonceMaxAge / time.Second))
}()

func nonceFor(secret []byte) string {
	sum := sha256.Sum256(secret)
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

// nonceSecret returns the S in the request's signed nonce cookie, if valid.
func nonceSecret(r *http.Request) ([]byte, bool) {
	c, err := r.Cookie(nonceCookieName())
	if err != nil {
		return nil, false
	}
	var secret []byte
	if err := nonceCodec.Decode(nonceCookieName(), c.Value, &secret); err != nil || len(secret) == 0 {
		return nil, false
	}
	return secret, true
}

// setNonceCookie sets the nonce cookie and returns N for the page. A valid
// cookie's S is reused (and re-issued with a fresh timestamp), so a second tab
// doesn't invalidate the first tab's pending sign-in; S is minted only when the
// cookie is missing or invalid.
func setNonceCookie(w http.ResponseWriter, r *http.Request) (string, error) {
	secret, ok := nonceSecret(r)
	if !ok {
		secret = make([]byte, 32)
		if _, err := rand.Read(secret); err != nil {
			return "", err
		}
	}
	enc, err := nonceCodec.Encode(nonceCookieName(), secret)
	if err != nil {
		return "", err
	}
	http.SetCookie(w, &http.Cookie{
		Name:     nonceCookieName(),
		Value:    enc,
		Path:     "/",
		MaxAge:   int(nonceMaxAge / time.Second),
		HttpOnly: true,
		Secure:   config.IsProd,
		SameSite: http.SameSiteLaxMode,
	})
	return nonceFor(secret), nil
}

// checkNonce reports whether claim equals the nonce for the S in the request's
// signed cookie. Missing, tampered or expired cookies fail.
func checkNonce(r *http.Request, claim string) bool {
	secret, ok := nonceSecret(r)
	if !ok || claim == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(nonceFor(secret)), []byte(claim)) == 1
}

// --- Handlers ---

type loginPageData struct {
	ClientID, Nonce, Next string
	Dev                   bool
}

func (a *authServer) loginPage(w http.ResponseWriter, r *http.Request) {
	loginHeaders(w)
	noStore(w)
	next := validateNext(r.URL.Query().Get("next"))
	if _, _, ok := readSession(r, a.now()); ok {
		http.Redirect(w, r, next, http.StatusSeeOther)
		return
	}
	nonce, err := setNonceCookie(w, r)
	if err != nil {
		log.Println("member auth: nonce:", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := loginTmpl.Execute(w, loginPageData{
		ClientID: config.WebsiteGoogleClientID,
		Nonce:    nonce,
		Next:     next,
		Dev:      devLogin(),
	}); err != nil {
		log.Println("member auth: render login:", err)
	}
}

// loginAssetsHandler serves /login/assets/*.
func loginAssetsHandler() http.Handler {
	files := http.StripPrefix("/login/assets/", http.FileServer(http.FS(loginAssets)))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/") || strings.HasSuffix(r.URL.Path, "/login.html") {
			http.NotFound(w, r)
			return
		}
		loginHeaders(w)
		w.Header().Set("Cache-Control", "public, max-age=300")
		files.ServeHTTP(w, r)
	})
}

// sameOrigin requires a browser request from this server's own origin. Used by
// the endpoints that mint sessions or send mail: nothing else may call them.
func sameOrigin(w http.ResponseWriter, r *http.Request) bool {
	if r.Header.Get("Origin") != authBaseURL() {
		http.Error(w, "forbidden origin", http.StatusForbidden)
		return false
	}
	return true
}

// isWebsiteHost reports whether the request is for the website members origin.
// Only meaningful in prod; locally every request shares one host.
func isWebsiteHost(r *http.Request) bool {
	return config.IsProd && r.Host == "members.directactioneverywhere.com"
}

// websiteHostAllows lists what members.directactioneverywhere.com serves. The
// legacy members.dxesf.org pages (roster, reminder, the OAuth callback at
// /auth, ...) don't carry our CSP, so they must not exist on this origin.
func websiteHostAllows(path string) bool {
	return path == "/login" || strings.HasPrefix(path, "/login/assets/") ||
		strings.HasPrefix(path, "/auth/") || strings.HasPrefix(path, "/api/") ||
		path == "/health"
}

// websiteHostMiddleware scopes the website host to the sign-in surface (404
// for everything else, "/" redirects to /login) and sets HSTS there. Other
// hosts, and local development, are untouched.
func websiteHostMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isWebsiteHost(r) {
			next.ServeHTTP(w, r)
			return
		}
		w.Header().Set("Strict-Transport-Security", "max-age=31536000")
		switch {
		case r.URL.Path == "/":
			http.Redirect(w, r, "/login", http.StatusFound)
		case websiteHostAllows(r.URL.Path):
			next.ServeHTTP(w, r)
		default:
			http.NotFound(w, r)
		}
	})
}
