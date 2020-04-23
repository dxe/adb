package model

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strings"
	"time"
)

type FacebookPage struct {
	ID    int     `db:"id"`
	Name  string  `db:"name"`
	Lat   float64 `db:"lat"`
	Lng   float64 `db:"lng"`
	Token string  `db:"token"`
}

type FacebookResponseJSON struct {
	Data []FacebookEventJSON `json:"data"`
}

// fb event schema: https://developers.facebook.com/docs/graph-api/reference/event/
type FacebookEventJSON struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	StartTime       string            `json:"start_time"`
	EndTime         string            `json:"end_time"`
	AttendingCount  int               `json:"attending_count"`
	InterestedCount int               `json:"interested_count"`
	IsCanceled      bool              `json:"is_canceled"`
	Place           FacebookPlaceJSON `json:"place"`
	Cover           FacebookCoverJSON `json:"cover"`
}

type FacebookPlaceJSON struct {
	Name     string               `json:"name"`
	Location FacebookLocationJSON `json:"location"`
}

type FacebookLocationJSON struct {
	City    string  `json:"city"`
	State   string  `json:"state"`
	Country string  `json:"country"`
	Street  string  `json:"street"`
	Zip     string  `json:"zip"`
	Lat     float64 `json:"latitude"`
	Lng     float64 `json:"longitude"`
}

type FacebookCoverJSON struct {
	Source string `json:"source"`
}

type FacebookEventOutput struct {
	ID              int       `db:"id"`
	PageID          int       `db:"page_id"`
	Name            string    `db:"name"`
	Description     string    `db:"description"`
	StartTime       time.Time `db:"start_time"`
	EndTime         time.Time `db:"end_time"`
	LocationName    string    `db:"location_name"`
	LocationCity    string    `db:"location_city"`
	LocationCountry string    `db:"location_country"`
	LocationState   string    `db:"location_state"`
	LocationAddress string    `db:"location_address"`
	LocationZip     string    `db:"location_zip"`
	Lat             float64   `db:"lat"`
	Lng             float64   `db:"lng"`
	Cover           string    `db:"cover"`
	AttendingCount  int       `db:"attending_count"`
	InterestedCount int       `db:"interested_count"`
	IsCanceled      bool      `db:"is_canceled"`
	LastUpdate      time.Time `db:"last_update"`
}

func GetFacebookPages(db *sqlx.DB) ([]FacebookPage, error) {
	query := `SELECT id, name, lat, lng, token FROM fb_pages`

	var pages []FacebookPage
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}

	return pages, nil
}

func GetFacebookEvents(db *sqlx.DB, pageID int, startTime string, endTime string) ([]FacebookEventOutput, error) {
	query := `SELECT id, page_id, name, description, start_time, end_time, location_name,
		location_country, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update FROM fb_events`

	whereClause := []string{}
	if startTime != "" {
		whereClause = append(whereClause, "start_time >= '"+startTime+"'")
	}
	if endTime != "" {
		whereClause = append(whereClause, "end_time <= '"+endTime+"'")
	}
	if len(whereClause) != 0 {
		query += " WHERE " + strings.Join(whereClause, " AND ")
	}

	query += " ORDER BY start_time"

	var events []FacebookEventOutput
	err := db.Select(&events, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select events")
	}

	return events, nil
}

func InsertFacebookEvent(db *sqlx.DB, event FacebookEventJSON, page FacebookPage) (err error) {
	// parse fb's datetimes
	fbTimeLayout := "2006-01-02T15:04:05-0700"
	startTime, err := time.Parse(fbTimeLayout, event.StartTime)
	startTime = startTime.UTC()
	endTime, err := time.Parse(fbTimeLayout, event.EndTime)
	endTime = endTime.UTC()
	// insert into database
	_, err = db.Exec(`REPLACE INTO fb_events (id, page_id, name, description, start_time, end_time,
		location_name, location_city, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update) VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())`,
		event.ID, page.ID, event.Name, event.Description, startTime.Format("2006-01-02T15:04:05"),
		endTime.Format("2006-01-02T15:04:05"), event.Place.Name, event.Place.Location.City,
		event.Place.Location.Country, event.Place.Location.State, event.Place.Location.Street,
		event.Place.Location.Zip, event.Place.Location.Lat, event.Place.Location.Lng, event.Cover.Source,
		event.AttendingCount, event.InterestedCount, event.IsCanceled)
	if err != nil {
		return errors.Wrap(err, "failed to insert event")
	}
	return nil
}