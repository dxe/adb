package persistence

import "fmt"

// joinSpec represents a SQL join clause
type joinSpec struct {
	// Key is a SQL correlation name (alias of joined table in the query) as well as unique identifier to avoid
	// performing the same join twice.
	Key string
	SQL string
}

// joinRegistry manages a collection of joins, ensuring each join is only added once.
type joinRegistry struct {
	joins map[string]string
}

func newJoinRegistry() *joinRegistry {
	return &joinRegistry{
		joins: make(map[string]string),
	}
}

func (r *joinRegistry) registerJoin(spec joinSpec) {
	if _, exists := r.joins[spec.Key]; !exists {
		r.joins[spec.Key] = spec.SQL
	}
}

func (r *joinRegistry) getJoins() []string {
	joins := make([]string, 0, len(r.joins))
	for _, sql := range r.joins {
		joins = append(joins, sql)
	}
	return joins
}

const (
	firstEventSubqueryKey  = "first_event_subquery"
	lastEventSubqueryKey   = "last_event_subquery"
	totalEventsSubqueryKey = "total_events_subquery"
	chapterKey             = "chapter"
)

var (
	firstEventSubqueryJoin = joinSpec{
		Key: firstEventSubqueryKey,
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT activist_id, first_event_date, event_name
	FROM (
		SELECT
			ea.activist_id,
			e.date as first_event_date,
			e.name as event_name,
			ROW_NUMBER() OVER (PARTITION BY ea.activist_id ORDER BY e.date ASC) as rn
		FROM event_attendance ea
		JOIN events e ON e.id = ea.event_id
	) ranked
	WHERE rn = 1
) %s ON %s.activist_id = %s.id`, firstEventSubqueryKey, firstEventSubqueryKey, activistTableAlias),
	}

	lastEventSubqueryJoin = joinSpec{
		Key: lastEventSubqueryKey,
		// Note: a more efficient query could be used when only last_event_date is needed.
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT activist_id, last_event_date, event_name
	FROM (
		SELECT
			ea.activist_id,
			e.date as last_event_date,
			e.name as event_name,
			ROW_NUMBER() OVER (PARTITION BY ea.activist_id ORDER BY e.date DESC) as rn
		FROM event_attendance ea
		JOIN events e ON e.id = ea.event_id
	) ranked
	WHERE rn = 1
) %s ON %s.activist_id = %s.id`, lastEventSubqueryKey, lastEventSubqueryKey, activistTableAlias),
	}

	totalEventsSubqueryJoin = joinSpec{
		Key: totalEventsSubqueryKey,
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT ea.activist_id, COUNT(DISTINCT ea.event_id) as event_count
	FROM event_attendance ea
	GROUP BY ea.activist_id
) %s ON %s.activist_id = %s.id`, totalEventsSubqueryKey, totalEventsSubqueryKey, activistTableAlias),
	}

	chapterJoin = joinSpec{
		Key: chapterKey,
		SQL: fmt.Sprintf("LEFT JOIN fb_pages %s ON %s.chapter_id = %s.chapter_id", chapterKey, chapterKey, activistTableAlias),
	}
)
