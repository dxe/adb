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
	_ "github.com/mattn/go-sqlite3"
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

func GetEvents(db *sqlx.DB) ([]Event, error) {
	return getEvents(db, 0)
}

func GetEvent(db *sqlx.DB, eventID int) (Event, error) {
	if eventID == 0 {
		return Event{}, errors.New("EventID for GetEvent cannot be zero")
	}
	events, err := getEvents(db, eventID)
	if err != nil {
		return Event{}, nil
	} else if len(events) == 0 {
		return Event{}, errors.New("Could not find any events")
	} else if len(events) > 1 {
		return Event{}, errors.New("Found too many events")
	}
	return events[0], nil
}

func getEvents(db *sqlx.DB, eventID int) ([]Event, error) {
	var events []Event
	var err error
	if eventID == 0 {
		err = db.Select(&events, `SELECT id, name, date, event_type
FROM events`)
	} else {
		err = db.Select(&events, `SELECT id, name, date, event_type
FROM events WHERE id = $1`, eventID)
	}
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
  et.event_id = $1`, events[i].ID)
		if err != nil {
			return nil, err
		}
		events[i].Attendees = attendees
	}
	return events, nil
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
	ID        int           `db:"id"`
	Name      string        `db:"name"`
	Email     string        `db:"email"`
	ChapterID sql.NullInt64 `db:"chapter_id"`
	Phone     string        `db:"phone"`
	Location  string        `db:"location"`
	Facebook  string        `db:"facebook"`
}

func GetUser(db *sqlx.DB, name string) (User, error) {
	var user User
	err := db.Get(&user, `SELECT
  id, name, email, chapter_id, phone, location, facebook
FROM activists
WHERE
  name = $1`, name)
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
	_, err = db.Exec("INSERT INTO activists (name) VALUES ($1)", name)
	if err != nil {
		return User{}, nil
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

func InsertEvent(db *sqlx.DB, event Event) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}
	res, err := tx.NamedExec(`INSERT INTO events (name, date, event_type)
VALUES (:name, :date, :event_type)`, event)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	for _, u := range event.Attendees {
		res, err = tx.Exec(`INSERT INTO event_attendance (activist_id, event_id)
VALUES ($1, $2)`, u.ID, id)
		if err != nil {
			return err
		}
	}
	return tx.Commit()
}
