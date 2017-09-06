package model

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestAutocompleteActivistsHandler(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	_, err := GetOrCreateActivist(db, "Activist One")
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetOrCreateActivist(db, "Activist Two")
	if err != nil {
		t.Fatal(err)
	}

	gotNames := GetAutocompleteNames(db)
	wantNames := []string{"Activist One", "Activist Two"}

	if len(gotNames) != len(wantNames) {
		t.Fatalf("gotNames and wantNames must have the same length.")
	}

	for i := range wantNames {
		if gotNames[i] != wantNames[i] {
			t.Fatalf("gotNames and wantNames must be equal")
		}
	}
}

func TestGetActivistEventData(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "Test Activist")
	require.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-04-15")
	require.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-04-16")
	require.NoError(t, err)
	d3, err := time.Parse("2006-01-02", "2017-04-17")
	require.NoError(t, err)

	// These events are intentionally out of order
	insertEvents := []Event{{
		ID:             1,
		EventName:      "event one",
		EventDate:      d2,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
	}, {
		ID:             2,
		EventName:      "event two",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
	}, {
		ID:             3,
		EventName:      "event three",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
	}, {
		ID:             4,
		EventName:      "event four",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
	}}
	mustInsertAllEvents(t, db, insertEvents)

	d, err := a1.GetActivistEventData(db)
	require.NoError(t, err)

	d.FirstEvent.Equal(d1)
	d.LastEvent.Equal(d3)
	require.Equal(t, d.TotalEvents, 4)
}

func TestGetActivistEventData_noEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "Test Activist")
	require.NoError(t, err)

	d, err := a1.GetActivistEventData(db)
	require.NoError(t, err)

	require.Equal(t, d, ActivistEventData{
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

func TestHideActivist(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Test that deleting activists works
	a1, err := GetOrCreateActivist(db, "Test Activist")
	require.NoError(t, err)

	a2, err := GetOrCreateActivist(db, "Another Test Activist")
	require.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-01-15")
	require.NoError(t, err)

	eventID, err := InsertUpdateEvent(db, Event{
		EventName:      "my event",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1, a2},
	})

	require.NoError(t, HideActivist(db, a1.ID))

	// Hidden activists should not show up in the autocompleted names
	names := GetAutocompleteNames(db)
	require.Equal(t, len(names), 1)
	require.Equal(t, names[0], a2.Name)

	// Hidden activists should not show up in GetActivistsJSON unless
	// Hidden = true.
	unhiddenActivists, err := GetActivistsJSON(db, GetActivistOptions{})
	require.NoError(t, err)
	require.Equal(t, len(unhiddenActivists), 1)
	require.Equal(t, unhiddenActivists[0].ID, a2.ID)

	hiddenActivists, err := GetActivistsJSON(db, GetActivistOptions{Hidden: true})
	require.NoError(t, err)
	require.Equal(t, len(hiddenActivists), 1)
	require.Equal(t, hiddenActivists[0].ID, a1.ID)

	// Hidden activists should show up in GetActivistJSON
	a1JSON, err := GetActivistJSON(db, GetActivistOptions{ID: a1.ID})
	require.NoError(t, err)
	require.Equal(t, a1JSON.ID, a1.ID)

	// Hidden activists *should* show up in the event attendance
	event, err := GetEvent(db, GetEventOptions{EventID: eventID})
	require.NoError(t, err)
	require.Equal(t, len(event.Attendees), 2)
	require.Equal(t, event.Attendees[0], a1.Name)
	require.Equal(t, event.Attendees[1], a2.Name)

	attendanceNames, err := GetEventAttendance(db, eventID)
	require.NoError(t, err)
	require.Equal(t, len(attendanceNames), 2)
	require.Equal(t, attendanceNames[0], a1.Name)
	require.Equal(t, attendanceNames[1], a2.Name)
}

func TestMergeActivist(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Test that deleting activists works
	a1, err := GetOrCreateActivist(db, "Test Activist")
	require.NoError(t, err)

	a2, err := GetOrCreateActivist(db, "Another Test Activist")
	require.NoError(t, err)

	a3, err := GetOrCreateActivist(db, "A Third Test Activist")
	require.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-04-15")
	require.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-04-16")
	require.NoError(t, err)
	d3, err := time.Parse("2006-01-02", "2017-04-17")
	require.NoError(t, err)

	insertEvents := []Event{{
		ID:             1,
		EventName:      "event one",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1, a3},
	}, {
		ID:             2,
		EventName:      "event two",
		EventDate:      d2,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1, a2, a3},
	}, {
		ID:             3,
		EventName:      "event three",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a2, a3},
	}}
	mustInsertAllEvents(t, db, insertEvents)

	require.NoError(t, MergeActivist(db, a1.ID, a2.ID))

	e1, err := GetEvent(db, GetEventOptions{EventID: 1})
	require.NoError(t, err)
	require.Equal(t, len(e1.Attendees), 2)
	require.Equal(t, e1.Attendees[0], a2.Name)
	require.Equal(t, e1.Attendees[1], a3.Name)

	e2, err := GetEvent(db, GetEventOptions{EventID: 2})
	require.NoError(t, err)
	require.Equal(t, len(e2.Attendees), 2)
	require.Equal(t, e2.Attendees[0], a2.Name)
	require.Equal(t, e2.Attendees[1], a3.Name)

	e3, err := GetEvent(db, GetEventOptions{EventID: 3})
	require.NoError(t, err)
	require.Equal(t, len(e3.Attendees), 2)
	require.Equal(t, e3.Attendees[0], a2.Name)
	require.Equal(t, e3.Attendees[1], a3.Name)
}
