package model

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func stringListToMap(l []string) map[string]struct{} {
	m := map[string]struct{}{}
	for _, i := range l {
		m[i] = struct{}{}
	}
	return m
}

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

	require.Equal(t, stringListToMap(gotNames), stringListToMap(wantNames),
		"gotNames and wantNames must be equal")
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

func TestGetActivistsJSON_RestrictDates(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "A")
	require.NoError(t, err)

	a2, err := GetOrCreateActivist(db, "B")
	require.NoError(t, err)

	a3, err := GetOrCreateActivist(db, "C")
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
		AddedAttendees: []Activist{a2, a3},
	}, {
		ID:             3,
		EventName:      "event three",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a2},
	}}
	mustInsertAllEvents(t, db, insertEvents)

	activists, err := GetActivistsJSON(db, GetActivistOptions{
		Order: DescOrder,
	})
	require.NoError(t, err)
	assertActivistJSONSliceContainsNames(t, activists, []string{"A", "B", "C"})

	activists, err = GetActivistsJSON(db, GetActivistOptions{
		Order:             DescOrder,
		LastEventDateFrom: "2017-04-17",
	})
	require.NoError(t, err)
	assertActivistJSONSliceContainsNames(t, activists, []string{"B"})

	activists, err = GetActivistsJSON(db, GetActivistOptions{
		Order:             DescOrder,
		LastEventDateFrom: "2017-04-16",
		LastEventDateTo:   "2017-04-17",
	})
	require.NoError(t, err)
	assertActivistJSONSliceContainsNames(t, activists, []string{"B", "C"})
}

func TestGetActivistsJSON_OrderField(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "A")
	require.NoError(t, err)

	a2, err := GetOrCreateActivist(db, "B")
	require.NoError(t, err)

	a3, err := GetOrCreateActivist(db, "C")
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
		AddedAttendees: []Activist{a2, a3},
	}, {
		ID:             3,
		EventName:      "event three",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a2},
	}}
	mustInsertAllEvents(t, db, insertEvents)

	activists, err := GetActivistsJSON(db, GetActivistOptions{
		Order:      AscOrder,
		OrderField: "a.name",
	})
	require.NoError(t, err)
	assertActivistJSONSliceContainsOrderedNames(t, activists, []string{"A", "B", "C"})

	activists, err = GetActivistsJSON(db, GetActivistOptions{
		Order:      DescOrder,
		OrderField: "last_event",
	})
	require.NoError(t, err)
	assertActivistJSONSliceContainsOrderedNames(t, activists, []string{"B", "C", "A"})
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

// Not Specfiying a starting name with ascending order
// and no limit, returns all activists
func TestActivistRange_noNameOrLimitAscOrder_returnsAllActivists(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activistsToInsert := []string{"A", "B", "C", "D"}
	insertTestActivists(t, db, activistsToInsert)
	activistOptions := ActivistRangeOptionsJSON{
		Order: AscOrder,
	}
	fetchedActivists, err := GetActivistRangeJSON(db, activistOptions)

	require.NoError(t, err)
	require.Equal(t, len(activistsToInsert), len(fetchedActivists))

	// For this test, fetched activists should be in the same order
	// as the activistsToInsert slice
	for idx, a := range fetchedActivists {
		require.Equal(t, activistsToInsert[idx], a.Name)
	}

}

// Not specifying a starting name with descending order
// and no limit, returns all activists in descending order
func TestActivistRange_noNameOrLimitDescOrder_returnsAllActivists(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activistsToInsert := []string{"A", "B", "C", "D"}
	insertTestActivists(t, db, activistsToInsert)
	activistOptions := ActivistRangeOptionsJSON{
		Order: DescOrder,
	}
	fetchedActivists, err := GetActivistRangeJSON(db, activistOptions)

	require.NoError(t, err)
	require.Equal(t, len(activistsToInsert), len(fetchedActivists))

	// For this test, fetched activists should be in reverse order as
	// the activistsToInsert slice
	reverseIdx := len(activistsToInsert) - 1
	for idx, a := range fetchedActivists {
		require.Equal(t, activistsToInsert[reverseIdx-idx], a.Name)
	}

}

// No limit, ascending order, specified name
// returns all activists with names greater than specified name
func TestActivistRange_NameNoLimitAscOrder_returnsSubsetOfActivists(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activistsToInsert := []string{"A", "B", "C", "D", "E", "F"}
	insertTestActivists(t, db, activistsToInsert)
	activistOptions := ActivistRangeOptionsJSON{
		Name:  "A",
		Order: AscOrder,
	}
	fetchedActivists, err := GetActivistRangeJSON(db, activistOptions)

	require.NoError(t, err)
	require.Equal(t, 5, len(fetchedActivists))

	insertedActivistsSubset := activistsToInsert[1:]
	for idx, a := range fetchedActivists {
		require.Equal(t, insertedActivistsSubset[idx], a.Name)
	}

	// If specified name is last, then result should be nil
	activistOptions.Name = "F"
	fetchedActivists, err = GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Nil(t, fetchedActivists)

}

// No limit, descending order, specified name
// returns all acitivists with names less than specified name
func TestActivistRange_NameNoLimitDescOrder_returnsSubsetOfActivists(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activistsToInsert := []string{"A", "B", "C", "D", "E", "F"}
	insertTestActivists(t, db, activistsToInsert)
	activistOptions := ActivistRangeOptionsJSON{
		Name:  "F",
		Order: DescOrder,
	}
	fetchedActivists, err := GetActivistRangeJSON(db, activistOptions)

	require.NoError(t, err)
	require.Equal(t, 5, len(fetchedActivists))

	reverseStringSlice(activistsToInsert)
	activistsToInsert = activistsToInsert[1:]
	for idx, a := range fetchedActivists {
		require.Equal(t, activistsToInsert[idx], a.Name)
	}

	// If specified name is last, then result is nil
	activistOptions.Name = "A"
	fetchedActivists, err = GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Nil(t, fetchedActivists)

}

// Limit < 0 behaves as if no limit was specified
func TestActivistRange_nonPositiveLimit_behavesAsNoLimit(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activistsToInsert := []string{"A", "B", "C", "D"}
	insertTestActivists(t, db, activistsToInsert)
	activistOptions := ActivistRangeOptionsJSON{
		Order: AscOrder,
		Limit: -42,
	}
	fetchedActivists, err := GetActivistRangeJSON(db, activistOptions)

	require.NoError(t, err)
	require.Equal(t, len(activistsToInsert), len(fetchedActivists))
}

// Specifying limit restricts number of returned entries
func TestActivistRange_NameAndLimitAscOrder_returnsSubsetOfActivists(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activistsToInsert := []string{"A", "B", "C", "D", "E", "F"}
	insertTestActivists(t, db, activistsToInsert)
	activistOptions := ActivistRangeOptionsJSON{
		Order: AscOrder,
		Limit: 20,
	}
	// Should get all activists back since Limit > Number of activists
	fetchedActivists, err := GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Equal(t, len(activistsToInsert), len(fetchedActivists))

	activistOptions.Limit = 2
	fetchedActivists, err = GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Equal(t, 2, len(fetchedActivists))

	for idx, a := range fetchedActivists {
		require.Equal(t, activistsToInsert[idx], a.Name)
	}

	activistOptions.Name = "F"
	fetchedActivists, err = GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Nil(t, fetchedActivists)
}

// Specifying limit restricts number of returned entries
func TestActivistRange_NameAndLimitDescOrder_returnsSubsetofActivists(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activistsToInsert := []string{"A", "B", "C", "D", "E", "F"}
	insertTestActivists(t, db, activistsToInsert)
	activistOptions := ActivistRangeOptionsJSON{
		Order: DescOrder,
		Limit: 20,
	}
	// Should get all activists back since 20 > Number of activists
	fetchedActivists, err := GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Equal(t, len(activistsToInsert), len(fetchedActivists))

	activistOptions.Limit = 2
	fetchedActivists, err = GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Equal(t, 2, len(fetchedActivists))

	reverseStringSlice(activistsToInsert)
	for idx, a := range fetchedActivists {
		require.Equal(t, activistsToInsert[idx], a.Name)
	}

	activistOptions.Name = "A"
	fetchedActivists, err = GetActivistRangeJSON(db, activistOptions)
	require.NoError(t, err)
	require.Nil(t, fetchedActivists)
}

func assertActivistJSONSliceContainsNames(t *testing.T, activists []ActivistJSON, names []string) {
	activistNames := map[string]struct{}{}
	for _, a := range activists {
		activistNames[a.Name] = struct{}{}
	}
	namesMap := map[string]struct{}{}
	for _, n := range names {
		namesMap[n] = struct{}{}
	}
	require.Equalf(t, activistNames, namesMap,
		"expected names to exist in activist map:\nactivists: %v\nnames: %v",
		activists, names)
}

func assertActivistJSONSliceContainsOrderedNames(t *testing.T, activists []ActivistJSON, names []string) {
	activistNames := []string{}
	for _, a := range activists {
		activistNames = append(activistNames, a.Name)
	}
	require.Equalf(t, activistNames, names,
		"expected names to equal activists names\nactivists: %v\nnames: %v",
		activists, names)
}

func insertTestActivists(t *testing.T, db *sqlx.DB, names []string) []Activist {
	var activists []Activist = make([]Activist, len(names))
	for idx, name := range names {
		activist, err := GetOrCreateActivist(db, name)
		require.NoError(t, err)
		activists[idx] = activist
	}
	return activists
}

func reverseStringSlice(s []string) {
	for left, right := 0, len(s)-1; left < right; left, right = left+1, right-1 {
		s[left], s[right] = s[right], s[left]
	}
}
