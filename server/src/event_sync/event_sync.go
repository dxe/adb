package event_sync

import (
	"errors"
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
		log.Println("Error getting FB pages from database:", err)
		return
	}
	if pages == nil {
		log.Println("There are no Facebook pages to sync.")
		return
	}
	for _, page := range pages {
		err := syncFacebookEventsForPage(db, page)
		if err != nil {
			log.Println("Sync failed for page", page.Name, err.Error())
		}
	}
}

func syncFacebookEventsForPage(db *sqlx.DB, page model.ChapterWithToken) error {
	log.Println("Getting FB events from", page.Name, "(", page.ID, ")")

	events, err := getFacebookEvents(page) // gets events from FB API
	if err != nil {
		return err
	}
	if len(events) == 0 {
		log.Println("INFO: No events returned from Facebook.")
		return nil
	}

	for _, event := range events {
		err := parseAndUpsertFacebookEvent(db, event, page)
		if err != nil {
			log.Println(err.Error()) // print the error, but keep trying for the rest of the events
		}
	}
	return nil
}

func parseAndUpsertFacebookEvent(db *sqlx.DB, event FacebookEvent, page model.ChapterWithToken) error {
	// if event has event_times, then insert the sub-events instead
	if event.EventTimes != nil {
		for _, subEvent := range event.EventTimes {
			err := parseAndUpsertFacebookEvent(db, subEvent, page)
			if err != nil {
				return errors.New("failed to insert FB sub-events for: " + event.Name + ": " + err.Error())
			}
		}
		return nil
	}

	// if event has no event_times (sub-events), then just insert the event
	parsedEvent, err := parseFacebookEvent(event, page)
	if err != nil {
		return errors.New("failed to parse FB event: " + event.Name + ": " + err.Error())
	}
	err = model.UpsertExternalEvent(db, parsedEvent)
	if err != nil {
		return errors.New("failed to insert event into database: " + event.Name + ": " + err.Error())
	}
	return nil
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
	eventId, err := strconv.ParseInt(fbEvent.ID, 10, 64)
	if err != nil {
		return model.ExternalEvent{}, err
	}
	parsedEvent := model.ExternalEvent{
		ID:              strconv.FormatInt(eventId, 10),
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
		log.Println("There are no Eventbrite pages to sync.")
		return
	}
	// for each page
	for _, page := range pages {

		log.Println("Starting Eventbrite sync for", page.Name, "(", page.EventbriteID, ")")

		ebEvents, err := getUpcomingEventsFromEventbrite(page)
		if err != nil {
			log.Println("ERROR:", err)
		}

		dbEvents, err := model.GetExternalEvents(db, page.ID, now, time.Time{})
		if err != nil {
			log.Println("ERROR:", err)
		}

		err = addEventbriteDataToExistingEvents(db, ebEvents, dbEvents)
		if err != nil {
			log.Println("ERROR:", err)
		}

		// read events from db again since we may have just added the EB information to them
		dbEvents, err = model.GetExternalEvents(db, page.ID, now, time.Time{})
		if err != nil {
			log.Println("ERROR:", err)
		}

		err = createOrUpdateEventbriteEvents(db, page, dbEvents, ebEvents)
		if err != nil {
			log.Println("ERROR:", err)
		}
	}
}

func getUpcomingEventsFromEventbrite(chapter model.ChapterWithToken) ([]EventbriteEvent, error) {
	path := eventbriteAPIBaseURL + "/organizations/" + chapter.EventbriteID +
		"/events?status=live&page_size=200&token=" + chapter.EventbriteToken

	var events EventbriteEvents
	err := getAPI(path, &events)
	if err != nil {
		return []EventbriteEvent{}, errors.New("failed to get upcoming events from Eventbrite: " + err.Error())
	}
	return events.Events, nil
}

// addEventbriteDataToExistingEvents attempts to add the Eventbrite URL and Eventbrite ID to events in the database that
// have a matching name and date but do not already have an Eventbrite URL and ID.
func addEventbriteDataToExistingEvents(db *sqlx.DB, ebEvents []EventbriteEvent, dbEvents []model.ExternalEvent) error {
	// create a map of the Facebook events to quickly look an event up by Eventbrite URL
	fbEventsMap := make(map[string]model.ExternalEvent)
	for _, fbEvent := range dbEvents {
		if fbEvent.EventbriteURL != "" {
			fbEventsMap[fbEvent.EventbriteURL] = fbEvent
		}
	}
	// loop through events
	for _, ebEvent := range ebEvents {
		if _, ok := fbEventsMap[ebEvent.URL]; !ok {
			log.Println("Attempting to add Eventbrite details to DB for", ebEvent.Name.Text)
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

func createOrUpdateEventbriteEvents(db *sqlx.DB, chapter model.ChapterWithToken, dbEvents []model.ExternalEvent, ebEvents []EventbriteEvent) error {

	// create a map of the Eventbrite events to quickly look an event up by URL
	ebEventsMap := make(map[string]EventbriteEvent)
	for _, ebEvent := range ebEvents {
		if ebEvent.URL != "" {
			ebEventsMap[ebEvent.URL] = ebEvent
		}
	}

	for _, dbEvent := range dbEvents {
		err := createOrUpdateEventbriteEvent(db, chapter, dbEvent, ebEventsMap)
		if err != nil {
			return err
		}
	}

	return nil
}

func createOrUpdateEventbriteEvent(db *sqlx.DB, chapter model.ChapterWithToken, dbEvent model.ExternalEvent, ebEventsMap map[string]EventbriteEvent) error {
	// if there is no Eventbrite URL in the database, we need to create the event on Eventbrite
	if dbEvent.EventbriteURL == "" {
		ebEvent, err := createEventbriteEvent(dbEvent, chapter)
		if err != nil {
			return err
		}
		// update the database w/ the information from Eventbrite
		event := model.ExternalEvent{
			ID:            dbEvent.ID,
			EventbriteURL: ebEvent.URL,
			EventbriteID:  ebEvent.ID,
		}
		err = model.AddEventbriteDetailsToEventByID(db, event)
		if err != nil {
			return err
		}
		return nil
	}

	// event has Eventbrite URL in the database, so first make sure that it's still listed on Eventbrite
	if _, ok := ebEventsMap[dbEvent.EventbriteURL]; !ok {
		log.Println("WARNING:", dbEvent.Name, "was deleted by a human on Eventbrite, but an Eventbrite URL still exists in our database!")
		// TODO: We should probably remove the Eventbrite ID & URL from the database since it no longer exists.
		return nil
	}

	// event still exists on Eventbrite where we expect it, so check if it need any updates
	// TODO: move this to an updateEventbriteEvent function
	ebEvent := ebEventsMap[dbEvent.EventbriteURL]
	if ebEvent.Name.Text != dbEvent.Name || ebEvent.Start.UTC != dbEvent.StartTime.Format(time.RFC3339) {
		log.Println("We need to update the event name & start time on Eventbrite for", ebEvent.Name)
		// TODO: update the event name, summary (w/ the name), and start time
	}
	// TODO: check if location changed
	// TODO: Compare description (Create new structured content version is it doesn't match)
	// TODO: Check if there are any cancelled events in the database. If they are still on Eventbrite, cancel them.
	return nil
}

func createEventbriteEvent(event model.ExternalEvent, chapter model.ChapterWithToken) (EventbriteEvent, error) {
	log.Println("Creating event on Eventbrite:", event.Name)

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
