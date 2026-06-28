package model

// This file re-exports the activist query slice that now lives in
// github.com/dxe/adb/pkg/activists so it can be shared with standalone jobs
// (e.g. AWS Lambdas) without depending on the server module. The canonical
// definitions live in pkg/activists; these aliases keep the existing
// model.* API stable for the server, transport, and persistence layers.
//
// Error sentinels (ErrNotFound, ErrValidation, ValidationErrorf) live in
// pkg/shared so errors.Is(err, model.ErrValidation) keeps working across the
// model, transport, and activists layers (they share one sentinel value).

import (
	"github.com/dxe/adb/pkg/activists"
	"github.com/dxe/adb/pkg/shared"
)

// Cross-cutting error sentinels (defined in pkg/shared).
type ValidationError = shared.ValidationError

var (
	ErrNotFound      = shared.ErrNotFound
	ErrValidation    = shared.ErrValidation
	ValidationErrorf = shared.ValidationErrorf
)

// ActionEventTypesSQL is the SQL list of "action" event types (defined in pkg/shared).
const ActionEventTypesSQL = shared.ActionEventTypesSQL

// Row types.
type (
	Activist               = activists.Activist
	Coords                 = activists.Coords
	ActivistEventData      = activists.ActivistEventData
	ActivistMembershipData = activists.ActivistMembershipData
	ActivistConnectionData = activists.ActivistConnectionData
	ActivistAddress        = activists.ActivistAddress
	ActivistExtra          = activists.ActivistExtra
)

// Column metadata.
type (
	ActivistColumnName = activists.ActivistColumnName
	ActivistColumnInfo = activists.ActivistColumnInfo
)

var ActivistColumns = activists.ActivistColumns

// Query option/shape/result types.
type (
	QueryActivistOptions          = activists.QueryActivistOptions
	QueryActivistShape            = activists.QueryActivistShape
	QueryActivistFilters          = activists.QueryActivistFilters
	ActivistSortOptions           = activists.ActivistSortOptions
	ActivistSortColumn            = activists.ActivistSortColumn
	QueryActivistResult           = activists.QueryActivistResult
	QueryActivistResultPagination = activists.QueryActivistResultPagination
)

// Filter types and helpers.
type (
	DateOnly            = activists.DateOnly
	DateRangeFilter     = activists.DateRangeFilter
	IntRangeFilter      = activists.IntRangeFilter
	NameFilter          = activists.NameFilter
	ActivistLevelFilter = activists.ActivistLevelFilter
	SourceFilter        = activists.SourceFilter
	TrainingFilter      = activists.TrainingFilter
)

var (
	ValidActivistLevels  = activists.ValidActivistLevels
	ValidTrainingColumns = activists.ValidTrainingColumns
)

const (
	ProspectFilterChapterMember = activists.ProspectFilterChapterMember
	ProspectFilterOrganizer     = activists.ProspectFilterOrganizer
)

// Activist API column-name constants.
const (
	ColChapterID             = activists.ColChapterID
	ColID                    = activists.ColID
	ColName                  = activists.ColName
	ColPreferredName         = activists.ColPreferredName
	ColPronouns              = activists.ColPronouns
	ColDOB                   = activists.ColDOB
	ColEmail                 = activists.ColEmail
	ColPhone                 = activists.ColPhone
	ColFacebook              = activists.ColFacebook
	ColLanguage              = activists.ColLanguage
	ColAccessibility         = activists.ColAccessibility
	ColLocation              = activists.ColLocation
	ColStreetAddress         = activists.ColStreetAddress
	ColCity                  = activists.ColCity
	ColState                 = activists.ColState
	ColLat                   = activists.ColLat
	ColLng                   = activists.ColLng
	ColActivistLevel         = activists.ColActivistLevel
	ColSource                = activists.ColSource
	ColHiatus                = activists.ColHiatus
	ColConnector             = activists.ColConnector
	ColTraining0             = activists.ColTraining0
	ColTraining1             = activists.ColTraining1
	ColTraining4             = activists.ColTraining4
	ColTraining5             = activists.ColTraining5
	ColTraining6             = activists.ColTraining6
	ColConsentQuiz           = activists.ColConsentQuiz
	ColTrainingProtest       = activists.ColTrainingProtest
	ColDevAppDate            = activists.ColDevAppDate
	ColDevAppType            = activists.ColDevAppType
	ColDevQuiz               = activists.ColDevQuiz
	ColDevInterest           = activists.ColDevInterest
	ColCMFirstEmail          = activists.ColCMFirstEmail
	ColCMApprovalEmail       = activists.ColCMApprovalEmail
	ColProspectOrganizer     = activists.ColProspectOrganizer
	ColProspectChapterMbr    = activists.ColProspectChapterMbr
	ColReferralFriends       = activists.ColReferralFriends
	ColReferralApply         = activists.ColReferralApply
	ColReferralOutlet        = activists.ColReferralOutlet
	ColInterestDate          = activists.ColInterestDate
	ColAssignedTo            = activists.ColAssignedTo
	ColFollowupDate          = activists.ColFollowupDate
	ColMPI                   = activists.ColMPI
	ColNotes                 = activists.ColNotes
	ColVisionWall            = activists.ColVisionWall
	ColVotingAgreement       = activists.ColVotingAgreement
	ColChapterName           = activists.ColChapterName
	ColFirstEvent            = activists.ColFirstEvent
	ColFirstEventName        = activists.ColFirstEventName
	ColLastEvent             = activists.ColLastEvent
	ColLastEventName         = activists.ColLastEventName
	ColTotalEvents           = activists.ColTotalEvents
	ColLastAction            = activists.ColLastAction
	ColMonthsSinceLastAction = activists.ColMonthsSinceLastAction
	ColTotalPoints           = activists.ColTotalPoints
	ColActive                = activists.ColActive
	ColStatus                = activists.ColStatus
	ColLastConnection        = activists.ColLastConnection
	ColGeoCircles            = activists.ColGeoCircles
	ColAssignedToName        = activists.ColAssignedToName
	ColTotalInteractions     = activists.ColTotalInteractions
	ColLastInteractionDate   = activists.ColLastInteractionDate
	ColMPPRequirements       = activists.ColMPPRequirements
)
