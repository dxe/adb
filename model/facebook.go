package model

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"strconv"
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

// TODO: replace this with a new struct called "ChapterWithToken" that is similar to the Chapter struct (or maybe just have one struct using "omitempty")
type FacebookPageOutput struct {
	ID         int     `db:"id"`
	ChapterID  int     `db:"chapter_id"`
	Name       string  `db:"name"`
	Flag       string  `db:"flag"`
	FbURL      string  `db:"fb_url"`
	TwitterURL string  `db:"twitter_url"`
	InstaURL   string  `db:"insta_url"`
	Email      string  `db:"email"`
	Region     string  `db:"region"`
	Lat        float64 `db:"lat"`
	Lng        float64 `db:"lng"`
	Distance   float32 `db:"distance"`
	Token      string  `db:"token"`
	LastUpdate string  `db:"last_update"`
}

// used by public API
type Chapter struct {
	ID         int     `db:"chapter_id"`
	FacebookID int     `db:"id"`
	Name       string  `db:"name"`
	Flag       string  `db:"flag"`
	FbURL      string  `db:"fb_url"`
	TwitterURL string  `db:"twitter_url"`
	InstaURL   string  `db:"insta_url"`
	Email      string  `db:"email"`
	Region     string  `db:"region"`
	Lat        float64 `db:"lat"`
	Lng        float64 `db:"lng"`
}

// used for making api calls to facebook, not for responding to our api requests
func GetFacebookPagesWithTokens(db *sqlx.DB) ([]FacebookPage, error) {
	query := `SELECT id, name, lat, lng, token FROM fb_pages WHERE token <> '' and id <> 0`

	var pages []FacebookPage
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}

	return pages, nil
}

// for the chapter management admin page on the ADB itself -- NOTE THAT THIS RETURNS TOKENS, SO IT SHOULD NOT BE MADE PUBLIC
func GetAllChapters(db *sqlx.DB) ([]FacebookPageOutput, error) {
	query := `SELECT fb_pages.id, chapter_id, fb_pages.name, flag, fb_url, twitter_url, insta_url, email, region, fb_pages.lat, fb_pages.lng, token, IFNULL(MAX(last_update),'') as last_update
		FROM fb_pages
		LEFT JOIN fb_events on fb_pages.id = fb_events.page_id
		GROUP BY fb_pages.chapter_id
		ORDER BY name`
	var pages []FacebookPageOutput
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}
	return pages, nil
}

// for the chapter management admin page on the ADB itself
func GetChapterByID(db *sqlx.DB, id int) (FacebookPageOutput, error) {
	query := `SELECT id, chapter_id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng, token
		FROM fb_pages
		WHERE chapter_id = ?`
	var pages []FacebookPageOutput
	err := db.Select(&pages, query, id)
	if err != nil {
		return FacebookPageOutput{}, errors.Wrap(err, "failed to select page")
	}
	if len(pages) == 0 {
		return FacebookPageOutput{}, errors.New("Could not find page")
	}
	if len(pages) > 1 {
		return FacebookPageOutput{}, errors.New("Found too many pages")
	}
	return pages[0], nil
}

// for the chapter management admin page on the ADB itself
func UpdateChapter(db *sqlx.DB, page FacebookPageOutput) error {
	_, err := db.NamedExec(`UPDATE fb_pages
		SET id = :id,
		name = :name,
		flag = :flag,
		fb_url = :fb_url,
		insta_url = :insta_url,
		twitter_url = :twitter_url,
		email = :email,
		region = :region,
		lat = :lat,
		lng = :lng,
		token = :token
		WHERE chapter_id = :chapter_id`, page)
	if err != nil {
		return errors.Wrapf(err, "failed to update chapter %d", page.ID)
	}
	return nil
}

// for the chapter management admin page on the ADB itself
func DeleteChapter(db *sqlx.DB, page FacebookPageOutput) error {
	_, err := db.NamedExec(`DELETE FROM fb_pages
		WHERE chapter_id = :chapter_id`, page)
	if err != nil {
		return errors.Wrapf(err, "failed to delete chapter %d", page.ID)
	}
	return nil
}

// for the chapter management admin page on the ADB itself
func InsertChapter(db *sqlx.DB, page FacebookPageOutput) error {
	_, err := db.NamedExec(`INSERT INTO fb_pages ( id, name, flag, fb_url, insta_url, twitter_url, email, region, lat, lng, token )
		VALUES ( :id, :name, :flag, :fb_url, :insta_url, :twitter_url, :email, :region, :lat, :lng, :token )`, page)
	if err != nil {
		return errors.Wrapf(err, "failed to insert chapter %d", page.ID)
	}
	return nil
}

// returns all public chapter data for public API consumption
func GetAllChaptersWithoutTokens(db *sqlx.DB) ([]Chapter, error) {
	query := `SELECT chapter_id, id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng
		FROM fb_pages
		ORDER BY name`
	var chapters []Chapter
	err := db.Select(&chapters, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select chapters")
	}
	return chapters, nil
}

// TODO: update this to use the Chapter struct instead of FacebookPageOutput (will also need to update website)
func FindNearestFacebookPages(db *sqlx.DB, lat float64, lng float64) ([]FacebookPageOutput, error) {
	query := `SELECT id, name, flag, fb_url, region, (3959*acos(cos(radians(` + fmt.Sprintf("%f", lat) + `))*cos(radians(lat))* 
		cos(radians(lng)-radians(` + fmt.Sprintf("%f", lng) + `))+sin(radians(` + fmt.Sprintf("%f", lat) + `))* 
		sin(radians(lat)))) AS distance
		FROM fb_pages
		ORDER BY distance
		LIMIT 3`
	var pages []FacebookPageOutput
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}
	return pages, nil
}

// returns pages grouped by region
// TODO: update this to use the Chapter struct instead of FacebookPageOutput (will also need to update website)
func GetAllFBPagesByRegion(db *sqlx.DB) (map[string][]FacebookPageOutput, error) {
	query := `SELECT id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng
		FROM fb_pages
		ORDER BY name`
	var pages []FacebookPageOutput
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}
	regions := make(map[string][]FacebookPageOutput)
	for _, p := range pages {
		regions[p.Region] = append(regions[p.Region], p)
	}
	//return pages grouped into regions, nil
	return regions, nil
}

func GetFacebookEvents(db *sqlx.DB, pageID int, startTime string, endTime string) ([]FacebookEventOutput, error) {
	query := `SELECT id, page_id, name, description, start_time, end_time, location_name,
		location_country, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update FROM fb_events`

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

	var events []FacebookEventOutput
	err := db.Select(&events, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select events")
	}

	return events, nil
}

func GetOnlineFacebookEvents(db *sqlx.DB, startTime string, endTime string) ([]FacebookEventOutput, error) {
	// TODO: move these page IDs to config variables?
	query := `SELECT id, page_id, name, description, start_time, end_time, location_name,
		location_country, location_country, location_state, location_address, location_zip,
		lat, lng, cover, attending_count, interested_count, is_canceled, last_update FROM fb_events
		WHERE is_canceled = 0 and ((page_id = 1377014279263790 and location_name = 'Online') or page_id = 287332515138353)`

	if startTime != "" {
		query += " and start_time >= '" + startTime + "'"
	}
	if endTime != "" {
		// we actually want to show events which have a START time before the query's end time
		// otherwise really long (or recurring) events could be hidden
		query += " and start_time <= '" + endTime + "'"
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
