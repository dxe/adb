package activists

import (
	"database/sql"
	"slices"
	"time"

	"github.com/dxe/adb/pkg/shared"

	"github.com/go-sql-driver/mysql"
)

// ActivistColumnName is a column name in the API layer, not in the database
// (see DbCol in model.ActivistColumns for those).
type ActivistColumnName string

// Column name constants for activist fields. These are used in the API.
// Not to be confused with database column names.
const (
	ColChapterID ActivistColumnName = "chapter_id"
	ColID        ActivistColumnName = "id"

	ColName          ActivistColumnName = "name"
	ColPreferredName ActivistColumnName = "preferred_name"
	ColPronouns      ActivistColumnName = "pronouns"
	ColDOB           ActivistColumnName = "dob"

	ColEmail    ActivistColumnName = "email"
	ColPhone    ActivistColumnName = "phone"
	ColFacebook ActivistColumnName = "facebook"

	ColLanguage      ActivistColumnName = "language"
	ColAccessibility ActivistColumnName = "accessibility"

	ColLocation      ActivistColumnName = "location"
	ColStreetAddress ActivistColumnName = "street_address"
	ColCity          ActivistColumnName = "city"
	ColState         ActivistColumnName = "state"
	ColLat           ActivistColumnName = "lat"
	ColLng           ActivistColumnName = "lng"

	ColActivistLevel   ActivistColumnName = "activist_level"
	ColSource          ActivistColumnName = "source"
	ColHiatus          ActivistColumnName = "hiatus"
	ColConnector       ActivistColumnName = "connector"
	ColTraining0       ActivistColumnName = "training0"
	ColTraining1       ActivistColumnName = "training1"
	ColTraining4       ActivistColumnName = "training4"
	ColTraining5       ActivistColumnName = "training5"
	ColTraining6       ActivistColumnName = "training6"
	ColConsentQuiz     ActivistColumnName = "consent_quiz"
	ColTrainingProtest ActivistColumnName = "training_protest"

	ColDevAppDate  ActivistColumnName = "dev_application_date"
	ColDevAppType  ActivistColumnName = "dev_application_type"
	ColDevQuiz     ActivistColumnName = "dev_quiz"
	ColDevInterest ActivistColumnName = "dev_interest"

	ColCMFirstEmail       ActivistColumnName = "cm_first_email"
	ColCMApprovalEmail    ActivistColumnName = "cm_approval_email"
	ColProspectOrganizer  ActivistColumnName = "prospect_organizer"
	ColProspectChapterMbr ActivistColumnName = "prospect_chapter_member"
	ColReferralFriends    ActivistColumnName = "referral_friends"
	ColReferralApply      ActivistColumnName = "referral_apply"
	ColReferralOutlet     ActivistColumnName = "referral_outlet"
	ColInterestDate       ActivistColumnName = "interest_date"
	ColAssignedTo         ActivistColumnName = "assigned_to"
	ColFollowupDate       ActivistColumnName = "followup_date"

	ColMPI   ActivistColumnName = "mpi"
	ColNotes ActivistColumnName = "notes"

	ColVisionWall      ActivistColumnName = "vision_wall"
	ColVotingAgreement ActivistColumnName = "voting_agreement"

	// Read-only / computed columns (used in SELECT queries but not writable).
	ColChapterName           ActivistColumnName = "chapter_name"
	ColFirstEvent            ActivistColumnName = "first_event"
	ColFirstEventName        ActivistColumnName = "first_event_name"
	ColLastEvent             ActivistColumnName = "last_event"
	ColLastEventName         ActivistColumnName = "last_event_name"
	ColTotalEvents           ActivistColumnName = "total_events"
	ColLastAction            ActivistColumnName = "last_action"
	ColMonthsSinceLastAction ActivistColumnName = "months_since_last_action"
	ColTotalPoints           ActivistColumnName = "total_points"
	ColActive                ActivistColumnName = "active"
	ColStatus                ActivistColumnName = "status"
	ColLastConnection        ActivistColumnName = "last_connection"
	ColGeoCircles            ActivistColumnName = "geo_circles"
	ColAssignedToName        ActivistColumnName = "assigned_to_name"
	ColTotalInteractions     ActivistColumnName = "total_interactions"
	ColLastInteractionDate   ActivistColumnName = "last_interaction_date"
	ColMPPRequirements       ActivistColumnName = "mpp_requirements"
)

type Activist struct {
	Email           string         `db:"email"`
	EmailUpdated    time.Time      `db:"email_updated"`
	Facebook        string         `db:"facebook"`
	Hidden          bool           `db:"hidden"`
	HiddenUpdated   mysql.NullTime `db:"hidden_updated"`
	ID              int            `db:"id"`
	Location        sql.NullString `db:"location"`
	LocationUpdated time.Time      `db:"location_updated"`
	Name            string         `db:"name"`
	NameUpdated     time.Time      `db:"name_updated"`
	PreferredName   string         `db:"preferred_name"`
	Phone           string         `db:"phone"`
	PhoneUpdated    time.Time      `db:"phone_updated"`
	Pronouns        string         `db:"pronouns"`
	Language        string         `db:"language"`
	Accessibility   string         `db:"accessibility"`
	Birthday        sql.NullString `db:"dob"`
	Coords
	ChapterID   int    `db:"chapter_id"`
	ChapterName string `db:"chapter_name"`
}

type Coords struct {
	Lat float64 `db:"lat"`
	Lng float64 `db:"lng"`
}

type ActivistEventData struct {
	FirstEvent            mysql.NullTime `db:"first_event"`
	LastEvent             mysql.NullTime `db:"last_event"`
	FirstEventName        string         `db:"first_event_name"`
	LastEventName         string         `db:"last_event_name"`
	LastAction            mysql.NullTime `db:"last_action"`
	MonthsSinceLastAction int            `db:"months_since_last_action"`
	TotalEvents           int            `db:"total_events"`
	TotalPoints           int            `db:"total_points"`
	Active                bool           `db:"active"`
	// Status is computed in SQL by the "status" column in persistence/activist_columns.go.
	// Must be kept in sync with getStatus() below.
	Status string `db:"status"`
}

type ActivistMembershipData struct {
	ActivistLevel string `db:"activist_level"`
	// May be updated by https://github.com/dxe/dxe-db-functions
	DateOrganizer sql.NullTime `db:"date_organizer"`
	Source        string       `db:"source"`
	Hiatus        bool         `db:"hiatus"`
}

type ActivistConnectionData struct {
	Connector string `db:"connector"`
	// May be updated by https://github.com/dxe/dxe-db-functions
	Training0 sql.NullString `db:"training0"`
	// May be updated by https://github.com/dxe/dxe-db-functions
	Training1       sql.NullString `db:"training1"`
	Training4       sql.NullString `db:"training4"`
	Training5       sql.NullString `db:"training5"`
	Training6       sql.NullString `db:"training6"`
	ConsentQuiz     sql.NullString `db:"consent_quiz"`
	TrainingProtest sql.NullString `db:"training_protest"`
	ApplicationDate mysql.NullTime `db:"dev_application_date"`
	ApplicationType string         `db:"dev_application_type"`
	Quiz            sql.NullString `db:"dev_quiz"`
	DevInterest     string         `db:"dev_interest"`

	CMFirstEmail          sql.NullString `db:"cm_first_email"`
	CMApprovalEmail       sql.NullString `db:"cm_approval_email"`
	ProspectOrganizer     bool           `db:"prospect_organizer"`
	ProspectChapterMember bool           `db:"prospect_chapter_member"`
	LastConnection        sql.NullString `db:"last_connection"`
	ReferralFriends       string         `db:"referral_friends"`
	ReferralApply         string         `db:"referral_apply"`
	ReferralOutlet        string         `db:"referral_outlet"`
	InterestDate          sql.NullString `db:"interest_date"`
	MPI                   bool           `db:"mpi"`
	Notes                 sql.NullString `db:"notes"`
	VisionWall            string         `db:"vision_wall"`
	MPPRequirements       string         `db:"mpp_requirements"`
	VotingAgreement       bool           `db:"voting_agreement"`
	ActivistAddress
	AddressUpdated      time.Time      `db:"address_updated"`
	GeoCircles          string         `db:"geo_circles"`
	AssignedTo          int            `db:"assigned_to"`
	AssignedToName      string         `db:"assigned_to_name"`
	FollowupDate        sql.NullString `db:"followup_date"`
	TotalInteractions   int            `db:"total_interactions"`
	LastInteractionDate string         `db:"last_interaction_date"`
}

type ActivistAddress struct {
	StreetAddress string `db:"street_address"`
	City          string `db:"city"`
	State         string `db:"state"`
}

type ActivistExtra struct {
	Activist
	ActivistEventData
	ActivistMembershipData
	ActivistConnectionData
}

type QueryActivistOptions struct {
	// This model is currently shared with the transport layer and treated as part of the frontend API.
	// Introduce transport DTOs when the wire format needs to differ from internal semantics.

	Shape QueryActivistShape `json:"shape"`

	// Cursor pointing to last item in previous page (base 64 encoding of values of sort columns and ID).
	// Must be a value returned by QueryActivistResultPagination.NextCursor.
	// If empty, the first page of results will be returned.
	// If invalid, an error is returned.
	After string `json:"after"`
}

// QueryActivistShape is the query-shape portion of an activist query: which
// columns to return, which rows to include, and what order.
type QueryActivistShape struct {
	Columns []ActivistColumnName `json:"columns"`
	Filters QueryActivistFilters `json:"filters"`
	Sort    ActivistSortOptions  `json:"sort"`
}

type ActivistSortOptions struct {
	SortColumns []ActivistSortColumn `json:"sort_columns"`
}

type ActivistSortColumn struct {
	ColumnName ActivistColumnName `json:"column_name"`
	Desc       bool               `json:"desc"`
}

type QueryActivistResult struct {
	Activists  []ActivistExtra               `json:"activists"`
	Pagination QueryActivistResultPagination `json:"pagination"`
}

type QueryActivistResultPagination struct {
	// An opaque string if more results are available; otherwise, the empty string.
	NextCursor string `json:"next_cursor"`
}

func (s *QueryActivistShape) NormalizeAndValidate() error {
	// TODO: remove invalid characters from s.Filters.Name.NameContains

	if len(s.Columns) == 0 {
		return shared.ValidationErrorf("must request at least one column")
	}

	if s.Filters.ChapterId == 0 && !slices.Contains(s.Columns, ColChapterName) {
		return shared.ValidationErrorf("must choose 'chapter_name' column when not filtering by chapter ID.")
	}

	if err := s.Filters.Validate(); err != nil {
		return err
	}

	if len(s.Sort.SortColumns) > 2 {
		return shared.ValidationErrorf("cannot sort by more than 2 columns")
	}

	for i, sc := range s.Sort.SortColumns {
		if sc.ColumnName == ColID {
			if i != len(s.Sort.SortColumns)-1 {
				return shared.ValidationErrorf("'id' must be the last sort column if present")
			}
			if sc.Desc {
				return shared.ValidationErrorf("'id' cannot be sorted in descending order")
			}
		}
	}

	return nil
}
