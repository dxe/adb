package persistence

import (
	"fmt"

	"github.com/dxe/adb/model"
)

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
	firstEventSubqueryKey      = "first_event_subquery"
	lastEventSubqueryKey       = "last_event_subquery"
	totalEventsSubqueryKey     = "total_events_subquery"
	chapterKey                 = "chapter"
	lastActionSubqueryKey      = "last_action_subquery"
	totalPointsSubqueryKey     = "total_points_subquery"
	lastConnectionSubqueryKey  = "last_connection_subquery"
	geoCirclesSubqueryKey      = "geo_circles_subquery"
	assignedToUserKey          = "assigned_to_user"
	interactionsSubqueryKey    = "interactions_subquery"
	mppRequirementsSubqueryKey = "mpp_requirements_subquery"

	// Community event types used for mpp_requirements column.
	communityEventTypes = "'community', 'training', 'circle'"
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

	lastActionSubqueryJoin = joinSpec{
		Key: lastActionSubqueryKey,
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT
		ea.activist_id,
		MAX(e.date) as last_action_date,
		TIMESTAMPDIFF(MONTH, DATE_FORMAT(MAX(e.date), '%%Y-%%m-01'), NOW()) as months_since_last_action
	FROM event_attendance ea
	JOIN events e ON e.id = ea.event_id
	WHERE LOWER(e.event_type) IN (%s)
	GROUP BY ea.activist_id
) %s ON %s.activist_id = %s.id`, model.ActionEventTypesSQL, lastActionSubqueryKey, lastActionSubqueryKey, activistTableAlias),
	}

	totalPointsSubqueryJoin = joinSpec{
		Key: totalPointsSubqueryKey,
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT ea.activist_id, COUNT(ea.event_id) as total_points
	FROM event_attendance ea
	JOIN events e ON e.id = ea.event_id
	WHERE e.date BETWEEN (NOW() - INTERVAL 30 DAY) AND NOW()
	GROUP BY ea.activist_id
) %s ON %s.activist_id = %s.id`, totalPointsSubqueryKey, totalPointsSubqueryKey, activistTableAlias),
	}

	lastConnectionSubqueryJoin = joinSpec{
		Key: lastConnectionSubqueryKey,
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT ea.activist_id, MAX(e.date) as last_connection_date
	FROM event_attendance ea
	JOIN events e ON e.id = ea.event_id
	WHERE LOWER(e.event_type) = 'connection'
	GROUP BY ea.activist_id
) %s ON %s.activist_id = %s.id`, lastConnectionSubqueryKey, lastConnectionSubqueryKey, activistTableAlias),
	}

	geoCirclesSubqueryJoin = joinSpec{
		Key: geoCirclesSubqueryKey,
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT cm.activist_id, GROUP_CONCAT(c.name) as circle_names
	FROM circle_members cm
	JOIN circles c ON cm.circle_id = c.id
	WHERE c.type = 2
	GROUP BY cm.activist_id
) %s ON %s.activist_id = %s.id`, geoCirclesSubqueryKey, geoCirclesSubqueryKey, activistTableAlias),
	}

	assignedToUserJoin = joinSpec{
		Key: assignedToUserKey,
		SQL: fmt.Sprintf("LEFT JOIN adb_users %s ON %s.id = %s.assigned_to", assignedToUserKey, assignedToUserKey, activistTableAlias),
	}

	interactionsSubqueryJoin = joinSpec{
		Key: interactionsSubqueryKey,
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT
		activist_id,
		COUNT(id) as interaction_count,
		DATE_FORMAT(MAX(timestamp), '%%Y-%%m-%%d') as last_interaction_date
	FROM interactions
	GROUP BY activist_id
) %s ON %s.activist_id = %s.id`, interactionsSubqueryKey, interactionsSubqueryKey, activistTableAlias),
	}

	mppRequirementsSubqueryJoin = joinSpec{
		Key: mppRequirementsSubqueryKey,
		// Performance note: The WHERE clause using YEAR() and MONTH() functions prevents index usage on e.date.
		// If performance becomes a concern, consider replacing with:
		// WHERE e.date >= DATE_FORMAT(NOW(), '%Y-%m-01') AND e.date < DATE_FORMAT(NOW(), '%Y-%m-01') + INTERVAL 1 MONTH
		SQL: fmt.Sprintf(`
LEFT JOIN (
	SELECT
		ea.activist_id,
		MAX(CASE WHEN LOWER(e.event_type) IN (%s) THEN 1 ELSE 0 END) as has_da,
		MAX(CASE WHEN LOWER(e.event_type) IN (%s) THEN 1 ELSE 0 END) as has_community
	FROM event_attendance ea
	JOIN events e ON e.id = ea.event_id
	WHERE YEAR(e.date) = YEAR(NOW()) AND MONTH(e.date) = MONTH(NOW())
	GROUP BY ea.activist_id
) %s ON %s.activist_id = %s.id`, model.ActionEventTypesSQL, communityEventTypes, mppRequirementsSubqueryKey, mppRequirementsSubqueryKey, activistTableAlias),
	}
)
