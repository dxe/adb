package members

import (
	"bytes"
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidateNext(t *testing.T) {
	cases := []struct{ in, want string }{
		{"https://www.directactioneverywhere.com/account", "https://www.directactioneverywhere.com/account"},
		{"https://directactioneverywhere.com/", "https://directactioneverywhere.com/"},
		{"https://staging.directactioneverywhere.com/account?x=1#y", "https://staging.directactioneverywhere.com/account?x=1#y"},
		{"http://localhost:3000/account", "http://localhost:3000/account"},
		{"http://localhost:4173/account", "http://localhost:4173/account"},
		{"", fallbackNextURL},
		{"/account", fallbackNextURL},
		{"//www.directactioneverywhere.com/account", fallbackNextURL},
		{"https://evil.com/", fallbackNextURL},
		{"http://www.directactioneverywhere.com/account", fallbackNextURL}, // wrong scheme for origin
		{"https://www.directactioneverywhere.com.evil.com/", fallbackNextURL},
		{"https://evil.com/https://www.directactioneverywhere.com", fallbackNextURL},
		{"https://www.directactioneverywhere.com@evil.com/", fallbackNextURL},
		{"https://evil.com@www.directactioneverywhere.com/", fallbackNextURL}, // userinfo
		{"https://www.directactioneverywhere.com:evil@evil.com/", fallbackNextURL},
		{"https://www.directactioneverywhere.com\\@evil.com/", fallbackNextURL},
		{"https://www.directactioneverywhere.com:8443/", fallbackNextURL},
		{"http://localhost:3001/", fallbackNextURL},
		{"javascript:alert(1)", fallbackNextURL},
		{"data:text/html,hi", fallbackNextURL},
		{"ftp://www.directactioneverywhere.com/", fallbackNextURL},
		{" https://www.directactioneverywhere.com/", fallbackNextURL},
		{"https://www.directactioneverywhere.com/\r\nX: y", fallbackNextURL},
		{"https://www.DirectActionEverywhere.com/", fallbackNextURL},
		{"https://www.directactioneverywhere.com/" + string(make([]byte, 600)), fallbackNextURL},
	}
	for _, c := range cases {
		assert.Equal(t, c.want, validateNext(c.in), "input %q", c.in)
	}
}

func TestOriginAllowed(t *testing.T) {
	assert.True(t, originAllowed(""))
	assert.True(t, originAllowed("https://www.directactioneverywhere.com"))
	assert.True(t, originAllowed("http://localhost:3000"))
	assert.False(t, originAllowed("https://evil.com"))
	assert.False(t, originAllowed("null"))
	assert.False(t, originAllowed("https://www.directactioneverywhere.com.evil.com"))
	assert.False(t, originAllowed("https://www.directactioneverywhere.com/"))
}

func TestCORSHandler(t *testing.T) {
	called := 0
	h := corsHandler(func(w http.ResponseWriter, r *http.Request) {
		called++
		w.WriteHeader(http.StatusNoContent)
	})
	do := func(method, origin string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(method, "/api/me", nil)
		if origin != "" {
			req.Header.Set("Origin", origin)
		}
		rec := httptest.NewRecorder()
		h(rec, req)
		return rec
	}

	rec := do(http.MethodOptions, "https://www.directactioneverywhere.com")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "https://www.directactioneverywhere.com", rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", rec.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, rec.Header().Get("Access-Control-Allow-Headers"), "Content-Type")
	assert.Equal(t, 0, called)

	rec = do(http.MethodOptions, "https://evil.com")
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))

	rec = do(http.MethodPost, "https://evil.com")
	assert.Equal(t, http.StatusForbidden, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, 0, called)

	rec = do(http.MethodPost, "http://localhost:4173")
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Equal(t, "http://localhost:4173", rec.Header().Get("Access-Control-Allow-Origin"))

	rec = do(http.MethodPost, "") // Non-browser client.
	assert.Equal(t, http.StatusNoContent, rec.Code)
	assert.Empty(t, rec.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, 2, called)
}

func TestClientIP(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.RemoteAddr = "10.0.0.5:1234"
	assert.Equal(t, "10.0.0.5", clientIP(req))
	req.Header.Set("X-Forwarded-For", "6.6.6.6, 203.0.113.9")
	assert.Equal(t, "203.0.113.9", clientIP(req), "ALB appends the real client; earlier hops are spoofable")

	// A client-sent duplicate header line cannot win: the ALB's line is last.
	req.Header = http.Header{}
	req.Header.Add("X-Forwarded-For", "6.6.6.6")
	req.Header.Add("X-Forwarded-For", "1.1.1.1, 203.0.113.9")
	assert.Equal(t, "203.0.113.9", clientIP(req))
}

func TestNormalizeEmail(t *testing.T) {
	got, ok := normalizeEmail("  Jane@Example.ORG ")
	assert.True(t, ok)
	assert.Equal(t, "jane@example.org", got)
	for _, bad := range []string{"jane@acmé.org", "jané@acme.org", "jane@acme\u200b.org", "ｊane@acme.org", "", "nope", "Jane <jane@example.org>", "a@b.c, d@e.f"} {
		_, ok := normalizeEmail(bad)
		assert.False(t, ok, bad)
	}
}

func TestRateLimiter(t *testing.T) {
	l := newRateLimiter()
	now := time.Now()
	for i := 0; i < 5; i++ {
		assert.True(t, l.allow("a", 5, time.Hour, now))
	}
	assert.False(t, l.allow("a", 5, time.Hour, now.Add(time.Minute)))
	assert.True(t, l.allow("b", 5, time.Hour, now))
	assert.True(t, l.allow("a", 5, time.Hour, now.Add(61*time.Minute)))
}

func setProd(t *testing.T, prod bool) {
	t.Helper()
	old := config.IsProd
	config.IsProd = prod
	t.Cleanup(func() { config.IsProd = old })
}

func TestLocalhostOriginsOnlyOutsideProd(t *testing.T) {
	setProd(t, false)
	assert.True(t, isAllowedOrigin("http://localhost:3000"))
	assert.Equal(t, "http://localhost:4173/account", validateNext("http://localhost:4173/account"))

	setProd(t, true)
	assert.True(t, isAllowedOrigin("https://www.directactioneverywhere.com"))
	assert.False(t, isAllowedOrigin("http://localhost:3000"))
	assert.False(t, isAllowedOrigin("http://localhost:4173"))
	assert.False(t, originAllowed("http://localhost:3000"))
	assert.Equal(t, fallbackNextURL, validateNext("http://localhost:3000/account"))
}

func setDevLogin(t *testing.T, on bool) {
	t.Helper()
	old := config.MembersDevLogin
	config.MembersDevLogin = on
	t.Cleanup(func() { config.MembersDevLogin = old })
}

// A server that isn't prod but lacks MEMBERS_DEV_LOGIN (e.g. a production task
// that lost PROD=true) gets none of the dev conveniences.
func TestNoDevLoginWithoutExplicitFlag(t *testing.T) {
	setProd(t, false)
	setDevLogin(t, false)
	assert.False(t, devLogin())
	assert.False(t, isAllowedOrigin("http://localhost:3000"))
	assert.Equal(t, fallbackNextURL, validateNext("http://localhost:3000/account"))

	oldHost := config.SMTPHost
	config.SMTPHost = ""
	t.Cleanup(func() { config.SMTPHost = oldHost })
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })
	assert.Error(t, sendLoginEmail("a@example.org", "https://x/?token=SECRET"))
	assert.NotContains(t, buf.String(), "SECRET")

	a := &authServer{now: time.Now, verifyGoogle: func(context.Context, string) (string, bool, string, error) {
		return "", false, "", errors.New("invalid")
	}}
	req := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(`{"credential":"dev"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", authBaseURL())
	rec := httptest.NewRecorder()
	a.googleToken(rec, req)
	assert.Equal(t, http.StatusUnauthorized, rec.Code)
	assert.False(t, hasSession(rec))
}

func TestSessionCookieName(t *testing.T) {
	setProd(t, false)
	assert.Equal(t, "dxe_member_session", sessionCookieName())
	setProd(t, true)
	assert.Equal(t, "__Host-dxe_member_session", sessionCookieName())
}

func TestSendLoginEmail_ProdWithoutSMTPErrorsAndDoesNotLogLink(t *testing.T) {
	setProd(t, true)
	oldHost := config.SMTPHost
	config.SMTPHost = ""
	t.Cleanup(func() { config.SMTPHost = oldHost })
	var buf bytes.Buffer
	log.SetOutput(&buf)
	t.Cleanup(func() { log.SetOutput(os.Stderr) })

	err := sendLoginEmail("a@example.org", "https://members.example/auth/email/confirm?token=SECRET")
	assert.Error(t, err)
	assert.NotContains(t, buf.String(), "SECRET")

	setProd(t, false)
	assert.NoError(t, sendLoginEmail("a@example.org", "https://x/?token=DEVLINK"))
	assert.Contains(t, buf.String(), "DEVLINK")
}

func TestConfirmPagesDenyFraming(t *testing.T) {
	check := func(rec *httptest.ResponseRecorder) {
		assert.Equal(t, "DENY", rec.Header().Get("X-Frame-Options"))
		assert.Equal(t, "frame-ancestors 'none'", rec.Header().Get("Content-Security-Policy"))
	}
	// Invalid-token paths need no DB rows, only a store that finds nothing.
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	a := &authServer{store: memberStore{db}, now: time.Now}

	rec := httptest.NewRecorder()
	a.emailConfirmPage(rec, httptest.NewRequest(http.MethodGet, "/auth/email/confirm?token=nope", nil))
	assert.Equal(t, http.StatusGone, rec.Code)
	check(rec)

	raw, err := a.store.createLoginToken(context.Background(), "a@example.org", fallbackNextURL, time.Now())
	require.NoError(t, err)
	rec = httptest.NewRecorder()
	a.emailConfirmPage(rec, httptest.NewRequest(http.MethodGet, "/auth/email/confirm?token="+raw, nil))
	assert.Equal(t, http.StatusOK, rec.Code)
	check(rec)
	// Not no-referrer: that makes browsers send `Origin: null` on the form's
	// own POST, which emailConfirm rejects.
	assert.Equal(t, "same-origin", rec.Header().Get("Referrer-Policy"))

	// The form's POST carries this host's own origin and must be accepted.
	req0 := httptest.NewRequest(http.MethodPost, "/auth/email/confirm", strings.NewReader("token="+raw))
	req0.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req0.Header.Set("Origin", authBaseURL())
	rec = httptest.NewRecorder()
	a.emailConfirm(rec, req0)
	assert.Equal(t, http.StatusSeeOther, rec.Code)

	req := httptest.NewRequest(http.MethodPost, "/auth/email/confirm", strings.NewReader("token=nope"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec = httptest.NewRecorder()
	a.emailConfirm(rec, req)
	assert.Equal(t, http.StatusGone, rec.Code)
	check(rec)
}
