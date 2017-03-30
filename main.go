package main

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/directactioneverywhere/adb/model"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	_ "github.com/mattn/go-sqlite3"
)

func router() *mux.Router {
	main := MainController{db: model.NewDB("adb.db")}

	router := mux.NewRouter()
	router.HandleFunc("/", main.IndexHandler)
	router.HandleFunc("/get_autocomplete_activist_names", main.AutocompleteActivistsHandler)
	router.HandleFunc("/update_event", main.UpdateEventHandler)
	router.HandleFunc("/list_events", main.ListEventsHandler)

	router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return router
}

type MainController struct {
	db *sqlx.DB
}

func (c MainController) ListEventsHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New()
	events, err := GetEvents(c.db)
	if err != nil {
		panic(err)
	}
	r.JSON(w, http.StatusOK, map[string]interface{}{
		"events": events,
	})
}

func (c MainController) IndexHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{
		Layout: "layout",
	})
	r.HTML(w, http.StatusOK, "event_new", nil)
}

func (c MainController) AutocompleteActivistsHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New()
	names := getAutocompleteNames(c.db)
	r.JSON(w, http.StatusOK, map[string][]string{
		"activist_names": names,
	})
}

func (c MainController) UpdateEventHandler(w http.ResponseWriter, req *http.Request) {
	event, err := cleanEventData(c.db, req.Body)
	if err != nil {
		panic(err)
	}

	err = InsertEvent(c.db, event)
	if err != nil {
		panic(err)
	}
}

func GetEvents(db *sqlx.DB) ([]Event, error) {
	var events []Event
	// TODO: this doesn't work!
	// 	err := DB.Select(&events, `SELECT (name, date, type)
	// FROM events`)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	return events, nil
}

func getAutocompleteNames(db *sqlx.DB) []string {
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
func (et EventType) Scan(src interface{}) error {
	et = EventType(src.([]uint8))

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

func cleanEventData(db *sqlx.DB, body io.Reader) (Event, error) {
	var raw RawEvent
	err := json.NewDecoder(body).Decode(&raw)
	if err != nil {
		return Event{}, err
	}

	// Strip spaces from front and back of all fields.
	var e Event
	e.EventName = strings.TrimSpace(raw.EventName)
	t, err := time.Parse(EventTypeLayout, raw.EventDate)
	if err != nil {
		return Event{}, err
	}
	e.EventDate = t
	eventType, err := getEventType(raw.EventType)
	if err != nil {
		return Event{}, err
	}
	e.EventType = eventType

	e.Attendees = []User{}
	for _, attendee := range raw.Attendees {
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
	res, err := tx.NamedExec(`INSERT INTO events (name, date, type)
VALUES (:name, :date, :type)`, event)
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

type RawEvent struct {
	EventName string   `json:"event_name"`
	EventDate string   `json:"event_date"`
	EventType string   `json:"event_type"`
	Attendees []string `json:"attendees"`
}

type Event struct {
	EventName string    `db:"name"`
	EventDate time.Time `db:"date"`
	EventType EventType `db:"type"`
	Attendees []User
}

func main() {
	n := negroni.New()

	n.Use(negroni.NewRecovery())
	n.Use(negroni.NewLogger())

	r := router()
	n.UseHandler(r)

	fmt.Println("Listening on localhost:8080")
	http.ListenAndServe(":8080", n)
}
