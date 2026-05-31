package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// timeLayout is the wall-clock layout (HH:MM) accepted from the form's
// time inputs for start_time / end_time.
const timeLayout string = "15:04"

/** Constant and Global Variable Definitions */

const EventDateLayout string = "2006-01-02"
const ActionEventTypesSQL string = "'Outreach', 'Action', 'Campaign Action', 'Animal Care', 'Frontline Surveillance'"

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

	// Advance-event fields. All optional; empty/zero values keep the
	// existing attendance flow unchanged.
	IsOnline    bool   `json:"is_online"`
	Description string `json:"description,omitempty"`
	StartTime   string `json:"start_time,omitempty"` // local wall-clock "HH:MM"
	EndTime     string `json:"end_time,omitempty"`   // local wall-clock "HH:MM"
	Timezone    string `json:"timezone,omitempty"`   // IANA zone, e.g. "America/Los_Angeles"
	IsPublic    bool   `json:"is_public"`

	// Location. Submitted as a Google Place (no free-text); resolved place
	// fields are echoed back for display when reading an event. Null/omitted
	// for attendance and online events.
	Location *LocationJSON `json:"location,omitempty"`
}

// LocationJSON is the deduped Google Place attached to an event.
type LocationJSON struct {
	LocationID       int      `json:"id,omitempty"`
	GooglePlaceID    string   `json:"google_place_id,omitempty"`
	Name             string   `json:"name,omitempty"`
	FormattedAddress string   `json:"formatted_address,omitempty"`
	Lat              *float64 `json:"lat,omitempty"`
	Lng              *float64 `json:"lng,omitempty"`
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
	AttendeePhones        []string
	AttendeeIDs           []int
	AttendeeMissingEmails []string   // Used for sending event surveys
	AddedAttendees        []Activist // Used for Updating Events
	DeletedAttendees      []Activist // Used for Updating Events
	CircleID              int        `db:"circle_id"`
	ChapterID             int        `db:"chapter_id"`

	// Advance-event columns.
	LocationID  *int           `db:"location_id"`
	IsOnline    bool           `db:"is_online"`
	Description sql.NullString `db:"description"`
	StartTime   sql.NullString `db:"start_time"`
	EndTime     sql.NullString `db:"end_time"`
	Timezone    string         `db:"timezone"`
	IsPublic    bool           `db:"is_public"`

	// Resolved location fields from LEFT JOIN locations (null when the event
	// has no location). Read-only — not written back to the events table.
	LocationPlaceID          sql.NullString  `db:"location_google_place_id"`
	LocationName             sql.NullString  `db:"location_name"`
	LocationFormattedAddress sql.NullString  `db:"location_formatted_address"`
	LocationLat              sql.NullFloat64 `db:"location_lat"`
	LocationLng              sql.NullFloat64 `db:"location_lng"`
}

func (event *Event) ToJSON() EventJSON {
	j := EventJSON{
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

		IsOnline:    event.IsOnline,
		Description: event.Description.String,
		StartTime:   event.StartTime.String,
		EndTime:     event.EndTime.String,
		Timezone:    event.Timezone,
		IsPublic:    event.IsPublic,
	}
	// Only attach a location when the event resolves to one (LEFT JOIN match).
	if event.LocationID != nil {
		j.Location = &LocationJSON{
			LocationID:       *event.LocationID,
			GooglePlaceID:    event.LocationPlaceID.String,
			Name:             event.LocationName.String,
			FormattedAddress: event.LocationFormattedAddress.String,
			Lat:              nullFloatToPtr(event.LocationLat),
			Lng:              nullFloatToPtr(event.LocationLng),
		}
	}
	return j
}

func nullFloatToPtr(n sql.NullFloat64) *float64 {
	if !n.Valid {
		return nil
	}
	v := n.Float64
	return &v
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
	// IncludeAttendeeEmails fetches the email column in the attendance
	// query and populates Event.AttendeeEmails. By default emails are
	// dropped at the SQL layer so they never leave the database — opt in
	// only when the caller is authorized to see attendee PII.
	IncludeAttendeeEmails bool
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
		return Event{}, err
	} else if len(events) == 0 {
		return Event{}, errors.Wrapf(ErrNotFound, "could not find event with id %d", options.EventID)
	} else if len(events) > 1 {
		return Event{}, errors.Errorf("found too many events with id %d", options.EventID)
	}
	return events[0], nil
}

func getEvents(db *sqlx.DB, options GetEventOptions) ([]Event, error) {
	query := `SELECT e.id, e.name, e.date, e.event_type, e.survey_sent, e.suppress_survey, e.circle_id, e.chapter_id,
e.location_id, e.is_online, e.description, e.start_time, e.end_time, e.timezone, e.is_public,
l.google_place_id AS location_google_place_id,
l.name AS location_name,
l.formatted_address AS location_formatted_address,
l.lat AS location_lat,
l.lng AS location_lng
FROM events e
LEFT JOIN locations l ON e.location_id = l.id `

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
		where("LOWER(e.event_type) in (" + ActionEventTypesSQL + ")")
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

	emailCol := "''"
	if options.IncludeAttendeeEmails {
		emailCol = "a.email"
	}
	attendanceQuery, attendanceArgs, err := sqlx.In(`
SELECT
  ea.event_id,
  a.name as activist_name,
  `+emailCol+` as activist_email,
  a.phone as activist_phone,
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
		ActivistPhone string `db:"activist_phone"`
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
		if options.IncludeAttendeeEmails {
			events[i].AttendeeEmails = append(events[i].AttendeeEmails, a.ActivistEmail)
		}
		events[i].AttendeePhones = append(events[i].AttendeePhones, a.ActivistPhone)
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
		_ = tx.Rollback()
		return errors.Wrapf(err, "failed to delete event attendance for event %d", eventID)
	}

	_, err = tx.Exec(`DELETE FROM events
WHERE id = ? AND chapter_id = ?`, eventID, chapterID)
	if err != nil {
		_ = tx.Rollback()
		return errors.Wrapf(err, "failed to delete event %d", eventID)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
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
	res, err := tx.NamedExec(`INSERT INTO events
(name, date, event_type, suppress_survey, circle_id, chapter_id,
 location_id, is_online, description, start_time, end_time, timezone, is_public)
VALUES (:name, :date, :event_type, :suppress_survey, :circle_id, :chapter_id,
 :location_id, :is_online, :description, :start_time, :end_time, :timezone, :is_public)`, event)
	if err != nil {
		_ = tx.Rollback()
		return 0, errors.Wrap(err, "failed to insert event")
	}
	id, err := res.LastInsertId()
	if err != nil {
		_ = tx.Rollback()
		return 0, errors.Wrap(err, "failed to get inserted event id")
	}
	event.ID = int(id)

	if err := insertEventAttendance(tx, event); err != nil {
		_ = tx.Rollback()
		return 0, errors.Wrap(err, "failed to insert event attendance")
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
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
		_ = tx.Rollback()
		return 0, errors.Wrap(err, "failed to get event count")
	}
	if eventCount == 0 {
		_ = tx.Rollback()
		return 0, errors.Errorf("Event with id %d does not exist", event.ID)
	}

	// Update the event
	_, err = tx.NamedExec(`UPDATE events
SET
  name = :name,
  date = :date,
  event_type = :event_type,
  suppress_survey = :suppress_survey,
  circle_id = :circle_id,
  location_id = :location_id,
  is_online = :is_online,
  description = :description,
  start_time = :start_time,
  end_time = :end_time,
  timezone = :timezone,
  is_public = :is_public
WHERE
  id = :id`, event)
	if err != nil {
		_ = tx.Rollback()
		return 0, errors.Wrap(err, "failed to update event")
	}

	if err := insertEventAttendance(tx, event); err != nil {
		_ = tx.Rollback()
		return 0, errors.Wrap(err, "failed to insert event attendance")
	}
	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
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
		_ = tx.Rollback()
		return 0, errors.Wrap(err, "failed to update event")
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
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

	e.IsPublic = eventJSON.IsPublic
	e.IsOnline = eventJSON.IsOnline

	// Description: trim and reject dangerous characters; empty -> NULL.
	if err := checkForDangerousChars(eventJSON.Description); err != nil {
		return Event{}, err
	}
	e.Description = nullStringFromValue(eventJSON.Description)

	// Start/end times are local wall-clock "HH:MM"; validate when present.
	startTime, err := cleanEventTime(eventJSON.StartTime)
	if err != nil {
		return Event{}, fmt.Errorf("invalid start_time: %w", err)
	}
	e.StartTime = startTime
	endTime, err := cleanEventTime(eventJSON.EndTime)
	if err != nil {
		return Event{}, fmt.Errorf("invalid end_time: %w", err)
	}
	e.EndTime = endTime

	// Timezone: IANA zone name; validate when present.
	e.Timezone = strings.TrimSpace(eventJSON.Timezone)
	if e.Timezone != "" {
		if _, err := time.LoadLocation(e.Timezone); err != nil {
			return Event{}, fmt.Errorf("invalid timezone %q: %w", e.Timezone, err)
		}
	}

	// Location: a Google Place (no free-text). Only resolve a location when a
	// place is supplied and the event is not online; otherwise location_id is
	// NULL. This keeps attendance/online events location-less.
	if !e.IsOnline && eventJSON.Location != nil {
		placeID := strings.TrimSpace(eventJSON.Location.GooglePlaceID)
		if placeID != "" {
			locationID, err := GetOrCreateLocation(
				db,
				chapterID,
				placeID,
				strings.TrimSpace(eventJSON.Location.Name),
				strings.TrimSpace(eventJSON.Location.FormattedAddress),
				eventJSON.Location.Lat,
				eventJSON.Location.Lng,
			)
			if err != nil {
				return Event{}, err
			}
			e.LocationID = &locationID
		}
	}

	return e, nil
}

// cleanEventTime validates an optional local wall-clock "HH:MM" string,
// returning a NULL value when empty.
func cleanEventTime(raw string) (sql.NullString, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sql.NullString{}, nil
	}
	if _, err := time.Parse(timeLayout, raw); err != nil {
		return sql.NullString{}, err
	}
	return sql.NullString{String: raw, Valid: true}, nil
}

// nullStringFromValue trims a string and returns a NULL value when empty.
func nullStringFromValue(raw string) sql.NullString {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: raw, Valid: true}
}

// GetOrCreateLocation upserts a location keyed by (chapter_id, google_place_id)
// and returns its id. Mirrors GetOrCreateActivist. Place metadata (name,
// address, coordinates) is refreshed on conflict.
func GetOrCreateLocation(db *sqlx.DB, chapterID int, placeID, name, formattedAddress string, lat, lng *float64) (int, error) {
	tx, err := db.Beginx()
	if err != nil {
		return 0, fmt.Errorf("failed to create transaction: %w", err)
	}

	_, err = tx.Exec(`INSERT INTO locations (chapter_id, google_place_id, name, formatted_address, lat, lng)
VALUES (?, ?, ?, ?, ?, ?)
ON DUPLICATE KEY UPDATE name = VALUES(name), formatted_address = VALUES(formatted_address), lat = VALUES(lat), lng = VALUES(lng)`,
		chapterID, placeID, name, formattedAddress, lat, lng)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("failed to upsert location %s: %w", placeID, err)
	}

	var locationID int
	err = tx.Get(&locationID, `SELECT id FROM locations WHERE chapter_id = ? AND google_place_id = ?`, chapterID, placeID)
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("failed to get location %s: %w", placeID, err)
	}

	if err := tx.Commit(); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("failed to commit location %s: %w", placeID, err)
	}

	return locationID, nil
}

func cleanEventAttendanceData(db *sqlx.DB, attendees []string, chapterID int) ([]Activist, error) {
	activists := make([]Activist, len(attendees))
	caser := cases.Title(language.Und)

	for idx, attendee := range attendees {
		if err := checkForDangerousChars(attendee); err != nil {
			return []Activist{}, err
		}
		cleanAttendee := caser.String(strings.TrimSpace(attendee))
		activist, err := GetOrCreateActivist(db, cleanAttendee, chapterID)
		if err != nil {
			return []Activist{}, err
		}
		activists[idx] = activist
	}

	return activists, nil
}
