package event_sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"mime/multipart"
	"net/http"
	"path"
	"strconv"
	"strings"
)

type Image struct {
	Buffer []byte
	Name   string
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

	image.Buffer = make([]byte, resp.ContentLength)
	_, err = io.ReadFull(resp.Body, image.Buffer)
	if err != nil {
		return image, err
	}

	image.Name = path.Base(imageUrl)
	image.Name = image.Name[:strings.Index(image.Name, "?")]

	return image, nil
}

func getEventbriteUploadToken(token string) (EventbriteUploadData, error) {
	var data EventbriteUploadData

	url := "https://www.eventbriteapi.com/v3/media/upload/?type=image-event-logo" +
		"&token=" + token

	resp, err := http.Get(url)
	if err != nil {
		return data, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return data, errors.New("failed to get S3 details for uploading image. Status: " + strconv.Itoa(resp.StatusCode))
	}

	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return data, err
	}
	return data, nil
}

func uploadImageToEventbrite(image Image, data EventbriteUploadData) error {
	// create a buffer to store the form data that's needed for S3
	buf := new(bytes.Buffer)
	w := multipart.NewWriter(buf)
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
	resp, err := http.Post(data.UploadURL, w.FormDataContentType(), buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 204 {  // This endpoint returns 204 instead of 200.
		return errors.New("failed to upload image to S3 bucket. Status: " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}

func notifyEventbriteOfNewImage(token string, uploadToken string) (string, error) {
	url := "https://www.eventbriteapi.com/v3/media/upload/?token=" + token

	type topLeft struct {
		X int `json:"x"`
		Y int `json:"y"`
	}

	type cropMask struct {
		TopLeft topLeft `json:"top_left"`
		Width   int     `json:"width"`
		Height  int     `json:"height"`
	}

	type request struct {
		UploadToken string   `json:"upload_token"`
		CropMask    cropMask `json:"crop_mask"`
	}

	type response struct {
		ID string `json:"id"`
	}

	req, err := json.Marshal(request{
		UploadToken: uploadToken,
		CropMask: cropMask{
			TopLeft: topLeft{
				X: 1,
				Y: 1,
			},
			Width:  1280,
			Height: 640,
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
		return "", errors.New("failed to notify Eventbrite of image upload. Status: " + strconv.Itoa(resp.StatusCode))
	}

	var respData response
	err = json.NewDecoder(resp.Body).Decode(&respData)
	if err != nil {
		return "", err
	}
	return respData.ID, nil
}
