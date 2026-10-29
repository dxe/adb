package members

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/dxe/adb/testdb"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMagicLinkFlow(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	_, err := db.Exec(`INSERT INTO activists (name, email, chapter_id, hidden) VALUES
		('Hidden Dupe', 'jane@example.org', 1, 1), ('Jane Doe', 'Jane@Example.org', 1, 0)`)
	require.NoError(t, err)

	links := make(chan string, 1)
	a := &authServer{
		store:        memberStore{db},
		emailLimiter: newRateLimiter(),
		ipLimiter:    newRateLimiter(),
		now:          time.Now,
		sendEmail:    func(to, link string) error { links <- link; return nil },
	}
	r := mux.NewRouter()
	r.HandleFunc("/auth/email", a.emailStart).Methods(http.MethodPost)
	r.HandleFunc("/auth/email/confirm", a.emailConfirmPage).Methods(http.MethodGet)
	r.HandleFunc("/auth/email/confirm", a.emailConfirm).Methods(http.MethodPost)
	r.HandleFunc("/auth/logout", corsHandler(a.logout)).Methods(http.MethodPost, http.MethodOptions)
	r.HandleFunc("/api/me", corsHandler(a.me)).Methods(http.MethodGet, http.MethodOptions)
	do := func(req *http.Request) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec
	}

	// Signed out.
	rec := do(httptest.NewRequest(http.MethodGet, "/api/me", nil))
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.JSONEq(t, `{}`, rec.Body.String())

	// Request a link; a bad next falls back.
	req := httptest.NewRequest(http.MethodPost, "/auth/email",
		strings.NewReader(`{"email":"Jane@Example.org","next":"https://evil.com/"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", authBaseURL())
	rec = do(req)
	require.Equal(t, http.StatusNoContent, rec.Code)
	link := <-links
	u, err := url.Parse(link)
	require.NoError(t, err)
	token := u.Query().Get("token")
	require.NotEmpty(t, token)

	// GET does not consume.
	for i := 0; i < 2; i++ {
		rec = do(httptest.NewRequest(http.MethodGet, "/auth/email/confirm?token="+token, nil))
		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Contains(t, rec.Body.String(), "jane@example.org")
	}

	// A hostile Origin cannot confirm.
	form := url.Values{"token": {token}}.Encode()
	bad := httptest.NewRequest(http.MethodPost, "/auth/email/confirm", strings.NewReader(form))
	bad.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	bad.Header.Set("Origin", "https://evil.com")
	assert.Equal(t, http.StatusForbidden, do(bad).Code)

	// POST consumes, sets the cookie, redirects to the fallback.
	post := httptest.NewRequest(http.MethodPost, "/auth/email/confirm", strings.NewReader(form))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = do(post)
	require.Equal(t, http.StatusSeeOther, rec.Code)
	assert.Equal(t, fallbackNextURL, rec.Header().Get("Location"))
	cookies := rec.Result().Cookies()
	require.Len(t, cookies, 1)
	c := cookies[0]
	assert.Equal(t, sessionCookieName(), c.Name)
	assert.True(t, c.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	assert.Equal(t, "/", c.Path)

	// Reuse fails.
	post = httptest.NewRequest(http.MethodPost, "/auth/email/confirm", strings.NewReader(form))
	post.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	assert.Equal(t, http.StatusGone, do(post).Code)

	// /api/me matches the visible activist case-insensitively.
	me := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	me.Header.Set("Origin", "http://localhost:3000")
	me.AddCookie(c)
	rec = do(me)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "http://localhost:3000", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.JSONEq(t, `{"email":"jane@example.org","name":"Jane Doe","activistId":2}`, rec.Body.String())

	// Logout.
	out := httptest.NewRequest(http.MethodPost, "/auth/logout", nil)
	out.AddCookie(c)
	rec = do(out)
	assert.Equal(t, http.StatusNoContent, rec.Code)
	cleared := rec.Result().Cookies()
	require.Len(t, cleared, 1)
	assert.Equal(t, sessionCookieName(), cleared[0].Name)
	assert.Equal(t, -1, cleared[0].MaxAge) // Stateless: logout only clears the cookie.
}

func TestGoogleToken(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	_, err := db.Exec(`INSERT INTO activists (name, email, chapter_id, hidden) VALUES ('Gina Google', 'gina@example.org', 1, 0)`)
	require.NoError(t, err)

	// The page's nonce, and the cookie that backs it.
	nonceRec := httptest.NewRecorder()
	claimNonce, err := setNonceCookie(nonceRec, httptest.NewRequest(http.MethodGet, "/login", nil))
	require.NoError(t, err)
	nonceCookie := nonceRec.Result().Cookies()[0]

	a := &authServer{
		store: memberStore{db},
		now:   time.Now,
		verifyGoogle: func(_ context.Context, cred string) (string, bool, string, error) {
			switch cred {
			case "good":
				return "Gina@Example.org", true, claimNonce, nil
			case "unverified":
				return "gina@example.org", false, claimNonce, nil
			}
			return "", false, "", errors.New("bad token")
		},
	}
	h := a.googleToken
	post := func(origin, contentType, body string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(body))
		req.AddCookie(nonceCookie)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		if contentType != "" {
			req.Header.Set("Content-Type", contentType)
		}
		rec := httptest.NewRecorder()
		h(rec, req)
		return rec
	}
	site := authBaseURL()
	const j = "application/json"

	assert.Equal(t, http.StatusForbidden, post("https://evil.com", j, `{"credential":"good"}`).Code)
	assert.Equal(t, http.StatusForbidden, post("https://www.directactioneverywhere.com", j, `{"credential":"good"}`).Code, "website origin")
	assert.Equal(t, http.StatusForbidden, post("", j, `{"credential":"good"}`).Code, "missing Origin")
	assert.Equal(t, http.StatusUnsupportedMediaType, post(site, "", `{"credential":"good"}`).Code)
	assert.Equal(t, http.StatusUnsupportedMediaType, post(site, "text/plain", `{"credential":"good"}`).Code)
	assert.Equal(t, http.StatusBadRequest, post(site, j, `not json`).Code)
	assert.Equal(t, http.StatusBadRequest, post(site, j, `{}`).Code)

	rec := post(site, j, `{"credential":"garbage"}`)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
	assert.True(t, clearedNonce(rec))
	assert.Equal(t, http.StatusUnauthorized, post(site, j, `{"credential":"unverified"}`).Code)

	// Happy path with a fake verified token: cookie set, /api/me-shaped body.
	rec = post(site, j, `{"credential":"good"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.JSONEq(t, `{"email":"gina@example.org","name":"Gina Google","activistId":1}`, rec.Body.String())
	cs := rec.Result().Cookies()
	require.Len(t, cs, 2) // Session, and the cleared nonce.
	assert.True(t, clearedNonce(rec))
	var sess *http.Cookie
	for _, c := range cs {
		if c.Name == sessionCookieName() {
			sess = c
		}
	}
	require.NotNil(t, sess)
	email, _, ok := readSession(reqWith(sess), time.Now())
	assert.True(t, ok)
	assert.Equal(t, "gina@example.org", email)

	// Dev credential (tests run with PROD unset).
	rec = post(site, j, `{"credential":"dev"}`)
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Contains(t, rec.Body.String(), "test-dev@directactioneverywhere.com")
	require.Len(t, rec.Result().Cookies(), 2)
}

func TestMe_ActivistLookupIsExact(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	_, err := db.Exec(`INSERT INTO activists (name, email, chapter_id, hidden) VALUES
		('Accent Person', 'jane@acmé.org', 1, 0), ('Plain Person', 'sam@acme.org', 1, 0)`)
	require.NoError(t, err)
	a := &authServer{store: memberStore{db}, now: time.Now}

	me := func(email string) string {
		rec := httptest.NewRecorder()
		a.writeMe(rec, httptest.NewRequest(http.MethodGet, "/api/me", nil), email)
		return rec.Body.String()
	}
	// The unaccented session email must not match the accented activist.
	assert.JSONEq(t, `{"email":"jane@acme.org","name":null,"activistId":null}`, me("jane@acme.org"))
	// An exact (case-insensitive) match still works.
	assert.JSONEq(t, `{"email":"sam@acme.org","name":"Plain Person","activistId":2}`, me("sam@acme.org"))
}
