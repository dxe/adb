package model

import (
	"database/sql/driver"
	"encoding/json"
	"io"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

/** Constant and Global Variable Definitions */

const EventDateLayout string = "2006-01-02"

var EventTypes map[string]bool = map[string]bool{
	"Action":                 true,
	"Campaign Action":        true,
	"Community":              true,
	"Frontline Surveillance": true,
	"Meeting":                true,
	"Outreach":               true,
	"Animal Care":            true,
	"Training":               true,
	"Connection":             true,
}

/** Type Definitions */

type EventType string

/* TODO Restructure this struct */
type EventJSON struct {
	EventID          int      `json:"event_id"`
	EventName        string   `json:"event_name"`
	EventDate        string   `json:"event_date"`
	EventType        string   `json:"event_type"`
	Attendees        []string `json:"attendees"` // For displaying all event attendees
	AttendeeEmails   []string `json:"attendee_emails"`
	AttendeeIDs      []int    `json:"attendee_ids"`
	AddedAttendees   []string `json:"added_attendees"`   // Used for Updating Events
	DeletedAttendees []string `json:"deleted_attendees"` // Used for Updating Events
	SuppressSurvey   bool     `json:"suppress_survey"`
	CircleID         int      `json:"circle_id"`
	ChapterID        int      `json:"chapter_id"`
}

/* TODO Restructure this Struct */
type Event struct {
	ID                    int       `db:"id"`
	EventName             string    `db:"name"`
	EventDate             time.Time `db:"date"`
	EventType             EventType `db:"event_type"`
	SurveySent            int       `db:"survey_sent"` // Used for sending event surveys
	SuppressSurvey        bool      `db:"suppress_survey"`
	Attendees             []string  // For retrieving all event attendees
	AttendeeEmails        []string
	AttendeeIDs           []int
	AttendeeMissingEmails []string   // Used for sending event surveys
	AddedAttendees        []Activist // Used for Updating Events
	DeletedAttendees      []Activist // Used for Updating Events
	CircleID              int        `db:"circle_id"`
	ChapterID             int        `db:"chapter_id"`
}

func (event *Event) ToJSON() EventJSON {
	return EventJSON{
		EventID:        event.ID,
		EventName:      event.EventName,
		EventDate:      event.EventDate.Format(EventDateLayout),
		EventType:      string(event.EventType),
		Attendees:      event.Attendees,
		AttendeeEmails: event.AttendeeEmails,
		AttendeeIDs:    event.AttendeeIDs,
		SuppressSurvey: event.SuppressSurvey,
		CircleID:       event.CircleID,
		ChapterID:      event.ChapterID,
	}
}

type GetEventOptions struct {
	EventID   int
	ChapterID int
	// NOTE: don't pass user input to OrderBy, cause that could
	// cause a SQL injection.
	OrderBy        string
	DateFrom       string
	DateTo         string
	EventType      string
	EventNameQuery string
	EventActivist  string
	SurveySent     string
	SuppressSurvey string
}

/** Functions and Methods */

func GetEventsJSON(db *sqlx.DB, options GetEventOptions) ([]EventJSON, error) {
	dbEvents, err := GetEvents(db, options)

	if err != nil {
		return nil, err
	}

	events := make([]EventJSON, 0, len(dbEvents))
	for _, event := range dbEvents {
		events = append(events, event.ToJSON())
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
	query := `SELECT e.id, e.name, e.date, e.event_type, e.survey_sent, e.suppress_survey, e.circle_id, e.chapter_id FROM events e `

	// Items in whereClause are added to the query in order, separated by ' AND '.
	var whereClause []string
	var queryArgs []interface{}

	where := func(clause string, args ...interface{}) {
		whereClause = append(whereClause, clause)
		queryArgs = append(queryArgs, args...)
	}

	if options.EventActivist != "" {
		// If we're filtering with an activist name, we need
		// to join a couple tables which makes this slightly
		// more complicated.
		query += `
JOIN (event_attendance ea, activists a)
ON (e.id = ea.event_id AND ea.activist_id = a.id)
`
		where("a.name = ?", options.EventActivist)
	}

	if options.EventID != 0 {
		where("e.id = ?", options.EventID)
	}
	if options.ChapterID != 0 {
		where("e.chapter_id = ?", options.ChapterID)
	}
	if options.DateFrom != "" {
		where("e.date >= ?", options.DateFrom)
	}
	if options.DateTo != "" {
		where("e.date <= ?", options.DateTo)
	}
	if options.SurveySent != "" {
		where("e.survey_sent = ?", options.SurveySent)
	}
	if options.SuppressSurvey != "" {
		where("e.suppress_survey = ?", options.SuppressSurvey)
	}
	if options.EventType == "noConnections" {
		where("e.event_type <> 'Connection'")
	} else if options.EventType == "mpiDA" {
		where("e.event_type in ('Outreach', 'Action', 'Campaign Action', 'Animal Care', 'Frontline Surveillance')")
	} else if options.EventType == "mpiCOM" {
		where("e.event_type in ('Community', 'Training', 'Circle')")
	} else if options.EventType != "" {
		where("e.event_type like ?", options.EventType)
	}
	if options.EventNameQuery != "" {
		where("MATCH (e.name) AGAINST (?)", options.EventNameQuery)
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
		return nil, errors.Wrap(err, "failed to select events")
	}
	if len(events) == 0 {
		return nil, nil
	}

	// Create a map of eventIDs to their index in `events` so we can easily add all
	// attendance to them.
	eventIDToIndex := map[int]int{}
	// Create a list of eventIDs so we can pass them into the all
	// attendance query.
	var eventIDs []int
	for i, e := range events {
		eventIDs = append(eventIDs, e.ID)
		eventIDToIndex[e.ID] = i
	}

	attendanceQuery, attendanceArgs, err := sqlx.In(`
SELECT
  ea.event_id,
  a.name as activist_name,
  a.email as activist_email,
  a.id as activist_id
FROM activists a
JOIN event_attendance ea
  ON a.id = ea.activist_id
WHERE
  ea.event_id IN (?)`, eventIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "could not create sqlx.In query")
	}

	attendanceQuery = db.Rebind(attendanceQuery)
	type Attendance struct {
		EventID       int    `db:"event_id"`
		ActivistName  string `db:"activist_name"`
		ActivistEmail string `db:"activist_email"`
		ActivistID    int    `db:"activist_id"`
	}
	var allAttendance []Attendance
	err = db.Select(&allAttendance, attendanceQuery, attendanceArgs...)
	if err != nil {
		return nil, errors.Wrapf(err, "could not make all attendance query")
	}

	for _, a := range allAttendance {
		i := eventIDToIndex[a.EventID]
		events[i].Attendees = append(events[i].Attendees, a.ActivistName)
		events[i].AttendeeEmails = append(events[i].AttendeeEmails, a.ActivistEmail)
		events[i].AttendeeIDs = append(events[i].AttendeeIDs, a.ActivistID)
	}

	return events, nil
}

/* Get attendance for a single event
 * Returns a zero-value slice if query returns no results
 */
func GetEventAttendance(db *sqlx.DB, eventID int) ([]string, error) {
	var attendees []string
	err := db.Select(&attendees, `SELECT a.name FROM activists a
    JOIN event_attendance ea on a.id = ea.activist_id WHERE ea.event_id = ?`, eventID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get attendees for event %d", eventID)
	}
	return attendees, nil
}

func DeleteEvent(db *sqlx.DB, eventID int, chapterID int) error {
	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}
	_, err = tx.Exec(`DELETE FROM event_attendance
WHERE event_id = ?`, eventID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to delete event attendance for event %d", eventID)
	}

	_, err = tx.Exec(`DELETE FROM events
WHERE id = ? AND chapter_id = ?`, eventID, chapterID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to delete event %d", eventID)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to commit event %d", eventID)
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
		return 0, errors.Wrap(err, "failed to create transaction")
	}
	res, err := tx.NamedExec(`INSERT INTO events (name, date, event_type, suppress_survey, circle_id, chapter_id)
VALUES (:name, :date, :event_type, :suppress_survey, :circle_id, :chapter_id)`, event)
	if err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to insert event")
	}
	id, err := res.LastInsertId()
	if err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to get inserted event id")
	}
	event.ID = int(id)

	if err := insertEventAttendance(tx, event); err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to insert event attendance")
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed insert event transaction")
	}
	return int(id), nil
}

func updateEvent(db *sqlx.DB, event Event) (eventID int, err error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, errors.Wrap(err, "failed to update event")
	}
	// Error out if the event doesn't exist for this chapter.
	var eventCount int
	err = tx.Get(&eventCount, `SELECT count(*) FROM events WHERE id = ? AND chapter_id = ?`, event.ID, event.ChapterID)
	if err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to get event count")
	}
	if eventCount == 0 {
		tx.Rollback()
		return 0, errors.Errorf("Event with id %d does not exist", event.ID)
	}

	// Update the event
	_, err = tx.NamedExec(`UPDATE events
SET
  name = :name,
  date = :date,
  event_type = :event_type,
  suppress_survey = :suppress_survey,
  circle_id = :circle_id
WHERE
  id = :id`, event)
	if err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to update event")
	}

	if err := insertEventAttendance(tx, event); err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to insert event attendance")
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to commit update event")
	}
	return event.ID, nil
}

func UpdateEventSurveyStatus(db *sqlx.DB, event Event) (eventID int, err error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, errors.Wrap(err, "failed to update event")
	}

	// Update the event
	_, err = tx.NamedExec(`UPDATE events
SET
  survey_sent = :survey_sent
WHERE
  id = :id`, event)
	if err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to update event")
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "failed to commit update event")
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
			return errors.Wrap(err, "failed to delete attendees")
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
			return errors.Wrap(err, "failed to insert attendees")
		}
	}
	return nil
}

func CleanEventData(db *sqlx.DB, body io.Reader, chapterID int) (Event, error) {
	var eventJSON EventJSON
	err := json.NewDecoder(body).Decode(&eventJSON)
	if err != nil {
		return Event{}, err
	}

	// Strip spaces from front and back of all fields.
	var e Event
	e.ID = eventJSON.EventID

	if err := checkForDangerousChars(eventJSON.EventName); err != nil {
		return Event{}, err
	}

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

	addedAttendees, err := cleanEventAttendanceData(db, eventJSON.AddedAttendees, chapterID)
	if err != nil {
		return Event{}, err
	}

	deletedAttendees, err := cleanEventAttendanceData(db, eventJSON.DeletedAttendees, chapterID)
	if err != nil {
		return Event{}, err
	}

	e.AddedAttendees = addedAttendees
	e.DeletedAttendees = deletedAttendees

	e.SuppressSurvey = eventJSON.SuppressSurvey

	e.CircleID = eventJSON.CircleID

	e.ChapterID = chapterID

	return e, nil
}

func cleanEventAttendanceData(db *sqlx.DB, attendees []string, chapterID int) ([]Activist, error) {
	activists := make([]Activist, len(attendees))

	for idx, attendee := range attendees {
		if err := checkForDangerousChars(attendee); err != nil {
			return []Activist{}, err
		}
		cleanAttendee := strings.Title(strings.TrimSpace(attendee))
		activist, err := GetOrCreateActivist(db, cleanAttendee, chapterID)
		if err != nil {
			return []Activist{}, err
		}
		activists[idx] = activist
	}

	return activists, nil
}
