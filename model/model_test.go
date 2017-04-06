package model

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func newTestDB() *sqlx.DB {
	// TODO: Don't use the normal dev database for tests lol
	db := NewDB("adb_user:adbpassword@/adb_db?parseTime=true")
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

	gotEvents, err := GetEvents(db)
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
