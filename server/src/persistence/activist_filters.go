package persistence

import (
	"fmt"
	"time"
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
	Name string
}

func (f *nameFilter) getJoins() []joinSpec {
	return nil
}

func (f *nameFilter) buildWhere() []queryClause {
	if f.Name == "" {
		return nil
	}
	return []queryClause{{
		sql:  fmt.Sprintf("%s.name LIKE ?", activistTableAlias),
		args: []any{"%" + f.Name + "%"},
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
	After  *time.Time
	Before *time.Time
}

func (f *lastEventFilter) getJoins() []joinSpec {
	return []joinSpec{lastEventSubqueryJoin}
}

func (f *lastEventFilter) buildWhere() []queryClause {
	var clauses []queryClause

	if f.After != nil {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s.last_event_date > ?", lastEventSubqueryJoin.Key),
			args: []any{f.After},
		})
	}
	if f.Before != nil {
		clauses = append(clauses, queryClause{
			sql:  fmt.Sprintf("%s.last_event_date < ?", lastEventSubqueryJoin.Key),
			args: []any{f.Before},
		})
	}

	return clauses
}
