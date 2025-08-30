package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestGetEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1 := Activist{Name: "Hello", ChapterID: 1, Email: "test1@example.org", Phone: "123-456-7890"}
	a2 := Activist{Name: "Hi", ChapterID: 1, Email: "test2@example.org", Phone: "888-888-8888"}
	a1ID, err := CreateActivist(db, ActivistExtra{Activist: a1})
	require.NoError(t, err)
	a2ID, err := CreateActivist(db, ActivistExtra{Activist: a2})
	require.NoError(t, err)
	a1.ID = a1ID
	a2.ID = a2ID

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	require.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-01-16")
	require.NoError(t, err)
	var wantEvents = []Event{{
		ID:             1,
		EventName:      "event one",
		EventDate:      d1,
		EventType:      "Working Group",
		Attendees:      []string{"Hello"},
		AttendeeEmails: []string{"test1@example.org"},
		AttendeePhones: []string{"123-456-7890"},
		AttendeeIDs:    []int{a1.ID},
		ChapterID:      1,
	}, {
		ID:             2,
		EventName:      "event two",
		EventDate:      d2,
		EventType:      "Protest",
		Attendees:      []string{"Hello", "Hi"},
		AttendeeEmails: []string{"test1@example.org", "test2@example.org"},
		AttendeePhones: []string{"123-456-7890", "888-888-8888"},
		AttendeeIDs:    []int{a1.ID, a2.ID},
		ChapterID:      1,
	}}

	for _, e := range wantEvents {
		insert := Event{
			EventName: e.EventName,
			EventDate: e.EventDate,
			EventType: e.EventType,
			ChapterID: 1,
		}
		if e.ID == 1 {
			insert.AddedAttendees = []Activist{a1}
		} else if e.ID == 2 {
			insert.AddedAttendees = []Activist{a1, a2}
		}

		_, err := InsertUpdateEvent(db, insert)
		if err != nil {
			t.Fatal(err)
		}
	}

	gotEvents, err := GetEvents(db, GetEventOptions{})
	require.NoError(t, err)

	require.Len(t, wantEvents, 2)
	require.Len(t, gotEvents, 2)

	for i := range wantEvents {
		// We need to check time equality separately b/c
		// require.EqualValues doesn't call EventDate.Equal.
		require.True(t, wantEvents[i].EventDate.Equal(gotEvents[i].EventDate))

		wantEvents[i].EventDate = time.Time{}
		gotEvents[i].EventDate = time.Time{}
		require.EqualValues(t, wantEvents[i], gotEvents[i])
	}
}

func TestGetEvents_orderBy(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "Hello", SFBayChapterIdDevTest)
	require.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	require.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-01-16")
	require.NoError(t, err)
	var wantEvents = []Event{{
		ID:             1,
		EventName:      "earlier event",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
		ChapterID:      1,
	}, {
		ID:             2,
		EventName:      "later event",
		EventDate:      d2,
		EventType:      "Protest",
		AddedAttendees: []Activist{a1},
		ChapterID:      1,
	}}

	for _, e := range wantEvents {
		_, err := InsertUpdateEvent(db, Event{
			EventName:      e.EventName,
			EventDate:      e.EventDate,
			EventType:      e.EventType,
			AddedAttendees: e.AddedAttendees,
			ChapterID:      1,
		})
		require.NoError(t, err)
	}

	gotEvents, err := GetEvents(db, GetEventOptions{
		OrderBy: "e.date DESC",
	})
	require.NoError(t, err)

	require.Len(t, gotEvents, 2)

	// "later event" must be listed first
	require.Equal(t, gotEvents[0].EventName, "later event")
	require.Equal(t, gotEvents[1].EventName, "earlier event")
}

func TestInsertUpdateEvent(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "Hello", SFBayChapterIdDevTest)
	require.NoError(t, err)
	a2, err := GetOrCreateActivist(db, "Hi", SFBayChapterIdDevTest)
	require.NoError(t, err)

	event := Event{
		EventName:      "event one",
		EventDate:      time.Now(),
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
		ChapterID:      1,
	}

	eventID, err := InsertUpdateEvent(db, event)
	require.NoError(t, err)
	require.Equal(t, eventID, 1)

	var events []Event
	require.NoError(t,
		db.Select(&events, "select * from events where name = 'event one'"))

	require.Equal(t, len(events), 1)

	var attendees []int
	require.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))
	require.Equal(t, len(attendees), 1)

	event.ID = 1
	event.AddedAttendees = []Activist{a1, a2}

	eventID, err = InsertUpdateEvent(db, event)
	require.NoError(t, err)
	require.Equal(t, eventID, 1)

	events = nil
	require.NoError(t,
		db.Select(&events, "select * from events where name = 'event one'"))

	require.Equal(t, len(events), 1)

	attendees = nil
	require.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))
	require.Equal(t, len(attendees), 2)
}

func TestInsertUpdateEvent_noDuplicateAttendees(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "Hello", SFBayChapterIdDevTest)
	require.NoError(t, err)

	event := Event{
		EventName:      "event one",
		EventDate:      time.Now(),
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1, a1},
		ChapterID:      1,
	}

	eventID, err := InsertUpdateEvent(db, event)
	require.NoError(t, err)
	require.Equal(t, eventID, 1)

	var attendees []int
	require.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))
	require.Equal(t, len(attendees), 1)
}

func TestDeleteEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Set up two events
	a1, err := GetOrCreateActivist(db, "Hello", SFBayChapterIdDevTest)
	require.NoError(t, err)
	a2, err := GetOrCreateActivist(db, "Hi", SFBayChapterIdDevTest)
	require.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	require.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-01-16")
	require.NoError(t, err)
	var wantEvents = []Event{{
		ID:             1,
		EventName:      "event one",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
		ChapterID:      1,
	}, {
		ID:             2,
		EventName:      "event two",
		EventDate:      d2,
		EventType:      "Protest",
		AddedAttendees: []Activist{a1, a2},
		ChapterID:      1,
	}}

	for _, e := range wantEvents {
		_, err := InsertUpdateEvent(db, Event{
			EventName:      e.EventName,
			EventDate:      e.EventDate,
			EventType:      e.EventType,
			AddedAttendees: e.AddedAttendees,
			ChapterID:      1,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Delete the first event
	err = DeleteEvent(db, 1, 1)
	require.NoError(t, err)

	gotEvents, err := GetEvents(db, GetEventOptions{})
	if err != nil {
		t.Fatal(err)
	}

	require.Len(t, gotEvents, 1)

	// Make sure we got the 2nd event back
	gotEvent := gotEvents[0]
	wantEvent := wantEvents[1]

	require.True(t, wantEvent.EventDate.Equal(gotEvent.EventDate))
	gotEvent.EventDate = time.Time{}
	wantEvent.EventDate = time.Time{}

	// Make sure that no attendance exists for the first event.
	var attendees []int
	require.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))

	require.Len(t, attendees, 0)
}

func TestCleanEventAttendanceData(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	testAttendees := []string{"New Person", "Another person", "A third person"}

	gotActivists, err := cleanEventAttendanceData(db, testAttendees, 1)
	require.NoError(t, err)

	gotActivistNames := map[string]struct{}{}
	for _, a := range gotActivists {
		gotActivistNames[a.Name] = struct{}{}
	}

	wantActivistNames := map[string]struct{}{
		"New Person":     struct{}{},
		"Another Person": struct{}{},
		"A Third Person": struct{}{},
	}
	require.Equal(t, gotActivistNames, wantActivistNames)
}
