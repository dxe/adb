package event_sync

import (
	"encoding/json"
	"errors"
	"fmt"
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

func getUpcomingEventsFromEventbrite(chapter model.ChapterWithToken) ([]model.EventbriteEventJSON, error) {
	url := "https://www.eventbriteapi.com/v3/organizations/" + chapter.EventbriteID +
		"/events?status=live&page_size=200&token=" + chapter.EventbriteToken

	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("failed to get upcoming events from Eventbrite")
	}
	// read the response & decode the json data
	data := model.EventbriteResponseJSON{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return nil, err
	}
	return data.Events, nil
}

// This function will always need to match all existing events since the Facebook Sync will wipe out the data.
func addEventbriteDataToExistingEvents(db *sqlx.DB, ebEvents []model.EventbriteEventJSON) error {
	if len(ebEvents) > 0 {
		// loop through events
		for _, ebEvent := range ebEvents {
				err := model.AddEventbriteDetailsToEventByName(db, ebEvent)
				if err != nil {
					return err
				}
		}
		return nil
	}
	return errors.New("found no upcoming events listed on Eventbrite")
}

func createOrUpdateEventbriteEvents(db *sqlx.DB, chapter model.ChapterWithToken, fbEvents []model.ExternalEvent, ebEvents []model.EventbriteEventJSON) error {

	// create a map of the Eventbrite events to quickly look an event up by URL
	ebEventsMap := make(map[string]model.EventbriteEventJSON)
	for _, ebEvent := range ebEvents {
		if ebEvent.URL != "" {
			ebEventsMap[ebEvent.URL] = ebEvent
		}
	}

	for _, fbEvent := range fbEvents {

		// filter the events to find which ones don't yet have an Eventbrite ID in the database
		if fbEvent.EventbriteURL == "" {
			eventbriteID, eventbriteURL, err := createEventbriteEvent(fbEvent, chapter)
			if err != nil {
				return err
			}
			err = model.AddEventbriteDetailsToEventByID(db, fbEvent.ID, eventbriteID, eventbriteURL)
			if err != nil {
				return err
			}
		} else {
			// event already exists on Eventbrite, so check if it needs to be updated
			ebEvent := ebEventsMap[fbEvent.EventbriteURL]
			if ebEvent.Name.Text != fbEvent.Name || ebEvent.Start.UTC != fbEvent.StartTime.Format(time.RFC3339) {
				// TODO: update the event name, summary (w/ the name), and start time
				fmt.Println("We need to update the event name & start time on Eventbrite.")
			}
			// TODO: Compare description (Create new structured content version is it doesn't match)
			// TODO: Check if there are any cancelled event in the database. If they are still on Eventbrite, cancel them.  https://www.eventbriteapi.com/v3/events/EVENT_ID/cancel/
		}
	}

	return nil
}

func createEventbriteEvent(event model.ExternalEvent, chapter model.ChapterWithToken) (string, string, error) {
	fmt.Println("Creating event on Eventbrite:", event.Name)

	// create venue
	venueID, err := createEventbriteVenue(event, chapter)
	if err != nil {
		return "", "", err
	}

	imageID, err := createEventbriteImage(event, chapter)
	if err != nil {
		return "", "", err
	}

	eventID, eventURL, err := addEventToEventbrite(event, chapter, venueID, imageID)
	if err != nil {
		return "", "", err
	}

	// add ticket class
	err = addEventTicketClass(eventID, chapter.EventbriteToken)
	if err != nil {
		return "", "", err
	}

	// add description
	err = updateEventDescription(eventID, event.Description, chapter.EventbriteToken)
	if err != nil {
		return "", "", err
	}

	// publish event
	err = publishEvent(eventID, chapter.EventbriteToken)
	if err != nil {
		return "", "", err
	}

	return eventID, eventURL, nil
}

func syncEventbriteEvents(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in Eventbrite sync.", r)
		}
	}()

	now := time.Now().UTC().Format(time.RFC3339)

	// get pages from database
	pages, err := model.GetChaptersWithEventbriteTokens(db)
	if err != nil {
		log.Println("ERROR: Failed to get chapters with Eventbrite tokens from database.")
		return
	}
	if pages == nil {
		// stop if no pages in database
		log.Println("ERROR: There are no Eventbrite pages to sync.")
		return
	}
	// for each page
	for _, page := range pages {

		log.Println("Starting Eventbrite sync for ", page.Name, "(", page.EventbriteID, ")")

		ebEvents, err := getUpcomingEventsFromEventbrite(page)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		err = addEventbriteDataToExistingEvents(db, ebEvents)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		fbEvents, err := model.GetFacebookEvents(db, page.ID, now, "", false)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		err = createOrUpdateEventbriteEvents(db, page, fbEvents, ebEvents)
		if err != nil {
			fmt.Println("ERROR:", err)
		}
	}
}

// Sync events every 60 minutes.
// Should be run in a goroutine.
func StartExternalEventSync(db *sqlx.DB) {

	for {
		log.Println("Starting Facebook event sync")
		syncFacebookEvents(db)
		log.Println("Finished Facebook event sync")

		log.Println("Starting Eventbrite event sync")
		syncEventbriteEvents(db)
		log.Println("Finished Eventbrite event sync")

		time.Sleep(60 * time.Minute)
	}
}
