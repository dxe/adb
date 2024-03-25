package model

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestInsertFacebookEvent(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	event := ExternalEvent{
		ID:              "1111111111",
		PageID:          123123123123,
		Name:            "Test Event 1",
		Description:     "This is a test event.",
		StartTime:       time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		LocationName:    "Berkeley Animal Rights Center",
		LocationCity:    "Berkeley",
		LocationCountry: "United States",
		LocationState:   "CA",
		LocationAddress: "123 Channing Way",
		LocationZip:     "94703",
		Lat:             1.000,
		Lng:             1.000,
		Cover:           "http://not-a-real-link",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		Featured:        false,
	}

	err := InsertExternalEvent(db, event)
	require.NoError(t, err)

	var events []ExternalEvent
	require.NoError(t,
		db.Select(&events, "select id, page_id, name from fb_events where name = 'Test Event 1'"))

	require.Equal(t, len(events), 1)

}

func TestGetFacebookEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	const pageBerkeley int = 1377014279263790
	const pageOther int = 456456456456

	event1 := ExternalEvent{
		ID:              "1111111111",
		PageID:          pageBerkeley,
		Name:            "Test Event 1",
		Description:     "This is a test event in Berkeley.",
		StartTime:       time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		LocationName:    "Berkeley Animal Rights Center",
		LocationCity:    "Berkeley",
		LocationCountry: "United States",
		LocationState:   "CA",
		LocationAddress: "123 Channing Way",
		LocationZip:     "94703",
		Lat:             1.000,
		Lng:             1.000,
		Cover:           "http://not-a-real-link",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		Featured:        false,
	}

	event2 := ExternalEvent{
		ID:              "2222222222",
		PageID:          pageOther,
		Name:            "Test Event 2",
		Description:     "This is a test event in NY.",
		StartTime:       time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		LocationName:    "Not Berkeley Animal Rights Center",
		LocationCity:    "New York",
		LocationCountry: "United States",
		LocationState:   "NY",
		LocationAddress: "123 Main St",
		LocationZip:     "10258",
		Lat:             1.000,
		Lng:             1.000,
		Cover:           "http://not-a-real-link",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		Featured:        false,
	}

	event3 := ExternalEvent{
		ID:              "3333333333",
		PageID:          pageOther,
		Name:            "Test Event 3",
		Description:     "This is a test event in NY at a later date.",
		StartTime:       time.Date(2020, 2, 1, 11, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2020, 2, 1, 13, 0, 0, 0, time.UTC),
		LocationName:    "Not Berkeley Animal Rights Center",
		LocationCity:    "New York",
		LocationCountry: "United States",
		LocationState:   "NY",
		LocationAddress: "123 Main St",
		LocationZip:     "10258",
		Lat:             1.000,
		Lng:             1.000,
		Cover:           "http://not-a-real-link",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		Featured:        false,
	}

	event4 := ExternalEvent{
		ID:              "4444444444",
		PageID:          pageOther,
		Name:            "Test Event 4",
		Description:     "This is a test event that was cancelled.",
		StartTime:       time.Date(2020, 2, 1, 11, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2020, 2, 1, 13, 0, 0, 0, time.UTC),
		LocationName:    "Not Berkeley Animal Rights Center",
		LocationCity:    "New York",
		LocationCountry: "United States",
		LocationState:   "NY",
		LocationAddress: "123 Main St",
		LocationZip:     "10258",
		Lat:             1.000,
		Lng:             1.000,
		Cover:           "http://not-a-real-link",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      true,
		Featured:        false,
	}

	event5 := ExternalEvent{
		ID:              "5555555555",
		PageID:          pageBerkeley,
		Name:            "Test Event 5",
		Description:     "This is an online event hosted by Berkeley.",
		StartTime:       time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		LocationName:    "Online",
		LocationCity:    "",
		LocationCountry: "",
		LocationState:   "",
		LocationAddress: "",
		LocationZip:     "",
		Lat:             1.000,
		Lng:             1.000,
		Cover:           "http://not-a-real-link",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		Featured:        false,
	}

	err := InsertExternalEvent(db, event1)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event2)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event3)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event4)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event5)
	require.NoError(t, err)

	var events []ExternalEvent

	const queryTimeLayout string = "2006-01-02T15:04"

	// get events for specific chapter, excluding cancelled events
	queryStartTime, err := time.Parse(queryTimeLayout, "2019-12-01T00:00")
	require.NoError(t, err)
	queryEndTime, err := time.Parse(queryTimeLayout, "2020-03-01T00:00")
	require.NoError(t, err)
	events, err = GetExternalEvents(db, 456456456456, queryStartTime, queryEndTime, false)
	require.Equal(t, len(events), 2)
	require.Equal(t, events[0].PageID, 456456456456)

	// get events filtered by date for specific chapter
	queryStartTime, err = time.Parse(queryTimeLayout, "2019-12-01T00:00")
	require.NoError(t, err)
	queryEndTime, err = time.Parse(queryTimeLayout, "2020-01-15T00:00")
	require.NoError(t, err)
	events, err = GetExternalEvents(db, 456456456456, queryStartTime, queryEndTime, false)
	require.Equal(t, len(events), 1)
	require.Equal(t, events[0].PageID, 456456456456)

	// get online events
	queryStartTime, err = time.Parse(queryTimeLayout, "2019-12-01T00:00")
	require.NoError(t, err)
	queryEndTime, err = time.Parse(queryTimeLayout, "2020-01-15T00:00")
	require.NoError(t, err)
	events, err = GetExternalEvents(db, 0, queryStartTime, queryEndTime, true)
	require.Equal(t, len(events), 1)
	require.Equal(t, events[0].ID, "5555555555")
}
