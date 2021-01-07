package event_sync

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

func getFacebookEvents(page model.ChapterWithToken) []model.FacebookEventJSON {
	url := "https://graph.facebook.com/v4.0/" + strconv.Itoa(page.ID) + "/events?include_canceled=1&fields=name,start_time,end_time,cover,attending_count,description,place,interested_count,is_canceled,event_times,is_online&limit=50&access_token=" + page.Token

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// TODO: we should handle errors better so we don't stop all pages from syncing
	if resp.StatusCode != http.StatusOK {
		log.Println("ERROR getting events from", page.Name, err)
		return nil
	}
	// read the response & decode the json data
	data := model.FacebookResponseJSON{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Println("ERROR getting events from", page.Name, err)
		return nil
	}

	return data.Data
}

func getFacebookEvent(page model.ChapterWithToken, eventID string) model.FacebookEventJSON {
	url := "https://graph.facebook.com/v4.0/" + eventID + "?fields=name,start_time,end_time,cover,attending_count,description,place,interested_count,is_canceled,event_times,is_online&limit=50&access_token=" + page.Token

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// TODO: we should handle errors better so we don't stop all pages from syncing
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}
	// read the response & decode the json data
	data := model.FacebookEventJSON{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	return data
}

func syncFacebookEvents(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in Facebook event sync", r)
		}
	}()

	// get pages from database
	pages, err := model.GetChaptersWithFacebookTokens(db)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}
	if pages == nil {
		// stop if no pages in database
		log.Println("There are no Facebook pages to sync.")
		return
	}
	// for each page, get event data
	for _, page := range pages {

		log.Println("Getting FB events from", page.Name, "(", page.ID, ")")

		// make call to fb api
		events := getFacebookEvents(page)

		if len(events) > 0 {

			// loop through events
			for _, event := range events {

				// if event has event_times, then we need to find the sub-events instead
				if event.EventTimes != nil {
					for _, subEvent := range event.EventTimes {
						// make api call for subEvent
						subEventData := getFacebookEvent(page, subEvent.ID)
						// insert (replace into) database
						err = model.InsertFacebookEvent(db, subEventData, page)
						if err != nil {
							log.Println("ERROR:", err)
						}
					}
					continue
				}
				// insert (replace into) database
				err = model.InsertFacebookEvent(db, event, page)
				if err != nil {
					log.Println("ERROR:", err)
				}
			}
		} else {
			log.Println("No events returned for", page.Name)
		}
	}
}

func getEventbriteEvents(chapter model.ChapterWithToken) []model.EventbriteEventJSON {
	// TODO: move eventbrite organizationId and token to config
	url := "https://www.eventbriteapi.com/v3/organizations/" + chapter.EventbriteID +
		"/events?status=live&page_size=200&token=" + chapter.EventbriteToken

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	// TODO: we should handle errors better so we don't stop all pages from syncing
	if resp.StatusCode != http.StatusOK {
		panic(resp.StatusCode)
	}
	// read the response & decode the json data
	data := model.EventbriteResponseJSON{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		panic(err)
	}
	return data.Events
}

func syncEventbriteEvents(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in Eventbrite event sync", r)
		}
	}()

	// get pages from database
	pages, err := model.GetChaptersWithEventbriteTokens(db)
	if err != nil {
		log.Println("ERROR:", err)
		return
	}
	if pages == nil {
		// stop if no pages in database
		log.Println("There are no Eventbrite pages to sync.")
		return
	}
	// for each page, get event data
	for _, page := range pages {

		log.Println("Getting EB events from", page.Name, "(", page.EventbriteID, ")")

		// make call to fb api
		events := getEventbriteEvents(page)

		if len(events) > 0 {
			// loop through events
			for _, event := range events {
				println("Syncing EB event:", event.Name.Text, " ", "at", event.Start.UTC)
				err = model.AddEventbriteDetailsToEvent(db, event)
				if err != nil {
					log.Println("ERROR:", err)
				}
			}
		} else {
			log.Println("No events returned for", page.Name)
		}
	}
}

// Get events from Facebook every 15 minutes.
// Should be run in a goroutine.
func StartFacebookSync(db *sqlx.DB) {

	for {
		log.Println("Starting Facebook event sync")
		syncFacebookEvents(db)
		log.Println("Finished Facebook event sync")
		log.Println("Starting Eventbrite event sync")
		syncEventbriteEvents(db)
		log.Println("Finished Eventbrite event sync")
		time.Sleep(60 * time.Minute)
		time.Sleep(60 * time.Minute)
	}
}
