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
	"id":                      {expr: fmt.Sprintf("%s.id", activistTableAlias)},
	"name":                    {expr: fmt.Sprintf("%s.name", activistTableAlias)},
	"preferred_name":          {expr: fmt.Sprintf("%s.preferred_name", activistTableAlias)},
	"email":                   {expr: fmt.Sprintf("LOWER(%s.email)", activistTableAlias), alias: "email"},
	"phone":                   {expr: fmt.Sprintf("%s.phone", activistTableAlias)},
	"pronouns":                {expr: fmt.Sprintf("%s.pronouns", activistTableAlias)},
	"language":                {expr: fmt.Sprintf("%s.language", activistTableAlias)},
	"accessibility":           {expr: fmt.Sprintf("%s.accessibility", activistTableAlias)},
	"dob":                     {expr: fmt.Sprintf("%s.dob", activistTableAlias)},
	"facebook":                {expr: fmt.Sprintf("%s.facebook", activistTableAlias)},
	"location":                {expr: fmt.Sprintf("%s.location", activistTableAlias)},
	"street_address":          {expr: fmt.Sprintf("%s.street_address", activistTableAlias)},
	"city":                    {expr: fmt.Sprintf("%s.city", activistTableAlias)},
	"state":                   {expr: fmt.Sprintf("%s.state", activistTableAlias)},
	"lat":                     {expr: fmt.Sprintf("%s.lat", activistTableAlias)},
	"lng":                     {expr: fmt.Sprintf("%s.lng", activistTableAlias)},
	"chapter_id":              {expr: fmt.Sprintf("%s.chapter_id", activistTableAlias)},
	"activist_level":          {expr: fmt.Sprintf("%s.activist_level", activistTableAlias)},
	"source":                  {expr: fmt.Sprintf("%s.source", activistTableAlias)},
	"hiatus":                  {expr: fmt.Sprintf("%s.hiatus", activistTableAlias)},
	"connector":               {expr: fmt.Sprintf("%s.connector", activistTableAlias)},
	"training0":               {expr: fmt.Sprintf("%s.training0", activistTableAlias)},
	"training1":               {expr: fmt.Sprintf("%s.training1", activistTableAlias)},
	"training4":               {expr: fmt.Sprintf("%s.training4", activistTableAlias)},
	"training5":               {expr: fmt.Sprintf("%s.training5", activistTableAlias)},
	"training6":               {expr: fmt.Sprintf("%s.training6", activistTableAlias)},
	"consent_quiz":            {expr: fmt.Sprintf("%s.consent_quiz", activistTableAlias)},
	"training_protest":        {expr: fmt.Sprintf("%s.training_protest", activistTableAlias)},
	"dev_application_date":    {expr: fmt.Sprintf("%s.dev_application_date", activistTableAlias)},
	"dev_application_type":    {expr: fmt.Sprintf("%s.dev_application_type", activistTableAlias)},
	"dev_quiz":                {expr: fmt.Sprintf("%s.dev_quiz", activistTableAlias)},
	"dev_interest":            {expr: fmt.Sprintf("%s.dev_interest", activistTableAlias)},
	"cm_first_email":          {expr: fmt.Sprintf("%s.cm_first_email", activistTableAlias)},
	"cm_approval_email":       {expr: fmt.Sprintf("%s.cm_approval_email", activistTableAlias)},
	"prospect_organizer":      {expr: fmt.Sprintf("%s.prospect_organizer", activistTableAlias)},
	"prospect_chapter_member": {expr: fmt.Sprintf("%s.prospect_chapter_member", activistTableAlias)},
	"referral_friends":        {expr: fmt.Sprintf("%s.referral_friends", activistTableAlias)},
	"referral_apply":          {expr: fmt.Sprintf("%s.referral_apply", activistTableAlias)},
	"referral_outlet":         {expr: fmt.Sprintf("%s.referral_outlet", activistTableAlias)},
	"interest_date":           {expr: fmt.Sprintf("%s.interest_date", activistTableAlias)},
	"mpi":                     {expr: fmt.Sprintf("%s.mpi", activistTableAlias)},
	"notes":                   {expr: fmt.Sprintf("%s.notes", activistTableAlias)},
	"vision_wall":             {expr: fmt.Sprintf("%s.vision_wall", activistTableAlias)},
	"voting_agreement":        {expr: fmt.Sprintf("%s.voting_agreement", activistTableAlias)},
	"assigned_to":             {expr: fmt.Sprintf("%s.assigned_to", activistTableAlias)},
	"followup_date":           {expr: fmt.Sprintf("DATE_FORMAT(%s.followup_date, '%%Y-%%m-%%d')", activistTableAlias), alias: "followup_date"},
}

func getColumnSpec(colName model.ActivistColumnName) *activistColumn {
	if col, ok := simpleColumns[colName]; ok {
		return &col
	}

	switch colName {
	case "chapter_name":
		return &activistColumn{
			joins: []joinSpec{chapterJoin},
			expr:  fmt.Sprintf("COALESCE(%s.name, '<invalid chapter id>')", chapterJoin.Key),
			alias: "chapter_name",
		}
	case "first_event":
		// TODO: rename to first_event_date once legacy activist query is removed
		return &activistColumn{
			joins: []joinSpec{firstEventSubqueryJoin},
			expr:  fmt.Sprintf("%s.first_event_date", firstEventSubqueryJoin.Key),
			alias: "first_event",
		}
	case "first_event_name":
		return &activistColumn{
			joins: []joinSpec{firstEventSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.event_name, 'n/a')", firstEventSubqueryJoin.Key),
			alias: "first_event_name",
		}
	case "last_event":
		// TODO: rename to last_event_date once legacy activist query is removed
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			expr:  fmt.Sprintf("%s.last_event_date", lastEventSubqueryJoin.Key),
			alias: "last_event",
		}
	case "last_event_name":
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.event_name, 'n/a')", lastEventSubqueryJoin.Key),
			alias: "last_event_name",
		}
	case "total_events":
		return &activistColumn{
			joins: []joinSpec{totalEventsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.event_count, 0)", totalEventsSubqueryJoin.Key),
			alias: "total_events",
		}
	case "last_action":
		return &activistColumn{
			joins: []joinSpec{lastActionSubqueryJoin},
			expr:  fmt.Sprintf("%s.last_action_date", lastActionSubqueryJoin.Key),
			alias: "last_action",
		}
	case "months_since_last_action":
		return &activistColumn{
			joins: []joinSpec{lastActionSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.months_since_last_action, 9999)", lastActionSubqueryJoin.Key),
			alias: "months_since_last_action",
		}
	case "total_points":
		return &activistColumn{
			joins: []joinSpec{totalPointsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.total_points, 0)", totalPointsSubqueryJoin.Key),
			alias: "total_points",
		}
	case "active":
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			expr:  fmt.Sprintf("IF(%s.last_event_date >= NOW() - INTERVAL 30 DAY, 1, 0)", lastEventSubqueryJoin.Key),
			alias: "active",
		}
	case "status":
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
	case "last_connection":
		return &activistColumn{
			joins: []joinSpec{lastConnectionSubqueryJoin},
			expr:  fmt.Sprintf("%s.last_connection_date", lastConnectionSubqueryJoin.Key),
			alias: "last_connection",
		}
	case "geo_circles":
		return &activistColumn{
			joins: []joinSpec{geoCirclesSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.circle_names, '')", geoCirclesSubqueryJoin.Key),
			alias: "geo_circles",
		}
	case "assigned_to_name":
		return &activistColumn{
			joins: []joinSpec{assignedToUserJoin},
			expr:  fmt.Sprintf("COALESCE(%s.name, '')", assignedToUserJoin.Key),
			alias: "assigned_to_name",
		}
	case "total_interactions":
		return &activistColumn{
			joins: []joinSpec{interactionsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.interaction_count, 0)", interactionsSubqueryJoin.Key),
			alias: "total_interactions",
		}
	case "last_interaction_date":
		return &activistColumn{
			joins: []joinSpec{interactionsSubqueryJoin},
			expr:  fmt.Sprintf("COALESCE(%s.last_interaction_date, '')", interactionsSubqueryJoin.Key),
			alias: "last_interaction_date",
		}
	case "mpp_requirements":
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
