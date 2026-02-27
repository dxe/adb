package persistence

import (
	"fmt"
	"strings"

	"github.com/dxe/adb/model"
)

type filter interface {
	buildWhere() []queryClause
	getJoins() []joinSpec
}

// chapterFilter filters activists by chapter.
type chapterFilter struct {
	ChapterId int
}

func (f *chapterFilter) getJoins() []joinSpec {
	return nil
}

func (f *chapterFilter) buildWhere() []queryClause {
	if f.ChapterId == 0 {
		return nil
	}
	return []queryClause{{
		sql:  fmt.Sprintf("%s.chapter_id = ?", activistTableAlias),
		args: []any{f.ChapterId},
	}}
}

// nameFilter filters activists by name using LIKE.
type nameFilter struct {
	NameContains string
}

func (f *nameFilter) getJoins() []joinSpec {
	return nil
}

func (f *nameFilter) buildWhere() []queryClause {
	if f.NameContains == "" {
		return nil
	}
	return []queryClause{{
		sql:  fmt.Sprintf("%s.name LIKE ?", activistTableAlias),
		args: []any{"%" + f.NameContains + "%"},
	}}
}

// hiddenFilter includes or excludes hidden activists.
type hiddenFilter struct{}

func (f *hiddenFilter) getJoins() []joinSpec {
	return nil
}

func (f *hiddenFilter) buildWhere() []queryClause {
	return []queryClause{{
		sql:  fmt.Sprintf("%s.hidden = false", activistTableAlias),
		args: nil,
	}}
}

// dateRangeFilter filters by a date column with optional NULL inclusion.
type dateRangeFilter struct {
	filter model.DateRangeFilter
	// joinSpec to use for the date column, or nil if on the main table.
	join *joinSpec
	// SQL expression for the date column (e.g. "a.interest_date" or "first_event_subquery.first_event_date").
	expr string
}

func (f *dateRangeFilter) getJoins() []joinSpec {
	if f.join != nil {
		return []joinSpec{*f.join}
	}
	return nil
}

func (f *dateRangeFilter) buildWhere() []queryClause {
	hasGte := !f.filter.Gte.IsZero()
	hasLt := !f.filter.Lt.IsZero()

	// OrNull: no range conditions, just IS NULL.
	if !hasGte && !hasLt && f.filter.OrNull {
		return []queryClause{{
			sql: fmt.Sprintf("%s IS NULL", f.expr),
		}}
	}

	// No OrNull: emit each bound as a separate AND clause.
	if !f.filter.OrNull {
		var clauses []queryClause
		if hasGte {
			clauses = append(clauses, queryClause{
				sql:  fmt.Sprintf("%s >= ?", f.expr),
				args: []any{f.filter.Gte.Format("2006-01-02")},
			})
		}
		if hasLt {
			clauses = append(clauses, queryClause{
				sql:  fmt.Sprintf("%s < ?", f.expr),
				args: []any{f.filter.Lt.Format("2006-01-02")},
			})
		}
		return clauses
	}

	// OrNull with range: ((range conditions) OR expr IS NULL).
	var rangeParts []string
	var args []any
	if hasGte {
		rangeParts = append(rangeParts, fmt.Sprintf("%s >= ?", f.expr))
		args = append(args, f.filter.Gte.Format("2006-01-02"))
	}
	if hasLt {
		rangeParts = append(rangeParts, fmt.Sprintf("%s < ?", f.expr))
		args = append(args, f.filter.Lt.Format("2006-01-02"))
	}
	return []queryClause{{
		sql:  fmt.Sprintf("((%s) OR %s IS NULL)", strings.Join(rangeParts, " AND "), f.expr),
		args: args,
	}}
}

// intRangeFilter filters by an integer column using COALESCE to treat NULL as 0.
type intRangeFilter struct {
	filter model.IntRangeFilter
	// joinSpec to use, or nil if on the main table.
	join *joinSpec
	// SQL expression for the column (will be wrapped in COALESCE).
	expr string
}

func (f *intRangeFilter) getJoins() []joinSpec {
	if f.join != nil {
		return []joinSpec{*f.join}
	}
	return nil
}

func (f *intRangeFilter) buildWhere() []queryClause {
	var clauses []queryClause
	coalesced := fmt.Sprintf("COALESCE(%s, 0)", f.expr)

	if f.filter.Gte != nil {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s >= ?", coalesced),
			args: []any{*f.filter.Gte},
		})
	}
	if f.filter.Lt != nil {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s < ?", coalesced),
			args: []any{*f.filter.Lt},
		})
	}

	return clauses
}

// activistLevelFilter filters by activist_level values.
type activistLevelFilter struct {
	Include []string
	Exclude []string
}

func (f *activistLevelFilter) getJoins() []joinSpec {
	return nil
}

func (f *activistLevelFilter) buildWhere() []queryClause {
	var clauses []queryClause

	if len(f.Include) > 0 {
		placeholders := make([]string, len(f.Include))
		args := make([]any, len(f.Include))
		for i, v := range f.Include {
			placeholders[i] = "?"
			args[i] = v
		}
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s.activist_level IN (%s)", activistTableAlias, strings.Join(placeholders, ",")),
			args: args,
		})
	}

	for _, v := range f.Exclude {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s.activist_level <> ?", activistTableAlias),
			args: []any{v},
		})
	}

	return clauses
}

// sourceFilter filters by the source column using LIKE patterns.
type sourceFilter struct {
	ContainsAny    []string
	NotContainsAny []string
}

func (f *sourceFilter) getJoins() []joinSpec {
	return nil
}

func (f *sourceFilter) buildWhere() []queryClause {
	var clauses []queryClause

	if len(f.ContainsAny) > 0 {
		parts := make([]string, len(f.ContainsAny))
		args := make([]any, len(f.ContainsAny))
		for i, v := range f.ContainsAny {
			parts[i] = fmt.Sprintf("%s.source LIKE ?", activistTableAlias)
			args[i] = "%" + v + "%"
		}
		clauses = append(clauses, queryClause{
			sql:  "(" + strings.Join(parts, " OR ") + ")",
			args: args,
		})
	}

	for _, v := range f.NotContainsAny {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s.source NOT LIKE ?", activistTableAlias),
			args: []any{"%" + v + "%"},
		})
	}

	return clauses
}

// trainingFilter filters by training column completion status.
// Column names are pre-validated by the model layer.
type trainingFilter struct {
	Completed    []string
	NotCompleted []string
}

func (f *trainingFilter) getJoins() []joinSpec {
	return nil
}

func (f *trainingFilter) buildWhere() []queryClause {
	var clauses []queryClause

	for _, col := range f.Completed {
		clauses = append(clauses, queryClause{
			sql: fmt.Sprintf("%s.%s IS NOT NULL", activistTableAlias, col),
		})
	}

	for _, col := range f.NotCompleted {
		clauses = append(clauses, queryClause{
			sql: fmt.Sprintf("%s.%s IS NULL", activistTableAlias, col),
		})
	}

	return clauses
}

// assignedToFilter filters by the assigned_to column.
type assignedToFilter struct {
	AssignedTo int // -1 = any assignee, >0 = specific user ID
}

func (f *assignedToFilter) getJoins() []joinSpec {
	return nil
}

func (f *assignedToFilter) buildWhere() []queryClause {
	if f.AssignedTo == -1 {
		return []queryClause{{
			sql: fmt.Sprintf("%s.assigned_to <> 0", activistTableAlias),
		}}
	}
	return []queryClause{{
		sql:  fmt.Sprintf("%s.assigned_to = ?", activistTableAlias),
		args: []any{f.AssignedTo},
	}}
}

// followupsFilter filters by followup_date.
type followupsFilter struct {
	Followups string // "all", "due", "upcoming"
}

func (f *followupsFilter) getJoins() []joinSpec {
	return nil
}

func (f *followupsFilter) buildWhere() []queryClause {
	switch f.Followups {
	case "all":
		return []queryClause{{
			sql: fmt.Sprintf("%s.followup_date IS NOT NULL", activistTableAlias),
		}}
	case "due":
		return []queryClause{{
			sql: fmt.Sprintf("DATE(%s.followup_date) <= CURRENT_DATE", activistTableAlias),
		}}
	case "upcoming":
		return []queryClause{{
			sql: fmt.Sprintf("DATE(%s.followup_date) > CURRENT_DATE", activistTableAlias),
		}}
	default:
		return nil
	}
}

// prospectFilter filters by prospect flags.
type prospectFilter struct {
	Prospect string // "chapter_member" or "organizer"
}

func (f *prospectFilter) getJoins() []joinSpec {
	return nil
}

func (f *prospectFilter) buildWhere() []queryClause {
	switch f.Prospect {
	case "chapter_member":
		return []queryClause{{
			sql: fmt.Sprintf("%s.prospect_chapter_member = true", activistTableAlias),
		}}
	case "organizer":
		return []queryClause{{
			sql: fmt.Sprintf("%s.prospect_organizer = true", activistTableAlias),
		}}
	default:
		return nil
	}
}
