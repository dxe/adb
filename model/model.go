package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

type EventJSON struct {
	EventID   int      `json:"event_id"`
	EventName string   `json:"event_name"`
	EventDate string   `json:"event_date"`
	EventType string   `json:"event_type"`
	Attendees []string `json:"attendees"`
}

type Event struct {
	ID        int       `db:"id"`
	EventName string    `db:"name"`
	EventDate time.Time `db:"date"`
	EventType EventType `db:"event_type"`
	Attendees []User
}

func GetEventsJSON(db *sqlx.DB, dateFrom string, dateTo string, eventType string) ([]EventJSON, error) {
	dbEvents, err := GetEvents(db, GetEventOptions{
		OrderBy:   "date DESC",
		DateFrom:  dateFrom,
		DateTo:    dateTo,
		EventType: eventType,
	})
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

type GetEventOptions struct {
	EventID int
	// NOTE: don't pass user input to OrderBy, cause that could
	// cause a SQL injection.
	OrderBy   string
	DateFrom  string
	DateTo    string
	EventType string
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
	query := `SELECT id, name, date, event_type FROM events`

	if options.EventID != 0 {
		query += ` WHERE id = ?`
		queryArgs = append(queryArgs, options.EventID)
	}

	if (options.EventID == 0) && ((options.DateFrom != "") || (options.DateTo != "")) {
		/* Only get events in date range if event id not provided */
		switch rangeType := checkValidDateRange(options.DateFrom, options.DateTo); rangeType {
		case -1:
			//TODO Maybe handle this differently? Returning no events now
			return make([]Event, 0), nil
		case 1:
			query += ` WHERE date >= ?`
			queryArgs = append(queryArgs, options.DateFrom)
		case 2:
			query += ` WHERE date <= ?`
			queryArgs = append(queryArgs, options.DateTo)
		case 3:
			query += ` WHERE date >= ? AND date <= ?`
			queryArgs = append(queryArgs, options.DateFrom, options.DateTo)
		}
	}

	if options.EventType != "" {
		query += ` AND event_type LIKE ?`
		queryArgs = append(queryArgs, options.EventType)
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
a.id, a.name, a.email, a.chapter_id, a.phone, a.location, a.facebook
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

/**
* checkValidDateRange - return a number based on date range
* return -1 : dateFrom >= dateTo or date strings cannot be parsed
* return 1: dateFrom is specified by dateTo is empty
* return 2: dateFrom is empty and dateTo is specified
* return 3: Both dateFrom and dateTo are specified and valid
 */
func checkValidDateRange(dateFromStr string, dateToStr string) int {
	if dateToStr == "" {
		return 1
	}
	if dateFromStr == "" {
		return 2
	}
	/* Both dates are non-empty so make sure the range is valid */
	dateLayout := "2006-01-02"
	dateFrom, errFrom := time.Parse(dateLayout, dateFromStr)
	dateTo, errTo := time.Parse(dateLayout, dateToStr)
	if (errFrom != nil) || (errTo != nil) {
		/* Invalid date string */
		return -1
	}
	if dateFrom.After(dateTo) {
		return -1
	}
	return 3
}

func GetAutocompleteNames(db *sqlx.DB) []string {
	type Name struct {
		Name string `db:"name"`
	}
	names := []Name{}
	err := db.Select(&names, "SELECT name FROM activists ORDER BY name ASC")
	if err != nil {
		// TODO: return error
		panic(err)
	}

	ret := []string{}
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

var EventTypes map[string]bool = map[string]bool{
	"Working Group": true,
	"Community":     true,
	"Protest":       true,
	"Outreach":      true,
	"Key Event":     true,
}

var EventDateLayout string = "2006-01-02"

type EventType string

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

type User struct {
	ID        int            `db:"id"`
	Name      string         `db:"name"`
	Email     string         `db:"email"`
	ChapterID sql.NullInt64  `db:"chapter_id"`
	Phone     string         `db:"phone"`
	Location  sql.NullString `db:"location"`
	Facebook  string         `db:"facebook"`
}

func GetUser(db *sqlx.DB, name string) (User, error) {
	var user User
	err := db.Get(&user, `SELECT
id, name, email, chapter_id, phone, location, facebook
FROM activists
WHERE
  name = ?`, name)
	if user.ID == 0 || err != nil {
		return User{}, err
	}

	return user, nil
}

func GetOrCreateUser(db *sqlx.DB, name string) (User, error) {
	user, err := GetUser(db, name)
	if err == nil {
		// We got a valid user, return them.
		return user, nil
	}

	// There was an error, so try inserting the user first.
	_, err = db.Exec("INSERT INTO activists (name) VALUES (?)", name)
	if err != nil {
		return User{}, err
	}

	return GetUser(db, name)
}

func CleanEventData(db *sqlx.DB, body io.Reader) (Event, error) {
	var eventJSON EventJSON
	err := json.NewDecoder(body).Decode(&eventJSON)
	if err != nil {
		return Event{}, err
	}

	// Strip spaces from front and back of all fields.
	var e Event
	e.ID = eventJSON.EventID

	e.EventName = strings.TrimSpace(eventJSON.EventName)
	t, err := time.Parse(EventDateLayout, eventJSON.EventDate)
	if err != nil {
		return Event{}, err
	}
	e.EventDate = t
	eventType, err := getEventType(eventJSON.EventType)
	if err != nil {
		return Event{}, err
	}
	e.EventType = eventType

	e.Attendees = []User{}
	for _, attendee := range eventJSON.Attendees {
		user, err := GetOrCreateUser(db, strings.TrimSpace(attendee))
		if err != nil {
			return Event{}, err
		}
		e.Attendees = append(e.Attendees, user)
	}

	return e, nil
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

	if err := insertEventAttendance(tx, int(id), event.Attendees); err != nil {
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

	if err := insertEventAttendance(tx, event.ID, event.Attendees); err != nil {
		tx.Rollback()
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, err
	}
	return event.ID, nil
}

func insertEventAttendance(tx *sqlx.Tx, eventID int, attendees []User) error {
	// First, delete all previous attendees for the event.
	_, err := tx.Exec(`DELETE FROM event_attendance
WHERE event_id = ?`, eventID)
	if err != nil {
		return err
	}
	seen := map[int]bool{}
	// Then re-add all attendees.
	for _, u := range attendees {
		// Ignore duplicates
		if _, exists := seen[u.ID]; exists {
			continue
		}
		seen[u.ID] = true
		_, err = tx.Exec(`INSERT INTO event_attendance (activist_id, event_id)
VALUES (?, ?)`, u.ID, eventID)
		if err != nil {
			return err
		}
	}
	return nil
}
