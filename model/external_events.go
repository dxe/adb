package model

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strconv"
	"time"
)

type FacebookResponseJSON struct {
	Data []FacebookEventJSON `json:"data"`
}

// fb event schema: https://developers.facebook.com/docs/graph-api/reference/event/
type FacebookEventJSON struct {
	ID              string              `json:"id"`
	Name            string              `json:"name"`
	Description     string              `json:"description"`
	StartTime       string              `json:"start_time"`
	EndTime         string              `json:"end_time"`
	AttendingCount  int                 `json:"attending_count"`
	InterestedCount int                 `json:"interested_count"`
	IsCanceled      bool                `json:"is_canceled"`
	IsOnline        bool                `json:"is_online"`
	Place           FacebookPlaceJSON   `json:"place"`
	Cover           FacebookCoverJSON   `json:"cover"`
	EventTimes      []FacebookEventJSON `json:"event_times"`
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

type EventbriteResponseJSON struct {
	Events []EventbriteEventJSON `json:"events"`
}

type EventbriteEventJSON struct {
	ID    string               `json:"id"`
	Name  EventbriteEventName  `json:"name"`
	URL   string               `json:"url"`
	Start EventbriteEventStart `json:"start"`
}

type EventbriteEventName struct {
	Text string `json:"text"`
	HTML string `json:"html"`
}

type EventbriteEventStart struct {
	TimeZone string `json:"timezone"`
	Local    string `json:"local"`
	UTC      string `json:"utc"`
}

type ExternalEventOutput struct {
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
	EventbriteURL   string    `db:"eventbrite_url"`
}

func GetFacebookEvents(db *sqlx.DB, pageID int, startTime string, endTime string) ([]ExternalEventOutput, error) {
	query := `SELECT id, page_id, name, start_time, end_time, location_name,
		location_country, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update, eventbrite_url FROM fb_events`

	query += " WHERE is_canceled = 0 and page_id = " + strconv.Itoa(pageID)

	if startTime != "" {
		query += " and start_time >= '" + startTime + "'"
	}
	if endTime != "" {
		// we actually want to show events which have a START time before the query's end time
		// otherwise really long (or recurring) events could be hidden
		query += " and start_time <= '" + endTime + "'"
	}

	query += " ORDER BY start_time"

	var events []ExternalEventOutput
	err := db.Select(&events, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select events")
	}

	return events, nil
}

func GetOnlineFacebookEvents(db *sqlx.DB, startTime string, endTime string) ([]ExternalEventOutput, error) {
	// TODO: move these page IDs to config variables?
	query := `SELECT id, page_id, name, start_time, end_time, location_name,
		location_country, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update, eventbrite_url FROM fb_events
		WHERE is_canceled = 0 and ((page_id = 1377014279263790 and location_name = 'Online') or page_id = 287332515138353)`

	if startTime != "" {
		query += " and start_time >= '" + startTime + "'"
	}
	if endTime != "" {
		// we actually want to show events which have a START time before the query's end time
		// otherwise really long (or recurring) events could be hidden
		query += " and start_time <= '" + endTime + "'"
	}

	query += " GROUP BY id ORDER BY start_time"

	var events []ExternalEventOutput
	err := db.Select(&events, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select events")
	}

	return events, nil
}

func InsertFacebookEvent(db *sqlx.DB, event FacebookEventJSON, page ChapterWithToken) (err error) {
	// parse fb's datetimes
	fbTimeLayout := "2006-01-02T15:04:05-0700"
	startTime, err := time.Parse(fbTimeLayout, event.StartTime)
	startTime = startTime.UTC()
	endTime, err := time.Parse(fbTimeLayout, event.EndTime)
	endTime = endTime.UTC()
	placeName := event.Place.Name
	if event.IsOnline {
		placeName = "Online"
	}
	// insert into database
	_, err = db.Exec(`REPLACE INTO fb_events (id, page_id, name, description, start_time, end_time,
		location_name, location_city, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update) VALUES
		(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, now())`,
		event.ID, page.ID, event.Name, event.Description, startTime.Format("2006-01-02T15:04:05"),
		endTime.Format("2006-01-02T15:04:05"), placeName, event.Place.Location.City,
		event.Place.Location.Country, event.Place.Location.State, event.Place.Location.Street,
		event.Place.Location.Zip, event.Place.Location.Lat, event.Place.Location.Lng, event.Cover.Source,
		event.AttendingCount, event.InterestedCount, event.IsCanceled)
	if err != nil {
		return errors.Wrap(err, "failed to insert event")
	}
	return nil
}

func AddEventbriteDetailsToEvent(db *sqlx.DB, event EventbriteEventJSON) error {
	_, err := db.NamedExec(`UPDATE fb_events
		SET eventbrite_id = :id, eventbrite_url = :url
		WHERE name = :name.text and left(start_time, 10) = left(:start.utc, 10)`, event)
	if err != nil {
		return errors.Wrap(err, "failed to update event")
	}
	return nil
}
