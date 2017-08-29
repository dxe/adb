package model

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

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

	if len(gotNames) != len(wantNames) {
		t.Fatalf("gotNames and wantNames must have the same length.")
	}

	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("gotNames and wantNames must be equal")
		}
	}
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
		ID:             1,
		EventName:      "event one",
		EventDate:      d2,
		EventType:      "Working Group",
		AddedAttendees: []User{u1},
	}, {
		ID:             2,
		EventName:      "event two",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []User{u1},
	}, {
		ID:             3,
		EventName:      "event three",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []User{u1},
	}, {
		ID:             4,
		EventName:      "event four",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []User{u1},
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
			EventName:      e.EventName,
			EventDate:      e.EventDate,
			EventType:      e.EventType,
			AddedAttendees: e.AddedAttendees,
		})
		if err != nil {
			t.Fatal(err)
		}
	}
}

func TestHideUser(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Test that deleting users works
	u1, err := GetOrCreateUser(db, "Test User")
	assert.NoError(t, err)

	u2, err := GetOrCreateUser(db, "Another Test User")
	assert.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	assert.NoError(t, err)

	eventID, err := InsertUpdateEvent(db, Event{
		EventName:      "my event",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []User{u1, u2},
	})

	assert.NoError(t, HideUser(db, u1.ID))

	// Hidden users should not show up in the autocompleted names
	names := GetAutocompleteNames(db)
	assert.Equal(t, len(names), 1)
	assert.Equal(t, names[0], u2.Name)

	// Hidden users should not show up in GetUsersJSON unless
	// Hidden = true.
	unhiddenUsers, err := GetUsersJSON(db, GetUserOptions{})
	assert.NoError(t, err)
	assert.Equal(t, len(unhiddenUsers), 1)
	assert.Equal(t, unhiddenUsers[0].ID, u2.ID)

	hiddenUsers, err := GetUsersJSON(db, GetUserOptions{Hidden: true})
	assert.NoError(t, err)
	assert.Equal(t, len(hiddenUsers), 1)
	assert.Equal(t, hiddenUsers[0].ID, u1.ID)

	// Hidden users should show up in GetUserJSON
	u1JSON, err := GetUserJSON(db, GetUserOptions{ID: u1.ID})
	assert.NoError(t, err)
	assert.Equal(t, u1JSON.ID, u1.ID)

	// Hidden users *should* show up in the event attendance
	event, err := GetEvent(db, GetEventOptions{EventID: eventID})
	assert.NoError(t, err)
	assert.Equal(t, len(event.Attendees), 2)
	assert.Equal(t, event.Attendees[0].ID, u1.ID)
	assert.Equal(t, event.Attendees[1].ID, u2.ID)

	attendanceNames, err := GetEventAttendance(db, eventID)
	assert.NoError(t, err)
	assert.Equal(t, len(attendanceNames), 2)
	assert.Equal(t, attendanceNames[0], u1.Name)
	assert.Equal(t, attendanceNames[1], u2.Name)
}

func TestMergeUser(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Test that deleting users works
	u1, err := GetOrCreateUser(db, "Test User")
	assert.NoError(t, err)

	u2, err := GetOrCreateUser(db, "Another Test User")
	assert.NoError(t, err)

	u3, err := GetOrCreateUser(db, "A Third Test User")
	assert.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-04-15")
	assert.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-04-16")
	assert.NoError(t, err)
	d3, err := time.Parse("2006-01-02", "2017-04-17")
	assert.NoError(t, err)

	insertEvents := []Event{{
		ID:             1,
		EventName:      "event one",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []User{u1, u3},
	}, {
		ID:             2,
		EventName:      "event two",
		EventDate:      d2,
		EventType:      "Working Group",
		AddedAttendees: []User{u1, u2, u3},
	}, {
		ID:             3,
		EventName:      "event three",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []User{u2, u3},
	}}
	mustInsertAllEvents(t, db, insertEvents)

	assert.NoError(t, MergeUser(db, u1.ID, u2.ID))

	e1, err := GetEvent(db, GetEventOptions{EventID: 1})
	assert.NoError(t, err)
	assert.Equal(t, len(e1.Attendees), 2)
	assert.Equal(t, e1.Attendees[0].ID, u2.ID)
	assert.Equal(t, e1.Attendees[1].ID, u3.ID)

	e2, err := GetEvent(db, GetEventOptions{EventID: 2})
	assert.NoError(t, err)
	assert.Equal(t, len(e2.Attendees), 2)
	assert.Equal(t, e2.Attendees[0].ID, u2.ID)
	assert.Equal(t, e2.Attendees[1].ID, u3.ID)

	e3, err := GetEvent(db, GetEventOptions{EventID: 3})
	assert.NoError(t, err)
	assert.Equal(t, len(e3.Attendees), 2)
	assert.Equal(t, e3.Attendees[0].ID, u2.ID)
	assert.Equal(t, e3.Attendees[1].ID, u3.ID)
}
