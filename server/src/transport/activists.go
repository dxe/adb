package transport

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dxe/adb/model"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
)

type QueryActivistResultJSON struct {
	Activists  []model.ActivistJSON    `json:"activists"`
	Pagination QueryActivistPagination `json:"pagination"`
}

type QueryActivistPagination struct {
	NextCursor string `json:"next_cursor"`
}

// ActivistPatchInput represents PATCH /api/activists/{id} request body fields.
// Only non-nil fields are applied.
type ActivistPatchInput struct {
	// Activist fields
	Email         *string `json:"email"`
	Facebook      *string `json:"facebook"`
	Name          *string `json:"name"`
	PreferredName *string `json:"preferred_name"`
	Phone         *string `json:"phone"`
	Pronouns      *string `json:"pronouns"`
	Language      *string `json:"language"`
	Accessibility *string `json:"accessibility"`
	Birthday      *string `json:"dob"`
	Location      *string `json:"location"`

	// ActivistMembershipData fields
	ActivistLevel *string `json:"activist_level"`
	Source        *string `json:"source"`
	Hiatus        *bool   `json:"hiatus"`

	// ActivistConnectionData fields (user-editable)
	Connector             *string `json:"connector"`
	Training0             *string `json:"training0"`
	Training1             *string `json:"training1"`
	Training4             *string `json:"training4"`
	Training5             *string `json:"training5"`
	Training6             *string `json:"training6"`
	ConsentQuiz           *string `json:"consent_quiz"`
	TrainingProtest       *string `json:"training_protest"`
	DevQuiz               *string `json:"dev_quiz"`
	DevInterest           *string `json:"dev_interest"`
	CMFirstEmail          *string `json:"cm_first_email"`
	CMApprovalEmail       *string `json:"cm_approval_email"`
	ProspectOrganizer     *bool   `json:"prospect_organizer"`
	ProspectChapterMember *bool   `json:"prospect_chapter_member"`
	ReferralFriends       *string `json:"referral_friends"`
	ReferralApply         *string `json:"referral_apply"`
	ReferralOutlet        *string `json:"referral_outlet"`
	InterestDate          *string `json:"interest_date"`
	MPI                   *bool   `json:"mpi"`
	Notes                 *string `json:"notes"`
	VisionWall            *string `json:"vision_wall"`
	VotingAgreement       *bool   `json:"voting_agreement"`
	StreetAddress         *string `json:"street_address"`
	City                  *string `json:"city"`
	State                 *string `json:"state"`
	AssignedTo            *int    `json:"assigned_to"`
	FollowupDate          *string `json:"followup_date"`
}

func nullableString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	v := strings.TrimSpace(*s)
	return sql.NullString{String: v, Valid: v != ""}
}

// ToPatchData converts transport PATCH input into model patch fields.
func (p ActivistPatchInput) ToPatchData() model.ActivistPatchData {
	var d model.ActivistPatchData

	addString := func(name model.ActivistColumnName, ptr *string) {
		if ptr != nil {
			d.Append(name, strings.TrimSpace(*ptr))
		}
	}
	addNullableString := func(name model.ActivistColumnName, ptr *string) {
		if ptr != nil {
			d.Append(name, nullableString(ptr))
		}
	}
	addBool := func(name model.ActivistColumnName, ptr *bool) {
		if ptr != nil {
			d.Append(name, *ptr)
		}
	}
	addInt := func(name model.ActivistColumnName, ptr *int) {
		if ptr != nil {
			d.Append(name, *ptr)
		}
	}

	addString(model.ColEmail, p.Email)
	addString(model.ColFacebook, p.Facebook)
	addString(model.ColName, p.Name)
	addString(model.ColPreferredName, p.PreferredName)
	addString(model.ColPhone, p.Phone)
	addString(model.ColPronouns, p.Pronouns)
	addString(model.ColLanguage, p.Language)
	addString(model.ColAccessibility, p.Accessibility)
	addNullableString(model.ColDOB, p.Birthday)
	addNullableString(model.ColLocation, p.Location)

	addString(model.ColActivistLevel, p.ActivistLevel)
	addString(model.ColSource, p.Source)
	addBool(model.ColHiatus, p.Hiatus)

	addString(model.ColConnector, p.Connector)
	addNullableString(model.ColTraining0, p.Training0)
	addNullableString(model.ColTraining1, p.Training1)
	addNullableString(model.ColTraining4, p.Training4)
	addNullableString(model.ColTraining5, p.Training5)
	addNullableString(model.ColTraining6, p.Training6)
	addNullableString(model.ColConsentQuiz, p.ConsentQuiz)
	addNullableString(model.ColTrainingProtest, p.TrainingProtest)
	addNullableString(model.ColDevQuiz, p.DevQuiz)
	addString(model.ColDevInterest, p.DevInterest)
	addNullableString(model.ColCMFirstEmail, p.CMFirstEmail)
	addNullableString(model.ColCMApprovalEmail, p.CMApprovalEmail)
	addBool(model.ColProspectOrganizer, p.ProspectOrganizer)
	addBool(model.ColProspectChapterMbr, p.ProspectChapterMember)
	addString(model.ColReferralFriends, p.ReferralFriends)
	addString(model.ColReferralApply, p.ReferralApply)
	addString(model.ColReferralOutlet, p.ReferralOutlet)
	addNullableString(model.ColInterestDate, p.InterestDate)
	addBool(model.ColMPI, p.MPI)
	addNullableString(model.ColNotes, p.Notes)
	addString(model.ColVisionWall, p.VisionWall)
	addBool(model.ColVotingAgreement, p.VotingAgreement)
	addString(model.ColStreetAddress, p.StreetAddress)
	addString(model.ColCity, p.City)
	addString(model.ColState, p.State)
	addInt(model.ColAssignedTo, p.AssignedTo)
	addNullableString(model.ColFollowupDate, p.FollowupDate)

	return d
}

func ActivistsSearchHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, repo model.ActivistRepository) {
	var options model.QueryActivistOptions
	if err := json.NewDecoder(r.Body).Decode(&options); err != nil && err != io.EOF {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	result, err := model.QueryActivists(authedUser, options, repo)
	if err != nil {
		if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, QueryActivistResultJSON{
		Activists: model.BuildActivistJSONArray(result.Activists),
		Pagination: QueryActivistPagination{
			NextCursor: result.Pagination.NextCursor,
		},
	})
}

func ActivistPatchHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, db *sqlx.DB, repo model.ActivistRepository, userRepo model.UserRepository) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	activistID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, http.StatusBadRequest, fmt.Errorf("invalid activist id %s: %w", rawID, err))
		return
	}

	var input ActivistPatchInput
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&input); err != nil {
		sendErrorMessage(w, http.StatusBadRequest, err)
		return
	}

	if err := model.PatchActivist(db, repo, userRepo, authedUser, activistID, input.ToPatchData()); err != nil {
		if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else if errors.Is(err, model.ErrNotFound) {
			sendErrorMessage(w, http.StatusNotFound, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	// Return updated activist.
	activist, err := model.GetActivistJSONForUser(db, authedUser, model.GetActivistOptions{ID: activistID})
	if err != nil {
		sendErrorMessage(w, http.StatusInternalServerError, err)
		return
	}

	writeJSON(w, map[string]any{
		"activist": activist,
	})
}

func ActivistGetHandler(w http.ResponseWriter, r *http.Request, authedUser model.ADBUser, db *sqlx.DB) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	activistID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, http.StatusBadRequest, fmt.Errorf("invalid activist id %s: %w", rawID, err))
		return
	}

	activist, err := model.GetActivistJSONForUser(db, authedUser, model.GetActivistOptions{ID: activistID})
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			sendErrorMessage(w, http.StatusNotFound, fmt.Errorf("no activist found with id %d", activistID))
		} else if errors.Is(err, model.ErrValidation) {
			sendErrorMessage(w, http.StatusBadRequest, err)
		} else {
			sendErrorMessage(w, http.StatusInternalServerError, err)
		}
		return
	}

	writeJSON(w, map[string]any{
		"activist": activist,
	})
}
