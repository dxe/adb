package event_sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/dxe/adb/model"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func createEventbriteVenue(event model.ExternalEvent, chapter model.ChapterWithToken) (string, error) {
	type address struct {
		Address1   string `json:"address_1"`
		City       string `json:"city"`
		Region     string `json:"region"`
		PostalCode string `json:"postal_code"`
		Country    string `json:"country"`
		Lat        string `json:"latitude"`
		Lng        string `json:"longitude"`
	}
	type venue struct {
		Name    string  `json:"name"`
		Address address `json:"address"`
	}
	type request struct {
		Venue venue `json:"venue"`
	}
	type response struct {
		ID string `json:"id"`
	}

	// online events don't have a venue
	if event.LocationName == "Online" {
		return "", nil
	}

	url := "https://www.eventbriteapi.com/v3/organizations/" + chapter.EventbriteID +
		"/venues/?token=" + chapter.EventbriteToken

	req, err := json.Marshal(request{
		venue{
			Name: event.LocationName,
			Address: address{
				Address1:   event.LocationAddress,
				City:       event.LocationCity,
				Region:     event.LocationState,
				PostalCode: event.LocationZip,
				Country:    "US", // TODO: determine this automatically if using for international chapters (must be 2 digit code)
				Lat:        strconv.FormatFloat(event.Lat, 'f', 6, 64),
				Lng:        strconv.FormatFloat(event.Lng, 'f', 6, 64),
			},
		},
	})
	if err != nil {
		return "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to create Venue on Eventbrite. Status: " + strconv.Itoa(resp.StatusCode))
	}

	var responseData response
	err = json.NewDecoder(resp.Body).Decode(&responseData)
	if err != nil {
		return "", err
	}
	return responseData.ID, nil

}

func createEventbriteImage(event model.ExternalEvent, chapter model.ChapterWithToken) (string, error) {
	image, err := downloadImageFromFacebook(event.Cover)
	if err != nil {
		return "", err
	}

	uploadData, err := getEventbriteUploadToken(chapter.EventbriteToken)
	if err != nil {
		return "", err
	}

	err = uploadImageToEventbrite(image, uploadData)
	if err != nil {
		return "", err
	}

	imageId, err := notifyEventbriteOfNewImage(chapter.EventbriteToken, uploadData.UploadToken)

	return imageId, nil
}

func addEventToEventbrite(event model.ExternalEvent, chapter model.ChapterWithToken, venueID string, imageID string) (string, string, error) {
	type name struct {
		HTML string `json:"html"`
	}

	type datetime struct {
		TimeZone string `json:"timezone"`
		UTC      string `json:"utc"`
	}

	type eventForReq struct {
		Name          name     `json:"name"`
		Summary       string   `json:"summary"`
		Start         datetime `json:"start"`
		End           datetime `json:"end"`
		VenueID       string   `json:"venue_id"`
		LogoID        string   `json:"logo_id"`
		Currency      string   `json:"currency"`
		OnlineEvent   bool     `json:"online_event"`
		Listed        bool     `json:"listed"`
		Shareable     bool     `json:"shareable"`
		ShowRemaining bool     `json:"show_remaining"`
		HideEndDate   bool     `json:"hide_end_date"`
		CategoryID    string   `json:"category_id"`
		SubcategoryID string   `json:"subcategory_id"`
	}

	type request struct {
		Event eventForReq `json:"event"`
	}

	type response struct {
		ID  string `json:"id"`
		URL string `json:"url"`
	}

	if event.EndTime.IsZero() {
		event.EndTime = event.StartTime.Add(time.Hour * 2)
	}

	url := "https://www.eventbriteapi.com/v3/organizations/" + chapter.EventbriteID +
		"/events/?token=" + chapter.EventbriteToken

	req, err := json.Marshal(request{
		Event: eventForReq{
			Name: name{
				HTML: event.Name,
			},
			Summary: event.Name,
			Start: datetime{
				TimeZone: "America/Los_Angeles", // TODO: change this to be flexible for other chapters
				UTC:      event.StartTime.UTC().Format(time.RFC3339),
			},
			End: datetime{
				TimeZone: "America/Los_Angeles", // TODO: change this to be flexible for other chapters
				UTC:      event.EndTime.UTC().Format(time.RFC3339),
			},
			VenueID:       venueID,
			LogoID:        imageID,
			Currency:      "USD",
			OnlineEvent:   event.LocationName == "Online",
			Listed:        true,
			Shareable:     true,
			ShowRemaining: false,
			HideEndDate:   true,
			CategoryID:    "111",   // Causes
			SubcategoryID: "11001", // Animal Welfare
		},
	})
	if err != nil {
		return "", "", err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", errors.New("failed to create event on Eventbrite. Status: " + strconv.Itoa(resp.StatusCode))
	}

	var respData response
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", "", err
	}
	return respData.ID, respData.URL, nil
}

func addEventTicketClass(eventId string, token string) error {

	type ticketClass struct {
		Name          string `json:"name"`
		QuantityTotal int    `json:"quantity_total"`
		Free          bool   `json:"free"`
	}

	type request struct {
		TicketClass ticketClass `json:"ticket_class"`
	}

	const ticketLimit = 500

	url := "https://www.eventbriteapi.com/v3/events/" + eventId +
		"/ticket_classes/?token=" + token

	req, err := json.Marshal(request{
		ticketClass{
			Name:          "General Admission",
			QuantityTotal: ticketLimit,
			Free:          true,
		},
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to create ticket class on Eventbrite. Status: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func getNextDescriptionID(eventID string, token string) (string, error) {
	url := "https://www.eventbriteapi.com/v3/events/" + eventID + "/structured_content/" +
		"?token=" + token

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", errors.New("failed to get event's structured content from Eventbrite. Status: " + strconv.Itoa(resp.StatusCode))
	}
	var versionData struct {
		VersionID string `json:"page_version_number"`
	}
	err = json.NewDecoder(resp.Body).Decode(&versionData)
	if err != nil {
		return "", err
	}
	currentVersion, err := strconv.Atoi(versionData.VersionID)
	if err != nil {
		return "", err
	}
	return strconv.Itoa(currentVersion + 1), nil
}

func updateEventDescription(eventID string, description string, token string) error {
	type moduleDataBody struct {
		Alignment string `json:"alignment"`
		Text      string `json:"text"`
	}

	type moduleData struct {
		Body moduleDataBody `json:"body"`
	}

	type module struct {
		Data moduleData `json:"data"`
		Type string     `json:"type"`
	}

	type request struct {
		Publish bool     `json:"publish"`
		Modules []module `json:"modules"`
	}

	versionID, err := getNextDescriptionID(eventID, token)
	if err != nil {
		return err
	}

	url := "https://www.eventbriteapi.com/v3/events/" + eventID + "/structured_content/" + versionID +
		"/?token=" + token

	eventDescriptionHtml := "<p>" + strings.Replace(description, "\n", "</p><p>", -1) + "</p>"

	req, err := json.Marshal(request{
		Publish: true,
		Modules: []module{
			{
				Data: moduleData{
					Body: moduleDataBody{
						Alignment: "left",
						Text:      eventDescriptionHtml,
					},
				},
				Type: "text",
			},
		},
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(req))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to update event description on Eventbrite. Status: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func publishEvent(eventID string, token string) error {
	url := "https://www.eventbriteapi.com/v3/events/" + eventID + "/publish/" +
		"?token=" + token

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(nil))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("failed to publish event on Eventbrite. Status: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
