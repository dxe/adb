package model

import (
	"encoding/json"
	"io"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

/** Constant and Variable Definitions */

const (
	working_group_db_value = 1
	committee_db_value     = 2
)

var WorkingGroupTypes map[int]string = map[int]string{
	working_group_db_value: "working_group",
	committee_db_value:     "committee",
}

var WorkingGroupTypeStringToInt map[string]int

func init() {
	WorkingGroupTypeStringToInt = make(map[string]int)
	for key := range WorkingGroupTypes {
		WorkingGroupTypeStringToInt[WorkingGroupTypes[key]] = key
	}
}

/** User-defined Types */

type WorkingGroup struct {
	ID              int    `db:"id"`
	Name            string `db:"name"`
	Type            int    `db:"type"`
	GroupEmail      string `db:"group_email"`
	Members         []WorkingGroupMember
	Visible         bool   `db:"visible"`
	Description     string `db:"description"`
	MeetingTime     string `db:"meeting_time"`
	MeetingLocation string `db:"meeting_location"`
	Coords          string `db:"coords"`
}

type WorkingGroupQueryOptions struct {
	GroupID   int
	GroupName string
}

type WorkingGroupMember struct {
	ActivistName           string `db:"activist_name"`
	ActivistID             int    `db:"activist_id"`
	ActivistEmail          string `db:"activist_email"`
	PointPerson            bool   `db:"point_person"`
	NonMemberOnMailingList bool   `db:"non_member_on_mailing_list"`
}

type WorkingGroupJSON struct {
	ID              int                      `json:"id"`
	Name            string                   `json:"name"`
	Type            string                   `json:"type"`
	Email           string                   `json:"email"`
	Members         []WorkingGroupMemberJSON `json:"members"`
	Visible         bool                     `json:"visible"`
	Description     string                   `json:"description"`
	MeetingTime     string                   `json:"meeting_time"`
	MeetingLocation string                   `json:"meeting_location"`
	Coords          string                   `json:"coords"`
}

type WorkingGroupMemberJSON struct {
	Name                   string `json:"name"`
	Email                  string `json:"email"`
	PointPerson            bool   `json:"point_person"`
	NonMemberOnMailingList bool   `json:"non_member_on_mailing_list"`
}

/** Functions and Methods */

func CreateWorkingGroup(db *sqlx.DB, workingGroup WorkingGroup) (int, error) {
	if workingGroup.ID != 0 {
		return 0, errors.New("Cannot Create a working group that already exists")
	}
	return createOrUpdateWorkingGroup(db, workingGroup)
}

func UpdateWorkingGroup(db *sqlx.DB, workingGroup WorkingGroup) (int, error) {
	if workingGroup.ID == 0 {
		return 0, errors.New("Unable to update working group if no working group id is provided")
	}
	return createOrUpdateWorkingGroup(db, workingGroup)
}

func createOrUpdateWorkingGroup(db *sqlx.DB, workingGroup WorkingGroup) (int, error) {
	// Check that required parameters are present
	if workingGroup.Name == "" {
		return 0, errors.New("WorkingGroup name for CreateWorkingGroup must not be zero-value")
	}
	if workingGroup.Type != working_group_db_value && workingGroup.Type != committee_db_value {
		return 0, errors.New("WorkingGroup type has to either be working group or committee")
	}

	var query string
	if workingGroup.ID == 0 {
		// Create working Group
		query = `
    INSERT INTO working_groups (name, type, group_email, visible, description, meeting_time, meeting_location, coords)
    VALUES (:name, :type, :group_email, :visible, :description, :meeting_time, :meeting_location, :coords)
    `
	} else {
		// Update existing working group
		query = `
UPDATE working_groups
SET
  name = :name,
  type = :type,
  group_email = :group_email,
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
	res, err := tx.NamedExec(query, workingGroup)
	if err != nil {
		tx.Rollback()
		return 0, errors.Wrap(err, "Failed to insert new working group")
	}

	if workingGroup.ID == 0 {
		id, err := res.LastInsertId()
		if err != nil {
			tx.Rollback()
			return 0, errors.Wrap(err, "Failed to get last inserted WorkingGroup ID")
		}
		workingGroup.ID = int(id)
	}

	if err := insertWorkingGroupMembers(tx, workingGroup); err != nil {
		tx.Rollback()
		return 0, errors.Wrapf(err, "Failed to insert members for working group %s", workingGroup.Name)
	}
	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, errors.Wrapf(err, "Failed to commit working group %s", workingGroup.Name)
	}
	return workingGroup.ID, nil
}

func insertWorkingGroupMembers(tx *sqlx.Tx, workingGroup WorkingGroup) error {
	if workingGroup.ID == 0 {
		return errors.New("Invalid WorkingGroup ID. ID's must be greater than 0")
	}
	// First drop all working group members for the working group.
	_, err := tx.Exec(`DELETE FROM working_group_members WHERE working_group_id = ?`, workingGroup.ID)
	if err != nil {
		return errors.Wrapf(err, "Failed to drop working groups for Working Group: %s", workingGroup.Name)
	}

	for _, m := range workingGroup.Members {
		if m.ActivistID < 1 {
			return errors.New("Invalid Activist ID; cannot add as a working group member")
		}
		_, err = tx.Exec(`INSERT INTO working_group_members (working_group_id, activist_id, point_person, non_member_on_mailing_list)
    VALUES (?, ?, ?, ?)`, workingGroup.ID, m.ActivistID, m.PointPerson, m.NonMemberOnMailingList)
		if err != nil {
			return errors.Wrapf(err, "Failed to insert %s into Working Group %s", m.ActivistName, workingGroup.Name)
		}
	}
	return nil
}

func CleanWorkingGroupData(db *sqlx.DB, body io.Reader) (WorkingGroup, error) {
	var workingGroupJSON WorkingGroupJSON
	err := json.NewDecoder(body).Decode(&workingGroupJSON)
	if err != nil {
		return WorkingGroup{}, err
	}

	if len(strings.TrimSpace(workingGroupJSON.Name)) == 0 {
		return WorkingGroup{}, errors.Errorf("Working group name must not be blank")
	}

	if !strings.Contains(workingGroupJSON.Email, "@") {
		return WorkingGroup{}, errors.Errorf("Working group email must contain @: %s", workingGroupJSON.Email)
	}

	if workingGroupJSON.Type == "" {
		return WorkingGroup{}, errors.New("Working group type can't be empty")
	}

	wgType, ok := WorkingGroupTypeStringToInt[workingGroupJSON.Type]
	if !ok {
		return WorkingGroup{}, errors.Errorf("Working group type doesn't exist: %s", workingGroupJSON.Type)
	}

	members := make([]WorkingGroupMember, 0, len(workingGroupJSON.Members))
	for _, m := range workingGroupJSON.Members {
		trimName := strings.TrimSpace(m.Name)
		if trimName == "" {
			return WorkingGroup{}, errors.New("Member name cannot be empty")
		}
		activist, err := GetActivist(db, strings.TrimSpace(m.Name))
		if err != nil {
			return WorkingGroup{}, err
		}
		members = append(members, WorkingGroupMember{
			ActivistName:           activist.Name,
			ActivistID:             activist.ID,
			ActivistEmail:          activist.Email,
			PointPerson:            m.PointPerson,
			NonMemberOnMailingList: m.NonMemberOnMailingList,
		})
	}

	return WorkingGroup{
		ID:              workingGroupJSON.ID,
		Name:            strings.TrimSpace(workingGroupJSON.Name),
		Type:            wgType,
		GroupEmail:      strings.TrimSpace(workingGroupJSON.Email),
		Members:         members,
		Visible:         workingGroupJSON.Visible,
		Description:     workingGroupJSON.Description,
		MeetingTime:     workingGroupJSON.MeetingTime,
		MeetingLocation: workingGroupJSON.MeetingLocation,
		Coords:          workingGroupJSON.Coords,
	}, nil
}

func DeleteWorkingGroup(db *sqlx.DB, workingGroupID int) error {
	if workingGroupID == 0 {
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
FROM working_group_members
WHERE working_group_id = ?`, workingGroupID)
		if err != nil {
			return errors.Wrapf(err, "Failed to get activists for working group: %d", workingGroupID)
		}

		if len(activistIDs) > 0 {
			return errors.New("Cannot delete working group because it has members associated with it")
		}
		_, err = tx.Exec(`
DELETE FROM working_groups
WHERE id = ?`, workingGroupID)
		if err != nil {
			return errors.Wrap(err, "Could not delete working group")
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

func GetWorkingGroupJSON(db *sqlx.DB, workingGroupID int) (WorkingGroupJSON, error) {
	wgs, err := getWorkingGroupsJSON(db, WorkingGroupQueryOptions{
		GroupID: workingGroupID,
	})
	if err != nil {
		return WorkingGroupJSON{}, err
	}
	if len(wgs) == 0 {
		return WorkingGroupJSON{}, errors.Errorf("Could not find working group with id: %d", workingGroupID)
	} else if len(wgs) > 1 {
		return WorkingGroupJSON{}, errors.Errorf("Found too many working groups with id: %d", workingGroupID)
	}
	return wgs[0], nil
}

func GetWorkingGroupsJSON(db *sqlx.DB, options WorkingGroupQueryOptions) ([]WorkingGroupJSON, error) {
	if options.GroupID != 0 {
		return nil, errors.New("Cannot include an ID in options")
	}
	if options.GroupName != "" {
		errorMsg := "Cannot include name in query options when fetching multiple working groups"
		return nil, errors.New(errorMsg)
	}

	return getWorkingGroupsJSON(db, options)
}

func getWorkingGroupsJSON(db *sqlx.DB, options WorkingGroupQueryOptions) ([]WorkingGroupJSON, error) {
	wgs, err := getWorkingGroups(db, options)
	if err != nil {
		return nil, err
	}

	wgsJSON := make([]WorkingGroupJSON, 0, len(wgs))
	for _, wg := range wgs {
		wgMembers := make([]WorkingGroupMemberJSON, 0, len(wg.Members))
		for _, member := range wg.Members {
			wgMembers = append(wgMembers, WorkingGroupMemberJSON{
				Name:                   member.ActivistName,
				Email:                  member.ActivistEmail,
				PointPerson:            member.PointPerson,
				NonMemberOnMailingList: member.NonMemberOnMailingList,
			})
		}
		wgsJSON = append(wgsJSON, WorkingGroupJSON{
			ID:              wg.ID,
			Name:            wg.Name,
			Type:            WorkingGroupTypes[wg.Type],
			Email:           wg.GroupEmail,
			Members:         wgMembers,
			Visible:         wg.Visible,
			Description:     wg.Description,
			MeetingTime:     wg.MeetingTime,
			MeetingLocation: wg.MeetingLocation,
			Coords:          wg.Coords,
		})
	}

	return wgsJSON, nil
}

func GetWorkingGroups(db *sqlx.DB, options WorkingGroupQueryOptions) ([]WorkingGroup, error) {
	if options.GroupID != 0 {
		return nil, errors.New("GetWorkingGroups: Cannot include an ID in options")
	}
	if options.GroupName != "" {
		errorMsg := "GetWorkingGroups: Cannot include name in query options when fetching multiple working groups"
		return nil, errors.New(errorMsg)
	}

	workingGroups, err := getWorkingGroups(db, options)
	if err != nil {
		return nil, errors.Wrapf(err, "GetWorkingGroups: Unable to retrieve working groups")
	}
	return workingGroups, nil
}

func GetWorkingGroup(db *sqlx.DB, options WorkingGroupQueryOptions) (WorkingGroup, error) {
	if options.GroupID == 0 && options.GroupName == "" {
		return WorkingGroup{}, errors.New("GetWorkingGroup: ID or Name required to fetch specific working group")
	}

	workingGroups, err := getWorkingGroups(db, options)
	if err != nil {
		return WorkingGroup{}, errors.Wrapf(err, "Error fetching working group with ID %d", options.GroupID)
	}
	if len(workingGroups) == 0 {
		return WorkingGroup{}, errors.Wrapf(err, "No working group with ID %d found", options.GroupID)
	}
	if len(workingGroups) > 1 {
		return WorkingGroup{}, errors.Wrapf(err, "Duplicate Working Groups with ID %d", options.GroupID)
	}
	return workingGroups[0], nil
}

func getWorkingGroups(db *sqlx.DB, options WorkingGroupQueryOptions) ([]WorkingGroup, error) {
	query := `
SELECT w.id, w.name, w.type, lower(w.group_email) as group_email, w.visible, w.description, w.meeting_time, w.meeting_location, w.coords FROM working_groups w
`

	var queryArgs []interface{}
	var whereClause []string

	if options.GroupID != 0 {
		whereClause = append(whereClause, "w.id = ?")
		queryArgs = append(queryArgs, options.GroupID)
	}

	if options.GroupName != "" {
		whereClause = append(whereClause, "w.name = ?")
		queryArgs = append(queryArgs, options.GroupName)
	}

	if len(whereClause) > 0 {
		query += ` WHERE ` + strings.Join(whereClause, " AND ")
	}

	query += ` ORDER BY w.name`

	var workingGroups []WorkingGroup
	if err := db.Select(&workingGroups, query, queryArgs...); err != nil {
		return []WorkingGroup{}, errors.Wrapf(err, "getWorkingGroups: Failed retrieving working groups from WorkingGroups table")
	}

	if err := fetchWorkingGroupMembers(db, workingGroups); err != nil {
		return []WorkingGroup{}, errors.Wrapf(err, "Failed to fetch working group members for group %s", options.GroupName)
	}

	return workingGroups, nil

}

func fetchWorkingGroupMembers(db *sqlx.DB, workingGroups []WorkingGroup) error {
	if len(workingGroups) == 0 {
		return nil
	}

	workingGroupIDToIndex := map[int]int{}
	var workingGroupIDs []int

	for i, w := range workingGroups {
		workingGroupIDs = append(workingGroupIDs, w.ID)
		workingGroupIDToIndex[w.ID] = i
	}
	membersQuery, membersArgs, err := sqlx.In(`
SELECT
  wm.working_group_id,
  a.name as activist_name,
  a.email as activist_email,
  a.id as activist_id,
  wm.point_person,
  wm.non_member_on_mailing_list
FROM activists a
JOIN working_group_members wm
  on a.id = wm.activist_id
WHERE
  wm.working_group_id IN (?)`, workingGroupIDs)
	if err != nil {
		return errors.Wrapf(err, "Could not create sqlx.In query for fetching working group members")
	}

	membersQuery = db.Rebind(membersQuery)
	var members []struct {
		GroupID int `db:"working_group_id"`
		WorkingGroupMember
	}
	if err := db.Select(&members, membersQuery, membersArgs...); err != nil {
		return errors.Wrapf(err, "Unable to fetch working group members")
	}

	for _, m := range members {
		idx := workingGroupIDToIndex[m.GroupID]
		workingGroups[idx].Members = append(workingGroups[idx].Members, m.WorkingGroupMember)
	}

	return nil

}
