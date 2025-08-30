package model

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"slices"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// TODO: consolidate these structs (will need to update some things in the database too)

// used by public API, which is polled by https://animalrightsmap.org
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
	Country    string  `db:"country"`
	Lat        float64 `db:"lat"`
	Lng        float64 `db:"lng"`
}

type ChapterWithMailingList struct {
	Chapter
	Distance          float32 `db:"distance"`
	MailingListType   string  `db:"ml_type"`
	MailingListRadius int     `db:"ml_radius"`
	MailingListID     string  `db:"ml_id"`
}

// used for internal Chapters page on ADB, syncing with FB and Eventbrite, and for displaying events on the website
//
// Deprecated in public API; use `Chapter` instead, which uses the real chapter ID in the `ID` field.
type ChapterWithToken struct {
	ID                   int          `db:"id"`         // Facebook page ID; `id` column on `fb_pages` table
	ChapterID            int          `db:"chapter_id"` // canonical ID of chapter recognized by ADB
	Name                 string       `db:"name"`
	Flag                 string       `db:"flag"`
	FbURL                string       `db:"fb_url"`
	TwitterURL           string       `db:"twitter_url"`
	InstaURL             string       `db:"insta_url"`
	Email                string       `db:"email"`
	Region               string       `db:"region"`
	Lat                  float64      `db:"lat"`
	Lng                  float64      `db:"lng"`
	Distance             float32      `db:"distance"`
	MailingListType      string       `db:"ml_type"`
	MailingListRadius    int          `db:"ml_radius"`
	MailingListID        string       `db:"ml_id"`
	Token                string       `db:"token,omitempty"`
	LastFBSync           string       `db:"last_update"`
	LastFBEvent          string       `db:"last_fb_event"`
	EventbriteID         string       `db:"eventbrite_id"`
	EventbriteToken      string       `db:"eventbrite_token"`
	Mentor               string       `db:"mentor"`
	Country              string       `db:"country"`
	Notes                string       `db:"notes"`
	LastContact          string       `db:"last_contact"`
	LastAction           string       `db:"last_action"`
	Organizers           Organizers   `db:"organizers"`
	LastCheckinEmailSent sql.NullTime `db:"last_checkin_email_sent"`
	EmailToken           string       `db:"email_token"`
}

type Organizer struct {
	Name      string
	Email     string
	Phone     string
	Facebook  string
	Instagram string
	Twitter   string
	Website   string
}

type Organizers []*Organizer

const SFBayChapterName = "SF Bay Area"
const SFBayChapterId = 47
const SFBayChapterIdStr = "47"
const SFBayChapterIdDevTest = 1
const SFBayChapterIdDevTestStr = "1"

func (o Organizers) Value() (driver.Value, error) {
	return json.Marshal(o)
}

func (o *Organizers) Scan(value interface{}) error {
	b, ok := value.([]byte)
	if !ok {
		return errors.New("type assertion to []byte failed")
	}
	return json.Unmarshal(b, &o)
}

// used for making api calls to facebook, not for responding to our api requests
func GetChaptersWithFacebookTokens(db *sqlx.DB) ([]ChapterWithToken, error) {
	query := `SELECT id, name, lat, lng, token FROM fb_pages WHERE token <> '' and id <> 0`

	var pages []ChapterWithToken
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}

	return pages, nil
}

func GetChaptersWithEventbriteTokens(db *sqlx.DB) ([]ChapterWithToken, error) {
	query := `SELECT id, name, eventbrite_id, eventbrite_token FROM fb_pages
		WHERE eventbrite_token <> '' and eventbrite_id <> ''`

	var pages []ChapterWithToken
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}

	return pages, nil
}

// for the chapter management admin page on the ADB itself – NOTE THAT THIS RETURNS TOKENS, SO IT SHOULD NOT BE MADE PUBLIC
func GetAllChapters(db *sqlx.DB) ([]ChapterWithToken, error) {
	query := `SELECT fb_pages.id, chapter_id, fb_pages.name, flag, fb_url, twitter_url, insta_url, email, region, fb_pages.lat, fb_pages.lng, token, fb_pages.eventbrite_id, eventbrite_token, ml_type, ml_radius, ml_id,
	
		@last_update := IFNULL((
		  SELECT MAX(last_update) AS last_update
		  FROM fb_events
		  WHERE fb_pages.id = fb_events.page_id    
		), "") AS last_update,
		
		@last_fb_event := IFNULL((
		  SELECT DATE(MAX(fb_events.start_time)) AS start_time
		  FROM fb_events
		  WHERE fb_pages.id = fb_events.page_id
		  AND fb_events.start_time < NOW()
		), "") AS last_fb_event,

		mentor, country, notes, last_contact, last_action, organizers, last_checkin_email_sent, IFNULL(email_token,"") as email_token
		
		FROM fb_pages
		ORDER BY name`
	var pages []ChapterWithToken
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}

	return pages, nil
}

// for the chapter management admin page on the ADB itself
func GetChapterWithTokenById(db *sqlx.DB, id int) (ChapterWithToken, error) {
	query := `SELECT fb_pages.id, chapter_id, fb_pages.name, flag, fb_url, twitter_url, insta_url, email, region, fb_pages.lat, fb_pages.lng, token, fb_pages.eventbrite_id, eventbrite_token, ml_type, ml_radius, ml_id,
	
		@last_update := IFNULL((
		  SELECT MAX(last_update) AS last_update
		  FROM fb_events
		  WHERE fb_pages.id = fb_events.page_id    
		), "") AS last_update,
		
		@last_fb_event := IFNULL((
		  SELECT DATE(MAX(fb_events.start_time)) AS start_time
		  FROM fb_events
		  WHERE fb_pages.id = fb_events.page_id
		  AND fb_events.start_time < NOW()
		), "") AS last_fb_event,

		mentor, country, notes, last_contact, last_action, organizers, last_checkin_email_sent, IFNULL(email_token,"") as email_token
		
		FROM fb_pages
		WHERE chapter_id = ?`
	var pages []ChapterWithToken
	err := db.Select(&pages, query, id)
	if err != nil {
		return ChapterWithToken{}, errors.Wrap(err, "failed to select page")
	}
	if len(pages) == 0 {
		return ChapterWithToken{}, errors.New("Could not find page")
	}
	if len(pages) > 1 {
		return ChapterWithToken{}, errors.New("Found too many pages")
	}
	return pages[0], nil
}

// for the chapter management admin page on the ADB itself
func UpdateChapter(db *sqlx.DB, page ChapterWithToken) (int, error) {
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
		token = :token,
		eventbrite_id = :eventbrite_id,
		eventbrite_token = :eventbrite_token,
		ml_type = :ml_type,
		ml_radius = :ml_radius,
		ml_id = :ml_id,
		mentor = :mentor,
		country = :country,
		notes = :notes,
		last_contact = :last_contact,
		last_action = :last_action,
		organizers = :organizers,
		email_token = :email_token,
		last_checkin_email_sent = :last_checkin_email_sent
		WHERE chapter_id = :chapter_id`, page)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to update chapter %d", page.ID)
	}
	return page.ChapterID, nil
}

// for the chapter management admin page on the ADB itself
func DeleteChapter(db *sqlx.DB, chapter int) error {
	// first make sure that there are no users associated w/ the chapter
	var userCount int
	err := db.QueryRow(`SELECT COUNT(*) from adb_users
		WHERE chapter_id = ?`, chapter).Scan(&userCount)
	if err != nil {
		return errors.Wrapf(err, "failed to count users for chapter %d", chapter)
	}
	if userCount > 0 {
		return errors.New("cannot delete chapter because users are associated with it")
	}

	_, err = db.Exec(`DELETE FROM fb_pages
		WHERE chapter_id = ?`, chapter)
	if err != nil {
		return errors.Wrapf(err, "failed to delete chapter %d", chapter)
	}
	return nil
}

// for the chapter management admin page on the ADB itself
func InsertChapter(db *sqlx.DB, page ChapterWithToken) (int, error) {
	res, err := db.NamedExec(`INSERT INTO fb_pages ( id, name, flag, fb_url, insta_url, twitter_url, email, region, lat, lng, mentor, country, notes, last_contact, last_action, organizers, email_token, ml_type, ml_radius, ml_id )
		VALUES ( :id, :name, :flag, :fb_url, :insta_url, :twitter_url, :email, :region, :lat, :lng, :mentor, :country, :notes, :last_contact, :last_action, :organizers, :email_token, :ml_type, :ml_radius, :ml_id )`, page)
	if err != nil {
		return 0, errors.Wrapf(err, "failed to insert chapter %d", page.ID)
	}
	insertedID, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("error getting ID after insert: %w", err)
	}
	return int(insertedID), nil
}

// returns all public chapter data for public API consumption
func GetAllChapterInfo(db *sqlx.DB) ([]Chapter, error) {
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

func CleanChapterData(db *sqlx.DB, body io.Reader) (ChapterWithToken, error) {
	var chapter ChapterWithToken
	err := json.NewDecoder(body).Decode(&chapter)
	if err != nil {
		return ChapterWithToken{}, err
	}

	chapter.Name = strings.TrimSpace(chapter.Name)

	// TODO: trim space off more fields

	return chapter, nil
}

func getFindNearestChaptersQuery(lat float64, lng float64) string {
	return `SELECT id, chapter_id, name, email, flag, fb_url, insta_url, twitter_url, region, country, ml_type, ml_radius, ml_id, (3959*acos(cos(radians(` + fmt.Sprintf("%f", lat) + `))*cos(radians(lat))*
		cos(radians(lng)-radians(` + fmt.Sprintf("%f", lng) + `))+sin(radians(` + fmt.Sprintf("%f", lat) + `))* 
		sin(radians(lat)))) AS distance
		FROM fb_pages
		WHERE region <> 'Online'
		ORDER BY distance
		LIMIT 3`
}

// Deprecated. Use FindNearestChaptersSortedByDistance instead.
func FindNearestChaptersSortedByDistanceDeprecated(db *sqlx.DB, lat float64, lng float64) ([]ChapterWithToken, error) {
	var pages []ChapterWithToken // we aren't actually getting tokens, but the website expects the FB ID to be in the ID field
	err := db.Select(&pages, getFindNearestChaptersQuery(lat, lng))
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}
	return pages, nil
}

func FindNearestChaptersSortedByDistance(db *sqlx.DB, lat float64, lng float64) ([]ChapterWithMailingList, error) {
	var pages []ChapterWithMailingList
	err := db.Select(&pages, getFindNearestChaptersQuery(lat, lng))
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select chapters")
	}
	return pages, nil
}

// returns pages grouped by region
// TODO: update this function (and the website) to handle data in the normal Chapter struct instead of w/ Token
func GetAllChaptersByRegion(db *sqlx.DB) (map[string][]ChapterWithToken, error) {
	query := `SELECT id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng
		FROM fb_pages
		ORDER BY name`
	var pages []ChapterWithToken // we aren't actually getting tokens, but the website expects the FB ID to be in the ID field
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
	}
	regions := make(map[string][]ChapterWithToken) // we aren't actually getting tokens, but the website expects the FB ID to be in the ID field
	for _, p := range pages {
		regions[p.Region] = append(regions[p.Region], p)
	}
	//return pages grouped into regions, nil
	return regions, nil
}

func GetChapterById(db *sqlx.DB, id int) (ChapterWithMailingList, error) {
	query := `SELECT chapter_id, id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng, ml_type, ml_radius, ml_id
		FROM fb_pages
		WHERE chapter_id = ?
		ORDER BY name`
	var chapter ChapterWithMailingList
	err := db.Get(&chapter, query, id)
	if err != nil {
		return ChapterWithMailingList{}, fmt.Errorf("failed to select chapter: %v", err)
	}
	return chapter, nil
}

const SFBayPageID = 1377014279263790
const NorthBayPageID = 495485410315891
const AlcPageID = 287332515138353

var BayAreaPages = []int{
	SFBayPageID,
	NorthBayPageID,
}

func IsBayAreaPage(pageId int) bool {
	return slices.Contains(BayAreaPages, pageId)
}
