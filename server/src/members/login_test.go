package members

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/testdb"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoginPage(t *testing.T) {
	old := config.WebsiteGoogleClientID
	config.WebsiteGoogleClientID = `client"><x`
	t.Cleanup(func() { config.WebsiteGoogleClientID = old })
	a := &authServer{now: time.Now}

	get := func(target string, cs ...*http.Cookie) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		for _, c := range cs {
			req.AddCookie(c)
		}
		rec := httptest.NewRecorder()
		a.loginPage(rec, req)
		return rec
	}

	next := "http://localhost:3000/account?a=1&b=%22x"
	rec := get("/login?next=" + url.QueryEscape(next))
	require.Equal(t, http.StatusOK, rec.Code)
	h := rec.Header()
	assert.Equal(t, loginCSP, h.Get("Content-Security-Policy"))
	assert.Equal(t, "DENY", h.Get("X-Frame-Options"))
	assert.Equal(t, "same-origin", h.Get("Referrer-Policy"))
	assert.Equal(t, "no-store", h.Get("Cache-Control"))
	assert.Equal(t, "nosniff", h.Get("X-Content-Type-Options"))
	assert.Contains(t, h.Get("Content-Type"), "text/html")

	body := rec.Body.String()
	// Data attributes are HTML-escaped; nothing from the inputs breaks out.
	assert.Contains(t, body, `data-client-id="client&#34;&gt;&lt;x"`)
	assert.NotContains(t, body, `client"><x`)
	assert.Contains(t, body, `data-next="http://localhost:3000/account?a=1&amp;b=%22x"`)
	assert.Contains(t, body, `id="dev-signin"`) // Tests run with MEMBERS_DEV_LOGIN.
	// No inline script, style or handlers.
	assert.NotContains(t, body, "<style")
	assert.NotContains(t, body, ` style=`)
	assert.NotContains(t, body, ` onclick=`)
	assert.NotContains(t, body, "<script>")
	assert.NotRegexp(t, `<script(?:[^>]*\s)?>[^<]`, body)

	// Nonce cookie: signed, HttpOnly, never holding the plain nonce.
	cs := rec.Result().Cookies()
	require.Len(t, cs, 1)
	c := cs[0]
	assert.Equal(t, nonceCookieName(), c.Name)
	assert.True(t, c.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	assert.Equal(t, "/", c.Path)
	assert.Equal(t, 600, c.MaxAge)
	nonce := between(body, `data-nonce="`, `"`)
	require.NotEmpty(t, nonce)
	assert.NotContains(t, c.Value, nonce)
	assert.True(t, checkNonce(reqWith(c), nonce))

	// A fresh nonce each render.
	rec2 := get("/login")
	assert.NotEqual(t, nonce, between(rec2.Body.String(), `data-nonce="`, `"`))
	assert.Contains(t, rec2.Body.String(), `data-next="`+fallbackNextURL+`"`)

	// Invalid next falls back.
	rec = get("/login?next=" + url.QueryEscape("https://evil.com/"))
	assert.Contains(t, rec.Body.String(), `data-next="`+fallbackNextURL+`"`)

	// Signed in: redirect, no page and no nonce cookie.
	sess := cookieFor(t, "a@example.org", time.Now())
	rec = get("/login?next="+url.QueryEscape("http://localhost:4173/x"), sess)
	assert.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, "http://localhost:4173/x", rec.Header().Get("Location"))
	assert.Empty(t, rec.Result().Cookies())
	rec = get("/login?next="+url.QueryEscape("https://evil.com/"), sess)
	assert.Equal(t, fallbackNextURL, rec.Header().Get("Location"))
}

func between(s, start, end string) string {
	i := strings.Index(s, start)
	if i < 0 {
		return ""
	}
	s = s[i+len(start):]
	return s[:strings.Index(s, end)]
}

func TestLoginAssets(t *testing.T) {
	h := loginAssetsHandler()
	for path, ctype := range map[string]string{
		"/login/assets/login.js":  "javascript",
		"/login/assets/login.css": "text/css",
		"/login/assets/logo.png":  "image/png",
	} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, http.StatusOK, rec.Code, path)
		assert.Contains(t, rec.Header().Get("Content-Type"), ctype, path)
		assert.Equal(t, loginCSP, rec.Header().Get("Content-Security-Policy"), path)
		assert.Equal(t, "nosniff", rec.Header().Get("X-Content-Type-Options"), path)
	}
	for _, path := range []string{"/login/assets/login.html", "/login/assets/", "/login/assets/nope.js"} {
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, path, nil))
		assert.Equal(t, http.StatusNotFound, rec.Code, path)
	}
}

func TestLoginRouteDispatch(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	r := mux.NewRouter()
	routeAuth(r, db, func(w http.ResponseWriter, _ *http.Request) { _, _ = w.Write([]byte("legacy")) })
	do := func(target, host string) string {
		req := httptest.NewRequest(http.MethodGet, target, nil)
		req.Host = host
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Body.String()
	}
	assert.Contains(t, do("/login", "localhost:8081"), "Sign in to Direct Action Everywhere")
	assert.Equal(t, "legacy", do("/login?legacy=1", "localhost:8081"))
	assert.Equal(t, "legacy", do("/login?force", "localhost:8081"))

	setProd(t, true)
	assert.Contains(t, do("/login", "members.directactioneverywhere.com"), "Sign in to Direct Action Everywhere")
	assert.Equal(t, "legacy", do("/login", "members.dxesf.org"))
	assert.Equal(t, "legacy", do("/login?force", "members.dxesf.org"))
}

// --- Nonce ---

type nonceFixture struct {
	t      *testing.T
	a      *authServer
	nonce  string
	cookie *http.Cookie
	claim  string // What the fake Google token will carry.
}

func newNonceFixture(t *testing.T) *nonceFixture {
	f := &nonceFixture{t: t}
	rec := httptest.NewRecorder()
	var err error
	f.nonce, err = setNonceCookie(rec)
	require.NoError(t, err)
	f.cookie = rec.Result().Cookies()[0]
	f.claim = f.nonce
	db := testdb.NewDB()
	t.Cleanup(func() { _ = db.Close() })
	f.a = &authServer{
		store: memberStore{db},
		now:   time.Now,
		verifyGoogle: func(context.Context, string) (string, bool, string, error) {
			return "nonce@example.org", true, f.claim, nil
		},
	}
	return f
}

func (f *nonceFixture) post(cookie *http.Cookie) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(`{"credential":"tok"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", authBaseURL())
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	f.a.googleToken(rec, req)
	return rec
}

func clearedNonce(rec *httptest.ResponseRecorder) bool {
	for _, c := range rec.Result().Cookies() {
		if c.Name == nonceCookieName() && c.MaxAge < 0 {
			return true
		}
	}
	return false
}

func hasSession(rec *httptest.ResponseRecorder) bool {
	for _, c := range rec.Result().Cookies() {
		if c.Name == sessionCookieName() && c.MaxAge > 0 {
			return true
		}
	}
	return false
}

func TestNonce_Match(t *testing.T) {
	f := newNonceFixture(t)
	rec := f.post(f.cookie)
	assert.Equal(t, http.StatusOK, rec.Code)
	assert.True(t, hasSession(rec))
	assert.True(t, clearedNonce(rec))
}

func TestNonce_Mismatch(t *testing.T) {
	f := newNonceFixture(t)
	f.claim = nonceFor([]byte("some other secret"))
	rec := f.post(f.cookie)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
	assert.True(t, clearedNonce(rec))
}

func TestNonce_MissingClaimOrCookie(t *testing.T) {
	f := newNonceFixture(t)
	rec := f.post(nil)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
	assert.True(t, clearedNonce(rec))

	f.claim = ""
	rec = f.post(f.cookie)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
	assert.True(t, clearedNonce(rec))
}

func TestNonce_Tampered(t *testing.T) {
	f := newNonceFixture(t)
	bad := *f.cookie
	bad.Value = f.cookie.Value[:len(f.cookie.Value)-3] + "AAA"
	rec := f.post(&bad)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
}

func TestNonce_Expired(t *testing.T) {
	f := newNonceFixture(t)
	// A cookie with the right secret behind it.
	secret := []byte("0123456789abcdef0123456789abcdef")
	old := *f.cookie
	enc, err := nonceCodec.Encode(nonceCookieName(), secret)
	require.NoError(t, err)
	old.Value = enc
	f.claim = nonceFor(secret)
	assert.Equal(t, http.StatusOK, f.post(&old).Code, "fresh cookie works")

	// securecookie rejects by its signed timestamp; use a codec that treats it
	// as already older than its max age.
	saved := nonceCodec
	t.Cleanup(func() { nonceCodec = saved })
	nonceCodec = nonceCodec.MaxAge(1)
	time.Sleep(2100 * time.Millisecond)
	rec := f.post(&old)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
}

// The cookie value must never be usable as the token's nonce claim, and
// neither may the secret itself: only SHA-256(S) is.
func TestNonce_CookieValueEqualsClaimRejected(t *testing.T) {
	f := newNonceFixture(t)
	f.claim = f.cookie.Value
	rec := f.post(f.cookie)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
	assert.NotEqual(t, f.nonce, f.cookie.Value)
	assert.False(t, strings.Contains(f.cookie.Value, f.nonce))
}

func TestNonce_SingleUse(t *testing.T) {
	f := newNonceFixture(t)
	assert.Equal(t, http.StatusOK, f.post(f.cookie).Code)
	// The browser dropped the cookie on the clearing Set-Cookie, but even a
	// replayed token with no cookie fails.
	assert.Equal(t, http.StatusUnauthorized, f.post(nil).Code)
}

func TestNonce_SessionCookieIsNotANonce(t *testing.T) {
	f := newNonceFixture(t)
	sess := cookieFor(t, "a@example.org", time.Now())
	forged := &http.Cookie{Name: nonceCookieName(), Value: sess.Value}
	f.claim = sess.Value
	assert.Equal(t, http.StatusUnauthorized, f.post(forged).Code)
}

// --- Origin rules ---

func TestSameOriginOnlyEndpoints(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	a := &authServer{store: memberStore{db}, emailLimiter: newRateLimiter(), ipLimiter: newRateLimiter(), now: time.Now,
		sendEmail: func(string, string) error { return nil }}
	r := mux.NewRouter()
	routeAuthHandlers(r, a)

	post := func(path, origin, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}
	const www = "https://www.directactioneverywhere.com"
	for _, path := range []string{"/auth/google/token", "/auth/email"} {
		for _, origin := range []string{www, "http://localhost:3000", "https://evil.com", "null", ""} {
			rec := post(path, origin, `{"email":"a@example.org","credential":"dev"}`)
			assert.Equal(t, http.StatusForbidden, rec.Code, "%s from %q", path, origin)
			assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
			assert.Empty(t, rec.Result().Cookies())
		}
		// Preflights get no CORS approval either.
		req := httptest.NewRequest(http.MethodOptions, path, nil)
		req.Header.Set("Origin", www)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"), path)
	}
	assert.Equal(t, http.StatusNoContent, post("/auth/email", authBaseURL(), `{"email":"a@example.org"}`).Code)

	// /api/me and logout still serve the website cross-origin.
	for _, c := range []struct{ method, path string }{{http.MethodGet, "/api/me"}, {http.MethodPost, "/auth/logout"}} {
		req := httptest.NewRequest(c.method, c.path, nil)
		req.Header.Set("Origin", www)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, www, rec.Header().Get("Access-Control-Allow-Origin"), c.path)
		assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"), c.path)
		assert.NotEqual(t, http.StatusForbidden, rec.Code, c.path)

		req = httptest.NewRequest(http.MethodOptions, c.path, nil)
		req.Header.Set("Origin", www)
		rec = httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code, c.path)
		assert.Equal(t, www, rec.Header().Get("Access-Control-Allow-Origin"), c.path)
	}
}
