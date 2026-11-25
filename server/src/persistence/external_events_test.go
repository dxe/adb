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

func TestExternalEvents_FeatureAndCancelRejectOtherChapters(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()
	repo := NewExternalEventRepository(db)

	chapterA := seedChapterWithEvent(t, db, 9001, "Chapter A", "1001")
	chapterB := seedChapterWithEvent(t, db, 9002, "Chapter B", "1002")

	// Chapter B can't touch chapter A's event.
	require.ErrorIs(t, repo.SetFeaturedForChapter(chapterB, "1001", true), model.ErrExternalEventNotFound)
	require.ErrorIs(t, repo.CancelForChapter(chapterB, "1001"), model.ErrExternalEventNotFound)

	var featured, canceled bool
	require.NoError(t, db.Get(&featured, `SELECT featured FROM fb_events WHERE id = '1001'`))
	require.NoError(t, db.Get(&canceled, `SELECT is_canceled FROM fb_events WHERE id = '1001'`))
	require.False(t, featured)
	require.False(t, canceled)

	// Chapter A can.
	require.NoError(t, repo.SetFeaturedForChapter(chapterA, "1001", true))
	require.NoError(t, db.Get(&featured, `SELECT featured FROM fb_events WHERE id = '1001'`))
	require.True(t, featured)

	require.NoError(t, repo.CancelForChapter(chapterA, "1001"))
	events, err := repo.ListUpcomingForChapter(chapterA, time.Now().UTC())
	require.NoError(t, err)
	require.Empty(t, events)
}
