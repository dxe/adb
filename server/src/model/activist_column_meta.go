package model

import (
	"database/sql"
	"fmt"

	"github.com/go-sql-driver/mysql"
)

// activistFieldSetter writes a typed value from a patch onto an ActivistExtra.
// Returns an error (not panic) if the value's runtime type is wrong for the field.
type activistFieldSetter func(*ActivistExtra, any) error

// setField returns a setter that asserts v is a T, then assigns it via set.
func setField[T any](set func(*ActivistExtra, T)) activistFieldSetter {
	return func(a *ActivistExtra, v any) error {
		t, ok := v.(T)
		if !ok {
			var zero T
			return fmt.Errorf("expected %T, got %T", zero, v)
		}
		set(a, t)
		return nil
	}
}

// ActivistColumnInfo contains metadata about an Activist column.
type ActivistColumnInfo struct {
	// Setter writes a typed patch value onto an ActivistExtra.
	Setter activistFieldSetter

	// DbCol is the database (SQL) column name.
	//
	// While `ActivistColumnName`s often coincide, these names are used by API
	// clients, while `DbCol` is used by the database layer for separation
	// of concerns.
	//
	// Kept in sync with struct db tags via
	// TestPatchActivist_UpdatesAllPatchableFields
	DbCol string

	// Nullable indicates the column is a NULLable SQL type.
	Nullable bool

	// BumpTimestamps lists timestamp columns to bump to NOW() when this
	// column's value changes.
	BumpTimestamps []string

	// UserPatchable indicates the column may be set via the PATCH API.
	UserPatchable bool
}

// ActivistColumns contains metadata about Activist columns.
var ActivistColumns = map[ActivistColumnName]ActivistColumnInfo{
	ColEmail: {
		Setter:         setField(func(a *ActivistExtra, v string) { a.Email = v }),
		DbCol:          "email",
		BumpTimestamps: []string{"email_updated"},
		UserPatchable:  true,
	},
	ColFacebook: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.Facebook = v }),
		DbCol:         "facebook",
		UserPatchable: true,
	},
	ColName: {
		Setter:         setField(func(a *ActivistExtra, v string) { a.Name = v }),
		DbCol:          "name",
		BumpTimestamps: []string{"name_updated"},
		UserPatchable:  true,
	},
	ColPreferredName: {
		Setter:         setField(func(a *ActivistExtra, v string) { a.PreferredName = v }),
		DbCol:          "preferred_name",
		BumpTimestamps: []string{"name_updated"},
		UserPatchable:  true,
	},
	ColPhone: {
		Setter:         setField(func(a *ActivistExtra, v string) { a.Phone = v }),
		DbCol:          "phone",
		BumpTimestamps: []string{"phone_updated"},
		UserPatchable:  true,
	},
	ColPronouns: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.Pronouns = v }),
		DbCol:         "pronouns",
		UserPatchable: true,
	},
	ColLanguage: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.Language = v }),
		DbCol:         "language",
		UserPatchable: true,
	},
	ColAccessibility: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.Accessibility = v }),
		DbCol:         "accessibility",
		UserPatchable: true,
	},
	ColDOB: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Birthday = v }),
		DbCol:         "dob",
		Nullable:      true,
		UserPatchable: true,
	},
	ColActivistLevel: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.ActivistLevel = v }),
		DbCol:         "activist_level",
		UserPatchable: true,
	},
	ColSource: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.Source = v }),
		DbCol:         "source",
		UserPatchable: true,
	},
	ColHiatus: {
		Setter:        setField(func(a *ActivistExtra, v bool) { a.Hiatus = v }),
		DbCol:         "hiatus",
		UserPatchable: true,
	},
	ColConnector: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.Connector = v }),
		DbCol:         "connector",
		UserPatchable: true,
	},
	ColTraining0: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Training0 = v }),
		DbCol:         "training0",
		Nullable:      true,
		UserPatchable: true,
	},
	ColTraining1: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Training1 = v }),
		DbCol:         "training1",
		Nullable:      true,
		UserPatchable: true,
	},
	ColTraining4: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Training4 = v }),
		DbCol:         "training4",
		Nullable:      true,
		UserPatchable: true,
	},
	ColTraining5: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Training5 = v }),
		DbCol:         "training5",
		Nullable:      true,
		UserPatchable: true,
	},
	ColTraining6: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Training6 = v }),
		DbCol:         "training6",
		Nullable:      true,
		UserPatchable: true,
	},
	ColConsentQuiz: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.ConsentQuiz = v }),
		DbCol:         "consent_quiz",
		Nullable:      true,
		UserPatchable: true,
	},
	ColTrainingProtest: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.TrainingProtest = v }),
		DbCol:         "training_protest",
		Nullable:      true,
		UserPatchable: true,
	},
	ColDevQuiz: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Quiz = v }),
		DbCol:         "dev_quiz",
		Nullable:      true,
		UserPatchable: true,
	},
	ColDevInterest: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.DevInterest = v }),
		DbCol:         "dev_interest",
		UserPatchable: true,
	},
	ColCMFirstEmail: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.CMFirstEmail = v }),
		DbCol:         "cm_first_email",
		Nullable:      true,
		UserPatchable: true,
	},
	ColCMApprovalEmail: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.CMApprovalEmail = v }),
		DbCol:         "cm_approval_email",
		Nullable:      true,
		UserPatchable: true,
	},
	ColProspectOrganizer: {
		Setter:        setField(func(a *ActivistExtra, v bool) { a.ProspectOrganizer = v }),
		DbCol:         "prospect_organizer",
		UserPatchable: true,
	},
	ColProspectChapterMbr: {
		Setter:        setField(func(a *ActivistExtra, v bool) { a.ProspectChapterMember = v }),
		DbCol:         "prospect_chapter_member",
		UserPatchable: true,
	},
	ColReferralFriends: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.ReferralFriends = v }),
		DbCol:         "referral_friends",
		UserPatchable: true,
	},
	ColReferralApply: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.ReferralApply = v }),
		DbCol:         "referral_apply",
		UserPatchable: true,
	},
	ColReferralOutlet: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.ReferralOutlet = v }),
		DbCol:         "referral_outlet",
		UserPatchable: true,
	},
	ColInterestDate: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.InterestDate = v }),
		DbCol:         "interest_date",
		Nullable:      true,
		UserPatchable: true,
	},
	ColMPI: {
		Setter: setField(func(a *ActivistExtra, v bool) { a.MPI = v }),
		DbCol:  "mpi",
	},
	ColNotes: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.Notes = v }),
		DbCol:         "notes",
		Nullable:      true,
		UserPatchable: true,
	},
	ColVisionWall: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.VisionWall = v }),
		DbCol:         "vision_wall",
		UserPatchable: true,
	},
	ColVotingAgreement: {
		Setter:        setField(func(a *ActivistExtra, v bool) { a.VotingAgreement = v }),
		DbCol:         "voting_agreement",
		UserPatchable: true,
	},
	ColStreetAddress: {
		Setter:         setField(func(a *ActivistExtra, v string) { a.StreetAddress = v }),
		DbCol:          "street_address",
		BumpTimestamps: []string{"address_updated", "location_updated"},
		UserPatchable:  true,
	},
	ColCity: {
		Setter:         setField(func(a *ActivistExtra, v string) { a.City = v }),
		DbCol:          "city",
		BumpTimestamps: []string{"address_updated", "location_updated"},
		UserPatchable:  true,
	},
	ColState: {
		Setter:         setField(func(a *ActivistExtra, v string) { a.State = v }),
		DbCol:          "state",
		BumpTimestamps: []string{"address_updated", "location_updated"},
		UserPatchable:  true,
	},
	ColLocation: {
		Setter:         setField(func(a *ActivistExtra, v sql.NullString) { a.Location = v }),
		DbCol:          "location",
		Nullable:       true,
		BumpTimestamps: []string{"location_updated"},
		UserPatchable:  true,
	},
	ColLat: {
		Setter:        setField(func(a *ActivistExtra, v float64) { a.Lat = v }),
		DbCol:         "lat",
		UserPatchable: false,
	},
	ColLng: {
		Setter:        setField(func(a *ActivistExtra, v float64) { a.Lng = v }),
		DbCol:         "lng",
		UserPatchable: false,
	},
	ColAssignedTo: {
		Setter:        setField(func(a *ActivistExtra, v int) { a.AssignedTo = v }),
		DbCol:         "assigned_to",
		UserPatchable: true,
	},
	ColFollowupDate: {
		Setter:        setField(func(a *ActivistExtra, v sql.NullString) { a.FollowupDate = v }),
		DbCol:         "followup_date",
		Nullable:      true,
		UserPatchable: true,
	},
	ColDevAppDate: {
		Setter:        setField(func(a *ActivistExtra, v mysql.NullTime) { a.ApplicationDate = v }),
		DbCol:         "dev_application_date",
		Nullable:      true,
		UserPatchable: false,
	},
	ColDevAppType: {
		Setter:        setField(func(a *ActivistExtra, v string) { a.ApplicationType = v }),
		DbCol:         "dev_application_type",
		UserPatchable: false,
	},
}
