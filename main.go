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

	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
	"github.com/unrolled/render"
	"github.com/urfave/negroni"

	_ "github.com/mattn/go-sqlite3"
)

func NewDB() *sqlx.DB {
	db, err := sqlx.Open("sqlite3", "adb.db")
	if err != nil {
		panic(err)
	}

	return db
}

var DB = NewDB()

func router() *mux.Router {
	router := mux.NewRouter()
	router.HandleFunc("/", IndexHandler)

	router.HandleFunc("/get_autocomplete_activist_names", AutocompleteActivistsHandler)

	router.HandleFunc("/update_event", UpdateEventHandler)

	router.PathPrefix("/static").Handler(http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return router
}

func IndexHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New(render.Options{
		Layout: "layout",
	})
	r.HTML(w, http.StatusOK, "event_new", nil)
}

func getAutocompleteNames() []string {
	type Name struct {
		Name string `db:"name"`
	}
	names := []Name{}
	err := DB.Select(&names, "SELECT name FROM activists")
	if err != nil {
		panic(err)
	}

	ret := []string{}
	for _, n := range names {
		ret = append(ret, n.Name)
	}
	return ret
}

func AutocompleteActivistsHandler(w http.ResponseWriter, req *http.Request) {
	r := render.New()
	names := getAutocompleteNames()
	r.JSON(w, http.StatusOK, map[string][]string{
		"activist_names": names,
	})
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

func GetUser(name string) (User, error) {
	var user User
	err := DB.Get(&user, `SELECT
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

func GetOrCreateUser(name string) (User, error) {
	user, err := GetUser(name)
	if err == nil {
		// We got a valid user, return them.
		return user, nil
	}

	// There was an error, so try inserting the user first.
	_, err = DB.Exec("INSERT INTO activists (name) VALUES ($1)", name)
	if err != nil {
		return User{}, nil
	}

	return GetUser(name)
}

func cleanEventData(body io.Reader) (Event, error) {
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
		user, err := GetOrCreateUser(strings.TrimSpace(attendee))
		if err != nil {
			return Event{}, err
		}
		e.Attendees = append(e.Attendees, user)
	}

	return e, nil
}

func InsertEvent(event Event) error {
	// TODO: wrap in transaction
	res, err := DB.NamedExec(`INSERT INTO events (name, date, type)
VALUES (:name, :date, :type)`, event)
	if err != nil {
		return err
	}
	id, err := res.LastInsertId()
	if err != nil {
		return err
	}
	for _, u := range event.Attendees {
		res, err = DB.Exec(`INSERT INTO event_attendance (activist_id, event_id)
VALUES ($1, $2)`, u.ID, id)
		if err != nil {
			return err
		}
	}
	return nil
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

func UpdateEventHandler(w http.ResponseWriter, req *http.Request) {
	event, err := cleanEventData(req.Body)
	if err != nil {
		panic(err)
	}

	err = InsertEvent(event)
	if err != nil {
		panic(err)
	}
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
