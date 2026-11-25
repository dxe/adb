package persistence

import (
	"testing"
	"time"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testdb"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// seedChapterWithEvent creates a chapter whose Facebook page is `pageID` and
// one upcoming event on that page. It returns the chapter ID.
func seedChapterWithEvent(t *testing.T, db *sqlx.DB, pageID int, name, eventID string) int {
	t.Helper()
	_, err := db.Exec(`INSERT INTO fb_pages (id, name, fb_url) VALUES (?, ?, '')`, pageID, name)
	require.NoError(t, err)
	var chapterID int
	require.NoError(t, db.Get(&chapterID, `SELECT chapter_id FROM fb_pages WHERE id = ?`, pageID))

	start := time.Now().UTC().Add(48 * time.Hour)
	require.NoError(t, model.UpsertExternalEvent(db, model.ExternalEvent{
		ID: eventID, PageID: pageID, Name: name + " event", StartTime: start, EndTime: start.Add(time.Hour),
	}))
	return chapterID
}

func TestExternalEvents_ListOnlyReturnsOwnChapter(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	repo := NewExternalEventRepository(db)

	chapterA := seedChapterWithEvent(t, db, 9001, "Chapter A", "1001")
	chapterB := seedChapterWithEvent(t, db, 9002, "Chapter B", "1002")

	events, err := repo.ListUpcomingForChapter(chapterA, time.Now().UTC())
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "1001", events[0].ID)

	events, err = repo.ListUpcomingForChapter(chapterB, time.Now().UTC())
	require.NoError(t, err)
	require.Len(t, events, 1)
	require.Equal(t, "1002", events[0].ID)
}

func TestExternalEvents_ListForChapterWithoutPageIsEmpty(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	repo := NewExternalEventRepository(db)

	seedChapterWithEvent(t, db, 9001, "Chapter A", "1001")
	// Chapter with no Facebook page (id 0).
	_, err := db.Exec(`INSERT INTO fb_pages (id, name, fb_url) VALUES (0, 'No Page', '')`)
	require.NoError(t, err)
	var chapterID int
	require.NoError(t, db.Get(&chapterID, `SELECT chapter_id FROM fb_pages WHERE name = 'No Page'`))

	events, err := repo.ListUpcomingForChapter(chapterID, time.Now().UTC())
	require.NoError(t, err)
	require.Empty(t, events)

	events, err = repo.ListUpcomingForChapter(999999, time.Now().UTC())
	require.NoError(t, err)
	require.Empty(t, events)
}

func TestExternalEvents_GetChapterIDAndWrites(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	repo := NewExternalEventRepository(db)

	chapterA := seedChapterWithEvent(t, db, 9001, "Chapter A", "1001")
	chapterB := seedChapterWithEvent(t, db, 9002, "Chapter B", "1002")

	got, err := repo.GetChapterID("1001")
	require.NoError(t, err)
	require.Equal(t, chapterA, got)
	got, err = repo.GetChapterID("1002")
	require.NoError(t, err)
	require.Equal(t, chapterB, got)
	_, err = repo.GetChapterID("424242")
	require.ErrorIs(t, err, model.ErrNotFound)

	var featured, canceled bool
	selectFlags := func() {
		require.NoError(t, db.Get(&featured, `SELECT featured FROM fb_events WHERE id = 1001`))
		require.NoError(t, db.Get(&canceled, `SELECT is_canceled FROM fb_events WHERE id = 1001`))
	}

	// The SQL itself refuses events on another chapter's page.
	require.ErrorIs(t, repo.SetFeatured(chapterB, "1001", true), model.ErrNotFound)
	require.ErrorIs(t, repo.Cancel(chapterB, "1001"), model.ErrNotFound)
	require.ErrorIs(t, repo.SetFeatured(chapterA, "424242", true), model.ErrNotFound)
	selectFlags()
	require.False(t, featured)
	require.False(t, canceled)

	require.NoError(t, repo.SetFeatured(chapterA, "1001", true))
	// Repeating a no-op write (MySQL reports 0 rows changed) is still success.
	require.NoError(t, repo.SetFeatured(chapterA, "1001", true))
	selectFlags()
	require.True(t, featured)

	require.NoError(t, repo.Cancel(chapterA, "1001"))
	require.NoError(t, repo.Cancel(chapterA, "1001"))
	selectFlags()
	require.True(t, canceled)
	events, err := repo.ListUpcomingForChapter(chapterA, time.Now().UTC())
	require.NoError(t, err)
	require.Empty(t, events)
}
