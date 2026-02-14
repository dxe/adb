package persistence

import (
	"fmt"
	"strings"

	"github.com/dxe/adb/model"
)

// activistColumn defines how to select a column, including any joins it requires.
type activistColumn struct {
	// sql is the SQL expression for this column in the SELECT clause.
	sql   string
	joins []joinSpec
}

// orderByExpr returns the SQL expression to use in ORDER BY and WHERE clauses.
// Strips the " as alias" suffix if present, since ORDER BY/WHERE need the raw expression.
func (c *activistColumn) orderByExpr() string {
	// Extract the part before " as " (case-insensitive).
	if idx := strings.Index(strings.ToLower(c.sql), " as "); idx != -1 {
		return strings.TrimSpace(c.sql[:idx])
	}
	return c.sql
}

var simpleColumns = map[model.ActivistColumnName]string{
	"id":                      fmt.Sprintf("%s.id", activistTableAlias),
	"name":                    fmt.Sprintf("%s.name", activistTableAlias),
	"preferred_name":          fmt.Sprintf("%s.preferred_name", activistTableAlias),
	"email":                   fmt.Sprintf("LOWER(%s.email) as email", activistTableAlias),
	"phone":                   fmt.Sprintf("%s.phone", activistTableAlias),
	"pronouns":                fmt.Sprintf("%s.pronouns", activistTableAlias),
	"language":                fmt.Sprintf("%s.language", activistTableAlias),
	"accessibility":           fmt.Sprintf("%s.accessibility", activistTableAlias),
	"dob":                     fmt.Sprintf("%s.dob", activistTableAlias),
	"facebook":                fmt.Sprintf("%s.facebook", activistTableAlias),
	"location":                fmt.Sprintf("%s.location", activistTableAlias),
	"street_address":          fmt.Sprintf("%s.street_address", activistTableAlias),
	"city":                    fmt.Sprintf("%s.city", activistTableAlias),
	"state":                   fmt.Sprintf("%s.state", activistTableAlias),
	"lat":                     fmt.Sprintf("%s.lat", activistTableAlias),
	"lng":                     fmt.Sprintf("%s.lng", activistTableAlias),
	"chapter_id":              fmt.Sprintf("%s.chapter_id", activistTableAlias),
	"activist_level":          fmt.Sprintf("%s.activist_level", activistTableAlias),
	"source":                  fmt.Sprintf("%s.source", activistTableAlias),
	"hiatus":                  fmt.Sprintf("%s.hiatus", activistTableAlias),
	"connector":               fmt.Sprintf("%s.connector", activistTableAlias),
	"training0":               fmt.Sprintf("%s.training0", activistTableAlias),
	"training1":               fmt.Sprintf("%s.training1", activistTableAlias),
	"training4":               fmt.Sprintf("%s.training4", activistTableAlias),
	"training5":               fmt.Sprintf("%s.training5", activistTableAlias),
	"training6":               fmt.Sprintf("%s.training6", activistTableAlias),
	"consent_quiz":            fmt.Sprintf("%s.consent_quiz", activistTableAlias),
	"training_protest":        fmt.Sprintf("%s.training_protest", activistTableAlias),
	"dev_application_date":    fmt.Sprintf("%s.dev_application_date", activistTableAlias),
	"dev_application_type":    fmt.Sprintf("%s.dev_application_type", activistTableAlias),
	"dev_quiz":                fmt.Sprintf("%s.dev_quiz", activistTableAlias),
	"dev_interest":            fmt.Sprintf("%s.dev_interest", activistTableAlias),
	"cm_first_email":          fmt.Sprintf("%s.cm_first_email", activistTableAlias),
	"cm_approval_email":       fmt.Sprintf("%s.cm_approval_email", activistTableAlias),
	"prospect_organizer":      fmt.Sprintf("%s.prospect_organizer", activistTableAlias),
	"prospect_chapter_member": fmt.Sprintf("%s.prospect_chapter_member", activistTableAlias),
	"referral_friends":        fmt.Sprintf("%s.referral_friends", activistTableAlias),
	"referral_apply":          fmt.Sprintf("%s.referral_apply", activistTableAlias),
	"referral_outlet":         fmt.Sprintf("%s.referral_outlet", activistTableAlias),
	"interest_date":           fmt.Sprintf("%s.interest_date", activistTableAlias),
	"mpi":                     fmt.Sprintf("%s.mpi", activistTableAlias),
	"notes":                   fmt.Sprintf("%s.notes", activistTableAlias),
	"vision_wall":             fmt.Sprintf("%s.vision_wall", activistTableAlias),
	"voting_agreement":        fmt.Sprintf("%s.voting_agreement", activistTableAlias),
	"assigned_to":             fmt.Sprintf("%s.assigned_to", activistTableAlias),
	"followup_date":           fmt.Sprintf("DATE_FORMAT(%s.followup_date, '%%Y-%%m-%%d') as followup_date", activistTableAlias),
}

func getColumnSpec(colName model.ActivistColumnName) *activistColumn {
	if sql, ok := simpleColumns[colName]; ok {
		return &activistColumn{sql: sql}
	}

	switch colName {
	case "chapter_name":
		return &activistColumn{
			joins: []joinSpec{chapterJoin},
			sql:   fmt.Sprintf("%s.name as chapter_name", chapterJoin.Key),
		}
	case "first_event":
		// TODO: rename to first_event_date once legacy activist query is removed
		return &activistColumn{
			joins: []joinSpec{firstEventSubqueryJoin},
			sql:   fmt.Sprintf("%s.first_event_date as first_event", firstEventSubqueryJoin.Key),
		}
	case "first_event_name":
		return &activistColumn{
			joins: []joinSpec{firstEventSubqueryJoin},
			sql:   fmt.Sprintf("COALESCE(%s.event_name, 'n/a') as first_event_name", firstEventSubqueryJoin.Key),
		}
	case "last_event":
		// TODO: rename to last_event_date once legacy activist query is removed
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			sql:   fmt.Sprintf("%s.last_event_date as last_event", lastEventSubqueryJoin.Key),
		}
	case "last_event_name":
		return &activistColumn{
			joins: []joinSpec{lastEventSubqueryJoin},
			sql:   fmt.Sprintf("COALESCE(%s.event_name, 'n/a') as last_event_name", lastEventSubqueryJoin.Key),
		}
	case "total_events":
		return &activistColumn{
			joins: []joinSpec{totalEventsSubqueryJoin},
			sql:   fmt.Sprintf("COALESCE(%s.event_count, 0) as total_events", totalEventsSubqueryJoin.Key),
		}
	}

	// TODO: Implement these columns with proper joins:
	// - last_action
	// - months_since_last_action
	// - total_points
	// - active
	// - status
	// - last_connection
	// - geo_circles
	// - assigned_to_name
	// - total_interactions
	// - last_interaction_date

	return nil
}
