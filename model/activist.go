package model

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
)

/** Constant and Variable Definitions */

const Duration60Days = 60 * 24 * time.Hour
const Duration90Days = 90 * 24 * time.Hour

const selectUserBaseQuery string = `
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

/** Type Definitions */

type User struct {
	ID               int            `db:"id"`
	Name             string         `db:"name"`
	Email            string         `db:"email"`
	Chapter          string         `db:"chapter"`
	Phone            string         `db:"phone"`
	Location         sql.NullString `db:"location"`
	Facebook         string         `db:"facebook"`
	LiberationPledge int            `db:"liberation_pledge"`
	Suspended        bool           `db:"suspended"`
}

type UserEventData struct {
	FirstEvent  *time.Time `db:"first_event"`
	LastEvent   *time.Time `db:"last_event"`
	TotalEvents int        `db:"total_events"`
	Status      string
}

type UserMembershipData struct {
	CoreStaff              int    `db:"core_staff"`
	ExcludeFromLeaderboard int    `db:"exclude_from_leaderboard"`
	GlobalTeamMember       int    `db:"global_team_member"`
	ActivistLevel          string `db:"activist_level"`
}

type UserExtra struct {
	User
	UserEventData
	UserMembershipData
}

type UserJSON struct {
	ID                     int    `json:"id"`
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	Chapter                string `json:"chapter"`
	Phone                  string `json:"phone"`
	Location               string `json:"location"`
	Facebook               string `json:"facebook"`
	FirstEvent             string `json:"first_event"`
	LastEvent              string `json:"last_event"`
	TotalEvents            int    `json:"total_events"`
	Status                 string `json:"status"`
	Core                   int    `json:"core_staff"`
	ExcludeFromLeaderboard int    `json:"exclude_from_leaderboard"`
	LiberationPledge       int    `json:"liberation_pledge"`
	GlobalTeamMember       int    `json:"global_team_member"`
	ActivistLevel          string `json:"activist_level"`
}

type GetUserOptions struct {
	ID        int
	Suspended bool
}

/** Functions and Methods */

func GetUsersJSON(db *sqlx.DB, options GetUserOptions) ([]UserJSON, error) {
	if options.ID != 0 {
		return nil, errors.New("GetUsersJSON: Cannot include ID in options")
	}
	return getUsersJSON(db, options)
}

func GetUserJSON(db *sqlx.DB, options GetUserOptions) (UserJSON, error) {
	if options.ID == 0 {
		return UserJSON{}, errors.New("GetUserJSON: Must include ID in options")
	}

	users, err := getUsersJSON(db, options)
	if err != nil {
		return UserJSON{}, err
	} else if len(users) == 0 {
		return UserJSON{}, errors.New("Could not find any users")
	} else if len(users) > 1 {
		return UserJSON{}, errors.New("Found too many users")
	}
	return users[0], nil
}

func getUsersJSON(db *sqlx.DB, options GetUserOptions) ([]UserJSON, error) {
	var usersJSON []UserJSON
	users, err := GetUsersExtra(db, options)
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
		location := ""
		if u.User.Location.Valid {
			location = u.User.Location.String
		}

		usersJSON = append(usersJSON, UserJSON{
			ID:            u.User.ID,
			Name:          u.User.Name,
			Email:         u.User.Email,
			Chapter:       u.User.Chapter,
			Phone:         u.User.Phone,
			Location:      location,
			Facebook:      u.User.Facebook,
			ActivistLevel: u.ActivistLevel,
			FirstEvent:    firstEvent,
			LastEvent:     lastEvent,
			TotalEvents:   u.UserEventData.TotalEvents,
			Status:        u.Status,
			Core:          u.CoreStaff,
			ExcludeFromLeaderboard: u.ExcludeFromLeaderboard,
			LiberationPledge:       u.LiberationPledge,
			GlobalTeamMember:       u.GlobalTeamMember,
		})
	}
	return usersJSON, nil
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

func GetUsers(db *sqlx.DB) ([]User, error) {
	return getUsers(db, "")
}

func getUsers(db *sqlx.DB, name string) ([]User, error) {
	var queryArgs []interface{}
	query := selectUserBaseQuery

	if name != "" {
		query += " WHERE name = ? "
		queryArgs = append(queryArgs, name)
	}

	query += " ORDER BY name "

	var users []User
	if err := db.Select(&users, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get users for %s", name)
	}

	return users, nil
}

func GetUsersExtra(db *sqlx.DB, options GetUserOptions) ([]UserExtra, error) {
	query := `
SELECT
  a.id,
  a.name,
  email,
  chapter,
  phone,
  location,
  facebook,
  activist_level,
  exclude_from_leaderboard,
  core_staff,
  global_team_member,
  liberation_pledge,
  MIN(e.date) AS first_event,
  MAX(e.date) AS last_event,
  COUNT(e.id) as total_events
FROM activists a

LEFT JOIN event_attendance ea
  ON ea.activist_id = a.id

LEFT JOIN events e
  ON ea.event_id = e.id
`
	var queryArgs []interface{}

	if options.ID != 0 {
		// retrieve specific user rather than all users
		query += " WHERE a.id = ? "
		queryArgs = append(queryArgs, options.ID)
	} else {
		// Only check filter by suspended if the userID isn't
		// supplied.
		if options.Suspended == true {
			query += " WHERE a.suspended = true "
		} else {
			query += " WHERE a.suspended = false "
		}
	}

	query += " GROUP BY a.id "

	var users []UserExtra
	if err := db.Select(&users, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get users extra for uid %d", options.ID)
	}

	for i := 0; i < len(users); i++ {
		u := users[i]
		users[i].Status = getStatus(u.FirstEvent, u.LastEvent, u.TotalEvents)
	}

	return users, nil
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
		return UserEventData{}, errors.Wrap(err, "failed to get user event data")
	}
	return data, nil
}

func GetOrCreateUser(db *sqlx.DB, name string) (User, error) {
	user, err := GetUser(db, name)
	if err == nil {
		// We got a valid user, return them.
		return user, nil
	}

	// There was an error, so try inserting the user first.
	// Wrap in transaction to avoid issue where a new user
	// is inserted successfully, but we are unable to retrieve
	// the new user, which will leave database in inconsistent state

	tx, err := db.Beginx()
	if err != nil {
		return User{}, errors.Wrap(err, "Failed to create transaction")
	}

	_, err = tx.Exec("INSERT INTO activists (name) VALUES (?)", name)
	if err != nil {
		tx.Rollback()
		return User{}, errors.Wrapf(err, "failed to insert user %s", name)
	}

	query := selectUserBaseQuery + " WHERE name = ? "

	var newUser User
	err = tx.Get(&newUser, query, name)

	if err != nil {
		tx.Rollback()
		return User{}, errors.Wrapf(err, "failed to get new user %s", name)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return User{}, errors.Wrapf(err, "failed to commit user %s", name)
	}

	return newUser, nil
}

func UpdateActivistData(db *sqlx.DB, user UserExtra) (int, error) {
	_, err := db.NamedExec(`UPDATE activists
SET
  name = :name,
  email = :email,
  chapter = :chapter,
  phone = :phone,
  location = :location,
  facebook = :facebook,
  activist_level = :activist_level,
  exclude_from_leaderboard = :exclude_from_leaderboard,
  core_staff = :core_staff,
  global_team_member = :global_team_member,
  liberation_pledge = :liberation_pledge
WHERE
id = :id`, user)

	if err != nil {
		return 0, errors.Wrap(err, "failed to update activist data")
	}
	return user.ID, nil
}

func SuspendUser(db *sqlx.DB, userID int) error {
	if userID == 0 {
		return errors.New("SuspendUser: userID cannot be 0")
	}
	var userCount int
	err := db.Get(&userCount, `SELECT count(*) FROM activists WHERE id = ?`, userID)
	if err != nil {
		return errors.Wrap(err, "failed to get user count")
	}
	if userCount == 0 {
		return errors.Errorf("User with id %d does not exist", userID)
	}

	_, err = db.Exec(`UPDATE activists SET suspended = true WHERE id = ?`, userID)
	if err != nil {
		return errors.Wrapf(err, "failed to update activist %d", userID)
	}
	return nil
}

func GetAutocompleteNames(db *sqlx.DB) []string {
	type Name struct {
		Name string `db:"name"`
	}
	names := []Name{}
	err := db.Select(&names, "SELECT name FROM activists WHERE suspended = 0 ORDER BY name ASC")
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

func CleanActivistData(body io.Reader) (UserExtra, error) {
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
			ActivistLevel:          userJSON.ActivistLevel,
		},
	}

	return userExtra, nil

}

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
