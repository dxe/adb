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
		location_country, location_country, location_state, location_address, location_zip,
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

func FeatureExternalEvent(db *sqlx.DB, eventId string, featured bool) error {
	_, err := db.Exec(`UPDATE fb_events
		SET featured = ?
		WHERE id = ?`, featured, eventId)
	if err != nil {
		return errors.Wrap(err, "failed to update event (failed to feature event)")
	}
	return nil
}

func CancelExternalEvent(db *sqlx.DB, eventId string) error {
	_, err := db.Exec(`UPDATE fb_events
		SET is_canceled = 1
		WHERE id = ?`, eventId)
	if err != nil {
		return errors.Wrap(err, "failed to update event")
	}
	return nil
}
