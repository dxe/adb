package event_sync

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

func postAPI(path string, req, resp interface{}) error {
	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(&req)
	if err != nil {
		return err
	}

	response, err := http.Post(path, "application/json", &body)
	if err != nil {
		return fmt.Errorf("error making POST request: %v", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		body := new(bytes.Buffer)
		body.ReadFrom(response.Body)
		return fmt.Errorf("request failed with status %v, body: %v", strconv.Itoa(response.StatusCode), body.String())
	}
	return json.NewDecoder(response.Body).Decode(&resp)
}

func getAPI(path string, resp interface{}) error {
	response, err := http.Get(path)
	if err != nil {
		return err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		body := new(bytes.Buffer)
		body.ReadFrom(response.Body)
		return fmt.Errorf("request failed with status %v, body: %v", strconv.Itoa(response.StatusCode), body.String())
	}
	return json.NewDecoder(response.Body).Decode(&resp)
}
