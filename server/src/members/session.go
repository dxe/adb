package members

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/dxe/adb/config"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

// Login-token storage and stateless signed-cookie sessions for the website
// member auth flow (see auth_flow.go). Login tokens are 256-bit crypto/rand
// values; only their SHA-256 hashes are stored. Sessions are not stored: they
// are signed gorilla/sessions cookies, like ADB admin auth, and are never
// revoked server-side.

const (
	sessionLifetime = 30 * 24 * time.Hour
	// The cookie is re-issued (sliding expiry) when older than this.
	sessionRefreshAfter = 24 * time.Hour
	loginTokenLifetime  = 15 * time.Minute
)

// sessionCookieName is __Host- prefixed in prod (Secure, Path=/, no Domain, as
// the prefix requires); dev runs over http, where __Host- cookies are refused.
func sessionCookieName() string {
	if config.IsProd {
		return "__Host-dxe_member_session"
	}
	return "dxe_member_session"
}

// errInvalidToken is returned when a login token is unknown, expired, or
// already used.
var errInvalidToken = errors.New("invalid or expired token")

// newToken returns a random 256-bit token (hex) and its SHA-256 hash.
func newToken() (raw string, hash []byte, err error) {
	var buf [32]byte
	if _, err := rand.Read(buf[:]); err != nil {
		return "", nil, err
	}
	raw = hex.EncodeToString(buf[:])
	return raw, hashToken(raw), nil
}

func hashToken(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}

type memberStore struct {
	db *sqlx.DB
}

// createLoginToken stores a new single-use login token for email and returns
// the raw token to put in the emailed link.
func (st memberStore) createLoginToken(ctx context.Context, email, nextURL string, now time.Time) (string, error) {
	raw, hash, err := newToken()
	if err != nil {
		return "", err
	}
	_, err = st.db.ExecContext(ctx,
		`INSERT INTO member_login_tokens (token_hash, email, next_url, created_at, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		hash, email, nextURL, now.UTC(), now.Add(loginTokenLifetime).UTC())
	if err != nil {
		return "", fmt.Errorf("insert login token: %w", err)
	}
	return raw, nil
}

// peekLoginToken returns the email a still-valid token belongs to without
// consuming it (link scanners prefetch GETs).
func (st memberStore) peekLoginToken(ctx context.Context, raw string, now time.Time) (email string, err error) {
	err = st.db.GetContext(ctx, &email,
		`SELECT email FROM member_login_tokens
		 WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?`,
		hashToken(raw), now.UTC())
	if errors.Is(err, sql.ErrNoRows) {
		return "", errInvalidToken
	}
	return email, err
}

// loginTokenNext returns where a token's link was going, whether or not it has
// been used or has expired (rows are deleted a day after expiry).
func (st memberStore) loginTokenNext(ctx context.Context, raw string) (string, error) {
	var next string
	err := st.db.GetContext(ctx, &next,
		`SELECT next_url FROM member_login_tokens WHERE token_hash = ?`, hashToken(raw))
	return next, err
}

// consumeLoginToken atomically marks the token used and returns its email and
// next URL. A second call with the same token, or a call after expiry,
// returns errInvalidToken.
func (st memberStore) consumeLoginToken(ctx context.Context, raw string, now time.Time) (email, nextURL string, err error) {
	hash := hashToken(raw)
	res, err := st.db.ExecContext(ctx,
		`UPDATE member_login_tokens SET used_at = ?
		 WHERE token_hash = ? AND used_at IS NULL AND expires_at > ?`,
		now.UTC(), hash, now.UTC())
	if err != nil {
		return "", "", err
	}
	if n, err := res.RowsAffected(); err != nil {
		return "", "", err
	} else if n != 1 {
		return "", "", errInvalidToken
	}
	var row struct {
		Email   string `db:"email"`
		NextURL string `db:"next_url"`
	}
	if err := st.db.GetContext(ctx, &row,
		`SELECT email, next_url FROM member_login_tokens WHERE token_hash = ?`, hash); err != nil {
		return "", "", err
	}
	return row.Email, row.NextURL, nil
}

// memberSessions is the members store, separate from admin auth's, signed with
// the same secret. Its codec MaxAge also rejects old cookies by signed timestamp.
var memberSessions = func() *sessions.CookieStore {
	store := sessions.NewCookieStore([]byte(config.CookieSecret))
	store.MaxAge(int(sessionLifetime / time.Second))
	return store
}()

// saveSession writes a signed session cookie for email, issued at now.
func saveSession(w http.ResponseWriter, r *http.Request, email string, now time.Time) error {
	// A decode error means a bad old cookie; Get returns a fresh session then.
	sess, _ := memberSessions.Get(r, sessionCookieName())
	sess.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   int(sessionLifetime / time.Second),
		HttpOnly: true,
		Secure:   config.IsProd,
		SameSite: http.SameSiteLaxMode,
	}
	sess.Values["email"] = email
	sess.Values["issued"] = now.Unix()
	return memberSessions.Save(r, w, sess)
}

// readSession returns the signed-in email and when the cookie was issued, or
// ok=false if the cookie is missing, tampered with, or older than 30 days.
func readSession(r *http.Request, now time.Time) (email string, issuedAt time.Time, ok bool) {
	if _, err := r.Cookie(sessionCookieName()); err != nil {
		return "", time.Time{}, false
	}
	sess, err := memberSessions.Get(r, sessionCookieName())
	if err != nil {
		return "", time.Time{}, false
	}
	email, _ = sess.Values["email"].(string)
	issued, _ := sess.Values["issued"].(int64)
	if email == "" || issued == 0 {
		return "", time.Time{}, false
	}
	issuedAt = time.Unix(issued, 0)
	if age := now.Sub(issuedAt); age >= sessionLifetime || age < -time.Minute {
		return "", time.Time{}, false
	}
	return email, issuedAt, true
}

// deleteExpired removes login tokens that expired before cutoff.
func (st memberStore) deleteExpired(ctx context.Context, cutoff time.Time) error {
	_, err := st.db.ExecContext(ctx, `DELETE FROM member_login_tokens WHERE expires_at <= ?`, cutoff.UTC())
	return err
}
