package model

import (
	"testing"

	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

func TestListActivistInteractions_ZeroIDReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	_, err := ListActivistInteractions(db, 0)
	require.Error(t, err)
}

func TestListActivistInteractions_EmptyForNewActivist(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	activist, err := GetOrCreateActivist(db, "New Activist", SFBayChapterIdDevTest)
	require.NoError(t, err)

	interactions, err := ListActivistInteractions(db, activist.ID)
	require.NoError(t, err)
	require.Empty(t, interactions)
}

func TestSaveInteraction_Insert(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	activist, err := GetOrCreateActivist(db, "Interacted Activist", SFBayChapterIdDevTest)
	require.NoError(t, err)

	err = SaveInteraction(db, Interaction{
		ActivistID: activist.ID,
		UserID:     DevTestUserId,
		Method:     "phone",
		Outcome:    "left voicemail",
		Notes:      "called at noon",
	})
	require.NoError(t, err)

	interactions, err := ListActivistInteractions(db, activist.ID)
	require.NoError(t, err)
	require.Len(t, interactions, 1)
	require.Equal(t, "phone", interactions[0].Method)
	require.Equal(t, "left voicemail", interactions[0].Outcome)
	require.Equal(t, "called at noon", interactions[0].Notes)
}

func TestSaveInteraction_Update(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	activist, err := GetOrCreateActivist(db, "Update Target", SFBayChapterIdDevTest)
	require.NoError(t, err)

	err = SaveInteraction(db, Interaction{
		ActivistID: activist.ID,
		UserID:     DevTestUserId,
		Method:     "email",
		Outcome:    "no reply",
	})
	require.NoError(t, err)

	interactions, err := ListActivistInteractions(db, activist.ID)
	require.NoError(t, err)
	require.Len(t, interactions, 1)

	interactions[0].Outcome = "replied"
	err = SaveInteraction(db, interactions[0])
	require.NoError(t, err)

	updated, err := ListActivistInteractions(db, activist.ID)
	require.NoError(t, err)
	require.Len(t, updated, 1)
	require.Equal(t, "replied", updated[0].Outcome)
}

func TestDeleteInteraction_ZeroIDReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	require.Error(t, DeleteInteraction(db, 0))
}

func TestDeleteInteraction_HappyPath(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	activist, err := GetOrCreateActivist(db, "Delete Target", SFBayChapterIdDevTest)
	require.NoError(t, err)

	err = SaveInteraction(db, Interaction{
		ActivistID: activist.ID,
		UserID:     DevTestUserId,
		Method:     "text",
		Outcome:    "no response",
	})
	require.NoError(t, err)

	interactions, err := ListActivistInteractions(db, activist.ID)
	require.NoError(t, err)
	require.Len(t, interactions, 1)

	require.NoError(t, DeleteInteraction(db, interactions[0].ID))

	remaining, err := ListActivistInteractions(db, activist.ID)
	require.NoError(t, err)
	require.Empty(t, remaining)
}
