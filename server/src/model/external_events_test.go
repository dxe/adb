package model

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// A start time before the default set by `makeExternalEvent`.
var beforeDefaultStartTime = time.Date(1999, time.January, 1, 0, 0, 0, 0, time.UTC)

const queryTimeLayout string = "2006-01-02T15:04"

func TestInsertFacebookEvent(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	event := makeExternalEvent("1",
		WithName("Test Event 1"),
		WithIsCanceled(false),
	)

	UpsertExternalEvents(t, db, event)

	var events []ExternalEvent
	require.NoError(t,
		db.Select(&events, "select id, page_id, name from fb_events where name = 'Test Event 1'"))

	require.Equal(t, len(events), 1)
}

func TestGetFacebookEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	const pageOther int = 456456456456

	sfBayEvent := makeExternalEvent("1",
		WithPageID(SFBayPageID),
		WithName("Test Event 1"),
		WithStartTime(time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC)),
		WithEndTime(time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC)),
	)

	firstOtherEvent := makeExternalEvent("2",
		WithPageID(pageOther),
		WithName("Test Event 2"),
		WithStartTime(time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC)),
		WithEndTime(time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC)),
	)

	secondOtherEvent := makeExternalEvent("3",
		WithPageID(pageOther),
		WithName("Test Event 3"),
		WithStartTime(time.Date(2020, 2, 1, 11, 0, 0, 0, time.UTC)),
		WithEndTime(time.Date(2020, 2, 1, 13, 0, 0, 0, time.UTC)),
	)

	cancelledEvent := makeExternalEvent("4",
		WithPageID(pageOther),
		WithName("Test Event 4"),
		WithStartTime(time.Date(2020, 2, 1, 11, 0, 0, 0, time.UTC)),
		WithEndTime(time.Date(2020, 2, 1, 13, 0, 0, 0, time.UTC)),
		WithIsCanceled(true),
	)

	onlineEvent := makeExternalEvent("5",
		WithPageID(SFBayPageID),
		WithName("Test Event 5"),
		WithStartTime(time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC)),
		WithEndTime(time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC)),
		WithLocationName("Online"),
	)

	UpsertExternalEvents(t, db,
		sfBayEvent,
		firstOtherEvent,
		secondOtherEvent,
		cancelledEvent,
		onlineEvent,
	)

	// get events for specific chapter, excluding cancelled events
	queryStartTime := ParseTime(t, "2019-12-01T00:00")
	queryEndTime := ParseTime(t, "2020-03-01T00:00")
	events1, err1 := GetExternalEvents(db, 456456456456, queryStartTime, queryEndTime)
	require.NoError(t, err1)
	require.Equal(t, 2, len(events1))
	require.Equal(t, 456456456456, events1[0].PageID)

	// get events filtered by date for specific chapter
	queryStartTime = ParseTime(t, "2019-12-01T00:00")
	queryEndTime = ParseTime(t, "2020-01-15T00:00")
	events2, err2 := GetExternalEvents(db, 456456456456, queryStartTime, queryEndTime)
	require.NoError(t, err2)
	require.Equal(t, 1, len(events2))
	require.Equal(t, 456456456456, events2[0].PageID)

	// get online events
	queryStartTime = ParseTime(t, "2019-12-01T00:00")
	queryEndTime = ParseTime(t, "2020-01-15T00:00")
	events3, err3 := GetExternalOnlineEvents(db, queryStartTime, queryEndTime)
	require.NoError(t, err3)
	require.Equal(t, 1, len(events3))
	require.Equal(t, "5", events3[0].ID)
}

func TestGetBayAreaFacebookEvents(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	const pageOther int = 456456456456

	event1 := makeExternalEvent("1", WithPageID(SFBayPageID))
	event2 := makeExternalEvent("2", WithPageID(NorthBayPageID))
	event3 := makeExternalEvent("3", WithPageID(pageOther))

	UpsertExternalEvents(t, db, event1, event2, event3)

	eventsSFBay, _, err1 := GetExternalEventsWithFallback(db, SFBayPageID, beforeDefaultStartTime, time.Time{})
	require.NoError(t, err1)
	require.Equal(t, 2, len(eventsSFBay))

	eventsNorthBay, _, err2 := GetExternalEventsWithFallback(db, NorthBayPageID, beforeDefaultStartTime, time.Time{})
	require.NoError(t, err2)
	require.Equal(t, 2, len(eventsNorthBay))
}

func ParseTime(t *testing.T, timeStr string) time.Time {
	parsedTime, err := time.Parse(queryTimeLayout, timeStr)
	require.NoError(t, err)
	return parsedTime
}

// ExternalEventOption is a function that sets a property on an ExternalEvent
type ExternalEventOption func(*ExternalEvent)

// makeExternalEvent creates an ExternalEvent with the given options and sets reasonable defaults for any options not specified
func makeExternalEvent(id string, options ...ExternalEventOption) ExternalEvent {
	event := ExternalEvent{
		ID:              id,
		PageID:          1234567890,
		Name:            "Event",
		Description:     "Description",
		StartTime:       time.Date(2020, 1, 1, 11, 0, 0, 0, time.UTC),
		EndTime:         time.Date(2020, 1, 1, 13, 0, 0, 0, time.UTC),
		LocationName:    "Location",
		LocationCity:    "City",
		LocationCountry: "Country",
		LocationState:   "State",
		LocationAddress: "123 Default St",
		LocationZip:     "00000",
		Lat:             0.0,
		Lng:             0.0,
		Cover:           "http://default-cover-link",
		AttendingCount:  0,
		InterestedCount: 0,
		IsCanceled:      false,
		Featured:        false,
	}

	for _, option := range options {
		option(&event)
	}

	return event
}

// Option functions to set properties on ExternalEvent

func WithPageID(pageID int) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.PageID = pageID
	}
}

func WithName(name string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.Name = name
	}
}

func WithDescription(description string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.Description = description
	}
}

func WithStartTime(startTime time.Time) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.StartTime = startTime
	}
}

func WithEndTime(endTime time.Time) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.EndTime = endTime
	}
}

func WithLocationName(locationName string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.LocationName = locationName
	}
}

func WithLocationCity(locationCity string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.LocationCity = locationCity
	}
}

func WithLocationCountry(locationCountry string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.LocationCountry = locationCountry
	}
}

func WithLocationState(locationState string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.LocationState = locationState
	}
}

func WithLocationAddress(locationAddress string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.LocationAddress = locationAddress
	}
}

func WithLocationZip(locationZip string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.LocationZip = locationZip
	}
}

func WithLat(lat float64) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.Lat = lat
	}
}

func WithLng(lng float64) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.Lng = lng
	}
}

func WithCover(cover string) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.Cover = cover
	}
}

func WithAttendingCount(attendingCount int) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.AttendingCount = attendingCount
	}
}

func WithInterestedCount(interestedCount int) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.InterestedCount = interestedCount
	}
}

func WithIsCanceled(isCanceled bool) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.IsCanceled = isCanceled
	}
}

func WithFeatured(featured bool) ExternalEventOption {
	return func(e *ExternalEvent) {
		e.Featured = featured
	}
}

func UpsertExternalEvents(t *testing.T, db *sqlx.DB, events ...ExternalEvent) {
	for _, event := range events {
		err := UpsertExternalEvent(db, event)
		require.NoError(t, err)
	}
}
