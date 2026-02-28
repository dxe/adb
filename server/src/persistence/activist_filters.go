package persistence

import (
	"fmt"

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

// lastEventFilter filters activists by their last event date.
type lastEventFilter struct {
	LastEventGte model.DateOnly
	LastEventLt  model.DateOnly
}

func (f *lastEventFilter) getJoins() []joinSpec {
	return []joinSpec{lastEventSubqueryJoin}
}

func (f *lastEventFilter) buildWhere() []queryClause {
	var clauses []queryClause

	if !f.LastEventGte.IsZero() {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s.last_event_date >= ?", lastEventSubqueryJoin.Key),
			args: []any{f.LastEventGte.Format("2006-01-02")},
		})
	}
	if !f.LastEventLt.IsZero() {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s.last_event_date < ?", lastEventSubqueryJoin.Key),
			args: []any{f.LastEventLt.Format("2006-01-02")},
		})
	}

	return clauses
}
