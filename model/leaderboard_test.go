package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

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
		ID:             1,
		EventName:      "event one",
		EventDate:      d2,
		EventType:      "Working Group",
		AddedAttendees: []User{u1},
	}, {
		ID:             2,
		EventName:      "event two",
		EventDate:      d2,
		EventType:      "Protest",
		AddedAttendees: []User{u1},
	}, {
		ID:             3,
		EventName:      "event three",
		EventDate:      d2,
		EventType:      "Key Event",
		AddedAttendees: []User{u1},
	}, {
		ID:             4,
		EventName:      "event four",
		EventDate:      d2,
		EventType:      "Community",
		AddedAttendees: []User{u1},
	}, {
		ID:             5,
		EventName:      "event five",
		EventDate:      d2,
		EventType:      "Outreach",
		AddedAttendees: []User{u1},
	}, {
		ID:             6,
		EventName:      "event six",
		EventDate:      d2,
		EventType:      "Sanctuary",
		AddedAttendees: []User{u1},
	}}

	for _, e := range wantEvents {
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
