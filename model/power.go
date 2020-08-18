package model

import (
	"github.com/jmoiron/sqlx"
)

/** Type Definitions */

type PowerHist struct {
	Month int
	Year  int
	Power int
}

/** Functions and Methods */

func GetPower(db *sqlx.DB) (string, error) {
	query := `
SELECT COUNT(id) AS movement_power_index
FROM activists
where mpi = 1
`
	var power string
	if err := db.Get(&power, query); err != nil {
		return "error", err
	}
	return power, nil
}

func GetActiveChapterMembers(db *sqlx.DB) (string, error) {
	query := `
SELECT
	count(id) active_chapter_members
FROM activists
where mpi = 1
and activist_level in ('chapter member','organizer')
`
	var members string
	if err := db.Get(&members, query); err != nil {
		return "error", err
	}
	return members, nil
}
