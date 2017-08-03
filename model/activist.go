package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
)

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
}

type UserEventData struct {
	FirstEvent  *time.Time `db:"first_event"`
	LastEvent   *time.Time `db:"last_event"`
	TotalEvents int        `db:"total_events"`
	Status      string
}

type UserMembershipData struct {
	CoreStaff              int `db:"core_staff"`
	ExcludeFromLeaderboard int `db:"exclude_from_leaderboard"`
	GlobalTeamMember       int `db:"global_team_member"`
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
}

/** Functions and Methods */

func GetUsersJSON(db *sqlx.DB) ([]UserJSON, error) {
	return getUsersJSON(db, 0)
}

func GetUserJSON(db *sqlx.DB, userID int) (UserJSON, error) {
	users, err := getUsersJSON(db, userID)
	if err != nil {
		return UserJSON{}, err
	}
	return users[0], nil
}

func getUsersJSON(db *sqlx.DB, userID int) ([]UserJSON, error) {
	var usersJSON []UserJSON
	users, err := GetUsersExtra(db, userID)
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
			ID:          u.User.ID,
			Name:        u.User.Name,
			Email:       u.User.Email,
			Chapter:     u.User.Chapter,
			Phone:       u.User.Phone,
			Location:    location,
			Facebook:    u.User.Facebook,
			FirstEvent:  firstEvent,
			LastEvent:   lastEvent,
			TotalEvents: u.UserEventData.TotalEvents,
			Status:      u.Status,
			Core:        u.CoreStaff,
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

func GetUsersExtra(db *sqlx.DB, userID int) ([]UserExtra, error) {
	query := `
SELECT
  a.id,
  a.name,
  email,
  chapter,
  phone,
  location,
  facebook,
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

	if userID != 0 {
		// retrieve specific user rather than all users
		query += " WHERE a.id = ? "
		queryArgs = append(queryArgs, userID)
	}
	query += " GROUP BY a.id "

	var users []UserExtra
	if err := db.Select(&users, query, queryArgs...); err != nil {
		return nil, err
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
		return UserEventData{}, err
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
	_, err = db.Exec("INSERT INTO activists (name) VALUES (?)", name)
	if err != nil {
		return User{}, err
	}

	return GetUser(db, name)
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
  exclude_from_leaderboard = :exclude_from_leaderboard,
  core_staff = :core_staff,
  global_team_member = :global_team_member,
  liberation_pledge = :liberation_pledge
WHERE
id = :id`, user)

	if err != nil {
		return 0, err
	}
	return user.ID, nil
}
