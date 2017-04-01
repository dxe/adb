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

type NewEvent struct {
	EventName string    `db:"name"`
	EventDate time.Time `db:"date"`
	EventType EventType `db:"event_type"`
	Attendees []User
}

type Event struct {
	ID        int       `db:"id"`
	EventName string    `db:"name"`
	EventDate time.Time `db:"date"`
	EventType EventType `db:"event_type"`
}

func GetEvents(db *sqlx.DB) ([]Event, error) {
	var events []Event
	err := db.Select(&events, `SELECT id, name, date, event_type
FROM events`)
	if err != nil {
		return nil, err
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

var EventTypeLayout string = "2006-01-02"

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
	City      string        `db:"city"`
	Zipcode   string        `db:"zipcode"`
	Country   string        `db:"country"`
	Facebook  string        `db:"facebook"`
}

func GetUser(db *sqlx.DB, name string) (User, error) {
	var user User
	err := db.Get(&user, `SELECT
  id, name, email, chapter_id, phone, city,
  zipcode, country, facebook
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

func CleanEventData(db *sqlx.DB, body io.Reader) (NewEvent, error) {
	var eventJSON EventJSON
	err := json.NewDecoder(body).Decode(&eventJSON)
	if err != nil {
		return NewEvent{}, err
	}

	// Strip spaces from front and back of all fields.
	var e NewEvent
	e.EventName = strings.TrimSpace(eventJSON.EventName)
	t, err := time.Parse(EventTypeLayout, eventJSON.EventDate)
	if err != nil {
		return NewEvent{}, err
	}
	e.EventDate = t
	eventType, err := getEventType(eventJSON.EventType)
	if err != nil {
		return NewEvent{}, err
	}
	e.EventType = eventType

	e.Attendees = []User{}
	for _, attendee := range eventJSON.Attendees {
		user, err := GetOrCreateUser(db, strings.TrimSpace(attendee))
		if err != nil {
			return NewEvent{}, err
		}
		e.Attendees = append(e.Attendees, user)
	}

	return e, nil
}

func InsertNewEvent(db *sqlx.DB, event NewEvent) error {
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
