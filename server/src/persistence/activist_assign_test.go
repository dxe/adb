package persistence

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

// TestAssignActivists_Repository exercises the bulk assign SQL against a real
// database: hidden activists must be left alone, and the returned count must be
// the rows the database matched.
func TestAssignActivists_Repository(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	repo := NewActivistRepository(db)
	userRepo := NewUserRepository(db)

	assignee, err := userRepo.CreateUser(model.ADBUser{
		Email:     "bulk-assignee@example.org",
		Name:      "Bulk Assignee",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.NoError(t, err)

	newActivist := func(t *testing.T, name string) int {
		t.Helper()
		id, err := model.CreateActivist(db, model.ActivistExtra{
			Activist: model.Activist{Name: name, ChapterID: model.SFBayChapterIdDevTest},
		})
		require.NoError(t, err)
		return id
	}
	assignedTo := func(t *testing.T, id int) int {
		t.Helper()
		activist, err := model.GetActivistExtra(db, id)
		require.NoError(t, err)
		return activist.AssignedTo
	}

	visibleID := newActivist(t, "Bulk Assign Visible")
	otherVisibleID := newActivist(t, "Bulk Assign Visible Two")
	hiddenID := newActivist(t, "Bulk Assign Hidden")
	_, err = db.Exec(`UPDATE activists SET hidden = 1 WHERE id = ?`, hiddenID)
	require.NoError(t, err)

	t.Run("GetActivistAssignInfo", func(t *testing.T) {
		const unknownID = 99999999
		infos, err := repo.GetActivistAssignInfo([]int{visibleID, hiddenID, unknownID})
		require.NoError(t, err)

		// Unknown ids are omitted rather than reported.
		byID := make(map[int]model.ActivistAssignInfo, len(infos))
		for _, info := range infos {
			byID[info.ID] = info
		}
		require.Len(t, byID, 2)
		require.Equal(t, model.SFBayChapterIdDevTest, byID[visibleID].ChapterID)
		require.False(t, byID[visibleID].Hidden)
		require.True(t, byID[hiddenID].Hidden)
	})

	t.Run("SkipsHiddenActivistsAndCountsMatchedRows", func(t *testing.T) {
		rows, err := repo.AssignActivists([]int{visibleID, otherVisibleID, hiddenID}, assignee.ID)
		require.NoError(t, err)
		require.Equal(t, int64(2), rows, "hidden activist should not be counted or assigned")

		require.Equal(t, assignee.ID, assignedTo(t, visibleID))
		require.Equal(t, assignee.ID, assignedTo(t, otherVisibleID))
		require.Equal(t, 0, assignedTo(t, hiddenID))
	})

	// The DSN sets clientFoundRows=true, so the count is rows matched rather
	// than rows changed: assigning the same activists again still reports 2.
	t.Run("CountsMatchedRowsNotChangedRows", func(t *testing.T) {
		rows, err := repo.AssignActivists([]int{visibleID, otherVisibleID}, assignee.ID)
		require.NoError(t, err)
		require.Equal(t, int64(2), rows)
	})

	// A repeated id matches one row, which is why callers report the database
	// count instead of the number of ids they sent.
	t.Run("DeduplicatesRepeatedIDs", func(t *testing.T) {
		rows, err := repo.AssignActivists([]int{visibleID, visibleID}, 0)
		require.NoError(t, err)
		require.Equal(t, int64(1), rows)
		require.Equal(t, 0, assignedTo(t, visibleID))
	})
}
