package facebook_events

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

func getFacebookEvents(page model.FacebookPage) []model.FacebookEventJSON {
	url := "https://graph.facebook.com/v4.0/" + strconv.Itoa(page.ID) + "/events?fields=name,start_time,end_time,cover,attending_count,description,place,interested_count,is_canceled&limit=50&access_token=" + page.Token

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}
	defer resp.Body.Close()
	// read the response
	body, err := ioutil.ReadAll(resp.Body)
	// unmarshal the json data
	data := model.FacebookResponseJSON{}
	err = json.Unmarshal(body, &data)
	if err != nil {
		panic(err.Error())
	}
	return data.Data
}

func syncFacebookEvents(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in Facebook event sync", r)
		}
	}()

	// get pages from database
	pages, err := model.GetFacebookPages(db)
	if err != nil {
		log.Println("ERROR:", err)
	}
	if pages == nil {
		// stop if no pages in database
		log.Println("There are no Facebook pages to sync.")
		return
	}
	// for each page, get event data
	for _, page := range pages {

		log.Println("Getting events from", page.Name, "(", page.ID, ")")

		// make call to fb api
		events := getFacebookEvents(page)

		// loop through events
		for _, event := range events {
			// insert (replace into) database
			_, err = model.InsertFacebookEvent(db, event, page)
			if err != nil {
				log.Println("ERROR:", err)
			}
		}
	}
}

// Get events from Facebook every 15 minutes.
// Should be run in a goroutine.
func StartFacebookSync(db *sqlx.DB) {

	// test getting events by page id
	//events, _ := model.GetFacebookEvents(db, 1377014279263790)
	//log.Println("events:")
	//log.Println(events)

	for {
		log.Println("Starting Facebook event sync")
		syncFacebookEvents(db)
		log.Println("Finished Facebook event sync")
		time.Sleep(15 * time.Minute)
	}
}