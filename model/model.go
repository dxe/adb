package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"github.com/jmoiron/sqlx"
	"io"
	"strconv"
	"strings"
	"time"
)

var DangerousCharacters = "<>&"

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

var Duration60Days = 60 * 24 * time.Hour
var Duration90Days = 90 * 24 * time.Hour

func getStatus(firstEvent *time.Time, lastEvent *time.Time, totalEvents int) string {
	if firstEvent == nil || lastEvent == nil {
		return "No attendance"
	}

	if time.Since(*lastEvent) > Duration60Days {
		return "Former"
	}
	if time.Since(*firstEvent) > Duration90Days && totalEvents > 5 {
		return "New"
	}
	return "Current"
}

type User struct {
	ID       int            `db:"id"`
	Name     string         `db:"name"`
	Email    string         `db:"email"`
	Chapter  string         `db:"chapter"`
	Phone    string         `db:"phone"`
	Location sql.NullString `db:"location"`
	Facebook string         `db:"facebook"`
}

type UserEventData struct {
	FirstEvent  *time.Time `db:"first_event"`
	LastEvent   *time.Time `db:"last_event"`
	TotalEvents int        `db:"total_events"`
	Status      string
}

type UserExtra struct {
	User
	UserEventData
}

type UserJSON struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	Chapter     string `json:"chapter"`
	Phone       string `json:"phone"`
	Location    string `json:"location"`
	Facebook    string `json:"facebook"`
	FirstEvent  string `json:"first_event"`
	LastEvent   string `json:"last_event"`
	TotalEvents int    `json:"total_events"`
	Status      string `json:"status"`
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
	users, err := GetUsersExtra(db)
	if err != nil {
		return nil, err
	}
	for _, u := range users {
		firstEvent := ""
		if u.UserEventData.FirstEvent != nil {
			firstEvent = u.UserEventData.FirstEvent.Format(EventDateLayout)
		}
		lastEvent := ""
		if u.UserEventData.LastEvent != nil {
			lastEvent = u.UserEventData.LastEvent.Format(EventDateLayout)
		}

		usersJSON = append(usersJSON, UserJSON{
			ID:          u.User.ID,
			Name:        u.User.Name,
			Email:       u.User.Email,
			Chapter:     u.User.Chapter,
			Phone:       u.User.Phone,
			Location:    u.User.Location.String,
			Facebook:    u.User.Facebook,
			FirstEvent:  firstEvent,
			LastEvent:   lastEvent,
			TotalEvents: u.UserEventData.TotalEvents,
			Status:      u.Status,
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
  id,
  name,
  email,
  chapter,
  phone,
  location,
  facebook
FROM activists
`

	if name != "" {
		query += " WHERE name = ? "
		queryArgs = append(queryArgs, name)
	}

	query += " ORDER BY name "

	var users []User
	if err := db.Select(&users, query, queryArgs...); err != nil {
		return nil, err
	}

	return users, nil
}

func GetUsersExtra(db *sqlx.DB) ([]UserExtra, error) {
	query := `
SELECT
  a.id,
  a.name,
  email,
  chapter,
  phone,
  location,
  facebook,
  MIN(e.date) AS first_event,
  MAX(e.date) AS last_event,
  COUNT(e.id) as total_events
FROM activists a

LEFT JOIN event_attendance ea
  ON ea.activist_id = a.id

LEFT JOIN events e
  ON ea.event_id = e.id

GROUP BY a.id
`
	var users []UserExtra
	if err := db.Select(&users, query); err != nil {
		return nil, err
	}

	for i := 0; i < len(users); i++ {
		u := users[i]
		users[i].Status = getStatus(u.FirstEvent, u.LastEvent, u.TotalEvents)
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

	addedAttendees, err := cleanEventAttendanceData(db, eventJSON.AddedAttendees)
	if err != nil {
		return Event{}, err
	}

	deletedAttendees, err := cleanEventAttendanceData(db, eventJSON.DeletedAttendees)
	if err != nil {
		return Event{}, err
	}

	e.AddedAttendees = addedAttendees
	e.DeletedAttendees = deletedAttendees

	return e, nil
}

func cleanEventAttendanceData(db *sqlx.DB, attendees []string) ([]User, error) {
	users := make([]User, len(attendees))

	for idx, attendee := range attendees {
		if err := checkForDangerousChars(attendee); err != nil {
			return []User{}, err
		}
		user, err := GetOrCreateUser(db, strings.TrimSpace(attendee))
		if err != nil {
			return []User{}, err
		}
		users[idx] = user
	}

	return users, nil

}

func checkForDangerousChars(data string) error {
	if strings.ContainsAny(data, DangerousCharacters) {
		return errors.New("Event name cannot include <, >, or &.")
	}
	return nil
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
	leaderboardUsersJSON := []LeaderboardUserJSON{}
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
	    MAX(CASE WHEN event_type = "outreach" or event_type = "sanctuary" or event_type = "community" THEN "1" ELSE "0" END) AS is_community
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

func GetPowerMTD(db *sqlx.DB) (int, error) {
	current_time := time.Now().Local()
	current_time_string := current_time.Format("2006-01")
	split_date := strings.Split(current_time_string, "-")

	year := split_date[0]
	month := split_date[1]

	query := `
SELECT COUNT(*) AS movement_power_index
FROM (
	SELECT
		activist_id,
		MAX(CASE WHEN event_type = "protest" or event_type = "key event" THEN "1" ELSE "0" END) AS is_protest,
	    MAX(CASE WHEN event_type = "outreach" or event_type = "sanctuary" or event_type = "community" THEN "1" ELSE "0" END) AS is_community,
        SUBSTR(e.date,1,4) AS year,
        SUBSTR(e.date,6,2) AS month
	FROM event_attendance ea
	JOIN events e ON ea.event_id = e.id
	GROUP BY activist_id, year, month
	HAVING is_protest = "1" AND is_community = "1" AND month = "` + month + `" AND year = "` + year + `"
) AS power_index
`
	var powerMTD int
	if err := db.Get(&powerMTD, query); err != nil {
		return 0, err
	}
	return powerMTD, nil
}

type PowerHist struct {
	Month int
	Year  int
	Power int
}

func GetPowerHistArray(db *sqlx.DB) ([]PowerHist, error) {
	current_time := time.Now().Local()
	current_time_string := current_time.Format("2006-01")
	split_date := strings.Split(current_time_string, "-")

	year := split_date[0]
	month := split_date[1]

	year_int, error := strconv.Atoi(year)
	if error != nil {
		return nil, error
	}
	month_int, error := strconv.Atoi(month)
	if error != nil {
		return nil, error
	}

	for i := 0; i < 12; i++ {
		if month_int == 1 {
			month_int = 12
			year_int -= 1
		} else {
			month_int -= 1
		}
	}

	var history []PowerHist

	for i := 0; i < 12; i++ {
		power, error := GetPowerHist(db, month_int, year_int)
		if error != nil {
			return nil, error
		}
		history = append(history, PowerHist{
			Month: month_int,
			Year:  year_int,
			Power: power,
		})
		if month_int == 12 {
			month_int = 1
			year_int += 1
		} else {
			month_int += 1
		}
	}
	return history, nil
}

func GetPowerHist(db *sqlx.DB, month int, year int) (int, error) {
	month_string := strconv.Itoa(month)
	if month < 10 {
		month_string = "0" + month_string
	}
	year_string := strconv.Itoa(year)
	query := `
SELECT COUNT(*) AS movement_power_index
FROM (
	SELECT
		activist_id,
		MAX(CASE WHEN event_type = "protest" or event_type = "key event" THEN "1" ELSE "0" END) AS is_protest,
	    MAX(CASE WHEN event_type = "outreach" or event_type = "sanctuary" or event_type = "community" THEN "1" ELSE "0" END) AS is_community,
        SUBSTR(e.date,1,4) AS year,
        SUBSTR(e.date,6,2) AS month
	FROM event_attendance ea
	JOIN events e ON ea.event_id = e.id
	GROUP BY activist_id, year, month
	HAVING is_protest = "1" AND is_community = "1" AND month = "` + month_string + `" AND year = "` + year_string + `"
) AS power_index
`
	var power int
	if err := db.Get(&power, query); err != nil {
		return 0, err
	}
	return power, nil
}
