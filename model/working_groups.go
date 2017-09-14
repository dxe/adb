package model

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
)

/** Constant and Variable Definitions */

const (
	working_group_db_value = 1
	committee_db_value     = 2
)

/** User-defined Types */

type WorkingGroup struct {
	ID            int            `db:"id"`
	Name          string         `db:"name"`
	Type          int            `db:"type"`
	GroupEmail    sql.NullString `db:"group_email"`
	PointPersonID int            `db:"point_person_id"`
	Members       []WorkingGroupMembers
}

type WorkingGroupQueryOptions struct {
	GroupID   int
	GroupName string
}

type WorkingGroupMembers struct {
	WorkingGroupID int    `db:"working_group_id"`
	ActivistName   string `db:"activist_name"`
	ActivistID     string `db:"activist_id"`
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
    INSERT INTO working_groups (name, type, group_email, point_person_id)
    VALUES (:name, :type, :group_email, :point_person_id)
    `
	} else {
		// Update existing working group
		query = `
UPDATE working_groups 
SET
  name = :name,
  type = :type,
  group_email = :group_email
  point_person_id = :point_person_id
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
	for _, m := range workingGroup.Members {
		_, err := tx.NamedExec(`INSERT INTO working_group_members (working_group_id, activist_id)
    VALUES (:working_group_id, :activist_id)`, m)
		if err != nil {
			errors.Wrapf(err, "Failed to insert %s into Working Group %s", m.ActivistName, workingGroup.Name)
		}
	}
	return nil
}

func GetWorkingGroups(db *sqlx.DB, options WorkingGroupQueryOptions) ([]WorkingGroup, error) {
	if options.GroupID != 0 {
		return []WorkingGroup{}, errors.New("GetWorkingGroups: Cannot include an ID in options")
	}

	workingGroups, err := getWorkingGroups(db, options)
	if err != nil {
		return []WorkingGroup{}, errors.Wrapf(err, "GetWorkingGroups: Unable to retrieve working groups")
	}
	return workingGroups, nil
}

func GetWorkingGroup(db *sqlx.DB, options WorkingGroupQueryOptions) (WorkingGroup, error) {
	if options.GroupID == 0 {
		return WorkingGroup{}, errors.New("GetWorkingGroup: ID required to fetch specific working group")
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
SELECT w.id, w.name, w.type, w.group_email, w.point_person_id FROM working_groups w
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
  a.id as activist_id
FROM activists a
JOIN working_group_members wm
  on a.id = wm.activist_id
WHERE
  wm.working_group_id IN (?)`, workingGroupIDs)
	if err != nil {
		return errors.Wrapf(err, "Could not create sqlx.In query for fetching working group members")
	}

	membersQuery = db.Rebind(membersQuery)
	var members []WorkingGroupMembers
	if err := db.Select(&members, membersQuery, membersArgs...); err != nil {
		return errors.Wrapf(err, "Unable to fetch working group members")
	}

	for _, m := range members {
		idx := workingGroupIDToIndex[m.WorkingGroupID]
		workingGroups[idx].Members = append(workingGroups[idx].Members, m)
	}

	return nil

}
