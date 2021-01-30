package event_sync

import (
	"errors"
	"io/ioutil"
	"net/http"
	"path"
	"strconv"
	"strings"

	"github.com/dxe/adb/model"
)

const facebookAPIBaseURL = "https://graph.facebook.com/v4.0"

type FacebookEvents struct {
	Data []FacebookEvent `json:"data"`
}

// fb event schema: https://developers.facebook.com/docs/graph-api/reference/event/
type FacebookEvent struct {
	ID              string          `json:"id"`
	Name            string          `json:"name"`
	Description     string          `json:"description"`
	StartTime       string          `json:"start_time"`
	EndTime         string          `json:"end_time"`
	AttendingCount  int             `json:"attending_count"`
	InterestedCount int             `json:"interested_count"`
	IsCanceled      bool            `json:"is_canceled"`
	IsOnline        bool            `json:"is_online"`
	Place           FacebookPlace   `json:"place"`
	Cover           FacebookCover   `json:"cover"`
	EventTimes      []FacebookEvent `json:"event_times"`
}

type FacebookPlace struct {
	Name     string           `json:"name"`
	Location FacebookLocation `json:"location"`
}

type FacebookLocation struct {
	City    string  `json:"city"`
	State   string  `json:"state"`
	Country string  `json:"country"`
	Street  string  `json:"street"`
	Zip     string  `json:"zip"`
	Lat     float64 `json:"latitude"`
	Lng     float64 `json:"longitude"`
}

type FacebookCover struct {
	Source string `json:"source"`
}

type Image struct {
	Buffer []byte
	Name   string
}

func getFacebookEvents(page model.ChapterWithToken) ([]FacebookEvent, error) {
	path := facebookAPIBaseURL + "/" + strconv.Itoa(page.ID) + "/events?include_canceled=1&fields=name,start_time,end_time,cover,attending_count,description,place,interested_count,is_canceled,event_times,is_online&limit=50&access_token=" + page.Token

	var events FacebookEvents
	err := callAPIGet(path, &events)
	if err != nil {
		return []FacebookEvent{}, errors.New("failed to get events from Facebook: " + err.Error())
	}

	return events.Data, nil
}

func getFacebookEvent(page model.ChapterWithToken, eventID string) (FacebookEvent, error) {
	path := facebookAPIBaseURL + "/" + eventID + "?fields=name,start_time,end_time,cover,attending_count,description,place,interested_count,is_canceled,event_times,is_online&limit=50&access_token=" + page.Token

	var event FacebookEvent
	err := callAPIGet(path, &event)
	if err != nil {
		return event, errors.New("failed to get individual event from Facebook: " + err.Error())
	}

	return event, nil
}

func downloadImageFromFacebook(imageUrl string) (Image, error) {
	var image Image

	resp, err := http.Get(imageUrl)
	if err != nil {
		return image, err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return image, errors.New("failed to get image from Facebook. Status: " + strconv.Itoa(resp.StatusCode))
	}

	image.Buffer, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return image, err
	}

	image.Name = path.Base(imageUrl)
	image.Name = image.Name[:strings.Index(image.Name, "?")]

	return image, nil
}
