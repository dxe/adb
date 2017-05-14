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
	query := `SELECT id, name, date, event_type FROM events `

	options.OrderBy = "date desc, id desc "

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
	ChapterID sql.NullString `db:"chapter_id"`
	Phone     string         `db:"phone"`
	Location  sql.NullString `db:"location"`
	Facebook  string         `db:"facebook"`
}

type UserJSON struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
	// TODO: Rename ChapterID to Chapter, make it a generic text
	// field.
	ChapterID   string `json:"chapter_id"`
	Phone       string `json:"phone"`
	Location    string `json:"location"`
	Facebook    string `json:"facebook"`
	FirstEvent  string `json:"firstevent"`
	LastEvent   string `json:"lastevent"`
	TotalEvents int    `json:"totalevents"`
}

type UserEventData struct {
	FirstEvent  *time.Time `db:"first_event"`
	LastEvent   *time.Time `db:"last_event"`
	TotalEvents int        `db:"total_events"`
}

func (u User) GetUserEventData(db *sqlx.DB) (UserEventData, error) {
	query := `
SELECT
  MIN(e.date) AS first_event,
  MAX(e.date) AS last_event,
  COUNT(*) as total_events
FROM events e
JOIN event_attendance
  ON event_attendance.event_id = e.id
WHERE
  event_attendance.activist_id = ?
`
	var data UserEventData
	if err := db.Get(&data, query, u.ID); err != nil {
		return UserEventData{}, err
	}
	return data, nil
}

func GetUsersJSON(db *sqlx.DB) ([]UserJSON, error) {
	var usersJSON []UserJSON
	users, err := GetUsers(db)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		eventData, err := u.GetUserEventData(db)
		if err != nil {
			return nil, err
		}
		firstEvent := ""
		if eventData.FirstEvent != nil {
			firstEvent = eventData.FirstEvent.Format(EventDateLayout)
		}
		lastEvent := ""
		if eventData.LastEvent != nil {
			lastEvent = eventData.LastEvent.Format(EventDateLayout)
		}

		usersJSON = append(usersJSON, UserJSON{
			ID:          u.ID,
			Name:        u.Name,
			Email:       u.Email,
			ChapterID:   u.ChapterID.String,
			Phone:       u.Phone,
			Location:    u.Location.String,
			Facebook:    u.Facebook,
			FirstEvent:  firstEvent,
			LastEvent:   lastEvent,
			TotalEvents: eventData.TotalEvents,
		})
	}
	return usersJSON, nil
}

func GetUsers(db *sqlx.DB) ([]User, error) {
	return getUsers(db, "")
}

func GetUser(db *sqlx.DB, name string) (User, error) {
	users, err := getUsers(db, name)
	if err != nil {
		return User{}, err
	} else if len(users) == 0 {
		return User{}, errors.New("Could not find any users")
	} else if len(users) > 1 {
		return User{}, errors.New("Found too many users")
	}
	return users[0], nil
}

func getUsers(db *sqlx.DB, name string) ([]User, error) {
	var queryArgs []interface{}
	query := `
SELECT
  a.id AS id,
  a.name AS name,
  email,
  c.name AS chapter_id,
  phone,
  location,
  facebook
FROM activists a

LEFT JOIN chapters c
  ON c.id = a.chapter_id
`

	if name != "" {
		query += "WHERE a.name = ? "
		queryArgs = append(queryArgs, name)
	}

	query += "ORDER BY a.name"

	var users []User
	if err := db.Select(&users, query, queryArgs...); err != nil {
		return nil, err
	}

	return users, nil
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

type LeaderboardUser struct {
	Name              string `db:"name"`
	FirstEvent        string `db:"first_event"`
	LastEvent         string `db:"last_event"`
	TotalEvents       int    `db:"total_events"`
	TotalEvents30Days int    `db:"total_events_30_days"`
	Points            int    `db:"points"`
}

type LeaderboardUserJSON struct {
	Name              string `json:"name"`
	FirstEvent        string `json:"first_event"`
	LastEvent         string `json:"last_event"`
	TotalEvents       int    `json:"total_events"`
	TotalEvents30Days int    `json:"total_events_30_days"`
	Points            int    `json:"points"`
}

func GetLeaderboardUsersJSON(db *sqlx.DB) ([]LeaderboardUserJSON, error) {
	var leaderboardUsersJSON []LeaderboardUserJSON
	leaderboardUsers, err := GetLeaderboardUsers(db)
	if err != nil {
		return nil, err
	}
	for _, l := range leaderboardUsers {
		leaderboardUsersJSON = append(leaderboardUsersJSON, LeaderboardUserJSON{
			Name:              l.Name,
			FirstEvent:        l.FirstEvent,
			LastEvent:         l.LastEvent,
			TotalEvents:       l.TotalEvents,
			TotalEvents30Days: l.TotalEvents30Days,
			Points:            l.Points,
		})
	}
	return leaderboardUsersJSON, nil
}

func GetLeaderboardUsers(db *sqlx.DB) ([]LeaderboardUser, error) {
	query := `
SELECT
  IFNULL(a.name,"") AS name,
  IFNULL(first_event,"None") AS first_event,
  IFNULL(last_event,"None") AS last_event,
  IFNULL(total_events,0) AS total_events,
  IFNULL(total_events_30_days,0) AS total_events_30_days,
  IFNULL((IFNULL(protest_points,0) + IFNULL(wg_points,0) + IFNULL(community_points,0) + IFNULL(outreach_points,0) + IFNULL(sanctuary_points,0) + IFNULL(key_event_points,0)),0) AS points
FROM activists a

LEFT JOIN (
  SELECT ea.activist_id, MIN(e.date) AS "first_event"
  FROM event_attendance ea
  JOIN events e
    ON e.id = ea.event_id
  GROUP BY ea.activist_id
) AS firstevent
  ON a.id = firstevent.activist_id

LEFT JOIN (
  SELECT ea.activist_id, MAX(e.date) AS "last_event"
  FROM event_attendance ea
  JOIN events e
    ON e.id = ea.event_id
  GROUP BY ea.activist_id
) AS lastevent
  ON firstevent.activist_id = lastevent.activist_id

LEFT JOIN (
  SELECT activist_id, COUNT(event_id) AS "total_events"
  FROM event_attendance
  GROUP BY activist_id
) AS total
  ON firstevent.activist_id = total.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id) AS "total_events_30_days"
  FROM event_attendance ea JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
  GROUP BY activist_id
) AS total30
  ON firstevent.activist_id = total30.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*2 AS "protest_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    and e.event_type = "protest"
  GROUP BY activist_id
) AS protest
  ON firstevent.activist_id = protest.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id) AS "wg_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "working group"
  GROUP BY activist_id
) AS wg
  ON firstevent.activist_id = wg.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id) AS "community_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
  AND e.event_type = "community"
  GROUP BY activist_id
) AS community
  ON firstevent.activist_id = community.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*2 AS "outreach_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "outreach"
  GROUP BY activist_id
) AS outreach
  ON firstevent.activist_id = outreach.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*2 AS "sanctuary_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "sanctuary"
  GROUP BY activist_id
) AS sanctuary
  ON firstevent.activist_id = sanctuary.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*3 AS "key_event_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "key event"
  GROUP BY activist_id
) AS key_event
  ON firstevent.activist_id = key_event.activist_id

WHERE
	total_events_30_days > 0
	AND a.exclude_from_leaderboard <> 1

ORDER BY points DESC`

	var leaderboardUsers []LeaderboardUser
	if err := db.Select(&leaderboardUsers, query); err != nil {
		return nil, err
	}

	return leaderboardUsers, nil
}

func GetPower(db *sqlx.DB) (int, error) {
	query := `
SELECT COUNT(*) AS movement_power_index
FROM (
	SELECT
		activist_id,
		MAX(CASE WHEN event_type = "protest" or event_type = "key event" THEN "1" ELSE "0" END) AS is_protest,
	    MAX(CASE WHEN event_type = "community" THEN "1" ELSE "0" END) AS is_community
	FROM event_attendance ea
	JOIN events e ON ea.event_id = e.id
	WHERE e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
	GROUP BY activist_id
	HAVING is_protest = "1" AND is_community = "1"
) AS power_index
`
	var power int
	if err := db.Get(&power, query); err != nil {
		return 0, err
	}
	return power, nil
}
