package event_sync

import (
	"bytes"
	"errors"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/dxe/adb/model"
)

const eventbriteAPIBaseURL = "https://www.eventbriteapi.com/v3"

type EventbriteEvents struct {
	Events []EventbriteEvent `json:"events"`
}

type EventbriteEvent struct {
	ID            string             `json:"id,omitempty"`
	URL           string             `json:"url,omitempty"`
	Name          EventbriteName     `json:"name"`
	Summary       string             `json:"summary"`
	Start         EventbriteDatetime `json:"start"`
	End           EventbriteDatetime `json:"end"`
	VenueID       string             `json:"venue_id"`
	LogoID        string             `json:"logo_id"`
	Currency      string             `json:"currency"`
	OnlineEvent   bool               `json:"online_event"`
	Listed        bool               `json:"listed"`
	Shareable     bool               `json:"shareable"`
	ShowRemaining bool               `json:"show_remaining"`
	HideEndDate   bool               `json:"hide_end_date"`
	CategoryID    string             `json:"category_id"`
	SubcategoryID string             `json:"subcategory_id"`
}

type EventbriteName struct {
	Text string `json:"text,omitempty"`
	HTML string `json:"html"`
}

type EventbriteDatetime struct {
	TimeZone string `json:"timezone"`
	Local    string `json:"local,omitempty"`
	UTC      string `json:"utc"`
}

type EventbriteUploadData struct {
	UploadData struct {
		AWSAccessKeyID string `json:"AWSAccessKeyId"`
		Bucket         string `json:"bucket"`
		ACL            string `json:"acl"`
		Key            string `json:"key"`
		Signature      string `json:"signature"`
		Policy         string `json:"policy"`
	} `json:"upload_data"`
	UploadURL   string `json:"upload_url"`
	UploadToken string `json:"upload_token"`
}

type EmptyRequest struct { // blank request body
	Empty string `json:",omitempty"`
}

func createEventbriteVenue(event model.ExternalEvent, chapter model.ChapterWithToken) (string, error) {
	// online events don't have a venue
	if event.LocationName == "Online" {
		return "", nil
	}

	path := eventbriteAPIBaseURL + "/organizations/" + chapter.EventbriteID +
		"/venues/?token=" + chapter.EventbriteToken

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
		ID      string  `json:"id,omitempty"` // this value is excluded when requesting a new venue, but is used for the response
		Name    string  `json:"name"`
		Address address `json:"address"`
	}
	type reqBody struct {
		Venue venue `json:"venue"`
	}
	req := reqBody{
		venue{
			Name: event.LocationName,
			Address: address{
				Address1:   event.LocationAddress,
				City:       event.LocationCity,
				Region:     event.LocationState,
				PostalCode: event.LocationZip,
				Country:    "US", // TODO: determine this automatically if using for international chapters (must be 2-digit code)
				Lat:        strconv.FormatFloat(event.Lat, 'f', 6, 64),
				Lng:        strconv.FormatFloat(event.Lng, 'f', 6, 64),
			},
		},
	}

	var resp venue

	err := callAPIPost(path, &req, &resp)
	if err != nil {
		return "", errors.New("failed to create venue on Eventbrite: " + err.Error())
	}

	return resp.ID, nil
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

func addEventToEventbrite(event model.ExternalEvent, chapter model.ChapterWithToken, venueID string, imageID string) (EventbriteEvent, error) {

	type reqBody struct {
		Event EventbriteEvent `json:"event"`
	}

	if event.EndTime.IsZero() {
		event.EndTime = event.StartTime.Add(time.Hour * 2)
	}

	path := eventbriteAPIBaseURL + "/organizations/" + chapter.EventbriteID +
		"/events/?token=" + chapter.EventbriteToken

	req := reqBody{
		Event: EventbriteEvent{
			Name: EventbriteName{
				HTML: event.Name,
			},
			Summary: event.Name,
			Start: EventbriteDatetime{
				TimeZone: "America/Los_Angeles", // TODO: change this to be flexible for other chapters
				UTC:      event.StartTime.UTC().Format(time.RFC3339),
			},
			End: EventbriteDatetime{
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
	}

	var resp EventbriteEvent

	err := callAPIPost(path, &req, &resp)
	if err != nil {
		return resp, errors.New("failed to add event to Eventbrite: " + err.Error())
	}

	return resp, nil
}

func addEventTicketClass(eventId string, token string) error {

	type ticketClass struct {
		Name          string `json:"name"`
		QuantityTotal int    `json:"quantity_total"`
		Free          bool   `json:"free"`
	}

	type reqBody struct {
		TicketClass ticketClass `json:"ticket_class"`
	}

	const ticketLimit = 500

	path := eventbriteAPIBaseURL + "/events/" + eventId +
		"/ticket_classes/?token=" + token

	req := reqBody{
		ticketClass{
			Name:          "General Admission",
			QuantityTotal: ticketLimit,
			Free:          true,
		},
	}

	var resp interface{} // we don't care about the response data

	err := callAPIPost(path, &req, &resp)
	if err != nil {
		return errors.New("failed to add ticket class to Eventbrite: " + err.Error())
	}

	return nil
}

func getNextDescriptionID(eventID string, token string) (string, error) {
	path := eventbriteAPIBaseURL + "/events/" + eventID + "/structured_content/" +
		"?token=" + token

	var version struct {
		VersionID string `json:"page_version_number"`
	}

	err := callAPIGet(path, &version)
	if err != nil {
		return "", errors.New("failed to get next description ID from Eventbrite: " + err.Error())
	}

	currentVersion, err := strconv.Atoi(version.VersionID)
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

	type reqBody struct {
		Publish bool     `json:"publish"`
		Modules []module `json:"modules"`
	}

	versionID, err := getNextDescriptionID(eventID, token)
	if err != nil {
		return err
	}

	path := eventbriteAPIBaseURL + "/events/" + eventID + "/structured_content/" + versionID +
		"/?token=" + token

	eventDescriptionHtml := "<p>" + strings.Replace(description, "\n", "</p><p>", -1) + "</p>"

	req := reqBody{
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
	}

	var resp interface{} // we don't care about the response data

	err = callAPIPost(path, &req, &resp)
	if err != nil {
		return errors.New("failed to add description to Eventbrite: " + err.Error())
	}

	return nil
}

func publishEvent(eventID string, token string) error {
	path := eventbriteAPIBaseURL + "/events/" + eventID + "/publish/" +
		"?token=" + token

	var req EmptyRequest
	var resp interface{} // we don't care about the response data

	err := callAPIPost(path, &req, &resp)
	if err != nil {
		return errors.New("failed to add publish event on Eventbrite: " + err.Error())
	}

	return nil
}

func getEventbriteUploadToken(token string) (EventbriteUploadData, error) {
	path := eventbriteAPIBaseURL + "/media/upload/?type=image-event-logo" +
		"&token=" + token

	var resp EventbriteUploadData

	err := callAPIGet(path, &resp)
	if err != nil {
		return EventbriteUploadData{}, errors.New("failed to get S3 details for uploading image: " + err.Error())
	}

	return resp, nil
}

func uploadImageToEventbrite(image Image, data EventbriteUploadData) error {
	// create a buffer to store the form data that's needed for S3
	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	// create the fields
	formFields := map[string]string{
		"AWSAccessKeyId": data.UploadData.AWSAccessKeyID,
		"bucket":         data.UploadData.Bucket,
		"acl":            data.UploadData.ACL,
		"key":            data.UploadData.Key,
		"signature":      data.UploadData.Signature,
		"policy":         data.UploadData.Policy,
	}
	for key, val := range formFields {
		err := w.WriteField(key, val)
		if err != nil {
			return err
		}
	}
	// add the file
	part, err := w.CreateFormFile("file", image.Name)
	if err != nil {
		return err
	}
	_, err = part.Write(image.Buffer)
	if err != nil {
		return err
	}
	err = w.Close()
	if err != nil {
		return err
	}

	// upload it
	resp, err := http.Post(data.UploadURL, w.FormDataContentType(), &buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		return errors.New("failed to upload image to S3 bucket. Status: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func notifyEventbriteOfNewImage(token string, uploadToken string) (string, error) {
	path := eventbriteAPIBaseURL + "/media/upload/?token=" + token

	type topLeft struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	type cropMask struct {
		TopLeft topLeft `json:"top_left"`
		Width   int     `json:"width"`
		Height  int     `json:"height"`
	}

	type reqBody struct {
		UploadToken string   `json:"upload_token"`
		CropMask    cropMask `json:"crop_mask"`
	}

	type respBody struct {
		ID string `json:"id"`
	}

	req := reqBody{
		UploadToken: uploadToken,
		CropMask: cropMask{
			TopLeft: topLeft{
				X: 1,
				Y: 1,
			},
			Width:  1280,
			Height: 640,
		},
	}

	var resp respBody

	err := callAPIPost(path, &req, &resp)
	if err != nil {
		return "", errors.New("failed to notify Eventbrite of image upload: " + err.Error())
	}

	return resp.ID, nil
}
