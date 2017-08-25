package model

import (
	"fmt"
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

func TestDeleteUser(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Test that deleting users works
	u1, err := GetOrCreateUser(db, "Test User")
	assert.NoError(t, err)

	u2, err := GetOrCreateUser(db, "Another Test User")
	assert.NoError(t, err)

	assert.NoError(t, DeleteUser(db, u1.ID))

	users, err := GetUsers(db)
	assert.NoError(t, err)
	assert.Equal(t, len(users), 1)
	assert.Equal(t, users[0].ID, u2.ID)

	// Test that deleted users also have all of their event
	// attendance deleted.

	u3, err := GetOrCreateUser(db, "Yay Test User")
	assert.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	assert.NoError(t, err)

	eventID, err := InsertUpdateEvent(db, Event{
		EventName:      "event one",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []User{u2, u3},
	})

	assert.NoError(t, DeleteUser(db, u2.ID))

	e, err := GetEvent(db, GetEventOptions{EventID: eventID})
	assert.NoError(t, err)

	assert.Equal(t, len(e.Attendees), 1)
	assert.Equal(t, u3.ID, e.Attendees[0].ID)
}
