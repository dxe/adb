package model

import (
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
}

func TestCreateWorkingGroup_insertAndFetchWorkingGroupWithMembersByID(t *testing.T) {
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

func TestCreateWorkingGroup_insertAndFetchWorkingGroupWithMembersByNameAndID(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	workingGroup := WorkingGroup{
		Name: "The Citadel",
		Type: working_group_db_value,
	}

	activistsToInsert := []string{"Rick", "And", "Morty"}
	workingGroup.Members = insertActivists(t, db, activistsToInsert)
	id, err := CreateWorkingGroup(db, workingGroup)
	require.NoError(t, err)
	workingGroup.ID = id

	fetchedGroup, err := GetWorkingGroup(db, WorkingGroupQueryOptions{GroupName: "The Citadel"})
	require.NoError(t, err)
	validateReturnedWorkingGroup(t, workingGroup, fetchedGroup)

	fetchedGroup2, err := GetWorkingGroup(db, WorkingGroupQueryOptions{GroupName: "The Citadel", GroupID: id})
	require.NoError(t, err)
	validateReturnedWorkingGroup(t, workingGroup, fetchedGroup2)
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

	members := insertActivists(t, db, []string{"Whimsical Winterbottom"})
	members[0].PointPerson = true
	updatedGroupExpected := WorkingGroup{
		ID:         id,
		Name:       "Sanguine Salesman",
		Type:       working_group_db_value,
		GroupEmail: "foo@bar.com",
		Members:    members,
	}

	_, err = UpdateWorkingGroup(db, updatedGroupExpected)
	require.NoError(t, err)
	updatedGroupActual, err := GetWorkingGroup(db, WorkingGroupQueryOptions{GroupID: id})
	validateReturnedWorkingGroup(t, updatedGroupExpected, updatedGroupActual)
}

func TestUpdateWorkingGroup_updateMultipleGroups(t *testing.T) {
	db := newTestDB()
	defer db.Close()

	workingGroup1 := WorkingGroup{
		Name: "WG 1",
		Type: working_group_db_value,
	}

	workingGroup2 := WorkingGroup{
		Name: "WG 2",
		Type: working_group_db_value,
	}

	id1, err := CreateWorkingGroup(db, workingGroup1)
	require.NoError(t, err)
	id2, err := CreateWorkingGroup(db, workingGroup2)
	require.NoError(t, err)

	members1 := insertActivists(t, db, []string{"Anthony Abe", "Smithy Smith", "Rick Rickel"})
	members2 := insertActivists(t, db, []string{"The", "Seven", "Deadly", "Sins"})

	UpdatedExpected1 := WorkingGroup{
		ID:         id1,
		Name:       "WG 1",
		Type:       working_group_db_value,
		GroupEmail: "hello@hello.org",
		Members:    members1,
	}

	UpdatedExpected2 := WorkingGroup{
		ID:      id2,
		Name:    "WG 2",
		Type:    working_group_db_value,
		Members: members2,
	}

	_, err = UpdateWorkingGroup(db, UpdatedExpected1)
	require.NoError(t, err)
	_, err = UpdateWorkingGroup(db, UpdatedExpected2)
	require.NoError(t, err)

	updatedGroups, err := GetWorkingGroups(db, WorkingGroupQueryOptions{})
	require.NoError(t, err)

	for _, group := range updatedGroups {
		if group.ID == UpdatedExpected1.ID {
			validateReturnedWorkingGroup(t, UpdatedExpected1, group)
		} else {
			validateReturnedWorkingGroup(t, UpdatedExpected2, group)
		}
	}

}

func validateReturnedWorkingGroup(t *testing.T, inserted WorkingGroup, returned WorkingGroup) {
	require.Equal(t, inserted.ID, returned.ID)
	require.Equal(t, inserted.Name, returned.Name)
	require.Equal(t, inserted.Type, returned.Type)
	require.Equal(t, inserted.GroupEmail, returned.GroupEmail)
	require.Equal(t, len(inserted.Members), len(returned.Members))

	memberMap := make(map[int]WorkingGroupMember)
	for _, member := range inserted.Members {
		memberMap[member.ActivistID] = member
	}

	for _, member := range returned.Members {
		insertedMember, ok := memberMap[member.ActivistID]
		require.True(t, ok)
		require.Equal(t, insertedMember.ActivistName, member.ActivistName)
		require.Equal(t, insertedMember.PointPerson, member.PointPerson)
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
