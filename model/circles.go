package model

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

/** Constant and Variable Definitions */

const (
	circle_group_db_value     = 1
	geo_circle_group_db_value = 2
)

var CircleGroupTypes map[int]string = map[int]string{
	circle_group_db_value:     "circle",
	geo_circle_group_db_value: "geo-circle",
}

var CircleGroupTypeStringToInt map[string]int

func init() {
	CircleGroupTypeStringToInt = make(map[string]int)
	for key := range CircleGroupTypes {
		CircleGroupTypeStringToInt[CircleGroupTypes[key]] = key
	}
}

/** User-defined Types */

type CircleGroup struct {
	ID              int    `db:"id"`
	Name            string `db:"name"`
	Type            int    `db:"type"`
	Members         []CircleGroupMember
	Visible         bool   `db:"visible"`
	Description     string `db:"description"`
	MeetingTime     string `db:"meeting_time"`
	MeetingLocation string `db:"meeting_location"`
	Coords          string `db:"coords"`
	LastMeeting     string `db:"last_meeting"`
}

type CircleGroupQueryOptions struct {
	GroupID    int
	CircleType int
	PublicAPI  bool
}

type CircleGroupMember struct {
	ActivistName           string  `db:"activist_name"`
	ActivistID             int     `db:"activist_id"`
	ActivistEmail          string  `db:"activist_email"`
	Lat                    float64 `db:"lat"`
	Lng                    float64 `db:"lng"`
	PointPerson            bool    `db:"point_person"`
	NonMemberOnMailingList bool    `db:"non_member_on_mailing_list"`
}

type CircleGroupJSON struct {
	ID              int                     `json:"id"`
	Name            string                  `json:"name"`
	Type            string                  `json:"type"`
	Members         []CircleGroupMemberJSON `json:"members"`
	Visible         bool                    `json:"visible"`
	Description     string                  `json:"description"`
	MeetingTime     string                  `json:"meeting_time"`
	MeetingLocation string                  `json:"meeting_location"`
	Coords          string                  `json:"coords"`
	LastMeeting     string                  `json:"last_meeting"`
}

type CircleGroupMemberJSON struct {
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	PointPerson            bool   `json:"point_person"`
	NonMemberOnMailingList bool   `json:"non_member_on_mailing_list"`
}

/** Functions and Methods */

func CreateCircleGroup(db *sqlx.DB, circleGroup CircleGroup) (int, error) {
	if circleGroup.ID != 0 {
		return 0, errors.New("Cannot Create a Circle that already exists")
	}
	return createOrUpdateCircleGroup(db, circleGroup)
}

func UpdateCircleGroup(db *sqlx.DB, circleGroup CircleGroup) (int, error) {
	if circleGroup.ID == 0 {
		return 0, errors.New("Unable to update Circle if no Circle id is provided")
	}
	return createOrUpdateCircleGroup(db, circleGroup)
}

func createOrUpdateCircleGroup(db *sqlx.DB, circleGroup CircleGroup) (int, error) {
	// Check that required parameters are present
	if circleGroup.Name == "" {
		return 0, errors.New("Circle name must not be zero-value")
	}
	if circleGroup.Type != circle_group_db_value && circleGroup.Type != geo_circle_group_db_value {
		return 0, errors.New("Circle type must be circle or geo-circle")
	}

	var query string
	if circleGroup.ID == 0 {
		// Create Circle
		query = `
    INSERT INTO circles (name, type, visible, description, meeting_time, meeting_location, coords)
    VALUES (:name, :type, :visible, :description, :meeting_time, :meeting_location, :coords)
    `
	} else {
		// Update existing working group
		query = `
UPDATE circles
SET
  name = :name,
  type = :type,
  visible = :visible,
  description = :description,
  meeting_time = :meeting_time,
  meeting_location = :meeting_location,
  coords = :coords
WHERE
id = :id
`
	}
	tx, err := db.Beginx()
	if err != nil {
		return 0, errors.Wrap(err, "Failed to Create Transaction")
	}
	res, err := tx.NamedExec(query, circleGroup)
	if err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "Failed to insert or update circle")
	}

	if circleGroup.ID == 0 {
		id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return 0, errors.Wrap(err, "Failed to get last inserted Circle ID.")
		}
		circleGroup.ID = int(id)
	}

	if err := insertCircleGroupMembers(tx, circleGroup); err != nil {
		tx.Rollback()
		return 0, errors.Wrapf(err, "Please check for duplicate members. Failed to insert members for Circle %s.", circleGroup.Name)
	}

	// if geo-circle, calculate lat & lng based on members
	if circleGroup.Type == geo_circle_group_db_value {
		var circleSize int
		var totLat, totLng float64
		for _, m := range circleGroup.Members {
			if m.Lat != 0 && m.Lng != 0 {
				// Just averaging will have weird edge cases if lat/lng are near the equator or prime merridian, but should be fine for the US.
				// TODO: It would probably also be good to remove noise if someone's coords are very far away from the others.
				totLat += m.Lat
				totLng += m.Lng
				circleSize++
			}
		}
		avgLat := totLat / float64(circleSize)
		avgLng := totLng / float64(circleSize)
		circleGroup.Coords = fmt.Sprintf("%.6f, %.6f", avgLat, avgLng)
		_, err := tx.NamedExec(`UPDATE circles SET coords = :coords WHERE id = :id`, circleGroup)
		if err != nil {
			tx.Rollback()
			return 0, errors.Wrap(err, "Failed to update geo-circle coords")
		}
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, errors.Wrapf(err, "Failed to commit Circle %s", circleGroup.Name)
	}

	return circleGroup.ID, nil
}

func insertCircleGroupMembers(tx *sqlx.Tx, circleGroup CircleGroup) error {
	if circleGroup.ID == 0 {
		return errors.New("Invalid Circle ID. ID's must be greater than 0")
	}
	// First drop all members.
	_, err := tx.Exec(`DELETE FROM circle_members WHERE circle_id = ?`, circleGroup.ID)
	if err != nil {
		return errors.Wrapf(err, "Failed to drop members for circle: %s", circleGroup.Name)
	}

	for _, m := range circleGroup.Members {
		if m.ActivistID < 1 {
			return errors.New("Invalid Activist ID; cannot add as a circle member")
		}
		_, err = tx.Exec(`INSERT INTO circle_members (circle_id, activist_id, point_person, non_member_on_mailing_list)
    VALUES (?, ?, ?, ?)`, circleGroup.ID, m.ActivistID, m.PointPerson, m.NonMemberOnMailingList)
		if err != nil {
			return errors.Wrapf(err, "Failed to insert %s into Circle %s", m.ActivistName, circleGroup.Name)
		}
	}
	return nil
}

func CleanCircleGroupData(db *sqlx.DB, body io.Reader) (CircleGroup, error) {
	var circleGroupJSON CircleGroupJSON
	err := json.NewDecoder(body).Decode(&circleGroupJSON)
	if err != nil {
		return CircleGroup{}, err
	}

	if len(strings.TrimSpace(circleGroupJSON.Name)) == 0 {
		return CircleGroup{}, errors.Errorf("Circle name must not be blank")
	}

	if circleGroupJSON.Type == "" {
		return CircleGroup{}, errors.New("Circle type can't be empty")
	}

	cirType, ok := CircleGroupTypeStringToInt[circleGroupJSON.Type]
	if !ok {
		return CircleGroup{}, errors.Errorf("Circle type doesn't exist: %s", circleGroupJSON.Type)
	}

	members := make([]CircleGroupMember, 0, len(circleGroupJSON.Members))
	for _, m := range circleGroupJSON.Members {
		trimName := strings.TrimSpace(m.Name)
		if trimName == "" {
			return CircleGroup{}, errors.New("Member name cannot be empty")
		}
		activist, err := GetActivist(db, strings.TrimSpace(m.Name))
		if err != nil {
			return CircleGroup{}, err
		}
		members = append(members, CircleGroupMember{
			ActivistName:           activist.Name,
			ActivistID:             activist.ID,
			ActivistEmail:          activist.Email,
			Lat:                    activist.Lat,
			Lng:                    activist.Lng,
			PointPerson:            m.PointPerson,
			NonMemberOnMailingList: m.NonMemberOnMailingList,
		})
	}

	return CircleGroup{
		ID:              circleGroupJSON.ID,
		Name:            strings.TrimSpace(circleGroupJSON.Name),
		Type:            cirType,
		Members:         members,
		Visible:         circleGroupJSON.Visible,
		Description:     circleGroupJSON.Description,
		MeetingTime:     circleGroupJSON.MeetingTime,
		MeetingLocation: circleGroupJSON.MeetingLocation,
		Coords:          circleGroupJSON.Coords,
	}, nil
}

func DeleteCircleGroup(db *sqlx.DB, circleGroupID int) error {
	if circleGroupID == 0 {
		return errors.New("Working group ID can't be 0")
	}

	// Wrap everything in a transaction because we only want to
	// delete the working group if there are no users associated
	// with it.
	tx, err := db.Beginx()
	if err != nil {
		return errors.Wrap(err, "Failed to create transaction")
	}

	txFn := func() error {
		var activistIDs []int
		err = tx.Select(&activistIDs, `
SELECT activist_id
FROM circle_members
WHERE circle_id = ?`, circleGroupID)
		if err != nil {
			return errors.Wrapf(err, "Failed to get activists for circle: %d", circleGroupID)
		}

		if len(activistIDs) > 0 {
			return errors.New("Cannot delete circle because it has members associated with it")
		}
		_, err = tx.Exec(`
DELETE FROM circles
WHERE id = ?`, circleGroupID)
		if err != nil {
			return errors.Wrap(err, "Could not delete circle")
		}
		return nil
	}

	if err = txFn(); err != nil {
		tx.Rollback()
		return err
	}
	if err = tx.Commit(); err != nil {
		tx.Rollback()
		return errors.Wrap(err, "Error during commit")
	}
	return nil
}

func GetCircleGroupJSON(db *sqlx.DB, circleGroupID int) (CircleGroupJSON, error) {

	cirs, err := getCircleGroupsJSON(db, CircleGroupQueryOptions{
		GroupID: circleGroupID,
	})
	if err != nil {
		return CircleGroupJSON{}, err
	}
	if len(cirs) == 0 {
		return CircleGroupJSON{}, errors.Errorf("Could not find circle with id: %d", circleGroupID)
	} else if len(cirs) > 1 {
		return CircleGroupJSON{}, errors.Errorf("Found too many circles with id: %d", circleGroupID)
	}
	return cirs[0], nil
}

func GetCircleGroupsJSON(db *sqlx.DB, circleType int, publicAPI bool) ([]CircleGroupJSON, error) {
	return getCircleGroupsJSON(db, CircleGroupQueryOptions{
		CircleType: circleType,
		PublicAPI:  publicAPI,
	})
}

func getCircleGroupsJSON(db *sqlx.DB, options CircleGroupQueryOptions) ([]CircleGroupJSON, error) {

	cirs, err := getCircleGroups(db, options)
	if err != nil {
		return nil, err
	}

	cirsJSON := make([]CircleGroupJSON, 0, len(cirs))
	for _, cir := range cirs {
		cirMembers := make([]CircleGroupMemberJSON, 0, len(cir.Members))
		for _, member := range cir.Members {
			// don't include member info in public API unless it's the point person
			if !options.PublicAPI || member.PointPerson {
				cirMembers = append(cirMembers, CircleGroupMemberJSON{
					Name:                   member.ActivistName,
					Email:                  member.ActivistEmail,
					PointPerson:            member.PointPerson,
					NonMemberOnMailingList: member.NonMemberOnMailingList,
				})
			}
		}
		cirsJSON = append(cirsJSON, CircleGroupJSON{
			ID:              cir.ID,
			Name:            cir.Name,
			Type:            CircleGroupTypes[cir.Type],
			Members:         cirMembers,
			Visible:         cir.Visible,
			Description:     cir.Description,
			MeetingTime:     cir.MeetingTime,
			MeetingLocation: cir.MeetingLocation,
			Coords:          cir.Coords,
			LastMeeting:     cir.LastMeeting,
		})
	}

	return cirsJSON, nil
}

func GetCircleGroups(db *sqlx.DB, options CircleGroupQueryOptions) ([]CircleGroup, error) {
	if options.GroupID != 0 {
		return nil, errors.New("GetCircleGroups: Cannot include an ID in options")
	}

	circleGroups, err := getCircleGroups(db, options)
	if err != nil {
		return nil, errors.Wrapf(err, "GetCircleGroups: Unable to retrieve circles")
	}
	return circleGroups, nil
}

func GetCircleGroup(db *sqlx.DB, options CircleGroupQueryOptions) (CircleGroup, error) {
	if options.GroupID == 0 {
		return CircleGroup{}, errors.New("GetCircleGroup: ID or Name required to fetch specific circle")
	}

	circleGroups, err := getCircleGroups(db, options)
	if err != nil {
		return CircleGroup{}, errors.Wrapf(err, "Error fetching circle with ID %d", options.GroupID)
	}
	if len(circleGroups) == 0 {
		return CircleGroup{}, errors.Wrapf(err, "No circle with ID %d found", options.GroupID)
	}
	if len(circleGroups) > 1 {
		return CircleGroup{}, errors.Wrapf(err, "Duplicate circle with ID %d", options.GroupID)
	}
	return circleGroups[0], nil
}

func getCircleGroups(db *sqlx.DB, options CircleGroupQueryOptions) ([]CircleGroup, error) {
	query := `
SELECT w.id, w.name, w.type, w.visible, w.description, w.meeting_time, w.meeting_location, w.coords,
@lastMeeting := IFNULL((SELECT max(date) from events where circle_id = w.id),"") as last_meeting
FROM circles w
`

	var queryArgs []interface{}
	var whereClause []string

	if options.GroupID != 0 {
		whereClause = append(whereClause, "w.id = ?")
		queryArgs = append(queryArgs, options.GroupID)
	}

	if options.CircleType != 0 {
		whereClause = append(whereClause, "w.type = ?")
		queryArgs = append(queryArgs, options.CircleType)
	}

	if options.PublicAPI { // public API should only list visible circles
		whereClause = append(whereClause, "w.visible = 1")
	}

	if len(whereClause) > 0 {
		query += ` WHERE ` + strings.Join(whereClause, " AND ")
	}

	query += ` ORDER BY w.name`

	var circleGroups []CircleGroup
	if err := db.Select(&circleGroups, query, queryArgs...); err != nil {
		return []CircleGroup{}, errors.Wrapf(err, "getCircleGroups: Failed retrieving working groups from circles table")
	}

	// TODO(mdempsky): Use a JOIN instead of a second round-trip.
	if err := fetchCircleGroupMembers(db, circleGroups); err != nil {
		return []CircleGroup{}, errors.Wrapf(err, "Failed to fetch working group members for query: %#v", options)
	}

	return circleGroups, nil

}

func fetchCircleGroupMembers(db *sqlx.DB, circleGroups []CircleGroup) error {
	if len(circleGroups) == 0 {
		return nil
	}

	circleGroupIDToIndex := map[int]int{}
	var circleGroupIDs []int

	for i, w := range circleGroups {
		circleGroupIDs = append(circleGroupIDs, w.ID)
		circleGroupIDToIndex[w.ID] = i
	}
	membersQuery, membersArgs, err := sqlx.In(`
SELECT
  wm.circle_id,
  a.name as activist_name,
  a.email as activist_email,
  a.id as activist_id,
  a.lat as lat,
  a.lng as lng,
  wm.point_person,
  wm.non_member_on_mailing_list
FROM activists a
JOIN circle_members wm
  on a.id = wm.activist_id
WHERE
  wm.circle_id IN (?)`, circleGroupIDs)
	if err != nil {
		return errors.Wrapf(err, "Could not create sqlx.In query for fetching circle members")
	}

	membersQuery = db.Rebind(membersQuery)
	var members []struct {
		GroupID int `db:"circle_id"`
		CircleGroupMember
	}
	if err := db.Select(&members, membersQuery, membersArgs...); err != nil {
		return errors.Wrapf(err, "Unable to fetch circle members")
	}

	for _, m := range members {
		idx := circleGroupIDToIndex[m.GroupID]
		circleGroups[idx].Members = append(circleGroups[idx].Members, m.CircleGroupMember)
	}

	return nil

}
