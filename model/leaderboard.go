package model

import (
	"github.com/jmoiron/sqlx"
)

/** Type Definitions */

type LeaderboardActivist struct {
	Name              string `db:"name"`
	FirstEvent        string `db:"first_event"`
	LastEvent         string `db:"last_event"`
	TotalEvents       int    `db:"total_events"`
	TotalEvents30Days int    `db:"total_events_30_days"`
	Points            int    `db:"points"`
}

type LeaderboardActivistJSON struct {
	Name              string `json:"name"`
	FirstEvent        string `json:"first_event"`
	LastEvent         string `json:"last_event"`
	TotalEvents       int    `json:"total_events"`
	TotalEvents30Days int    `json:"total_events_30_days"`
	Points            int    `json:"points"`
}

/** Functions and Methods */

func GetLeaderboardActivistsJSON(db *sqlx.DB) ([]LeaderboardActivistJSON, error) {
	leaderboardActivistsJSON := []LeaderboardActivistJSON{}
	leaderboardActivists, err := GetLeaderboardActivists(db)
	if err != nil {
		return nil, err
	}
	for _, l := range leaderboardActivists {
		leaderboardActivistsJSON = append(leaderboardActivistsJSON, LeaderboardActivistJSON{
			Name:              l.Name,
			FirstEvent:        l.FirstEvent,
			LastEvent:         l.LastEvent,
			TotalEvents:       l.TotalEvents,
			TotalEvents30Days: l.TotalEvents30Days,
			Points:            l.Points,
		})
	}
	return leaderboardActivistsJSON, nil
}

func GetLeaderboardActivists(db *sqlx.DB) ([]LeaderboardActivist, error) {
	query := `
SELECT
  IFNULL(a.name,"") AS name,
  IFNULL(first_event,"None") AS first_event,
  IFNULL(last_event,"None") AS last_event,
  IFNULL(total_events,0) AS total_events,
  IFNULL(total_events_30_days,0) AS total_events_30_days,
  IFNULL((IFNULL(protest_points,0) + IFNULL(wg_points,0) + IFNULL(community_points,0) + IFNULL(outreach_points,0) + IFNULL(sanctuary_points,0) + IFNULL(key_event_points,0)),0) AS points
FROM activists a

LEFT JOIN (
  SELECT ea.activist_id, MIN(e.date) AS "first_event"
  FROM event_attendance ea
  JOIN events e
    ON e.id = ea.event_id
  GROUP BY ea.activist_id
) AS firstevent
  ON a.id = firstevent.activist_id

LEFT JOIN (
  SELECT ea.activist_id, MAX(e.date) AS "last_event"
  FROM event_attendance ea
  JOIN events e
    ON e.id = ea.event_id
  GROUP BY ea.activist_id
) AS lastevent
  ON firstevent.activist_id = lastevent.activist_id

LEFT JOIN (
  SELECT activist_id, COUNT(event_id) AS "total_events"
  FROM event_attendance
  GROUP BY activist_id
) AS total
  ON firstevent.activist_id = total.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id) AS "total_events_30_days"
  FROM event_attendance ea JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
  GROUP BY activist_id
) AS total30
  ON firstevent.activist_id = total30.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*2 AS "protest_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    and e.event_type = "protest"
  GROUP BY activist_id
) AS protest
  ON firstevent.activist_id = protest.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id) AS "wg_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "working group"
  GROUP BY activist_id
) AS wg
  ON firstevent.activist_id = wg.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id) AS "community_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
  AND e.event_type = "community"
  GROUP BY activist_id
) AS community
  ON firstevent.activist_id = community.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*2 AS "outreach_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "outreach"
  GROUP BY activist_id
) AS outreach
  ON firstevent.activist_id = outreach.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*2 AS "sanctuary_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "sanctuary"
  GROUP BY activist_id
) AS sanctuary
  ON firstevent.activist_id = sanctuary.activist_id

LEFT JOIN (
  SELECT ea.activist_id, COUNT(ea.event_id)*3 AS "key_event_points"
  FROM event_attendance ea
  JOIN events e
    ON ea.event_id = e.id
  WHERE
    e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
    AND e.event_type = "key event"
  GROUP BY activist_id
) AS key_event
  ON firstevent.activist_id = key_event.activist_id

WHERE
	total_events_30_days > 0
	AND a.exclude_from_leaderboard <> 1
	AND a.hidden <> 1

ORDER BY points DESC`

	var leaderboardActivists []LeaderboardActivist
	if err := db.Select(&leaderboardActivists, query); err != nil {
		return nil, err
	}

	return leaderboardActivists, nil
}
