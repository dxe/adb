package model

import (
	"strconv"
	"strings"
	"time"

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
SELECT COUNT(*) AS movement_power_index
FROM (
	SELECT
		activist_id,
		MAX(CASE WHEN event_type = "protest" or event_type = "key event" or event_type = "outreach" or event_type = "sanctuary" THEN "1" ELSE "0" END) AS is_protest,
	    MAX(CASE WHEN event_type = "community" THEN "1" ELSE "0" END) AS is_community
	FROM event_attendance ea
	JOIN events e ON ea.event_id = e.id
	JOIN activists a ON ea.activist_id = a.id
	WHERE e.date BETWEEN DATE_SUB(NOW(), INTERVAL 30 DAY) AND NOW()
		AND a.hidden <> 1
	GROUP BY activist_id
	HAVING is_protest = "1" AND is_community = "1"
) AS power_index
`
	var power string
	if err := db.Get(&power, query); err != nil {
		return "error", err
	}
	return power, nil
}

func GetPowerMTD(db *sqlx.DB) (int, error) {
	current_time := time.Now().Local()
	current_time_string := current_time.Format("2006-01")
	split_date := strings.Split(current_time_string, "-")

	year := split_date[0]
	month := split_date[1]

	query := `
SELECT COUNT(*) AS movement_power_index
FROM (
	SELECT
		activist_id,
		MAX(CASE WHEN event_type = "protest" or event_type = "key event" or event_type = "outreach" or event_type = "sanctuary" THEN "1" ELSE "0" END) AS is_protest,
	    MAX(CASE WHEN event_type = "community" THEN "1" ELSE "0" END) AS is_community,
        SUBSTR(e.date,1,4) AS year,
        SUBSTR(e.date,6,2) AS month
	FROM event_attendance ea
	JOIN events e ON ea.event_id = e.id
	JOIN activists a ON ea.activist_id = a.id
	WHERE a.hidden <> 1
	GROUP BY activist_id, year, month
	HAVING is_protest = "1" AND is_community = "1" AND month = "` + month + `" AND year = "` + year + `"
) AS power_index
`
	var powerMTD int
	if err := db.Get(&powerMTD, query); err != nil {
		return 0, err
	}
	return powerMTD, nil
}

func GetPowerHistArray(db *sqlx.DB) ([]PowerHist, error) {
	current_time := time.Now().Local()
	current_time_string := current_time.Format("2006-01")
	split_date := strings.Split(current_time_string, "-")

	year := split_date[0]
	month := split_date[1]

	year_int, error := strconv.Atoi(year)
	if error != nil {
		return nil, error
	}
	month_int, error := strconv.Atoi(month)
	if error != nil {
		return nil, error
	}

	for i := 0; i < 12; i++ {
		if month_int == 1 {
			month_int = 12
			year_int -= 1
		} else {
			month_int -= 1
		}
	}

	var history []PowerHist

	for i := 0; i < 12; i++ {
		power, error := GetPowerHist(db, month_int, year_int)
		if error != nil {
			return nil, error
		}
		history = append(history, PowerHist{
			Month: month_int,
			Year:  year_int,
			Power: power,
		})
		if month_int == 12 {
			month_int = 1
			year_int += 1
		} else {
			month_int += 1
		}
	}
	return history, nil
}

func GetPowerHist(db *sqlx.DB, month int, year int) (int, error) {
	month_string := strconv.Itoa(month)
	if month < 10 {
		month_string = "0" + month_string
	}
	year_string := strconv.Itoa(year)
	query := `
SELECT COUNT(*) AS movement_power_index
FROM (
	SELECT
		activist_id,
		MAX(CASE WHEN event_type = "protest" or event_type = "key event" or event_type = "sanctuary" or event_type = "outreach" THEN "1" ELSE "0" END) AS is_protest,
	    MAX(CASE WHEN event_type = "community" THEN "1" ELSE "0" END) AS is_community,
        SUBSTR(e.date,1,4) AS year,
        SUBSTR(e.date,6,2) AS month
	FROM event_attendance ea
	JOIN events e ON ea.event_id = e.id
	JOIN activists a ON ea.activist_id = a.id
	WHERE a.hidden <> 1
	GROUP BY activist_id, year, month
	HAVING is_protest = "1" AND is_community = "1" AND month = "` + month_string + `" AND year = "` + year_string + `"
) AS power_index
`
	var power int
	if err := db.Get(&power, query); err != nil {
		return 0, err
	}
	return power, nil
}
