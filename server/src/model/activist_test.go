package model

import (
	"database/sql"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
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

	_, err := GetOrCreateActivist(db, "Activist One", SFBayChapterIdDevTest)
	if err != nil {
		t.Fatal(err)
	}

	_, err = GetOrCreateActivist(db, "Activist Two", SFBayChapterIdDevTest)
	if err != nil {
		t.Fatal(err)
	}

	gotNames := GetAutocompleteNames(db, SFBayChapterIdDevTest)
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

	a1, err := GetOrCreateActivist(db, "Test Activist", SFBayChapterIdDevTest)
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

	require.Equal(t, d.FirstEvent.Valid, true)
	d.FirstEvent.Time.Equal(d1)
	require.Equal(t, d.LastEvent.Valid, true)
	d.LastEvent.Time.Equal(d3)
	require.Equal(t, d.TotalEvents, 4)
}

func TestGetActivistEventData_noEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "Test Activist", SFBayChapterIdDevTest)
	require.NoError(t, err)

	d, err := a1.GetActivistEventData(db)
	require.NoError(t, err)

	require.Equal(t, d, ActivistEventData{
		FirstEvent:  mysql.NullTime{},
		LastEvent:   mysql.NullTime{},
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

	a1, err := GetOrCreateActivist(db, "A", SFBayChapterIdDevTest)
	require.NoError(t, err)

	a2, err := GetOrCreateActivist(db, "B", SFBayChapterIdDevTest)
	require.NoError(t, err)

	a3, err := GetOrCreateActivist(db, "C", SFBayChapterIdDevTest)
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

	a1, err := GetOrCreateActivist(db, "A", SFBayChapterIdDevTest)
	require.NoError(t, err)

	a2, err := GetOrCreateActivist(db, "B", SFBayChapterIdDevTest)
	require.NoError(t, err)

	a3, err := GetOrCreateActivist(db, "C", SFBayChapterIdDevTest)
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

func TestGetActivistsJSON_FirstAndLastEvent(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	a1, err := GetOrCreateActivist(db, "A", SFBayChapterIdDevTest)
	require.NoError(t, err)

	d1, err := time.Parse("2006-01-02", "2017-04-15")
	require.NoError(t, err)
	d2, err := time.Parse("2006-01-02", "2017-04-16")
	require.NoError(t, err)
	d3, err := time.Parse("2006-01-02", "2017-04-17")
	require.NoError(t, err)

	insertEvents := []Event{{
		EventName:      "event one",
		EventDate:      d2,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
	}, {
		EventName:      "yo yo yo",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []Activist{},
	}, {
		EventName:      "heyo",
		EventDate:      d3,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
	}, {
		EventName:      "hello",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []Activist{},
	}, {
		EventName:      "hi there",
		EventDate:      d1,
		EventType:      "Working Group",
		AddedAttendees: []Activist{a1},
	}}
	mustInsertAllEvents(t, db, insertEvents)

	activists, err := GetActivistsJSON(db, GetActivistOptions{})
	require.NoError(t, err)

	gotActivist := activists[0]
	require.Equal(t, gotActivist.FirstEvent, "2017-04-15")
	require.Equal(t, gotActivist.LastEvent, "2017-04-17")

	require.Equal(t, gotActivist.FirstEventName, "2017-04-15 hi there")
	require.Equal(t, gotActivist.LastEventName, "2017-04-17 heyo")
}

func TestGetActivistsExtra(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	insertTestActivists(t, db, []string{"Alex Taylor"})

	activists, err := GetActivistsExtra(db, GetActivistOptions{})
	require.NoError(t, err)
	require.Equal(t, len(activists), 1)
	assert.Equal(t, "Alex Taylor", activists[0].Name)
}

func TestInsertActivist(t *testing.T) {
	t.Run("Minimum", func(t *testing.T) {
		db := newTestDB()
		defer db.Close()

		var activist ActivistExtra
		// Only Chapter and Name are set.
		// Time fields all have 0 values, which should not underflow SQL timestamp/date fields.
		activist.ChapterID = SFBayChapterId
		activist.Name = "Alex Taylor"
		id, errCreate := CreateActivist(db, activist)
		require.NoError(t, errCreate)
		assert.NotZero(t, id)

		inserted, errGet := GetActivistExtra(db, id)
		require.NoError(t, errGet)
		assert.Equal(t, "Alex Taylor", inserted.Name)
		assert.Equal(t, SFBayChapterId, inserted.ChapterID)
	})

	t.Run("Basic", func(t *testing.T) {
		db := newTestDB()
		defer db.Close()

		activist := NewActivistBuilder().
			WithChapterID(SFBayChapterId).
			WithName("Alexander Taylor").
			WithEmail("ataylor@example.org").
			WithPhone("510-555-5555").
			WithAddress("5 Animal Rights Way", "Berkeley", "CA").
			WithLocation(sql.NullString{String: "94103", Valid: true}).
			WithCoords(1, -1).
			Build()
		id, errCreate := CreateActivist(db, *activist)
		require.NoError(t, errCreate)
		assert.NotZero(t, id)

		inserted, errGet := GetActivistExtra(db, id)
		require.NoError(t, errGet)
		assert.Equal(t, "Alexander Taylor", inserted.Name)
		assert.Equal(t, SFBayChapterId, inserted.ChapterID)
		assert.Equal(t, "ataylor@example.org", inserted.Email)
		assert.Equal(t, "510-555-5555", inserted.Phone)
		assert.Equal(t, "5 Animal Rights Way", inserted.StreetAddress)
		assert.Equal(t, "Berkeley", inserted.City)
		assert.Equal(t, "CA", inserted.State)
		assert.Equal(t, sql.NullString{String: "94103", Valid: true}, inserted.Location)
		assert.Equal(t, Coords{Lat: 1, Lng: -1}, inserted.Coords)
	})
}

func TestUpdateActivist(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	activist := NewActivistBuilder().
		WithChapterID(SFBayChapterId).
		WithName("Alexander Taylor").
		WithEmail("ataylor@example.org").
		WithPhone("510-555-5555").
		WithAddress("5 Animal Rights Way", "Berkeley", "CA").
		WithLocation(sql.NullString{String: "94103", Valid: true}).
		WithCoords(1, -1).
		Build()
	id, errCreate := CreateActivist(db, *activist)
	require.NoError(t, errCreate)

	activist.ID = id
	activist.Name = "Alex Taylor"
	activist.Email = "ataylor2@example.org"
	activist.Phone = "510-111-1234"
	activist.ActivistAddress = ActivistAddress{
		"6 Animal Rights Way", "New York", "NY",
	}
	activist.Location = sql.NullString{String: "90001", Valid: true}
	activist.Coords = Coords{1, 2}

	UpdateActivistData(db, *activist, DevTestUserEmail)

	updatedActivist, err := GetActivistExtra(db, id)
	require.NoError(t, err)

	assert.Equal(t, SFBayChapterId, updatedActivist.ChapterID)
	assert.Equal(t, "Alex Taylor", updatedActivist.Name)
	assert.Equal(t, "ataylor2@example.org", updatedActivist.Email)
	assert.Equal(t, "510-111-1234", updatedActivist.Phone)
	assert.Equal(t, ActivistAddress{"6 Animal Rights Way", "New York", "NY"}, updatedActivist.ActivistAddress)
	assert.Equal(t, sql.NullString{String: "90001", Valid: true}, updatedActivist.Location)
	assert.Equal(t, Coords{Lat: 1, Lng: 2}, updatedActivist.Coords)
}

func TestHideActivist(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	// Test that deleting activists works
	a1, err := GetOrCreateActivist(db, "Test Activist", SFBayChapterIdDevTest)
	require.NoError(t, err)

	a2, err := GetOrCreateActivist(db, "Another Test Activist", SFBayChapterIdDevTest)
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
	names := GetAutocompleteNames(db, SFBayChapterIdDevTest)
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
	assertStringsSliceUnorderedEquals(t, event.Attendees, []string{a1.Name, a2.Name})

	attendanceNames, err := GetEventAttendance(db, eventID)
	require.NoError(t, err)
	assertStringsSliceUnorderedEquals(t, attendanceNames, []string{a1.Name, a2.Name})
}

func mustParseTime(t *testing.T, s string) time.Time {
	time, err := time.Parse("2006-01-02", s)
	require.NoError(t, err)
	return time
}

func TestMergeActivist(t *testing.T) {
	t.Run("ContactInfo", func(t *testing.T) {
		t.Run("MergesNewerValues", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().
				WithEmail("berkeley@example.org").
				WithPhone("510-555-5555").
				Build()
			a1.EmailUpdated = mustParseTime(t, "2025-01-02")
			a1.PhoneUpdated = mustParseTime(t, "2025-01-02")
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().
				WithEmail("old@example.org").
				WithPhone("510-555-0000").
				Build()
			a2.EmailUpdated = mustParseTime(t, "2025-01-01")
			a2.PhoneUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, "berkeley@example.org", a2.Email)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.EmailUpdated)
			assert.Equal(t, "510-555-5555", a2.Phone)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.PhoneUpdated)
		})

		t.Run("DoesNotMergeNewerButEmptyValues", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().
				WithEmail("").
				WithPhone("").
				Build()
			a1.EmailUpdated = mustParseTime(t, "2025-01-02") // Newer
			a1.PhoneUpdated = mustParseTime(t, "2025-01-02") // Newer
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().
				WithEmail("old@example.org").
				WithPhone("510-555-0000").
				Build()
			a2.EmailUpdated = mustParseTime(t, "2025-01-01")
			a2.PhoneUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, "old@example.org", a2.Email)
			assert.Equal(t, mustParseTime(t, "2025-01-01"), a2.EmailUpdated)
			assert.Equal(t, "510-555-0000", a2.Phone)
			assert.Equal(t, mustParseTime(t, "2025-01-01"), a2.PhoneUpdated)
		})

		t.Run("DoesNotMergeOlderValues", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().
				WithEmail("old@example.org").
				WithPhone("510-555-0000").
				Build()
			a1.EmailUpdated = mustParseTime(t, "2025-01-01")
			a1.PhoneUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().
				WithEmail("berkeley@example.org").
				WithPhone("510-555-5555").
				Build()
			a2.EmailUpdated = mustParseTime(t, "2025-01-02")
			a2.PhoneUpdated = mustParseTime(t, "2025-01-02")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, "berkeley@example.org", a2.Email)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.EmailUpdated)
			assert.Equal(t, "510-555-5555", a2.Phone)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.PhoneUpdated)
		})
	})

	t.Run("Address", func(t *testing.T) {
		t.Run("MergesNewerValues", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().
				WithAddress("100 Berkeley Way", "Berkeley", "CA").
				WithCoords(1, 2).
				Build()
			a1.AddressUpdated = mustParseTime(t, "2025-01-02")
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().
				WithAddress("200 Berkeley Way", "Berkeley", "CA").
				WithCoords(3, 4).
				Build()
			a2.AddressUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.AddressUpdated)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.LocationUpdated)
			assert.Equal(t, "100 Berkeley Way", a2.StreetAddress)
			assert.Equal(t, Coords{1, 2}, a2.Coords)
		})

		t.Run("DoesNotMergeNewerButEmptyValues", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().
				WithAddress("", "", "").
				WithCoords(0, 0).
				Build()
			a1.AddressUpdated = mustParseTime(t, "2025-01-02") // Newer
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().
				WithAddress("200 Berkeley Way", "Berkeley", "CA").
				WithCoords(3, 4).
				Build()
			a2.AddressUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, mustParseTime(t, "2025-01-01"), a2.AddressUpdated)
			assert.Equal(t, "200 Berkeley Way", a2.StreetAddress)
			assert.Equal(t, Coords{3, 4}, a2.Coords)
		})

		t.Run("DoesNotMergeOlderValues", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().
				WithAddress("100 Berkeley Way", "Berkeley", "CA").
				WithCoords(1, 2).
				Build()
			a1.AddressUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().
				WithAddress("200 Berkeley Way", "Berkeley", "CA").
				WithCoords(3, 4).
				Build()
			a2.AddressUpdated = mustParseTime(t, "2025-01-02")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.AddressUpdated)
			assert.Equal(t, "200 Berkeley Way", a2.StreetAddress)
			assert.Equal(t, Coords{3, 4}, a2.Coords)
		})

		t.Run("MergesAddressWhenCityMatches", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().WithAddress("100 Berkeley Way", "Berkeley", "CA").Build()
			a1.AddressUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().WithAddress("", "Berkeley", "CA").Build()
			a2.AddressUpdated = mustParseTime(t, "2025-01-02")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.AddressUpdated)
			assert.Equal(t, "100 Berkeley Way", a2.StreetAddress)
		})

		t.Run("DoesNotMergeAddressWhenCityNotMatched", func(t *testing.T) {
			db := newTestDB()
			defer db.Close()

			a1 := NewActivistBuilder().WithAddress("100 Berkeley Way", "Berkeley", "CA").Build()
			a1.AddressUpdated = mustParseTime(t, "2025-01-01")
			MustInsertActivistWithTimestamps(t, db, a1)

			a2 := NewActivistBuilder().WithAddress("", "New York", "NY").Build()
			a2.AddressUpdated = mustParseTime(t, "2025-01-02")
			MustInsertActivistWithTimestamps(t, db, a2)

			require.NoError(t, MergeActivist(db, a1.ID, a2.ID))
			a2 = MustGetActivist(t, db, a2.ID)
			assert.Equal(t, mustParseTime(t, "2025-01-02"), a2.AddressUpdated)
			assert.Equal(t, "", a2.StreetAddress)
		})
	})

	t.Run("MergesEvents", func(t *testing.T) {
		db := newTestDB()
		defer db.Close()

		// Test that deleting activists works
		a1, err := GetOrCreateActivist(db, "Test Activist", SFBayChapterIdDevTest)
		require.NoError(t, err)

		a2, err := GetOrCreateActivist(db, "Another Test Activist", SFBayChapterIdDevTest)
		require.NoError(t, err)

		a3, err := GetOrCreateActivist(db, "A Third Test Activist", SFBayChapterIdDevTest)
		require.NoError(t, err)

		d1 := mustParseTime(t, "2017-04-15")
		d2 := mustParseTime(t, "2017-04-16")
		d3 := mustParseTime(t, "2017-04-17")

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
	})
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

func assertStringsSliceUnorderedEquals(t *testing.T, s0, s1 []string) {
	m0 := map[string]struct{}{}
	for _, s := range s0 {
		m0[s] = struct{}{}
	}
	m1 := map[string]struct{}{}
	for _, s := range s1 {
		m1[s] = struct{}{}
	}
	require.Equalf(t, m0, m1,
		"expected s0 and s1 to be equal:\ns0: %v\ns1: %v",
		s0, s1)
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
		activist, err := GetOrCreateActivist(db, name, SFBayChapterIdDevTest)
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
