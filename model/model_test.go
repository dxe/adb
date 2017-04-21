package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func newTestDB() *sqlx.DB {
	db := NewDB("adb_user:adbpassword@/adb_test_db?parseTime=true")
	CreateDatabase(db)

	// Insert sample data
	db.MustExec(`TRUNCATE activists`)
	db.MustExec(`TRUNCATE events`)
	db.MustExec(`TRUNCATE event_attendance`)

	return db
}

func TestAutocompleteActivistsHandler(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	_, err := GetOrCreateUser(db, "User One")
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetOrCreateUser(db, "User Two")
	if err != nil {
		t.Fatal(err)
	}

	gotNames := GetAutocompleteNames(db)
	wantNames := []string{"User One", "User Two"}
	fmt.Println(gotNames, wantNames)
	if len(gotNames) != len(wantNames) {
		t.Fatalf("gotNames and wantNames must have the same length.")
	}

	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("gotNames and wantNames must be equal")
		}
	}
}

func TestGetEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	u1, err := GetOrCreateUser(db, "Hello")
	assert.NoError(t, err)
	u2, err := GetOrCreateUser(db, "Hi")
	assert.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	assert.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-01-16")
	assert.NoError(t, err)
	var wantEvents = []Event{{
		ID:        1,
		EventName: "event one",
		EventDate: d1,
		EventType: "Working Group",
		Attendees: []User{u1},
	}, {
		ID:        2,
		EventName: "event two",
		EventDate: d2,
		EventType: "Protest",
		Attendees: []User{u1, u2},
	}}

	for _, e := range wantEvents {
		_, err := InsertUpdateEvent(db, Event{
			EventName: e.EventName,
			EventDate: e.EventDate,
			EventType: e.EventType,
			Attendees: e.Attendees,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	gotEvents, err := GetEvents(db, GetEventOptions{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, wantEvents, 2)
	assert.Len(t, gotEvents, 2)

	for i := range wantEvents {
		// We need to check time equality separately b/c
		// assert.EqualValues doesn't call EventDate.Equal.
		assert.True(t, wantEvents[i].EventDate.Equal(gotEvents[i].EventDate))

		wantEvents[i].EventDate = time.Time{}
		gotEvents[i].EventDate = time.Time{}
		assert.EqualValues(t, wantEvents[i], gotEvents[i])
	}
}

func TestGetEvents_orderBy(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	u1, err := GetOrCreateUser(db, "Hello")
	assert.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	assert.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-01-16")
	assert.NoError(t, err)
	var wantEvents = []Event{{
		ID:        1,
		EventName: "earlier event",
		EventDate: d1,
		EventType: "Working Group",
		Attendees: []User{u1},
	}, {
		ID:        2,
		EventName: "later event",
		EventDate: d2,
		EventType: "Protest",
		Attendees: []User{u1},
	}}

	for _, e := range wantEvents {
		_, err := InsertUpdateEvent(db, Event{
			EventName: e.EventName,
			EventDate: e.EventDate,
			EventType: e.EventType,
			Attendees: e.Attendees,
		})
		assert.NoError(t, err)
	}

	gotEvents, err := GetEvents(db, GetEventOptions{
		OrderBy: "date DESC",
	})
	assert.NoError(t, err)

	assert.Len(t, gotEvents, 2)

	// "later event" must be listed first
	assert.Equal(t, gotEvents[0].EventName, "later event")
	assert.Equal(t, gotEvents[1].EventName, "earlier event")
}

func TestInsertUpdateEvent(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	u1, err := GetOrCreateUser(db, "Hello")
	assert.NoError(t, err)
	u2, err := GetOrCreateUser(db, "Hi")
	assert.NoError(t, err)

	event := Event{
		EventName: "event one",
		EventDate: time.Now(),
		EventType: "Working Group",
		Attendees: []User{u1},
	}

	eventID, err := InsertUpdateEvent(db, event)
	assert.NoError(t, err)
	assert.Equal(t, eventID, 1)

	var events []Event
	assert.NoError(t,
		db.Select(&events, "select * from events where name = 'event one'"))

	assert.Equal(t, len(events), 1)

	var attendees []int
	assert.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))
	assert.Equal(t, len(attendees), 1)

	event.ID = 1
	event.Attendees = []User{u1, u2}

	eventID, err = InsertUpdateEvent(db, event)
	assert.NoError(t, err)
	assert.Equal(t, eventID, 1)

	events = nil
	assert.NoError(t,
		db.Select(&events, "select * from events where name = 'event one'"))

	assert.Equal(t, len(events), 1)

	attendees = nil
	assert.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))
	assert.Equal(t, len(attendees), 2)
}

func TestInsertUpdateEvent_noDuplicateAttendees(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	u1, err := GetOrCreateUser(db, "Hello")
	assert.NoError(t, err)

	event := Event{
		EventName: "event one",
		EventDate: time.Now(),
		EventType: "Working Group",
		Attendees: []User{u1, u1},
	}

	eventID, err := InsertUpdateEvent(db, event)
	assert.NoError(t, err)
	assert.Equal(t, eventID, 1)

	var attendees []int
	assert.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))
	assert.Equal(t, len(attendees), 1)
}

func TestDeleteEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Set up two events
	u1, err := GetOrCreateUser(db, "Hello")
	assert.NoError(t, err)
	u2, err := GetOrCreateUser(db, "Hi")
	assert.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	assert.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-01-16")
	assert.NoError(t, err)
	var wantEvents = []Event{{
		ID:        1,
		EventName: "event one",
		EventDate: d1,
		EventType: "Working Group",
		Attendees: []User{u1},
	}, {
		ID:        2,
		EventName: "event two",
		EventDate: d2,
		EventType: "Protest",
		Attendees: []User{u1, u2},
	}}

	for _, e := range wantEvents {
		_, err := InsertUpdateEvent(db, Event{
			EventName: e.EventName,
			EventDate: e.EventDate,
			EventType: e.EventType,
			Attendees: e.Attendees,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

	// Delete the first event
	err = DeleteEvent(db, 1)
	assert.NoError(t, err)

	gotEvents, err := GetEvents(db, GetEventOptions{})
	if err != nil {
		t.Fatal(err)
	}

	assert.Len(t, gotEvents, 1)

	// Make sure we got the 2nd event back
	gotEvent := gotEvents[0]
	wantEvent := wantEvents[1]

	assert.True(t, wantEvent.EventDate.Equal(gotEvent.EventDate))
	gotEvent.EventDate = time.Time{}
	wantEvent.EventDate = time.Time{}

	// Make sure that no attendance exists for the first event.
	var attendees []int
	assert.NoError(t,
		db.Select(&attendees, "select activist_id from event_attendance where event_id = 1"))

	assert.Len(t, attendees, 0)
}

func TestGetUserEventData(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	u1, err := GetOrCreateUser(db, "Test User")
	assert.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-04-15")
	assert.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-04-16")
	assert.NoError(t, err)
	d3, err := time.Parse("2006-01-02", "2017-04-17")
	assert.NoError(t, err)

	// These events are intentionally out of order
	insertEvents := []Event{{
		ID:        1,
		EventName: "event one",
		EventDate: d2,
		EventType: "Working Group",
		Attendees: []User{u1},
	}, {
		ID:        2,
		EventName: "event two",
		EventDate: d1,
		EventType: "Working Group",
		Attendees: []User{u1},
	}, {
		ID:        3,
		EventName: "event three",
		EventDate: d3,
		EventType: "Working Group",
		Attendees: []User{u1},
	}, {
		ID:        4,
		EventName: "event four",
		EventDate: d3,
		EventType: "Working Group",
		Attendees: []User{u1},
	}}
	mustInsertAllEvents(t, db, insertEvents)

	d, err := u1.GetUserEventData(db)
	assert.NoError(t, err)

	d.FirstEvent.Equal(d1)
	d.LastEvent.Equal(d3)
	assert.Equal(t, d.TotalEvents, 4)
}

func TestGetUserEventData_noEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	u1, err := GetOrCreateUser(db, "Test User")
	assert.NoError(t, err)

	d, err := u1.GetUserEventData(db)
	assert.NoError(t, err)

	assert.Equal(t, d, UserEventData{
		FirstEvent:  nil,
		LastEvent:   nil,
		TotalEvents: 0,
	})
}

func mustInsertAllEvents(t *testing.T, db *sqlx.DB, events []Event) {
	for _, e := range events {
		_, err := InsertUpdateEvent(db, Event{
			EventName: e.EventName,
			EventDate: e.EventDate,
			EventType: e.EventType,
			Attendees: e.Attendees,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

// TODO: Finish this test.
func TestGetLeaderboardUsers(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Set up two events
	u1, err := GetOrCreateUser(db, "Hello")
	assert.NoError(t, err)

	d2, err := time.Parse("2006-01-02", "2017-04-16")
	assert.NoError(t, err)
	var wantEvents = []Event{{
		ID:        1,
		EventName: "event one",
		EventDate: d2,
		EventType: "Working Group",
		Attendees: []User{u1},
	}, {
		ID:        2,
		EventName: "event two",
		EventDate: d2,
		EventType: "Protest",
		Attendees: []User{u1},
	}, {
		ID:        3,
		EventName: "event three",
		EventDate: d2,
		EventType: "Key Event",
		Attendees: []User{u1},
	}, {
		ID:        4,
		EventName: "event four",
		EventDate: d2,
		EventType: "Community",
		Attendees: []User{u1},
	}, {
		ID:        5,
		EventName: "event five",
		EventDate: d2,
		EventType: "Outreach",
		Attendees: []User{u1},
	}, {
		ID:        6,
		EventName: "event six",
		EventDate: d2,
		EventType: "Sanctuary",
		Attendees: []User{u1},
	}}

	for _, e := range wantEvents {
		_, err := InsertUpdateEvent(db, Event{
			EventName: e.EventName,
			EventDate: e.EventDate,
			EventType: e.EventType,
			Attendees: e.Attendees,
		})
		if err != nil {
			t.Fatal(err)
		}
	}

}
