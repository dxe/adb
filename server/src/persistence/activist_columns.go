package persistence

import (
	"fmt"

	"github.com/dxe/adb/model"
)

// activistColumn defines how to select a column, including any joins it requires.
type activistColumn struct {
	// expr is the raw SQL expression (e.g. "a.name", "LOWER(a.email)").
	expr string
	// alias, if non-empty, means SELECT will output "expr as alias".
	// Needed when the expression doesn't naturally produce the expected column name.
	alias string
	joins []joinSpec
}

// selectExpr returns the SQL for the SELECT clause: "expr as alias" if aliased, otherwise just expr.
func (c *activistColumn) selectExpr() string {
	if c.alias != "" {
		return c.expr + " as " + c.alias
	}
	return c.expr
}

var simpleColumns = map[model.ActivistColumnName]activistColumn{
	model.ColID:             {expr: fmt.Sprintf("%s.id", activistTableAlias)},
	model.ColName:           {expr: fmt.Sprintf("%s.name", activistTableAlias)},
	model.ColPreferredName:  {expr: fmt.Sprintf("%s.preferred_name", activistTableAlias)},
	model.ColEmail:          {expr: fmt.Sprintf("LOWER(%s.email)", activistTableAlias), alias: "email"},
	model.ColPhone:          {expr: fmt.Sprintf("%s.phone", activistTableAlias)},
	model.ColPronouns:       {expr: fmt.Sprintf("%s.pronouns", activistTableAlias)},
	model.ColLanguage:       {expr: fmt.Sprintf("%s.language", activistTableAlias)},
	model.ColAccessibility:  {expr: fmt.Sprintf("%s.accessibility", activistTableAlias)},
	model.ColDOB:            {expr: fmt.Sprintf("%s.dob", activistTableAlias)},
	model.ColFacebook:       {expr: fmt.Sprintf("%s.facebook", activistTableAlias)},
	model.ColLocation:       {expr: fmt.Sprintf("%s.location", activistTableAlias)},
	model.ColStreetAddress:  {expr: fmt.Sprintf("%s.street_address", activistTableAlias)},
	model.ColCity:           {expr: fmt.Sprintf("%s.city", activistTableAlias)},
	model.ColState:          {expr: fmt.Sprintf("%s.state", activistTableAlias)},
	model.ColLat:            {expr: fmt.Sprintf("%s.lat", activistTableAlias)},
	model.ColLng:            {expr: fmt.Sprintf("%s.lng", activistTableAlias)},
	model.ColChapterID:      {expr: fmt.Sprintf("%s.chapter_id", activistTableAlias)},
	model.ColActivistLevel:  {expr: fmt.Sprintf("%s.activist_level", activistTableAlias)},
	model.ColSource:         {expr: fmt.Sprintf("%s.source", activistTableAlias)},
	model.ColHiatus:         {expr: fmt.Sprintf("%s.hiatus", activistTableAlias)},
	model.ColConnector:      {expr: fmt.Sprintf("%s.connector", activistTableAlias)},
	model.ColTraining0:      {expr: fmt.Sprintf("%s.training0", activistTableAlias)},
	model.ColTraining1:      {expr: fmt.Sprintf("%s.training1", activistTableAlias)},
	model.ColTraining4:      {expr: fmt.Sprintf("%s.training4", activistTableAlias)},
	model.ColTraining5:      {expr: fmt.Sprintf("%s.training5", activistTableAlias)},
	model.ColTraining6:      {expr: fmt.Sprintf("%s.training6", activistTableAlias)},
	model.ColConsentQuiz:    {expr: fmt.Sprintf("%s.consent_quiz", activistTableAlias)},
	model.ColTrainingProtest: {expr: fmt.Sprintf("%s.training_protest", activistTableAlias)},
	model.ColDevAppDate:     {expr: fmt.Sprintf("%s.dev_application_date", activistTableAlias)},
	model.ColDevAppType:     {expr: fmt.Sprintf("%s.dev_application_type", activistTableAlias)},
	model.ColDevQuiz:        {expr: fmt.Sprintf("%s.dev_quiz", activistTableAlias)},
	model.ColDevInterest:    {expr: fmt.Sprintf("%s.dev_interest", activistTableAlias)},
	model.ColCMFirstEmail:   {expr: fmt.Sprintf("%s.cm_first_email", activistTableAlias)},
	model.ColCMApprovalEmail: {expr: fmt.Sprintf("%s.cm_approval_email", activistTableAlias)},
	model.ColProspectOrganizer:  {expr: fmt.Sprintf("%s.prospect_organizer", activistTableAlias)},
	model.ColProspectChapterMbr: {expr: fmt.Sprintf("%s.prospect_chapter_member", activistTableAlias)},
	model.ColReferralFriends: {expr: fmt.Sprintf("%s.referral_friends", activistTableAlias)},
	model.ColReferralApply:  {expr: fmt.Sprintf("%s.referral_apply", activistTableAlias)},
	model.ColReferralOutlet: {expr: fmt.Sprintf("%s.referral_outlet", activistTableAlias)},
	model.ColInterestDate:   {expr: fmt.Sprintf("%s.interest_date", activistTableAlias)},
	model.ColMPI:            {expr: fmt.Sprintf("%s.mpi", activistTableAlias)},
	model.ColNotes:          {expr: fmt.Sprintf("%s.notes", activistTableAlias)},
	model.ColVisionWall:     {expr: fmt.Sprintf("%s.vision_wall", activistTableAlias)},
	model.ColVotingAgreement: {expr: fmt.Sprintf("%s.voting_agreement", activistTableAlias)},
	model.ColAssignedTo:     {expr: fmt.Sprintf("%s.assigned_to", activistTableAlias)},
	model.ColFollowupDate:   {expr: fmt.Sprintf("DATE_FORMAT(%s.followup_date, '%%Y-%%m-%%d')", activistTableAlias), alias: "followup_date"},
}

func getColumnSpec(colName model.ActivistColumnName) *activistColumn {
	if col, ok := simpleColumns[colName]; ok {
		return &col
	}

	switch colName {
	case model.ColChapterName:
		return &activistColumn{
			joins: []joinSpec{chapterJoin},
			expr:  fmt.Sprintf("COALESCE(%s.name, '<invalid chapter id>')", chapterJoin.Key),
			alias: "chapter_name",
		}
	case model.ColFirstEvent:
		// TODO: rename to first_event_date once legacy activist query is removed
		return &activistColumn{
			joins: []joinSpec{firstEventSubqueryJoin},
			expr:  fmt.Sprintf("%s.first_event_date", firstEventSubqueryJoin.Key),
			alias: "first_event",
		}
	case model.ColFirstEventName:
		return &activistColumn{
			joins: []joinSpec{firstEventSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.event_name, 'n/a')", firstEventSubqueryJoin.Key),
			alias: "first_event_name",
		}
	case model.ColLastEvent:
		// TODO: rename to last_event_date once legacy activist query is removed
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			expr:  fmt.Sprintf("%s.last_event_date", lastEventSubqueryJoin.Key),
			alias: "last_event",
		}
	case model.ColLastEventName:
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.event_name, 'n/a')", lastEventSubqueryJoin.Key),
			alias: "last_event_name",
		}
	case model.ColTotalEvents:
		return &activistColumn{
			joins: []joinSpec{totalEventsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.event_count, 0)", totalEventsSubqueryJoin.Key),
			alias: "total_events",
		}
	case model.ColLastAction:
		return &activistColumn{
			joins: []joinSpec{lastActionSubqueryJoin},
			expr:  fmt.Sprintf("%s.last_action_date", lastActionSubqueryJoin.Key),
			alias: "last_action",
		}
	case model.ColMonthsSinceLastAction:
		return &activistColumn{
			joins: []joinSpec{lastActionSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.months_since_last_action, 9999)", lastActionSubqueryJoin.Key),
			alias: "months_since_last_action",
		}
	case model.ColTotalPoints:
		return &activistColumn{
			joins: []joinSpec{totalPointsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.total_points, 0)", totalPointsSubqueryJoin.Key),
			alias: "total_points",
		}
	case model.ColActive:
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			expr:  fmt.Sprintf("IF(%s.last_event_date >= NOW() - INTERVAL 30 DAY, 1, 0)", lastEventSubqueryJoin.Key),
			alias: "active",
		}
	case model.ColStatus:
		// Must be kept in sync with getStatus() in model/activist.go.
		return &activistColumn{
			joins: []joinSpec{firstEventSubqueryJoin, lastEventSubqueryJoin, totalEventsSubqueryJoin},
			expr: fmt.Sprintf(`CASE
	WHEN %[1]s.first_event_date IS NULL OR %[2]s.last_event_date IS NULL THEN 'No attendance'
	WHEN %[2]s.last_event_date < NOW() - INTERVAL 60 DAY THEN 'Former'
	WHEN %[1]s.first_event_date > NOW() - INTERVAL 90 DAY AND COALESCE(%[3]s.event_count, 0) < 5 THEN 'New'
	ELSE 'Current'
END`, firstEventSubqueryJoin.Key, lastEventSubqueryJoin.Key, totalEventsSubqueryJoin.Key),
			alias: "status",
		}
	case model.ColLastConnection:
		return &activistColumn{
			joins: []joinSpec{lastConnectionSubqueryJoin},
			expr:  fmt.Sprintf("%s.last_connection_date", lastConnectionSubqueryJoin.Key),
			alias: "last_connection",
		}
	case model.ColGeoCircles:
		return &activistColumn{
			joins: []joinSpec{geoCirclesSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.circle_names, '')", geoCirclesSubqueryJoin.Key),
			alias: "geo_circles",
		}
	case model.ColAssignedToName:
		return &activistColumn{
			joins: []joinSpec{assignedToUserJoin},
			expr:  fmt.Sprintf("COALESCE(%s.name, '')", assignedToUserJoin.Key),
			alias: "assigned_to_name",
		}
	case model.ColTotalInteractions:
		return &activistColumn{
			joins: []joinSpec{interactionsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.interaction_count, 0)", interactionsSubqueryJoin.Key),
			alias: "total_interactions",
		}
	case model.ColLastInteractionDate:
		return &activistColumn{
			joins: []joinSpec{interactionsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.last_interaction_date, '')", interactionsSubqueryJoin.Key),
			alias: "last_interaction_date",
		}
	case model.ColMPPRequirements:
		return &activistColumn{
			joins: []joinSpec{mppRequirementsSubqueryJoin},
			expr: fmt.Sprintf(`CASE
	WHEN %[1]s.has_da = 1 AND %[1]s.has_community = 1 THEN 'Fulfilling requirements'
	WHEN %[1]s.has_da = 1 THEN 'Missing Community event'
	WHEN %[1]s.has_community = 1 THEN 'Missing DA event'
	ELSE 'Missing Community & DA events'
END`, mppRequirementsSubqueryJoin.Key),
			alias: "mpp_requirements",
		}
	}

	return nil
}
