package members

import (
	"context"
	"testing"
	"time"

	"github.com/dxe/adb/testdb"
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

func TestSession_LookupExpiryRefreshDelete(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	st := memberStore{db}
	ctx := context.Background()
	now := time.Now()

	raw, err := st.createSession(ctx, "a@example.org", now)
	require.NoError(t, err)

	email, refreshed, err := st.lookupSession(ctx, raw, now.Add(time.Hour))
	require.NoError(t, err)
	assert.Equal(t, "a@example.org", email)
	assert.False(t, refreshed)

	// After a day the expiry slides forward.
	email, refreshed, err = st.lookupSession(ctx, raw, now.Add(25*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, "a@example.org", email)
	assert.True(t, refreshed)

	// Past the original 30 days but within the extended window.
	email, _, err = st.lookupSession(ctx, raw, now.Add(sessionLifetime+time.Hour))
	require.NoError(t, err)
	assert.Equal(t, "a@example.org", email)

	// Fully expired (a fresh session that is never refreshed).
	raw2, err := st.createSession(ctx, "b@example.org", now)
	require.NoError(t, err)
	email, _, err = st.lookupSession(ctx, raw2, now.Add(sessionLifetime+time.Second))
	require.NoError(t, err)
	assert.Empty(t, email)

	// Unknown, and deleted.
	email, _, err = st.lookupSession(ctx, "nope", now)
	require.NoError(t, err)
	assert.Empty(t, email)
	require.NoError(t, st.deleteSession(ctx, raw))
	email, _, err = st.lookupSession(ctx, raw, now.Add(time.Hour))
	require.NoError(t, err)
	assert.Empty(t, email)
}
