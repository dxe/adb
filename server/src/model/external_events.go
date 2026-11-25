package model

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type ExternalEvent struct {
	ID              string    `db:"id"`
	PageID          int       `db:"page_id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	StartTime       time.Time `db:"start_time"`
	EndTime         time.Time `db:"end_time"`
	LocationName    string    `db:"location_name"`
	LocationCity    string    `db:"location_city"`
	LocationCountry string    `db:"location_country"`
	LocationState   string    `db:"location_state"`
	LocationAddress string    `db:"location_address"`
	LocationZip     string    `db:"location_zip"`
	Lat             float64   `db:"lat"`
	Lng             float64   `db:"lng"`
	Cover           string    `db:"cover"`
	AttendingCount  int       `db:"attending_count"`
	InterestedCount int       `db:"interested_count"`
	IsCanceled      bool      `db:"is_canceled"`
	LastUpdate      time.Time `db:"last_update"`
	EventbriteID    string    `db:"eventbrite_id"`
	EventbriteURL   string    `db:"eventbrite_url"`
	Featured        bool      `db:"featured"`
}

func GetExternalEventsWithFallback(db *sqlx.DB, pageID int, startTime time.Time, endTime time.Time) (events []ExternalEvent, localEventsFound bool, err error) {
	localEventsFound = false

	// run query to get local events
	if IsBayAreaPage(pageID) {
		// If one Bay Area page is chosen, combine events from all Bay Area pages
		events, err = GetExternalEventsForPages(db, BayAreaPages, startTime, endTime)
		if err != nil {
			return nil, false, err
		}
	} else {
		var err error
		events, err = GetExternalEvents(db, pageID, startTime, endTime)
		if err != nil {
			return nil, false, err
		}
	}

	// check if any local events were returned
	if len(events) > 0 {
		localEventsFound = true
	}

	if !localEventsFound {
		// get online SF Bay + ALOA events instead
		var err error
		events, err = GetExternalOnlineEvents(db, startTime, endTime)
		if err != nil {
			panic(err)
		}
	}

	return events, localEventsFound, nil
}

func GetExternalEvents(db *sqlx.DB, pageID int, startTime time.Time, endTime time.Time) ([]ExternalEvent, error) {
	return getExternalEvents(db, []int{pageID}, startTime, endTime, false)
}
func GetExternalEventsForPages(db *sqlx.DB, pageIDs []int, startTime time.Time, endTime time.Time) ([]ExternalEvent, error) {
	return getExternalEvents(db, pageIDs, startTime, endTime, false)
}
func GetExternalOnlineEvents(db *sqlx.DB, startTime time.Time, endTime time.Time) ([]ExternalEvent, error) {
	return getExternalEvents(db, []int{}, startTime, endTime, true)
}

func getExternalEvents(db *sqlx.DB, pageIDs []int, startTime time.Time, endTime time.Time, onlineOnly bool) ([]ExternalEvent, error) {
	query := `SELECT id, page_id, name, start_time, end_time, location_name,
		location_country, location_city, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update, eventbrite_id, eventbrite_url, description, featured FROM fb_events`

	query += " WHERE is_canceled = 0"

	if onlineOnly {
		// Show main chapter online events and ALC events
		query += fmt.Sprintf(" and ((page_id = %d and location_name = 'Online') or page_id = %d)", SFBayPageID, AlcPageID)
	} else {
		query += fmt.Sprintf(" and page_id in (%s)", intsToString(pageIDs))
	}

	if !startTime.IsZero() {
		query += fmt.Sprintf(" and if(end_time = '0001-01-01T00:00:00Z', start_time, end_time) >= '%s'", startTime.Format(time.RFC3339))
	}
	if !endTime.IsZero() {
		query += fmt.Sprintf(" and start_time < '%s'", endTime.Format(time.RFC3339))
	}

	query += " ORDER BY start_time"

	var events []ExternalEvent
	err := db.Select(&events, query)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select events")
	}

	// If there are multiple pages, they could be co-hosting the same events.
	// To keep the SQL query simple, deduplicate the events here.
	// (Deduplicating in the SQL query would require ANY(page_id) and grouping
	// by all other columns, or not retrieving any page_id.)
	if len(pageIDs) > 1 {
		events = deduplicateEvents(events)
	}

	return events, nil
}

func deduplicateEvents(events []ExternalEvent) []ExternalEvent {
	seen := make(map[string]bool)
	uniqueEvents := make([]ExternalEvent, 0, len(events))
	for _, event := range events {
		if !seen[event.ID] {
			seen[event.ID] = true
			uniqueEvents = append(uniqueEvents, event)
		}
	}
	return uniqueEvents
}

func intsToString(ints []int) string {
	strs := make([]string, len(ints))
	for i, id := range ints {
		strs[i] = strconv.Itoa(id)
	}
	return strings.Join(strs, ",")
}

func UpsertExternalEvent(db *sqlx.DB, event ExternalEvent) (err error) {
	sqlTimeLayout := "2006-01-02T15:04:05"

	// insert into database
	// TODO: we should store eventbrite event info in a separate table so that we can just do "REPLACE INTO here" instead of handling "ON DUPLICATE KEY"
	// TODO: used NamedExec here to make it more maintainable
	_, err = db.Exec(`INSERT INTO fb_events (id, page_id, name, description, start_time, end_time,
		location_name, location_city, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update, featured) VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now(), ?)
		ON DUPLICATE KEY UPDATE name=VALUES(name), description=VALUES(description),
		start_time=VALUES(start_time), end_time=VALUES(end_time), 
		location_name=VALUES(location_name), location_city=VALUES(location_city), location_country=VALUES(location_country), 
		location_state=VALUES(location_state), location_address=VALUES(location_address), location_zip=VALUES(location_zip), 
		lat=VALUES(lat), lng=VALUES(lng), cover=VALUES(cover), attending_count=VALUES(attending_count),
		interested_count=VALUES(interested_count), is_canceled=VALUES(is_canceled), last_update=VALUES(last_update)`,
		event.ID, event.PageID, event.Name, event.Description, event.StartTime.Format(sqlTimeLayout),
		event.EndTime.Format(sqlTimeLayout), event.LocationName, event.LocationCity,
		event.LocationCountry, event.LocationState, event.LocationAddress,
		event.LocationZip, event.Lat, event.Lng, event.Cover,
		event.AttendingCount, event.InterestedCount, event.IsCanceled, event.Featured)
	if err != nil {
		return errors.Wrap(err, "failed to insert event")
	}
	return nil
}

func AddEventbriteDetailsToEventByNameAndDate(db *sqlx.DB, event ExternalEvent) error {
	_, err := db.NamedExec(`UPDATE fb_events
		SET eventbrite_id = :eventbrite_id, eventbrite_url = :eventbrite_url
		WHERE name = :name and left(start_time, 10) = left(:start_time, 10)`, event)
	if err != nil {
		return errors.Wrap(err, "failed to update event")
	}
	return nil
}

func AddEventbriteDetailsToEventByID(db *sqlx.DB, event ExternalEvent) error {
	_, err := db.NamedExec(`UPDATE fb_events
		SET eventbrite_id = :eventbrite_id, eventbrite_url = :eventbrite_url
		WHERE id = :id`, event)
	if err != nil {
		return errors.Wrap(err, "failed to update event")
	}
	return nil
}

// ExternalEventRepository reads external (Facebook/Eventbrite) events by ADB
// chapter ID so callers don't need to know about Facebook page IDs.
type ExternalEventRepository interface {
	// ListUpcomingForChapter returns non-canceled events from the chapter's own
	// page that haven't ended as of `since`. It returns no events if the chapter
	// has no page.
	ListUpcomingForChapter(chapterID int, since time.Time) ([]ExternalEvent, error)

	// GetChapterID returns the ID of the chapter that owns the event's page, or
	// ErrNotFound.
	GetChapterID(eventID string) (int, error)

	// SetFeatured and Cancel only touch the event if it belongs to chapterID's
	// page; otherwise they return ErrNotFound. (A redundant check: callers are
	// expected to have authorized the user for chapterID already.)
	SetFeatured(chapterID int, eventID string, featured bool) error
	Cancel(chapterID int, eventID string) error
}

// ListExternalEventsForChapter returns upcoming events for a chapter, which
// the user must be an admin or organizer of.
func ListExternalEventsForChapter(repo ExternalEventRepository, user ADBUser, chapterID int, since time.Time) ([]ExternalEvent, error) {
	if !UserHasOrganizerAccess(user) {
		return nil, ValidationErrorf("lacking permission to view external events")
	}
	if err := CheckChapterAccess(user, chapterID); err != nil {
		return nil, err
	}
	return repo.ListUpcomingForChapter(chapterID, since)
}

func FeatureExternalEvent(repo ExternalEventRepository, user ADBUser, eventID string, featured bool) error {
	return modifyExternalEvent(repo, user, eventID, func(chapterID int) error {
		return repo.SetFeatured(chapterID, eventID, featured)
	})
}

func CancelExternalEvent(repo ExternalEventRepository, user ADBUser, eventID string) error {
	return modifyExternalEvent(repo, user, eventID, func(chapterID int) error {
		return repo.Cancel(chapterID, eventID)
	})
}

// modifyExternalEvent authorizes the user (admin, or organizer of the chapter
// that owns the event) and then runs the write.
func modifyExternalEvent(repo ExternalEventRepository, user ADBUser, eventID string, write func(chapterID int) error) error {
	if !UserHasOrganizerAccess(user) {
		return ValidationErrorf("lacking permission to modify external events")
	}
	chapterID, err := repo.GetChapterID(eventID)
	if err != nil {
		return err
	}
	if err := CheckChapterAccess(user, chapterID); err != nil {
		return err
	}
	return write(chapterID)
}
