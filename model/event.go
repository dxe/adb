package model

import (
	"database/sql/driver"
	"errors"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

/** Constant and Global Variable Definitions */

const EventDateLayout string = "2006-01-02"

var EventTypes map[string]bool = map[string]bool{
	"Working Group": true,
	"Community":     true,
	"Protest":       true,
	"Outreach":      true,
	"Key Event":     true,
	"Sanctuary":     true,
}

/** Type Definitions */

type EventType string

/* TODO Restructure this struct */
type EventJSON struct {
	EventID          int      `json:"event_id"`
	EventName        string   `json:"event_name"`
	EventDate        string   `json:"event_date"`
	EventType        string   `json:"event_type"`
	Attendees        []string `json:"attendees"`         // For displaying all event attendees
	AddedAttendees   []string `json:"added_attendees"`   // Used for Updating Events
	DeletedAttendees []string `json:"deleted_attendees"` // Used for Updating Events
}

/* TODO Restructure this Struct */
type Event struct {
	ID               int       `db:"id"`
	EventName        string    `db:"name"`
	EventDate        time.Time `db:"date"`
	EventType        EventType `db:"event_type"`
	Attendees        []User    // For retrieving all event attendees
	AddedAttendees   []User    // Used for Updating Events
	DeletedAttendees []User    // Used for Updating Events
}

type GetEventOptions struct {
	EventID int
	// NOTE: don't pass user input to OrderBy, cause that could
	// cause a SQL injection.
	OrderBy        string
	DateFrom       string
	DateTo         string
	EventType      string
	EventNameQuery string
}

/** Functions and Methods */

func GetEventsJSON(db *sqlx.DB, options GetEventOptions) ([]EventJSON, error) {
	dbEvents, err := GetEvents(db, options)

	if err != nil {
		return nil, err
	}

	events := make([]EventJSON, 0, len(dbEvents))
	for _, event := range dbEvents {
		attendees := make([]string, 0, len(event.Attendees))
		for _, user := range event.Attendees {
			attendees = append(attendees, user.Name)
		}
		events = append(events, EventJSON{
			EventID:   event.ID,
			EventName: event.EventName,
			EventDate: event.EventDate.Format(EventDateLayout),
			EventType: string(event.EventType),
			Attendees: attendees,
		})
	}
	return events, nil
}

func GetEvents(db *sqlx.DB, options GetEventOptions) ([]Event, error) {
	return getEvents(db, options)
}

func GetEvent(db *sqlx.DB, options GetEventOptions) (Event, error) {
	if options.EventID == 0 {
		return Event{}, errors.New("EventID for GetEvent cannot be zero")
	}
	events, err := getEvents(db, options)
	if err != nil {
		return Event{}, nil
	} else if len(events) == 0 {
		return Event{}, errors.New("Could not find any events")
	} else if len(events) > 1 {
		return Event{}, errors.New("Found too many events")
	}
	return events[0], nil
}

func getEvents(db *sqlx.DB, options GetEventOptions) ([]Event, error) {
	var queryArgs []interface{}
	query := `SELECT id, name, date, event_type FROM events `

	// Items in whereClause are added to the query in order, separated by ' AND '.
	var whereClause []string
	if options.EventID != 0 {
		whereClause = append(whereClause, "id = ?")
		queryArgs = append(queryArgs, options.EventID)
	}
	if options.DateFrom != "" {
		whereClause = append(whereClause, "date >= ?")
		queryArgs = append(queryArgs, options.DateFrom)
	}
	if options.DateTo != "" {
		whereClause = append(whereClause, "date <= ?")
		queryArgs = append(queryArgs, options.DateTo)
	}
	if options.EventType != "" {
		whereClause = append(whereClause, "event_type = ?")
		queryArgs = append(queryArgs, options.EventType)
	}
	if options.EventNameQuery != "" {
		whereClause = append(whereClause, "MATCH (name) AGAINST (?)")
		queryArgs = append(queryArgs, options.EventNameQuery)
	}

	// Add the where clauses to the query.
	if len(whereClause) != 0 {
		query += ` WHERE ` + strings.Join(whereClause, " AND ")
	}

	if options.OrderBy != "" {
		// Potentially sketchy sql injection...
		query += ` ORDER BY ` + options.OrderBy
	}

	var events []Event
	err := db.Select(&events, query, queryArgs...)
	if err != nil {
		return nil, err
	}

	// Get attendees
	for i := range events {
		var attendees []User
		err = db.Select(&attendees, `SELECT
a.id, a.name, a.email, a.chapter, a.phone, a.location, a.facebook
FROM activists a
JOIN event_attendance et
  ON a.id = et.activist_id
WHERE
  et.event_id = ?`, events[i].ID)
		if err != nil {
			return nil, err
		}
		events[i].Attendees = attendees
	}
	return events, nil
}

/* Get attendance for a single event
 * Returns a zero-value slice if query returns no results
 */
func GetEventAttendance(db *sqlx.DB, eventID int) ([]string, error) {
	var attendees []string
	err := db.Select(&attendees, `SELECT a.name FROM activists a 
    JOIN event_attendance et on a.id = et.activist_id WHERE et.event_id = ?`, eventID)
	if err != nil {
		return nil, err
	}
	return attendees, nil
}

func DeleteEvent(db *sqlx.DB, eventID int) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	_, err = tx.Exec(`DELETE FROM event_attendance
WHERE event_id = ?`, eventID)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(`DELETE FROM events
WHERE id = ?`, eventID)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return err
	}

	return nil
}

// Value implements the driver.Valuer interface
func (et EventType) Value() (driver.Value, error) {
	return string(et), nil
}

// Scan implements the sql.Scanner interface
func (et *EventType) Scan(src interface{}) error {
	*et = EventType(src.([]uint8))

	return nil
}

func getEventType(rawEventType string) (EventType, error) {
	rawEventType = strings.TrimSpace(rawEventType)
	if EventTypes[rawEventType] {
		return EventType(rawEventType), nil
	}
	return "", errors.New("Not a valid event type: " + rawEventType)
}

func InsertUpdateEvent(db *sqlx.DB, event Event) (eventID int, err error) {
	if event.ID == 0 {
		return insertEvent(db, event)
	}
	return updateEvent(db, event)
}

func insertEvent(db *sqlx.DB, event Event) (eventID int, err error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	res, err := tx.NamedExec(`INSERT INTO events (name, date, event_type)
VALUES (:name, :date, :event_type)`, event)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	event.ID = int(id)

	if err := insertEventAttendance(tx, event); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}
	return int(id), nil
}

func updateEvent(db *sqlx.DB, event Event) (eventID int, err error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, err
	}
	_, err = tx.NamedExec(`UPDATE events
SET
  name = :name,
  date = :date,
  event_type = :event_type
WHERE
  id = :id`, event)
	if err != nil {
		tx.Rollback()
		return 0, err
	}

	if err := insertEventAttendance(tx, event); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}
	return event.ID, nil
}

/* Changes: Delete removed activists from attendance and add new ones */
func insertEventAttendance(tx *sqlx.Tx, event Event) error {
	if event.ID == 0 {
		// Not a valid event id, so return an error
		return errors.New("Invalid event ID. Event ID's must be greater than 0.")
	}
	// First, remove deleted attendees.
	for _, u := range event.DeletedAttendees {
		_, err := tx.Exec(`DELETE FROM event_attendance WHERE event_id = ?
        AND activist_id = ?`, event.ID, u.ID)
		if err != nil {
			return err
		}
	}
	// Add new attendees to the event_attendance
	seen := map[int]bool{}
	for _, u := range event.AddedAttendees {
		// Ignore duplicates
		if _, exists := seen[u.ID]; exists {
			continue
		}
		seen[u.ID] = true
		// Insert new (activist_id, event_id) pairs to event_attendance table
		// For duplicates,  set activist_id equal to itself. In other words, do nothing
		_, err := tx.Exec(`INSERT INTO event_attendance (activist_id, event_id)
            VALUES(?,?) ON DUPLICATE KEY UPDATE activist_id = activist_id`, u.ID, event.ID)
		if err != nil {
			return err
		}
	}
	return nil
}
