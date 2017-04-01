package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestAutocompleteActivistsHandler(t *testing.T) {
	db := NewDB(":memory:")
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

func TestGetEvents(t *testing.T) {
	db := NewDB(":memory:")
	defer db.Close()

	var wantEvents = []Event{{
		ID:        1,
		EventName: "event one",
		EventDate: time.Now(),
		EventType: "Working Group",
	}, {
		ID:        2,
		EventName: "event two",
		EventDate: time.Now(),
		EventType: "Protest",
	}}

	for _, e := range wantEvents {
		err := InsertNewEvent(db, NewEvent{
			EventName: e.EventName,
			EventDate: e.EventDate,
			EventType: e.EventType,
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
