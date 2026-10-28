package members

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// Session and login-token storage for the website member auth flow
// (see auth_flow.go). Raw tokens are 256-bit crypto/rand values; only their
// SHA-256 hashes are stored.

const (
	sessionCookieName = "dxe_member_session"
	sessionLifetime   = 30 * 24 * time.Hour
	// Sessions are extended (sliding expiry) at most this often.
	sessionRefreshAfter = 24 * time.Hour
	loginTokenLifetime  = 15 * time.Minute
)

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

// createSession stores a new session for email and returns the raw cookie value.
func (st memberStore) createSession(ctx context.Context, email string, now time.Time) (string, error) {
	raw, hash, err := newToken()
	if err != nil {
		return "", err
	}
	_, err = st.db.ExecContext(ctx,
		`INSERT INTO member_sessions (id_hash, email, created_at, expires_at, last_seen_at)
		 VALUES (?, ?, ?, ?, ?)`,
		hash, email, now.UTC(), now.Add(sessionLifetime).UTC(), now.UTC())
	if err != nil {
		return "", fmt.Errorf("insert session: %w", err)
	}
	return raw, nil
}

// lookupSession returns the session's email, or ("", nil) if the session is
// unknown or expired. refreshed is true when the expiry was extended, in which
// case the caller should re-issue the cookie.
func (st memberStore) lookupSession(ctx context.Context, raw string, now time.Time) (email string, refreshed bool, err error) {
	hash := hashToken(raw)
	var row struct {
		Email      string    `db:"email"`
		LastSeenAt time.Time `db:"last_seen_at"`
	}
	err = st.db.GetContext(ctx, &row,
		`SELECT email, last_seen_at FROM member_sessions WHERE id_hash = ? AND expires_at > ?`,
		hash, now.UTC())
	if errors.Is(err, sql.ErrNoRows) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	if now.Sub(row.LastSeenAt) >= sessionRefreshAfter {
		if _, err := st.db.ExecContext(ctx,
			`UPDATE member_sessions SET last_seen_at = ?, expires_at = ? WHERE id_hash = ?`,
			now.UTC(), now.Add(sessionLifetime).UTC(), hash); err != nil {
			return "", false, err
		}
		refreshed = true
	}
	return row.Email, refreshed, nil
}

func (st memberStore) deleteSession(ctx context.Context, raw string) error {
	_, err := st.db.ExecContext(ctx, `DELETE FROM member_sessions WHERE id_hash = ?`, hashToken(raw))
	return err
}

// deleteExpired removes expired sessions and login tokens.
func (st memberStore) deleteExpired(ctx context.Context, now time.Time) error {
	if _, err := st.db.ExecContext(ctx, `DELETE FROM member_sessions WHERE expires_at <= ?`, now.UTC()); err != nil {
		return err
	}
	_, err := st.db.ExecContext(ctx, `DELETE FROM member_login_tokens WHERE expires_at <= ?`, now.UTC())
	return err
}
