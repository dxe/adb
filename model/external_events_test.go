package model

import (
	"testing"

	"github.com/dxe/adb/event_sync"
	"github.com/stretchr/testify/require"
)

func TestInsertFacebookEvent(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	page := ChapterWithToken{
		ID: 123123123123,
	}
	event := event_sync.FacebookEvent{
		ID:              "1111111111",
		Name:            "Test Event 1",
		Description:     "This is a test event.",
		StartTime:       "2020-01-01T11:00:00-0700",
		EndTime:         "2020-01-01T13:00:00-0700",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		IsOnline:        false,
		Place: event_sync.FacebookPlace{
			Name: "Berkeley Animal Rights Center",
			Location: event_sync.FacebookLocation{
				City:    "Berkeley",
				State:   "CA",
				Country: "United States",
				Street:  "123 Channing Way",
				Zip:     "94703",
				Lat:     1.000,
				Lng:     1.000,
			},
		},
		Cover: event_sync.FacebookCover{
			Source: "http://not-a-real-link",
		},
	}

	err := InsertExternalEvent(db, event, page)
	require.NoError(t, err)

	var events []ExternalEvent
	require.NoError(t,
		db.Select(&events, "select id, page_id, name from fb_events where name = 'Test Event 1'"))

	require.Equal(t, len(events), 1)

}

func TestGetFacebookEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	page1 := ChapterWithToken{
		ID: 1377014279263790,
	}
	page2 := ChapterWithToken{
		ID: 456456456456,
	}
	event1 := event_sync.FacebookEvent{
		ID:              "1111111111",
		Name:            "Test Event 1",
		Description:     "This is a test event in Berkeley.",
		StartTime:       "2020-01-01T11:00:00-0700",
		EndTime:         "2020-01-01T13:00:00-0700",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		IsOnline:        false,
		Place: event_sync.FacebookPlace{
			Name: "Berkeley Animal Rights Center",
			Location: event_sync.FacebookLocation{
				City:    "Berkeley",
				State:   "CA",
				Country: "United States",
				Street:  "123 Channing Way",
				Zip:     "94703",
				Lat:     1.000,
				Lng:     1.000,
			},
		},
		Cover: event_sync.FacebookCover{
			Source: "http://not-a-real-link",
		},
	}
	event2 := event_sync.FacebookEvent{
		ID:              "2222222222",
		Name:            "Test Event 2",
		Description:     "This is a test event in NY.",
		StartTime:       "2020-01-01T11:00:00-0700",
		EndTime:         "2020-01-01T13:00:00-0700",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		IsOnline:        false,
		Place: event_sync.FacebookPlace{
			Name: "Not Berkeley Animal Rights Center",
			Location: event_sync.FacebookLocation{
				City:    "New York",
				State:   "NY",
				Country: "United States",
				Street:  "123 Main St",
				Zip:     "10258",
				Lat:     1.000,
				Lng:     1.000,
			},
		},
		Cover: event_sync.FacebookCover{
			Source: "http://not-a-real-link",
		},
	}
	event3 := event_sync.FacebookEvent{
		ID:              "3333333333",
		Name:            "Test Event 3",
		Description:     "This is a test event in NY at a later date.",
		StartTime:       "2020-02-01T11:00:00-0700",
		EndTime:         "2020-02-01T13:00:00-0700",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		IsOnline:        false,
		Place: event_sync.FacebookPlace{
			Name: "Not Berkeley Animal Rights Center",
			Location: event_sync.FacebookLocation{
				City:    "New York",
				State:   "NY",
				Country: "United States",
				Street:  "123 Main St",
				Zip:     "10258",
				Lat:     1.000,
				Lng:     1.000,
			},
		},
		Cover: event_sync.FacebookCover{
			Source: "http://not-a-real-link",
		},
	}
	event4 := event_sync.FacebookEvent{
		ID:              "444444444",
		Name:            "Test Event 4",
		Description:     "This is a test event that was cancelled.",
		StartTime:       "2020-02-01T11:00:00-0700",
		EndTime:         "2020-02-01T13:00:00-0700",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      true,
		IsOnline:        false,
		Place: event_sync.FacebookPlace{
			Name: "Not Berkeley Animal Rights Center",
			Location: event_sync.FacebookLocation{
				City:    "New York",
				State:   "NY",
				Country: "United States",
				Street:  "123 Main St",
				Zip:     "10258",
				Lat:     1.000,
				Lng:     1.000,
			},
		},
		Cover: event_sync.FacebookCover{
			Source: "http://not-a-real-link",
		},
	}
	event5 := event_sync.FacebookEvent{
		ID:              "5555555555",
		Name:            "Test Event 5",
		Description:     "This is a online event in Berkeley.",
		StartTime:       "2020-01-01T11:00:00-0700",
		EndTime:         "2020-01-01T13:00:00-0700",
		AttendingCount:  25,
		InterestedCount: 50,
		IsCanceled:      false,
		IsOnline:        true,
		Place: event_sync.FacebookPlace{
			Name: "Berkeley Animal Rights Center",
			Location: event_sync.FacebookLocation{
				City:    "Berkeley",
				State:   "CA",
				Country: "United States",
				Street:  "123 Channing Way",
				Zip:     "94703",
				Lat:     1.000,
				Lng:     1.000,
			},
		},
		Cover: event_sync.FacebookCover{
			Source: "http://not-a-real-link",
		},
	}

	err := InsertExternalEvent(db, event1, page1)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event2, page2)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event3, page2)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event4, page2)
	require.NoError(t, err)

	err = InsertExternalEvent(db, event5, page1)
	require.NoError(t, err)

	var events []ExternalEvent

	// get events for specific chapter, excluding cancelled events
	events, err = GetExternalEvents(db, 456456456456, "2019-12-01T00:00", "2020-03-01T00:00", false)
	require.Equal(t, len(events), 2)
	require.Equal(t, events[0].PageID, 456456456456)

	// get events filtered by date for specific chapter
	events, err = GetExternalEvents(db, 456456456456, "2019-12-01T00:00", "2020-01-15T00:00", false)
	require.Equal(t, len(events), 1)
	require.Equal(t, events[0].PageID, 456456456456)

	// get online events
	events, err = GetExternalEvents(db, 0, "2019-12-01T00:00", "2020-01-15T00:00", true)
	require.Equal(t, len(events), 1)
	require.Equal(t, events[0].ID, 5555555555)
}
