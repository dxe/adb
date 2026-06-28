package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
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
	// older basic attendance flow unchanged.
	IsOnline    bool   `json:"is_online"`
	Description string `json:"description,omitempty"`
	StartTime   string `json:"start_time,omitempty"` // local wall-clock "HH:MM"
	EndTime     string `json:"end_time,omitempty"`   // local wall-clock "HH:MM"
	Timezone    string `json:"timezone,omitempty"`   // IANA zone, e.g. "America/Los_Angeles"
	IsPublic    bool   `json:"is_public"`

	// Location. A free-text name plus optional geo data (Google Place id and/or
	// coordinates). Null/omitted for attendance and online events.
	Location *LocationJSON `json:"location,omitempty"`
}

// LocationJSON is the location attached to an event: a free-text display name
// with optional geo data. Stored on the event itself, not in a shared table.
type LocationJSON struct {
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
	IsOnline    bool           `db:"is_online"`
	Description sql.NullString `db:"description"`
	StartTime   sql.NullString `db:"start_time"`
	EndTime     sql.NullString `db:"end_time"`
	Timezone    string         `db:"timezone"`
	IsPublic    bool           `db:"is_public"`

	// Location, stored denormalized on the event: a free-text display name
	// (always editable, never shared between events) plus optional geo data — a
	// Google Place id and/or coordinates. Online events leave these empty.
	LocationName    string          `db:"location_name"`
	LocationAddress string          `db:"location_address"`
	LocationPlaceID string          `db:"location_google_place_id"`
	LocationLat     sql.NullFloat64 `db:"location_lat"`
	LocationLng     sql.NullFloat64 `db:"location_lng"`
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
		// MySQL returns TIME as "HH:MM:SS"; the form/API contract is "HH:MM".
		StartTime: trimTimeSeconds(event.StartTime.String),
		EndTime:   trimTimeSeconds(event.EndTime.String),
		Timezone:  event.Timezone,
		IsPublic:  event.IsPublic,
	}
	// Attach a location whenever the event carries one. A name is always present
	// for an in-person event, but echo back any stored geo too.
	if event.LocationName != "" || event.LocationAddress != "" || event.LocationLat.Valid {
		j.Location = &LocationJSON{
			GooglePlaceID:    event.LocationPlaceID,
			Name:             event.LocationName,
			FormattedAddress: event.LocationAddress,
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
e.is_online, e.description, e.start_time, e.end_time, e.timezone, e.is_public,
e.location_name, e.location_address, e.location_google_place_id, e.location_lat, e.location_lng
FROM events e `

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
 is_online, description, start_time, end_time, timezone, is_public,
 location_name, location_address, location_google_place_id, location_lat, location_lng)
VALUES (:name, :date, :event_type, :suppress_survey, :circle_id, :chapter_id,
 :is_online, :description, :start_time, :end_time, :timezone, :is_public,
 :location_name, :location_address, :location_google_place_id, :location_lat, :location_lng)`, event)
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

	// Update the event.
	//
	// TODO(concurrency): these detail fields are last-write-wins. If two people
	// have the edit form open and both save, the later save silently clobbers the
	// earlier one. Attendees are unaffected (saved as an add/delete diff, see
	// insertEventAttendance), and detail co-editing is rare for our user base, so
	// this is deferred. The fix is optimistic locking: add an updated_at/version
	// column to events, have the form send the value it loaded, scope this UPDATE
	// with `AND updated_at = ?`, and return a 409 the form turns into a "this
	// event was changed by someone else, reload" prompt when 0 rows match.
	_, err = tx.NamedExec(`UPDATE events
SET
  name = :name,
  date = :date,
  event_type = :event_type,
  suppress_survey = :suppress_survey,
  circle_id = :circle_id,
  is_online = :is_online,
  description = :description,
  start_time = :start_time,
  end_time = :end_time,
  timezone = :timezone,
  is_public = :is_public,
  location_name = :location_name,
  location_address = :location_address,
  location_google_place_id = :location_google_place_id,
  location_lat = :location_lat,
  location_lng = :location_lng
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

	// Description: trim only; empty -> NULL. This is free-text prose, so common
	// characters like & are allowed (unlike the event name). It is stored via a
	// parameterized query and rendered React-escaped, so it carries no SQL- or
	// HTML-injection risk.
	e.Description = nullStringFromValue(eventJSON.Description)

	// Start/end times are local wall-clock "HH:MM"; validate when present.
	startTime, err := cleanEventTime(eventJSON.StartTime)
	if err != nil {
		return Event{}, ValidationErrorf("invalid start_time: %v", err)
	}
	e.StartTime = startTime
	endTime, err := cleanEventTime(eventJSON.EndTime)
	if err != nil {
		return Event{}, ValidationErrorf("invalid end_time: %v", err)
	}
	e.EndTime = endTime

	// Timezone: IANA zone name; validate when present.
	e.Timezone = strings.TrimSpace(eventJSON.Timezone)
	if e.Timezone != "" {
		if _, err := time.LoadLocation(e.Timezone); err != nil {
			return Event{}, ValidationErrorf("invalid timezone %q: %v", e.Timezone, err)
		}
	}

	// Location. An in-person event carries a free-text display name plus optional
	// geo data — a Google Place id and/or coordinates. The name is always stored
	// on the event itself (there is no shared place record), so editing it never
	// affects another event. Online events stay location-less.
	if !e.IsOnline && eventJSON.Location != nil {
		name := strings.TrimSpace(eventJSON.Location.Name)
		if name == "" {
			name = strings.TrimSpace(eventJSON.Location.FormattedAddress)
		}
		e.LocationName = name
		e.LocationAddress = strings.TrimSpace(eventJSON.Location.FormattedAddress)
		e.LocationPlaceID = strings.TrimSpace(eventJSON.Location.GooglePlaceID)
		if eventJSON.Location.Lat != nil {
			e.LocationLat = sql.NullFloat64{Float64: *eventJSON.Location.Lat, Valid: true}
		}
		if eventJSON.Location.Lng != nil {
			e.LocationLng = sql.NullFloat64{Float64: *eventJSON.Location.Lng, Valid: true}
		}
	}

	// A publicly listed event must form a coherent schedule. The form enforces
	// this, but guard server-side too so a malformed payload can't create a
	// broken public listing. (end_time before start_time is intentionally not
	// rejected here — see the form's note on overnight events.)
	if e.IsPublic {
		if !e.StartTime.Valid {
			return Event{}, ValidationErrorf("public events require a start time")
		}
		if e.Timezone == "" {
			return Event{}, ValidationErrorf("public events require a timezone")
		}
		if !e.IsOnline && e.LocationName == "" {
			return Event{}, ValidationErrorf("public in-person events require a location")
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

// trimTimeSeconds reduces a MySQL TIME value ("HH:MM:SS") to the "HH:MM"
// wall-clock string the form and API contract use. Leaves shorter/empty values
// untouched.
func trimTimeSeconds(t string) string {
	if len(t) >= 5 {
		return t[:5]
	}
	return t
}

// nullStringFromValue trims a string and returns a NULL value when empty.
func nullStringFromValue(raw string) sql.NullString {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: raw, Valid: true}
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
