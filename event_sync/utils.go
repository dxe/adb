package event_sync

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func callAPIPost(path string, req, resp interface{}) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(&req)
	if err != nil {
		return err
	}

	response, err := http.Post(path, "application/json", &body)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return errors.New("POST request failed. Status: " + strconv.Itoa(response.StatusCode))
	}

	err = json.NewDecoder(response.Body).Decode(&resp)
	if err != nil {
		return err
	}

	return nil
}

func callAPIGet(path string, resp interface{}) error {
	response, err := http.Get(path)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return errors.New("GET request failed. Status: " + strconv.Itoa(response.StatusCode))
	}
	err = json.NewDecoder(response.Body).Decode(&resp)

	return nil
}
