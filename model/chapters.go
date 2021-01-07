package model

import (
	"fmt"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

// TODO: consolidate these structs (will need to update some things in the database too)

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

// used for internal Chapters page on ADB, as well as syncing with FB and Eventbrite
type ChapterWithToken struct {
	ID              int     `db:"id"`
	ChapterID       int     `db:"chapter_id"`
	Name            string  `db:"name"`
	Flag            string  `db:"flag"`
	FbURL           string  `db:"fb_url"`
	TwitterURL      string  `db:"twitter_url"`
	InstaURL        string  `db:"insta_url"`
	Email           string  `db:"email"`
	Region          string  `db:"region"`
	Lat             float64 `db:"lat"`
	Lng             float64 `db:"lng"`
	Distance        float32 `db:"distance"`
	Token           string  `db:"token,omitempty"`
	LastUpdate      string  `db:"last_update"`
	EventbriteID    string  `db:"eventbrite_id"`
	EventbriteToken string  `db:"eventbrite_token"`
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

// for the chapter management admin page on the ADB itself -- NOTE THAT THIS RETURNS TOKENS, SO IT SHOULD NOT BE MADE PUBLIC
func GetAllChapters(db *sqlx.DB) ([]ChapterWithToken, error) {
	query := `SELECT fb_pages.id, chapter_id, fb_pages.name, flag, fb_url, twitter_url, insta_url, email, region, fb_pages.lat, fb_pages.lng, token, IFNULL(MAX(last_update),'') as last_update, eventbrite_id, eventbrite_token
		FROM fb_pages
		LEFT JOIN fb_events on fb_pages.id = fb_events.page_id
		GROUP BY fb_pages.chapter_id
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
func GetChapterByID(db *sqlx.DB, id int) (ChapterWithToken, error) {
	query := `SELECT id, chapter_id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng, token, eventbrite_id, eventbrite_token
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
func UpdateChapter(db *sqlx.DB, page ChapterWithToken) error {
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
		eventbrite_token = :eventbrite_token
		WHERE chapter_id = :chapter_id`, page)
	if err != nil {
		return errors.Wrapf(err, "failed to update chapter %d", page.ID)
	}
	return nil
}

// for the chapter management admin page on the ADB itself
func DeleteChapter(db *sqlx.DB, page ChapterWithToken) error {
	_, err := db.NamedExec(`DELETE FROM fb_pages
		WHERE chapter_id = :chapter_id`, page)
	if err != nil {
		return errors.Wrapf(err, "failed to delete chapter %d", page.ID)
	}
	return nil
}

// for the chapter management admin page on the ADB itself
func InsertChapter(db *sqlx.DB, page ChapterWithToken) error {
	_, err := db.NamedExec(`INSERT INTO fb_pages ( id, name, flag, fb_url, insta_url, twitter_url, email, region, lat, lng, token )
		VALUES ( :id, :name, :flag, :fb_url, :insta_url, :twitter_url, :email, :region, :lat, :lng, :token, eventbrite_id, eventbrite_token )`, page)
	if err != nil {
		return errors.Wrapf(err, "failed to insert chapter %d", page.ID)
	}
	return nil
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

// TODO: update this function (and the website) to handle data in the normal Chapter struct instead of w/ Token
func FindNearestChapters(db *sqlx.DB, lat float64, lng float64) ([]ChapterWithToken, error) {
	query := `SELECT id, name, flag, fb_url, region, (3959*acos(cos(radians(` + fmt.Sprintf("%f", lat) + `))*cos(radians(lat))* 
		cos(radians(lng)-radians(` + fmt.Sprintf("%f", lng) + `))+sin(radians(` + fmt.Sprintf("%f", lat) + `))* 
		sin(radians(lat)))) AS distance
		FROM fb_pages
		ORDER BY distance
		LIMIT 3`
	var pages []ChapterWithToken // we aren't actually getting tokens, but the website expects the FB ID to be in the ID field
	err := db.Select(&pages, query)
	if err != nil {
		// error
		return nil, errors.Wrap(err, "failed to select pages")
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
