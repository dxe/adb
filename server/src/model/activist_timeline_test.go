package model

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dxe/adb/pkg/shared"
	"github.com/dxe/adb/testdb"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func timelineAdmin() ADBUser {
	return ADBUser{ID: DevTestUserId, Email: DevTestUserEmail, Name: "Admin", Roles: []string{shared.RoleAdmin}}
}

func timelineOrganizer(chapterID int) ADBUser {
	return ADBUser{ID: DevTestUserId, Email: DevTestUserEmail, Name: "Dev User", Roles: []string{shared.RoleOrganizer}, ChapterID: chapterID}
}

func mustInsertTimelineEvent(t *testing.T, db *sqlx.DB, activistID int, name, date string) {
	t.Helper()
	eventDate, err := time.Parse(EventDateLayout, date)
	require.NoError(t, err)
	_, err = InsertUpdateEvent(db, Event{
		EventName:      name,
		EventDate:      eventDate,
		EventType:      "Working Group",
		ChapterID:      SFBayChapterIdDevTest,
		AddedAttendees: []Activist{{ID: activistID}},
	})
	require.NoError(t, err)
}

// mustSetTimelineEventStartTime gives the named event a wall-clock start time
// in the supplied zone, which InsertUpdateEvent does not take.
func mustSetTimelineEventStartTime(t *testing.T, db *sqlx.DB, name, startTime, timezone string) {
	t.Helper()
	res, err := db.Exec(`UPDATE events SET start_time = ?, timezone = ? WHERE name = ?`,
		startTime, timezone, name)
	require.NoError(t, err)
	affected, err := res.RowsAffected()
	require.NoError(t, err)
	require.EqualValues(t, 1, affected)
}

func mustInsertTimelineInteraction(t *testing.T, db *sqlx.DB, activistID int, timestamp, method string) {
	t.Helper()
	_, err := db.Exec(`INSERT INTO interactions (activist_id, user_id, timestamp, method, outcome, notes)
		VALUES (?, ?, ?, ?, 'outcome', 'notes')`, activistID, DevTestUserId, timestamp, method)
	require.NoError(t, err)
}

func TestGetActivistTimeline_MergesEventsAndInteractionsNewestFirst(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	activist, err := GetOrCreateActivist(db, "Timeline Activist", SFBayChapterIdDevTest)
	require.NoError(t, err)

	mustInsertTimelineEvent(t, db, activist.ID, "older event", "2024-05-01")
	mustInsertTimelineEvent(t, db, activist.ID, "newer event", "2024-05-10")
	mustInsertTimelineInteraction(t, db, activist.ID, "2024-05-05 12:00:00", "phone")
	mustInsertTimelineInteraction(t, db, activist.ID, "2024-05-20 12:00:00", "email")

	timeline, err := GetActivistTimeline(db, timelineAdmin(), activist.ID)
	require.NoError(t, err)
	require.False(t, timeline.Truncated)
	require.Len(t, timeline.Items, 4)

	type got struct{ typ, date string }
	var order []got
	for _, item := range timeline.Items {
		order = append(order, got{item.Type, item.Date})
	}
	require.Equal(t, []got{
		{TimelineItemTypeInteraction, "2024-05-20"},
		{TimelineItemTypeEvent, "2024-05-10"},
		{TimelineItemTypeInteraction, "2024-05-05"},
		{TimelineItemTypeEvent, "2024-05-01"},
	}, order)

	// Exactly one payload per item, matching its type.
	require.Nil(t, timeline.Items[1].Interaction)
	require.Equal(t, &ActivistTimelineEvent{
		Name: "newer event",
		Type: "Working Group",
	}, timeline.Items[1].Event)

	require.Nil(t, timeline.Items[2].Event)
	require.Equal(t, &ActivistTimelineInteraction{
		Method:   "phone",
		Outcome:  "outcome",
		Notes:    "notes",
		UserName: "Dev User",
	}, timeline.Items[2].Interaction)

	// Every item carries the instant it was ordered by; only the interaction's
	// time of day is real.
	require.Equal(t, "2024-05-05T12:00:00Z", timeline.Items[2].Timestamp)
	require.True(t, timeline.Items[2].HasTime)
	require.False(t, timeline.Items[1].HasTime)
}

func TestGetActivistTimeline_EventStartTimes(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	activist, err := GetOrCreateActivist(db, "Start Time Activist", SFBayChapterIdDevTest)
	require.NoError(t, err)

	// Same day: an event at 6:30pm, an event with no start time, and two
	// interactions bracketing the placeholder noon.
	mustInsertTimelineEvent(t, db, activist.ID, "evening event", "2024-05-01")
	mustSetTimelineEventStartTime(t, db, "evening event", "18:30:00", "America/Los_Angeles")
	mustInsertTimelineEvent(t, db, activist.ID, "untimed event", "2024-05-01")
	// 09:00 and 15:00 Pacific.
	mustInsertTimelineInteraction(t, db, activist.ID, "2024-05-01 16:00:00", "morning call")
	mustInsertTimelineInteraction(t, db, activist.ID, "2024-05-01 22:00:00", "afternoon call")

	timeline, err := GetActivistTimeline(db, timelineAdmin(), activist.ID)
	require.NoError(t, err)
	require.Len(t, timeline.Items, 4)

	// A start time places the event within its day rather than at either edge,
	// and an untimed event lands at noon: between the two calls, not below
	// both of them as a start-of-day placeholder would put it.
	type got struct {
		label     string
		timestamp string
		hasTime   bool
	}
	var order []got
	for _, item := range timeline.Items {
		label := ""
		if item.Event != nil {
			label = item.Event.Name
		} else {
			label = item.Interaction.Method
		}
		order = append(order, got{label, item.Timestamp, item.HasTime})
	}
	require.Equal(t, []got{
		{"evening event", "2024-05-02T01:30:00Z", true},
		{"afternoon call", "2024-05-01T22:00:00Z", true},
		{"untimed event", "2024-05-01T19:00:00Z", false},
		{"morning call", "2024-05-01T16:00:00Z", true},
	}, order)

	// The placeholder does not move the day the item is filed under, even
	// though its UTC instant lands on a different date.
	require.Equal(t, "2024-05-01", timeline.Items[2].Date)
}

func TestGetActivistTimeline_ChapterAccess(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	activist, err := GetOrCreateActivist(db, "Chapter Scoped", SFBayChapterIdDevTest)
	require.NoError(t, err)
	mustInsertTimelineInteraction(t, db, activist.ID, "2024-05-05 12:00:00", "phone")

	otherChapterID, err := InsertChapter(db, ChapterWithToken{Name: "Other Chapter"})
	require.NoError(t, err)
	require.NotEqual(t, SFBayChapterIdDevTest, otherChapterID)

	_, err = GetActivistTimeline(db, timelineOrganizer(otherChapterID), activist.ID)
	require.ErrorIs(t, err, ErrValidation)

	adminTimeline, err := GetActivistTimeline(db, timelineAdmin(), activist.ID)
	require.NoError(t, err)
	require.Len(t, adminTimeline.Items, 1)

	sameChapterTimeline, err := GetActivistTimeline(db, timelineOrganizer(SFBayChapterIdDevTest), activist.ID)
	require.NoError(t, err)
	require.Len(t, sameChapterTimeline.Items, 1)
}

func TestGetActivistTimeline_CapsAtMaxItems(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	activist, err := GetOrCreateActivist(db, "Busy Activist", SFBayChapterIdDevTest)
	require.NoError(t, err)

	// One extra beyond the cap, inserted as a single multi-row statement per
	// table so the test stays fast.
	const eventCount = ActivistTimelineMaxItems + 1
	start := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)
	var eventValues []string
	var eventArgs []interface{}
	for i := 0; i < eventCount; i++ {
		eventValues = append(eventValues, "(?, ?, ?, ?)")
		eventArgs = append(eventArgs,
			fmt.Sprintf("event %d", i),
			start.AddDate(0, 0, i).Format(EventDateLayout),
			"Working Group",
			SFBayChapterIdDevTest,
		)
	}
	_, err = db.Exec(`INSERT INTO events (name, date, event_type, chapter_id) VALUES `+
		strings.Join(eventValues, ","), eventArgs...)
	require.NoError(t, err)

	_, err = db.Exec(`INSERT INTO event_attendance (activist_id, event_id)
		SELECT ?, id FROM events`, activist.ID)
	require.NoError(t, err)

	timeline, err := GetActivistTimeline(db, timelineAdmin(), activist.ID)
	require.NoError(t, err)
	require.True(t, timeline.Truncated)
	require.Len(t, timeline.Items, ActivistTimelineMaxItems)
	// The newest item is kept and the oldest is dropped.
	require.Equal(t, start.AddDate(0, 0, eventCount-1).Format(EventDateLayout), timeline.Items[0].Date)
	require.Equal(t, start.AddDate(0, 0, 1).Format(EventDateLayout), timeline.Items[len(timeline.Items)-1].Date)
}

func TestGetActivistTimeline_NotFound(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	_, err := GetActivistTimeline(db, timelineAdmin(), 999999)
	require.ErrorIs(t, err, ErrNotFound)
}
