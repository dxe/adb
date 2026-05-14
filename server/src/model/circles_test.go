package model

import (
	"testing"

	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

func newCircle(name string) CircleGroup {
	return CircleGroup{Name: name, Type: circle_group_db_value}
}

func TestCreateCircleGroup_NonzeroIDReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	cg := newCircle("Test Circle")
	cg.ID = 5
	_, err := CreateCircleGroup(db, cg)
	require.Error(t, err)
}

func TestCreateCircleGroup_EmptyNameReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	_, err := CreateCircleGroup(db, CircleGroup{Type: circle_group_db_value})
	require.Error(t, err)
}

func TestCreateCircleGroup_InvalidTypeReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	_, err := CreateCircleGroup(db, CircleGroup{Name: "Test", Type: 99})
	require.Error(t, err)
}

func TestCreateCircleGroup_HappyPath(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	id, err := CreateCircleGroup(db, newCircle("Outreach Circle"))
	require.NoError(t, err)
	require.NotZero(t, id)

	fetched, err := GetCircleGroup(db, CircleGroupQueryOptions{GroupID: id})
	require.NoError(t, err)
	require.Equal(t, "Outreach Circle", fetched.Name)
	require.Equal(t, circle_group_db_value, fetched.Type)
}

func TestUpdateCircleGroup_ZeroIDReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	_, err := UpdateCircleGroup(db, newCircle("No ID"))
	require.Error(t, err)
}

func TestUpdateCircleGroup_HappyPath(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	id, err := CreateCircleGroup(db, newCircle("Before"))
	require.NoError(t, err)

	_, err = UpdateCircleGroup(db, CircleGroup{ID: id, Name: "After", Type: circle_group_db_value})
	require.NoError(t, err)

	fetched, err := GetCircleGroup(db, CircleGroupQueryOptions{GroupID: id})
	require.NoError(t, err)
	require.Equal(t, "After", fetched.Name)
}

func TestGetCircleGroups_ReturnsList(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	_, err := CreateCircleGroup(db, newCircle("Alpha"))
	require.NoError(t, err)
	_, err = CreateCircleGroup(db, newCircle("Beta"))
	require.NoError(t, err)

	groups, err := GetCircleGroups(db, CircleGroupQueryOptions{})
	require.NoError(t, err)
	require.Len(t, groups, 2)
}

func TestGetCircleGroups_RejectsGroupIDInOptions(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	_, err := GetCircleGroups(db, CircleGroupQueryOptions{GroupID: 1})
	require.Error(t, err)
}

func TestDeleteCircleGroup_ZeroIDReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	require.Error(t, DeleteCircleGroup(db, 0))
}

func TestDeleteCircleGroup_HappyPath(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	id, err := CreateCircleGroup(db, newCircle("Temporary"))
	require.NoError(t, err)

	require.NoError(t, DeleteCircleGroup(db, id))

	_, err = GetCircleGroup(db, CircleGroupQueryOptions{GroupID: id})
	require.ErrorContains(t, err, "no circle with ID")
}

func TestDeleteCircleGroup_WithMembersReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	activist, err := GetOrCreateActivist(db, "Circle Member", SFBayChapterIdDevTest)
	require.NoError(t, err)

	cg := newCircle("Has Members")
	cg.Members = []CircleGroupMember{{ActivistID: activist.ID}}
	id, err := CreateCircleGroup(db, cg)
	require.NoError(t, err)

	require.Error(t, DeleteCircleGroup(db, id))
}
