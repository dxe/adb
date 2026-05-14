package model

import (
	"fmt"
	"testing"

	"github.com/dxe/adb/testdb"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestGetAdminChapterByID(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	id, insErr := InsertChapter(db, ChapterWithToken{
		Name: "FooChapter",
	})
	if insErr != nil {
		t.Fatalf("error inserting chapter: %v", insErr)
	}

	chapter, getErr := GetAdminChapterById(db, id)
	if getErr != nil {
		t.Fatalf("error getting chapter: %v", getErr)
	}

	require.Equal(t, "FooChapter", chapter.Name)
}

func TestGetChapterByID(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	id, insErr := InsertChapter(db, ChapterWithToken{
		Name:       "FooChapter",
		Flag:       "x",
		FbURL:      "fb-foo",
		TwitterURL: "tw-foo",
		InstaURL:   "ig-foo",
		Email:      "foo@example.org",
		Region:     "North America",
		Lat:        0.01,
		Lng:        0.02,
	})
	if insErr != nil {
		t.Fatalf("error inserting chapter: %v", insErr)
	}

	// InsertChapter does not set Facebook page ID and mailing list fields as it
	// is designed for the UI.
	_, err := db.Exec(
		`UPDATE fb_pages SET id = 999, ml_type = 'mailfoo', ml_radius = 10, ml_id = '12348' WHERE chapter_id = ?`,
		id,
	)
	if err != nil {
		t.Fatalf("error setting chapter fields: %v", err)
	}

	chapter, getErr := GetChapterById(db, id)
	if getErr != nil {
		t.Fatalf("error getting chapter: %v", getErr)
	}

	require.Equal(t, 999, chapter.FacebookID)
	require.Equal(t, "FooChapter", chapter.Name)
	require.Equal(t, "x", chapter.Flag)
	require.Equal(t, "fb-foo", chapter.FbURL)
	require.Equal(t, "tw-foo", chapter.TwitterURL)
	require.Equal(t, "ig-foo", chapter.InstaURL)
	require.Equal(t, "foo@example.org", chapter.Email)
	require.Equal(t, "North America", chapter.Region)
	require.Equal(t, 0.01, chapter.Lat)
	require.Equal(t, 0.02, chapter.Lng)
	require.Equal(t, "mailfoo", chapter.MailingListType)
	require.Equal(t, 10, chapter.MailingListRadius)
	require.Equal(t, "12348", chapter.MailingListID)
}

func insertChapters(db *sqlx.DB, chapters []ChapterWithToken) []int {
	ids := []int{}
	for _, chapter := range chapters {
		id, err := InsertChapter(db, chapter)
		if err != nil {
			panic(fmt.Errorf("error inserting chapter: %v", err))
		}
		ids = append(ids, id)
	}
	return ids
}

func TestFindNearestChaptersSortedByDistanceDeprecated(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	ids := insertChapters(db, []ChapterWithToken{
		{
			Lat: -50,
			Lng: -25,
		},
		{
			Lat: -50,
			Lng: -50,
		},
	})

	chapters, err := FindNearestChaptersSortedByDistanceDeprecated(db, -50.01, -50.01)
	require.NoError(t, err)

	require.Equal(t, chapters[0].ChapterID, ids[1])
	require.Equal(t, chapters[1].ChapterID, ids[0])
}

func TestFindNearestChaptersSortedByDistance(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	ids := insertChapters(db, []ChapterWithToken{
		{
			Lat: -50,
			Lng: -25,
		},
		{
			Lat: -50,
			Lng: -50,
		},
	})

	chapters, err := FindNearestChaptersSortedByDistance(db, -50.01, -50.01)
	require.NoError(t, err)

	require.Equal(t, chapters[0].ID, ids[1])
	require.Equal(t, chapters[1].ID, ids[0])
}
