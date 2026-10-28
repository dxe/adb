package members

import (
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
	r.HandleFunc("/auth/email", corsHandler(a.emailStart)).Methods(http.MethodPost, http.MethodOptions)
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
	req.Header.Set("Origin", "http://localhost:3000")
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
	assert.Equal(t, sessionCookieName, c.Name)
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
	assert.Equal(t, http.StatusNoContent, do(out).Code)
	me = httptest.NewRequest(http.MethodGet, "/api/me", nil)
	me.AddCookie(c)
	assert.Equal(t, http.StatusUnauthorized, do(me).Code)
}
