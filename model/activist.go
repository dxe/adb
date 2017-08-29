package model

import (
	"database/sql"
	"encoding/json"
	"io"
	"time"

	"strings"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
)

/** Constant and Variable Definitions */

const Duration60Days = 60 * 24 * time.Hour
const Duration90Days = 90 * 24 * time.Hour

const selectActivistBaseQuery string = `
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

type Activist struct {
	ID               int            `db:"id"`
	Name             string         `db:"name"`
	Email            string         `db:"email"`
	Chapter          string         `db:"chapter"`
	Phone            string         `db:"phone"`
	Location         sql.NullString `db:"location"`
	Facebook         string         `db:"facebook"`
	LiberationPledge int            `db:"liberation_pledge"`
	Hidden           bool           `db:"hidden"`
}

type ActivistEventData struct {
	FirstEvent  *time.Time `db:"first_event"`
	LastEvent   *time.Time `db:"last_event"`
	TotalEvents int        `db:"total_events"`
	Status      string
}

type ActivistMembershipData struct {
	CoreStaff              int    `db:"core_staff"`
	ExcludeFromLeaderboard int    `db:"exclude_from_leaderboard"`
	GlobalTeamMember       int    `db:"global_team_member"`
	ActivistLevel          string `db:"activist_level"`
}

type ActivistExtra struct {
	Activist
	ActivistEventData
	ActivistMembershipData
}

type ActivistJSON struct {
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

type GetActivistOptions struct {
	ID     int
	Hidden bool
}

/** Functions and Methods */

func GetActivistsJSON(db *sqlx.DB, options GetActivistOptions) ([]ActivistJSON, error) {
	if options.ID != 0 {
		return nil, errors.New("GetActivistsJSON: Cannot include ID in options")
	}
	return getActivistsJSON(db, options)
}

func GetActivistJSON(db *sqlx.DB, options GetActivistOptions) (ActivistJSON, error) {
	if options.ID == 0 {
		return ActivistJSON{}, errors.New("GetActivistJSON: Must include ID in options")
	}

	activists, err := getActivistsJSON(db, options)
	if err != nil {
		return ActivistJSON{}, err
	} else if len(activists) == 0 {
		return ActivistJSON{}, errors.New("Could not find any activists")
	} else if len(activists) > 1 {
		return ActivistJSON{}, errors.New("Found too many activists")
	}
	return activists[0], nil
}

func getActivistsJSON(db *sqlx.DB, options GetActivistOptions) ([]ActivistJSON, error) {
	var activistsJSON []ActivistJSON
	activists, err := GetActivistsExtra(db, options)
	if err != nil {
		return nil, err
	}
	for _, u := range activists {
		firstEvent := ""
		if u.ActivistEventData.FirstEvent != nil {
			firstEvent = u.ActivistEventData.FirstEvent.Format(EventDateLayout)
		}
		lastEvent := ""
		if u.ActivistEventData.LastEvent != nil {
			lastEvent = u.ActivistEventData.LastEvent.Format(EventDateLayout)
		}
		location := ""
		if u.Activist.Location.Valid {
			location = u.Activist.Location.String
		}

		activistsJSON = append(activistsJSON, ActivistJSON{
			ID:            u.Activist.ID,
			Name:          u.Activist.Name,
			Email:         u.Activist.Email,
			Chapter:       u.Activist.Chapter,
			Phone:         u.Activist.Phone,
			Location:      location,
			Facebook:      u.Activist.Facebook,
			ActivistLevel: u.ActivistLevel,
			FirstEvent:    firstEvent,
			LastEvent:     lastEvent,
			TotalEvents:   u.ActivistEventData.TotalEvents,
			Status:        u.Status,
			Core:          u.CoreStaff,
			ExcludeFromLeaderboard: u.ExcludeFromLeaderboard,
			LiberationPledge:       u.LiberationPledge,
			GlobalTeamMember:       u.GlobalTeamMember,
		})
	}
	return activistsJSON, nil
}

func GetActivist(db *sqlx.DB, name string) (Activist, error) {
	activists, err := getActivists(db, name)
	if err != nil {
		return Activist{}, err
	} else if len(activists) == 0 {
		return Activist{}, errors.New("Could not find any activists")
	} else if len(activists) > 1 {
		return Activist{}, errors.New("Found too many activists")
	}
	return activists[0], nil
}

func GetActivists(db *sqlx.DB) ([]Activist, error) {
	return getActivists(db, "")
}

func getActivists(db *sqlx.DB, name string) ([]Activist, error) {
	var queryArgs []interface{}
	query := selectActivistBaseQuery

	if name != "" {
		query += " WHERE name = ? "
		queryArgs = append(queryArgs, name)
	}

	query += " ORDER BY name "

	var activists []Activist
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists for %s", name)
	}

	return activists, nil
}

func GetActivistsExtra(db *sqlx.DB, options GetActivistOptions) ([]ActivistExtra, error) {
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
		// retrieve specific activist rather than all activists
		query += " WHERE a.id = ? "
		queryArgs = append(queryArgs, options.ID)
	} else {
		// Only check filter by hidden if the activistID isn't
		// supplied.
		if options.Hidden == true {
			query += " WHERE a.hidden = true "
		} else {
			query += " WHERE a.hidden = false "
		}
	}

	query += " GROUP BY a.id "

	var activists []ActivistExtra
	if err := db.Select(&activists, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get activists extra for uid %d", options.ID)
	}

	for i := 0; i < len(activists); i++ {
		u := activists[i]
		activists[i].Status = getStatus(u.FirstEvent, u.LastEvent, u.TotalEvents)
	}

	return activists, nil
}

func (u Activist) GetActivistEventData(db *sqlx.DB) (ActivistEventData, error) {
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
	var data ActivistEventData
	if err := db.Get(&data, query, u.ID); err != nil {
		return ActivistEventData{}, errors.Wrap(err, "failed to get activist event data")
	}
	return data, nil
}

func GetOrCreateActivist(db *sqlx.DB, name string) (Activist, error) {
	activist, err := GetActivist(db, name)
	if err == nil {
		// We got a valid activist, return them.
		return activist, nil
	}

	// There was an error, so try inserting the activist first.
	// Wrap in transaction to avoid issue where a new activist
	// is inserted successfully, but we are unable to retrieve
	// the new activist, which will leave database in inconsistent state

	tx, err := db.Beginx()
	if err != nil {
		return Activist{}, errors.Wrap(err, "Failed to create transaction")
	}

	_, err = tx.Exec("INSERT INTO activists (name) VALUES (?)", name)
	if err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to insert activist %s", name)
	}

	query := selectActivistBaseQuery + " WHERE name = ? "

	var newActivist Activist
	err = tx.Get(&newActivist, query, name)

	if err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to get new activist %s", name)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return Activist{}, errors.Wrapf(err, "failed to commit activist %s", name)
	}

	return newActivist, nil
}

func UpdateActivistData(db *sqlx.DB, activist ActivistExtra) (int, error) {
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
id = :id`, activist)

	if err != nil {
		return 0, errors.Wrap(err, "failed to update activist data")
	}
	return activist.ID, nil
}

func HideActivist(db *sqlx.DB, activistID int) error {
	if activistID == 0 {
		return errors.New("HideActivist: activistID cannot be 0")
	}
	var activistCount int
	err := db.Get(&activistCount, `SELECT count(*) FROM activists WHERE id = ?`, activistID)
	if err != nil {
		return errors.Wrap(err, "failed to get activist count")
	}
	if activistCount == 0 {
		return errors.Errorf("Activist with id %d does not exist", activistID)
	}

	_, err = db.Exec(`UPDATE activists SET hidden = true WHERE id = ?`, activistID)
	if err != nil {
		return errors.Wrapf(err, "failed to update activist %d", activistID)
	}
	return nil
}

// Merge activistID into targetActivistID.
//  - The original activist is hidden
//  - All of the original activist's event attendance is updated to be the target activist.
func MergeActivist(db *sqlx.DB, originalActivistID, targetActivistID int) error {
	if originalActivistID == 0 {
		return errors.New("originalActivistID cannot be 0")
	}
	if targetActivistID == 0 {
		return errors.New("targetActivistID cannot be 0")
	}
	if originalActivistID == targetActivistID {
		return errors.New("originalActivist and targetActivist cannot be the same")
	}

	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "could not create transaction")
	}

	_, err = tx.Exec(`UPDATE activists SET hidden = true WHERE id = ?`, originalActivistID)
	if err != nil {
		tx.Rollback()
		return errors.Wrapf(err, "failed to hide original activist %d", originalActivistID)
	}

	err = updateMergedActivistData(tx, originalActivistID, targetActivistID, true)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = updateMergedActivistData(tx, originalActivistID, targetActivistID, false)
	if err != nil {
		tx.Rollback()
		return err
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrapf(err,
			"failed to commit merge activist transaction. original activist id: %d, target activist id: %d",
			originalActivistID, targetActivistID)
	}

	return nil
}

func updateMergedActivistData(tx *sqlx.Tx, originalActivistID int, targetActivistID int, originalActivistOnly bool) error {
	baseQuery := `
SELECT event_id
FROM event_attendance ea
WHERE
  activist_id = ?
  AND `

	subquery := `
EXISTS(
  SELECT ea2.event_id
  FROM event_attendance ea2
  WHERE ea2.activist_id = ?
    AND ea2.event_id = ea.event_id)`

	var eventQuery string
	if originalActivistOnly {
		eventQuery = baseQuery + " NOT " + subquery
	} else {
		eventQuery = baseQuery + subquery
	}

	var eventIDs []int
	err := tx.Select(&eventIDs, eventQuery, originalActivistID, targetActivistID)
	if err != nil {
		return errors.Wrapf(err,
			"failed to get original activist's events: %d, originalActivistOnly: %v",
			originalActivistID, originalActivistOnly)
	}

	// There's nothing to do if there are no events.
	if len(eventIDs) == 0 {
		return nil
	}

	var eaQuery string
	var eaArgs []interface{}
	if originalActivistOnly {
		eaQuery, eaArgs, err = sqlx.In(`
UPDATE event_attendance
SET activist_id = ?
WHERE
  activist_id = ?
  AND event_id IN (?)`,
			targetActivistID, originalActivistID, eventIDs)
	} else {
		eaQuery, eaArgs, err = sqlx.In(`
DELETE FROM event_attendance
WHERE
  activist_id = ?
  AND event_id IN (?)`, originalActivistID, eventIDs)
	}
	if err != nil {
		return errors.Wrapf(err, "could not create sqlx.IN query. originalActivistOnly: %v",
			originalActivistOnly)
	}
	eaQuery = tx.Rebind(eaQuery)
	_, err = tx.Exec(eaQuery, eaArgs...)
	if err != nil {
		return errors.Wrapf(err, "could not update event attendance for activist: %d",
			originalActivistID)
	}
	err = insertMergedActivistAttendance(tx, originalActivistID, targetActivistID, eventIDs, originalActivistOnly)
	if err != nil {
		return err
	}
	return nil
}

func insertMergedActivistAttendance(tx *sqlx.Tx, originalActivistID int, targetActivistID int, eventIDs []int, replacedWithTargetActivist bool) error {
	if len(eventIDs) == 0 {
		return nil
	}

	query := `INSERT INTO merged_activist_attendance (original_activist_id, target_activist_id, event_id, replaced_with_target_activist) VALUES`

	var queryValues []string
	var queryArgs []interface{}
	for _, eventID := range eventIDs {
		queryValues = append(queryValues, " (?, ?, ?, ?) ")
		queryArgs = append(queryArgs, originalActivistID, targetActivistID, eventID, replacedWithTargetActivist)
	}
	query += strings.Join(queryValues, ",")
	_, err := tx.Exec(query, queryArgs...)

	return errors.Wrapf(err, "could not insert merged_activist_attendance for originalActivistID: %d, targetActivistID: %d",
		originalActivistID, targetActivistID)
}

func GetAutocompleteNames(db *sqlx.DB) []string {
	type Name struct {
		Name string `db:"name"`
	}
	names := []Name{}
	err := db.Select(&names, "SELECT name FROM activists WHERE hidden = 0 ORDER BY name ASC")
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

func CleanActivistData(body io.Reader) (ActivistExtra, error) {
	var activistJSON ActivistJSON
	err := json.NewDecoder(body).Decode(&activistJSON)
	if err != nil {
		return ActivistExtra{}, err
	}

	// Check if name field contains dangerous input
	if err := checkForDangerousChars(activistJSON.Name); err != nil {
		return ActivistExtra{}, err
	}

	valid := true
	if activistJSON.Location == "" {
		// No location specified so insert null value into database
		valid = false
	}

	activistExtra := ActivistExtra{
		Activist: Activist{
			ID:               activistJSON.ID,
			Name:             activistJSON.Name,
			Email:            activistJSON.Email,
			Chapter:          activistJSON.Chapter,
			Phone:            activistJSON.Phone,
			Location:         sql.NullString{String: activistJSON.Location, Valid: valid},
			Facebook:         activistJSON.Facebook,
			LiberationPledge: activistJSON.LiberationPledge,
		},
		ActivistMembershipData: ActivistMembershipData{
			CoreStaff:              activistJSON.Core,
			ExcludeFromLeaderboard: activistJSON.ExcludeFromLeaderboard,
			GlobalTeamMember:       activistJSON.GlobalTeamMember,
			ActivistLevel:          activistJSON.ActivistLevel,
		},
	}

	return activistExtra, nil

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
