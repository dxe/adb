package model

import (
	"testing"

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
