package model

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
)

/** Constant and Global Variable Definitions */

const DangerousCharacters = "<>&"
const Duration60Days = 60 * 24 * time.Hour
const Duration90Days = 90 * 24 * time.Hour

/** Functions and Methods */

// Returns one of the following statuses:
//  - Current
//  - New
//  - Former
//  - No attendance
// Must be kept in sync with the list in frontend/ActivistList.vue
func getStatus(firstEvent *time.Time, lastEvent *time.Time, totalEvents int) string {
	if firstEvent == nil || lastEvent == nil {
		return "No attendance"
	}

	if time.Since(*lastEvent) > Duration60Days {
		return "Former"
	}
	if time.Since(*firstEvent) < Duration90Days && totalEvents < 5 {
		return "New"
	}
	return "Current"
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

func CleanActivistData(db *sqlx.DB, body io.Reader) (UserExtra, error) {
	var userJSON UserJSON
	err := json.NewDecoder(body).Decode(&userJSON)
	if err != nil {
		return UserExtra{}, err
	}

	// Check if name field contains dangerous input
	if err := checkForDangerousChars(userJSON.Name); err != nil {
		return UserExtra{}, err
	}

	valid := true
	if userJSON.Location == "" {
		// No location specified so insert null value into database
		valid = false
	}

	userExtra := UserExtra{
		User: User{
			ID:               userJSON.ID,
			Name:             userJSON.Name,
			Email:            userJSON.Email,
			Chapter:          userJSON.Chapter,
			Phone:            userJSON.Phone,
			Location:         sql.NullString{String: userJSON.Location, Valid: valid},
			Facebook:         userJSON.Facebook,
			LiberationPledge: userJSON.LiberationPledge,
		},
		UserMembershipData: UserMembershipData{
			CoreStaff:              userJSON.Core,
			ExcludeFromLeaderboard: userJSON.ExcludeFromLeaderboard,
			GlobalTeamMember:       userJSON.GlobalTeamMember,
		},
	}

	return userExtra, nil

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
		return errors.New("User input cannot include <, >, or &.")
	}
	return nil
}
