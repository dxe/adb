package persistence

import (
	"database/sql"
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/pkg/shared"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

// TestPatchActivist_UpdatesAllPatchableFields exercises every column in
// activistPatchSetters end-to-end through the real DBActivistRepository to
// confirm each value is persisted.
func TestPatchActivist_UpdatesAllPatchableFields(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	repo := NewActivistRepository(db)
	userRepo := NewUserRepository(db)

	// Create a user to assign the activist to (needed for ColAssignedTo validation).
	assignee, err := userRepo.CreateUser(model.ADBUser{
		Email:     "assignee@example.org",
		Name:      "Assignee",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.NoError(t, err)

	activistID, err := model.CreateActivist(db, model.ActivistExtra{
		Activist: model.Activist{
			Name:      "Initial Name",
			ChapterID: model.SFBayChapterIdDevTest,
		},
	})
	require.NoError(t, err)

	patch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
		{Name: model.ColEmail, Value: "patched@example.org"},
		{Name: model.ColFacebook, Value: "patched_fb"},
		{Name: model.ColName, Value: "Patched Name"},
		{Name: model.ColPreferredName, Value: "Patchy"},
		{Name: model.ColPhone, Value: "555-0100"},
		{Name: model.ColPronouns, Value: "they/them"},
		{Name: model.ColLanguage, Value: "Spanish"},
		{Name: model.ColAccessibility, Value: "wheelchair access"},
		{Name: model.ColDOB, Value: sql.NullString{String: "1990-01-01", Valid: true}},
		{Name: model.ColLocation, Value: sql.NullString{String: "94110", Valid: true}},
		{Name: model.ColActivistLevel, Value: "Chapter Member"},
		{Name: model.ColSource, Value: "patched_source"},
		{Name: model.ColHiatus, Value: true},
		{Name: model.ColConnector, Value: "patched_connector"},
		{Name: model.ColTraining0, Value: sql.NullString{String: "2026-01-01", Valid: true}},
		{Name: model.ColTraining1, Value: sql.NullString{String: "2026-01-02", Valid: true}},
		{Name: model.ColTraining4, Value: sql.NullString{String: "2026-01-04", Valid: true}},
		{Name: model.ColTraining5, Value: sql.NullString{String: "2026-01-05", Valid: true}},
		{Name: model.ColTraining6, Value: sql.NullString{String: "2026-01-06", Valid: true}},
		{Name: model.ColConsentQuiz, Value: sql.NullString{String: "Passed", Valid: true}},
		{Name: model.ColTrainingProtest, Value: sql.NullString{String: "2026-02-01", Valid: true}},
		{Name: model.ColDevQuiz, Value: sql.NullString{String: "Passed", Valid: true}},
		{Name: model.ColDevInterest, Value: "high"},
		{Name: model.ColCMFirstEmail, Value: sql.NullString{String: "2026-03-01", Valid: true}},
		{Name: model.ColCMApprovalEmail, Value: sql.NullString{String: "2026-03-15", Valid: true}},
		{Name: model.ColProspectOrganizer, Value: true},
		{Name: model.ColProspectChapterMbr, Value: true},
		{Name: model.ColReferralFriends, Value: "Alice"},
		{Name: model.ColReferralApply, Value: "online_form"},
		{Name: model.ColReferralOutlet, Value: "newsletter"},
		{Name: model.ColInterestDate, Value: sql.NullString{String: "2026-04-01", Valid: true}},
		{Name: model.ColNotes, Value: sql.NullString{String: "patched notes", Valid: true}},
		{Name: model.ColVisionWall, Value: "vision1"},
		{Name: model.ColVotingAgreement, Value: true},
		{Name: model.ColStreetAddress, Value: "1 Patched St"},
		{Name: model.ColCity, Value: "Patchville"},
		{Name: model.ColState, Value: "PA"},
		{Name: model.ColAssignedTo, Value: assignee.ID},
		{Name: model.ColFollowupDate, Value: sql.NullString{String: "2026-05-01", Valid: true}},
	}}

	organizer := model.ADBUser{
		ID:        assignee.ID + 1,
		Email:     "organizer@example.org",
		Name:      "Organizer",
		Roles:     []string{shared.RoleOrganizer},
		ChapterID: model.SFBayChapterIdDevTest,
	}

	err = model.PatchActivist(db, repo, userRepo, organizer, activistID, patch)
	require.NoError(t, err)

	updated, err := model.GetActivistsExtra(db, model.GetActivistOptions{ID: activistID})
	require.NoError(t, err)
	require.Len(t, updated, 1)
	a := updated[0]

	require.Equal(t, "patched@example.org", a.Email)
	require.Equal(t, "patched_fb", a.Facebook)
	require.Equal(t, "Patched Name", a.Name)
	require.Equal(t, "Patchy", a.PreferredName)
	require.Equal(t, "555-0100", a.Phone)
	require.Equal(t, "they/them", a.Pronouns)
	require.Equal(t, "Spanish", a.Language)
	require.Equal(t, "wheelchair access", a.Accessibility)
	require.Equal(t, sql.NullString{String: "1990-01-01", Valid: true}, a.Birthday)
	require.Equal(t, sql.NullString{String: "94110", Valid: true}, a.Location)
	require.Equal(t, "Chapter Member", a.ActivistLevel)
	require.Equal(t, "patched_source", a.Source)
	require.Equal(t, true, a.Hiatus)
	require.Equal(t, "patched_connector", a.Connector)
	require.Equal(t, sql.NullString{String: "2026-01-01", Valid: true}, a.Training0)
	require.Equal(t, sql.NullString{String: "2026-01-02", Valid: true}, a.Training1)
	require.Equal(t, sql.NullString{String: "2026-01-04", Valid: true}, a.Training4)
	require.Equal(t, sql.NullString{String: "2026-01-05", Valid: true}, a.Training5)
	require.Equal(t, sql.NullString{String: "2026-01-06", Valid: true}, a.Training6)
	require.Equal(t, sql.NullString{String: "Passed", Valid: true}, a.ConsentQuiz)
	require.Equal(t, sql.NullString{String: "2026-02-01", Valid: true}, a.TrainingProtest)
	require.Equal(t, sql.NullString{String: "Passed", Valid: true}, a.Quiz)
	require.Equal(t, "high", a.DevInterest)
	require.Equal(t, sql.NullString{String: "2026-03-01", Valid: true}, a.CMFirstEmail)
	require.Equal(t, sql.NullString{String: "2026-03-15", Valid: true}, a.CMApprovalEmail)
	require.Equal(t, true, a.ProspectOrganizer)
	require.Equal(t, true, a.ProspectChapterMember)
	require.Equal(t, "Alice", a.ReferralFriends)
	require.Equal(t, "online_form", a.ReferralApply)
	require.Equal(t, "newsletter", a.ReferralOutlet)
	require.Equal(t, sql.NullString{String: "2026-04-01", Valid: true}, a.InterestDate)
	require.Equal(t, sql.NullString{String: "patched notes", Valid: true}, a.Notes)
	require.Equal(t, "vision1", a.VisionWall)
	require.Equal(t, true, a.VotingAgreement)
	require.Equal(t, "1 Patched St", a.StreetAddress)
	require.Equal(t, "Patchville", a.City)
	require.Equal(t, "PA", a.State)
	require.Equal(t, assignee.ID, a.AssignedTo)
	require.Equal(t, sql.NullString{String: "2026-05-01", Valid: true}, a.FollowupDate)

	// Test that patch calls are idempotent: setting a column to its current
	// value should not cause an error, even though no rows are updated.
	//
	// MySQL's go-sql-driver reports RowsAffected as the number of rows whose values
	// actually changed, not the number of rows matched by the WHERE clause — unless
	// the DSN includes clientFoundRows=true. A same-value UPDATE therefore returns
	// RowsAffected==0 even though the row exists, which would cause the rows==0
	// guard in DBActivistRepository.PatchActivist to emit a false ErrNotFound.

	noOpPatch := model.ActivistPatchData{Fields: []model.ActivistPatchField{
		{Name: model.ColEmail, Value: a.Email},
	}}
	err = model.PatchActivist(db, repo, userRepo, devUser, activistID, noOpPatch)
	require.NoError(t, err)
}
