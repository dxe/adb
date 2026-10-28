package members

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dxe/adb/testdb"
	"github.com/gorilla/sessions"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewTokenHashes(t *testing.T) {
	raw1, hash1, err := newToken()
	require.NoError(t, err)
	raw2, _, err := newToken()
	require.NoError(t, err)
	assert.Len(t, raw1, 64)
	assert.NotEqual(t, raw1, raw2)
	assert.Len(t, hash1, 32)
	assert.Equal(t, hash1, hashToken(raw1))
	assert.NotEqual(t, hash1, hashToken(raw2))
}

func TestLoginToken_ConsumeOnceThenRejected(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	st := memberStore{db}
	ctx := context.Background()
	now := time.Now()

	raw, err := st.createLoginToken(ctx, "a@example.org", "http://localhost:3000/account", now)
	require.NoError(t, err)

	// Only the hash is stored.
	var n int
	require.NoError(t, db.Get(&n, `SELECT COUNT(*) FROM member_login_tokens WHERE token_hash = ?`, hashToken(raw)))
	assert.Equal(t, 1, n)

	// Peeking does not consume.
	email, err := st.peekLoginToken(ctx, raw, now)
	require.NoError(t, err)
	assert.Equal(t, "a@example.org", email)
	email, err = st.peekLoginToken(ctx, raw, now)
	require.NoError(t, err)
	assert.Equal(t, "a@example.org", email)

	email, next, err := st.consumeLoginToken(ctx, raw, now.Add(time.Minute))
	require.NoError(t, err)
	assert.Equal(t, "a@example.org", email)
	assert.Equal(t, "http://localhost:3000/account", next)

	_, _, err = st.consumeLoginToken(ctx, raw, now.Add(2*time.Minute))
	assert.ErrorIs(t, err, errInvalidToken)
	_, err = st.peekLoginToken(ctx, raw, now.Add(2*time.Minute))
	assert.ErrorIs(t, err, errInvalidToken)
}

func TestLoginToken_Expired(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	st := memberStore{db}
	ctx := context.Background()
	now := time.Now()

	raw, err := st.createLoginToken(ctx, "a@example.org", fallbackNextURL, now)
	require.NoError(t, err)

	late := now.Add(loginTokenLifetime + time.Second)
	_, err = st.peekLoginToken(ctx, raw, late)
	assert.ErrorIs(t, err, errInvalidToken)
	_, _, err = st.consumeLoginToken(ctx, raw, late)
	assert.ErrorIs(t, err, errInvalidToken)

	// Still valid just before expiry, and unknown tokens are rejected.
	_, _, err = st.consumeLoginToken(ctx, raw, now.Add(loginTokenLifetime-time.Second))
	assert.NoError(t, err)
	_, _, err = st.consumeLoginToken(ctx, "nope", now)
	assert.ErrorIs(t, err, errInvalidToken)
}

func cookieFor(t *testing.T, email string, issued time.Time) *http.Cookie {
	t.Helper()
	rec := httptest.NewRecorder()
	require.NoError(t, saveSession(rec, httptest.NewRequest(http.MethodGet, "/", nil), email, issued))
	cs := rec.Result().Cookies()
	require.Len(t, cs, 1)
	return cs[0]
}

func reqWith(c *http.Cookie) *http.Request {
	r := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	r.AddCookie(c)
	return r
}

func TestSessionCookie_RoundTrip(t *testing.T) {
	now := time.Now()
	c := cookieFor(t, "a@example.org", now)
	assert.Equal(t, sessionCookieName, c.Name)
	assert.True(t, c.HttpOnly)
	assert.Equal(t, http.SameSiteLaxMode, c.SameSite)
	assert.Equal(t, "/", c.Path)
	assert.Empty(t, c.Domain)
	assert.Equal(t, int(sessionLifetime/time.Second), c.MaxAge)

	email, issued, ok := readSession(reqWith(c), now.Add(time.Hour))
	require.True(t, ok)
	assert.Equal(t, "a@example.org", email)
	assert.Equal(t, now.Unix(), issued.Unix())

	_, _, ok = readSession(httptest.NewRequest(http.MethodGet, "/", nil), now)
	assert.False(t, ok)
}

func TestSessionCookie_TamperedRejected(t *testing.T) {
	now := time.Now()
	c := cookieFor(t, "a@example.org", now)
	c.Value = c.Value[:len(c.Value)-3] + "AAA"
	_, _, ok := readSession(reqWith(c), now)
	assert.False(t, ok)

	// A cookie signed with another key is rejected too.
	other := sessions.NewCookieStore([]byte("a-different-secret"))
	sess, _ := other.Get(httptest.NewRequest(http.MethodGet, "/", nil), sessionCookieName)
	sess.Values["email"] = "a@example.org"
	sess.Values["issued"] = now.Unix()
	rec := httptest.NewRecorder()
	require.NoError(t, other.Save(httptest.NewRequest(http.MethodGet, "/", nil), rec, sess))
	_, _, ok = readSession(reqWith(rec.Result().Cookies()[0]), now)
	assert.False(t, ok)
}

func TestSessionCookie_Expired(t *testing.T) {
	now := time.Now()
	c := cookieFor(t, "a@example.org", now.Add(-sessionLifetime-time.Hour))
	_, _, ok := readSession(reqWith(c), now)
	assert.False(t, ok)

	c = cookieFor(t, "a@example.org", now.Add(-sessionLifetime+time.Hour))
	_, _, ok = readSession(reqWith(c), now)
	assert.True(t, ok)
}

func TestSessionCookie_SlidingRefresh(t *testing.T) {
	a := &authServer{now: time.Now}
	me := func(c *http.Cookie) *httptest.ResponseRecorder {
		rec := httptest.NewRecorder()
		a.store = memberStore{testdb.NewDB()}
		defer func() { _ = a.store.db.Close() }()
		a.me(rec, reqWith(c))
		return rec
	}

	// Fresh cookie: not re-issued.
	rec := me(cookieFor(t, "a@example.org", time.Now().Add(-time.Hour)))
	require.Equal(t, http.StatusOK, rec.Code)
	assert.Empty(t, rec.Result().Cookies())

	// Older than a day: re-issued with a fresh issued-at.
	rec = me(cookieFor(t, "a@example.org", time.Now().Add(-48*time.Hour)))
	require.Equal(t, http.StatusOK, rec.Code)
	cs := rec.Result().Cookies()
	require.Len(t, cs, 1)
	_, issued, ok := readSession(reqWith(cs[0]), time.Now())
	require.True(t, ok)
	assert.WithinDuration(t, time.Now(), issued, time.Minute)

	// Tampered: 401 and cookie cleared.
	bad := cookieFor(t, "a@example.org", time.Now())
	bad.Value += "x"
	rec = me(bad)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
