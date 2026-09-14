package persistence

import (
	"errors"
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

// TestAssignActivists_Repository exercises the bulk assign SQL against a real
// database: the authorize callback must see the rows the locking read found and
// be able to veto the write, hidden activists must be left alone, and the
// returned count must be the rows the database matched.
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

	allow := func([]model.ActivistAssignInfo) error { return nil }

	// The locked rows handed to authorize are what the caller gets to check:
	// unknown ids are omitted rather than reported, and hidden is included.
	t.Run("AuthorizeSeesLockedRows", func(t *testing.T) {
		const unknownID = 99999999
		var infos []model.ActivistAssignInfo
		err := repo.AssignActivists([]int{visibleID, hiddenID, unknownID}, assignee.ID,
			func(locked []model.ActivistAssignInfo) error {
				infos = locked
				return errStopAssign
			})
		require.ErrorIs(t, err, errStopAssign)

		byID := make(map[int]model.ActivistAssignInfo, len(infos))
		for _, info := range infos {
			byID[info.ID] = info
		}
		require.Len(t, byID, 2)
		require.Equal(t, model.SFBayChapterIdDevTest, byID[visibleID].ChapterID)
		require.False(t, byID[visibleID].Hidden)
		require.True(t, byID[hiddenID].Hidden)
	})

	t.Run("AssignsWholeSet", func(t *testing.T) {
		require.NoError(t, repo.AssignActivists([]int{visibleID, otherVisibleID}, assignee.ID, allow))

		require.Equal(t, assignee.ID, assignedTo(t, visibleID))
		require.Equal(t, assignee.ID, assignedTo(t, otherVisibleID))
	})

	// An activist that moved out of the caller's chapter since the request
	// started is rejected by authorize against the locked rows, before any
	// activist is reassigned.
	t.Run("WritesNothingWhenAuthorizeRejects", func(t *testing.T) {
		err := repo.AssignActivists([]int{visibleID, otherVisibleID}, 0, func([]model.ActivistAssignInfo) error {
			return errStopAssign
		})
		require.ErrorIs(t, err, errStopAssign)

		require.Equal(t, assignee.ID, assignedTo(t, visibleID), "not written")
		require.Equal(t, assignee.ID, assignedTo(t, otherVisibleID), "not written")
	})

	// The DSN sets clientFoundRows=true, so the row count is rows matched
	// rather than rows changed: reassigning activists to the user they are
	// already assigned to still matches every row.
	t.Run("SucceedsWhenNothingChanges", func(t *testing.T) {
		require.NoError(t, repo.AssignActivists([]int{visibleID, otherVisibleID}, assignee.ID, allow))
	})

	// authorize is what rejects hidden activists; if one gets past it, the
	// UPDATE skips the row and the whole transaction is rolled back.
	t.Run("RollsBackWhenAnActivistIsHidden", func(t *testing.T) {
		err := repo.AssignActivists([]int{visibleID, otherVisibleID, hiddenID}, 0, allow)
		require.ErrorIs(t, err, model.ErrNotFound)
		require.Contains(t, err.Error(), "matched 2 of 3 activists")

		require.Equal(t, assignee.ID, assignedTo(t, visibleID), "rolled back")
		require.Equal(t, assignee.ID, assignedTo(t, otherVisibleID), "rolled back")
		require.Equal(t, 0, assignedTo(t, hiddenID))
	})

	t.Run("Unassigns", func(t *testing.T) {
		require.NoError(t, repo.AssignActivists([]int{visibleID, otherVisibleID}, 0, allow))

		require.Equal(t, 0, assignedTo(t, visibleID))
		require.Equal(t, 0, assignedTo(t, otherVisibleID))
	})
}

// errStopAssign stands in for an authorization failure found against the
// locked rows.
var errStopAssign = errors.New("stop assign")
