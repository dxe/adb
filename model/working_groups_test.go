package model

import (
	"database/sql"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

func TestCreateWorkingGroup_missingRequiredParameters_returnsError(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	workingGroup := WorkingGroup{
		Name: "foo",
	}
	_, err := CreateWorkingGroup(db, workingGroup)
	require.Error(t, err)

	workingGroup.Type = 3 // Valid values = 1 or 2
	_, err = CreateWorkingGroup(db, workingGroup)
	require.Error(t, err)

	workingGroup.Type = 1
	workingGroup.Name = ""
	_, err = CreateWorkingGroup(db, workingGroup)
	require.Error(t, err)

	workingGroup.ID = 2
	_, err = CreateWorkingGroup(db, workingGroup)
	require.Error(t, err)
}

func TestCreateWorkingGroup_allRequiredParametersPresent_returnsNoError(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	workingGroup := WorkingGroup{
		Name: "Tech (Best by a longshot)",
		Type: working_group_db_value,
	}

	_, err := CreateWorkingGroup(db, workingGroup)
	require.NoError(t, err)
}

func TestCreateWorkingGroup_insertAndFetchWorkingGroupNoMembers_returnsNoError(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	workingGroup := WorkingGroup{
		Name: "Tech FTW",
		Type: committee_db_value,
	}

	id, err := CreateWorkingGroup(db, workingGroup)
	require.NoError(t, err)
	workingGroup.ID = id

	fetchedGroup, err := GetWorkingGroup(db, WorkingGroupQueryOptions{GroupID: id})
	require.NoError(t, err)
	require.Equal(t, fetchedGroup, workingGroup)

	_, err = GetWorkingGroups(db, WorkingGroupQueryOptions{GroupID: id})
	require.Error(t, err)

	fetchedGroups, err := GetWorkingGroups(db, WorkingGroupQueryOptions{})
	require.NoError(t, err)
	require.Equal(t, fetchedGroups[0], workingGroup)

	fetchedGroups, err = GetWorkingGroups(db, WorkingGroupQueryOptions{GroupName: "Tech FTW"})
	require.NoError(t, err)
	require.Equal(t, fetchedGroups[0], workingGroup)
}

func TestCreateWorkingGroup_insertAndFetchWorkingGroupWithMembers(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	workingGroup := WorkingGroup{
		Name: "Emacs or Vim?",
		Type: working_group_db_value,
	}

	activistsToInsert := []string{"A", "B", "C", "D"}
	workingGroup.Members = insertActivists(t, db, activistsToInsert)
	id, err := CreateWorkingGroup(db, workingGroup)
	require.NoError(t, err)
	workingGroup.ID = id

	fetchedGroup, err := GetWorkingGroup(db, WorkingGroupQueryOptions{GroupID: id})
	require.NoError(t, err)
	validateReturnedWorkingGroup(t, workingGroup, fetchedGroup)

}

func TestUpdateWorkingGroup_updatePointPersonAndGroupEmail(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	workingGroup := WorkingGroup{
		Name: "Sanguine Salesman",
		Type: working_group_db_value,
	}

	id, err := CreateWorkingGroup(db, workingGroup)
	require.NoError(t, err)
	workingGroup.ID = id

	fetchedGroup, err := GetWorkingGroup(db, WorkingGroupQueryOptions{GroupID: id})
	require.NoError(t, err)
	validateReturnedWorkingGroup(t, workingGroup, fetchedGroup)

	member := insertActivists(t, db, []string{"Whimsical Winterbottom"})
	updatedGroupExpected := WorkingGroup{
		ID:            id,
		Name:          "Sanguine Salesman",
		Type:          working_group_db_value,
		GroupEmail:    sql.NullString{Valid: true, String: "foo@bar.com"},
		PointPersonID: member[0].ActivistID,
	}

	_, err = UpdateWorkingGroup(db, updatedGroupExpected)
	require.NoError(t, err)
	updatedGroupActual, err := GetWorkingGroup(db, WorkingGroupQueryOptions{GroupID: id})
	validateReturnedWorkingGroup(t, updatedGroupExpected, updatedGroupActual)
}

func validateReturnedWorkingGroup(t *testing.T, inserted WorkingGroup, returned WorkingGroup) {
	require.Equal(t, inserted.ID, returned.ID)
	require.Equal(t, inserted.Name, returned.Name)
	require.Equal(t, inserted.Type, returned.Type)
	require.Equal(t, inserted.GroupEmail, returned.GroupEmail)
	require.Equal(t, inserted.PointPersonID, returned.PointPersonID)
	require.Equal(t, len(inserted.Members), len(returned.Members))

	memberMap := make(map[int]string)
	for _, member := range inserted.Members {
		memberMap[member.ActivistID] = member.ActivistName
	}

	for _, member := range returned.Members {
		name, ok := memberMap[member.ActivistID]
		require.True(t, ok)
		require.Equal(t, name, member.ActivistName)
	}
}

func insertActivists(t *testing.T, db *sqlx.DB, names []string) []WorkingGroupMember {
	members := make([]WorkingGroupMember, len(names))
	for idx, a := range names {
		activist, err := GetOrCreateActivist(db, a)
		require.NoError(t, err)
		members[idx] = WorkingGroupMember{
			ActivistName: activist.Name,
			ActivistID:   activist.ID,
		}
	}
	return members
}
