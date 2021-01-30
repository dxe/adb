package event_sync

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

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
		events, err := getFacebookEvents(page)
		if err != nil {
			fmt.Println("ERROR: failed to get Facebook events. ", err.Error())
			continue
		}

		if len(events) > 0 {
			// loop through events
			for _, event := range events {
				// if event has event_times, then we need to find the sub-events instead
				if event.EventTimes != nil {
					for _, subEvent := range event.EventTimes {
						// make api call for subEvent
						subEventData, err := getFacebookEvent(page, subEvent.ID)
						if err != nil {
							fmt.Println("ERROR: failed to get individual Facebook event.", err.Error())
							continue
						}
						parsedEvent, err := parseFacebookEvent(subEventData, page)
						if err != nil {
							log.Println("ERROR: failed to parse FB event:", err)
						}
						err = model.InsertExternalEvent(db, parsedEvent)
						if err != nil {
							log.Println("ERROR:", err)
						}
					}
					continue
				}
				// insert into database
				parsedEvent, err := parseFacebookEvent(event, page)
				if err != nil {
					log.Println("ERROR: failed to parse FB event:", err)
				}
				err = model.InsertExternalEvent(db, parsedEvent)
				if err != nil {
					log.Println("ERROR:", err)
				}
			}
		} else {
			log.Println("No events returned for", page.Name)
		}
	}
}

func parseFacebookEvent(fbEvent FacebookEvent, page model.ChapterWithToken) (model.ExternalEvent, error) {
	fbTimeLayout := "2006-01-02T15:04:05-0700"
	startTime, err := time.Parse(fbTimeLayout, fbEvent.StartTime)
	startTime = startTime.UTC()
	endTime, err := time.Parse(fbTimeLayout, fbEvent.EndTime)
	endTime = endTime.UTC()
	placeName := fbEvent.Place.Name
	if fbEvent.IsOnline {
		placeName = "Online"
	}
	eventId, err := strconv.Atoi(fbEvent.ID)
	if err != nil {
		return model.ExternalEvent{}, err
	}
	parsedEvent := model.ExternalEvent{
		ID:              eventId,
		PageID:          page.ID,
		Name:            fbEvent.Name,
		Description:     fbEvent.Description,
		StartTime:       startTime,
		EndTime:         endTime,
		LocationName:    placeName,
		LocationCity:    fbEvent.Place.Location.City,
		LocationCountry: fbEvent.Place.Location.Country,
		LocationState:   fbEvent.Place.Location.State,
		LocationAddress: fbEvent.Place.Location.Street,
		LocationZip:     fbEvent.Place.Location.Zip,
		Lat:             fbEvent.Place.Location.Lat,
		Lng:             fbEvent.Place.Location.Lng,
		Cover:           fbEvent.Cover.Source,
		AttendingCount:  fbEvent.AttendingCount,
		InterestedCount: fbEvent.InterestedCount,
		IsCanceled:      fbEvent.IsCanceled,
	}
	return parsedEvent, nil
}

func getUpcomingEventsFromEventbrite(chapter model.ChapterWithToken) ([]EventbriteEvent, error) {
	path := eventbriteAPIBaseURL + "/organizations/" + chapter.EventbriteID +
		"/events?status=live&page_size=200&token=" + chapter.EventbriteToken

	var events EventbriteEvents
	err := callAPIGet(path, &events)
	if err != nil {
		return []EventbriteEvent{}, errors.New("failed to get upcoming events from Eventbrite: " + err.Error())
	}
	return events.Events, nil
}

// addEventbriteDataToExistingEvents attempts to add the Eventbrite URL and Eventbrite ID to events in the database that
// have a matching name and date but do not already have an Eventbrite URL and ID.
func addEventbriteDataToExistingEvents(db *sqlx.DB, ebEvents []EventbriteEvent, fbEvents []model.ExternalEvent) error {
	// create a map of the Facebook events to quickly look an event up by Eventbrite URL
	fbEventsMap := make(map[string]model.ExternalEvent)
	for _, fbEvent := range fbEvents {
		if fbEvent.EventbriteURL != "" {
			fbEventsMap[fbEvent.EventbriteURL] = fbEvent
		}
	}
	// loop through events
	for _, ebEvent := range ebEvents {
		if _, ok := fbEventsMap[ebEvent.URL]; !ok {
			fmt.Println("Attempting to add Eventbrite details to DB for", ebEvent.Name.Text)
			startTime, err := time.Parse(time.RFC3339, ebEvent.Start.UTC)
			if err != nil {
				return errors.New("failed to parse eventbrite start time: " + err.Error())
			}
			event := model.ExternalEvent{
				Name:          ebEvent.Name.Text,
				StartTime:     startTime,
				EventbriteURL: ebEvent.URL,
				EventbriteID:  ebEvent.ID,
			}
			err = model.AddEventbriteDetailsToEventByNameAndDate(db, event)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func createOrUpdateEventbriteEvents(db *sqlx.DB, chapter model.ChapterWithToken, fbEvents []model.ExternalEvent, ebEvents []EventbriteEvent) error {

	// create a map of the Eventbrite events to quickly look an event up by URL
	ebEventsMap := make(map[string]EventbriteEvent)
	for _, ebEvent := range ebEvents {
		if ebEvent.URL != "" {
			ebEventsMap[ebEvent.URL] = ebEvent
		}
	}

	for _, fbEvent := range fbEvents {

		// filter the events to find which ones don't yet have an Eventbrite ID in the database
		if fbEvent.EventbriteURL == "" {
			ebEvent, err := createEventbriteEvent(fbEvent, chapter)
			if err != nil {
				return err
			}
			event := model.ExternalEvent{
				ID:            fbEvent.ID,
				EventbriteURL: ebEvent.URL,
				EventbriteID:  ebEvent.ID,
			}
			err = model.AddEventbriteDetailsToEventByID(db, event)
			if err != nil {
				return err
			}
			continue
		}

		// check if event was cancelled or deleted from eventbrite
		if _, ok := ebEventsMap[fbEvent.EventbriteURL]; !ok {
			fmt.Println("WARNING:", fbEvent.Name, "was deleted by a human on Eventbrite, but an Eventbrite URL still exists in our database!")
			// TODO: We should probably remove the Eventbrite ID & URL from the database since it no longer exists.
			continue
		}

		// check if event is still on eventbrite but needs to be updated
		ebEvent := ebEventsMap[fbEvent.EventbriteURL]
		if ebEvent.Name.Text != fbEvent.Name || ebEvent.Start.UTC != fbEvent.StartTime.Format(time.RFC3339) {
			// TODO: update the event name, summary (w/ the name), and start time
			// TODO: try to reuse code from creating events to update them
			fmt.Println("We need to update the event name & start time on Eventbrite for", ebEvent.Name)
		}
		// TODO: check if location changed
		// TODO: Compare description (Create new structured content version is it doesn't match)
		// TODO: Check if there are any cancelled event in the database. If they are still on Eventbrite, cancel them.  https://www.eventbriteapi.com/v3/events/EVENT_ID/cancel/

	}

	return nil
}

func createEventbriteEvent(event model.ExternalEvent, chapter model.ChapterWithToken) (EventbriteEvent, error) {
	fmt.Println("Creating event on Eventbrite:", event.Name)

	var ebEvent EventbriteEvent

	// create venue
	venueID, err := createEventbriteVenue(event, chapter)
	if err != nil {
		return ebEvent, err
	}

	imageID, err := createEventbriteImage(event, chapter)
	if err != nil {
		return ebEvent, err
	}

	ebEvent, err = addEventToEventbrite(event, chapter, venueID, imageID)
	if err != nil {
		return ebEvent, err
	}

	// add ticket class
	err = addEventTicketClass(ebEvent.ID, chapter.EventbriteToken)
	if err != nil {
		return ebEvent, err
	}

	// add description
	err = updateEventDescription(ebEvent.ID, event.Description, chapter.EventbriteToken)
	if err != nil {
		return ebEvent, err
	}

	// publish event
	err = publishEvent(ebEvent.ID, chapter.EventbriteToken)
	if err != nil {
		return ebEvent, err
	}

	return ebEvent, nil
}

func syncEventbriteEvents(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in Eventbrite sync.", r)
		}
	}()

	now := time.Now().UTC()

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

		fbEvents, err := model.GetExternalEvents(db, page.ID, now, time.Time{}, false)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		err = addEventbriteDataToExistingEvents(db, ebEvents, fbEvents)
		if err != nil {
			fmt.Println("ERROR:", err)
		}

		// get Facebook events again since we may have just added the EB information to them
		fbEvents, err = model.GetExternalEvents(db, page.ID, now, time.Time{}, false)
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
