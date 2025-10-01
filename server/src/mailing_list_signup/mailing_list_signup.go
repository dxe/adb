package mailing_list_signup

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/dxe/adb/config"
)

// Corresponds to Signup defined by sign-up service:
// https://github.com/search?q=repo%3Adxe%2Fsignups-worker%20%22type%20Signup%20struct%20%7B%22&type=code
type Signup struct {
	Source          string
	Name            string
	Email           string
	Phone           string
	City            string
	State           string
	Zip             string
	Country         string
	Coords          string
	TargetChapterId int    `json:"target_chapter_id"`
	SourceChapterId int    `json:"source_chapter_id"`
	ActivistLevel   string `json:"activist_level"`
}

func Enqueue(signup Signup) error {
	if config.SignupURI == "" || config.SignupAPIKey == "" {
		return errors.New("mailing list signup URI or API key missing")
	}

	var body bytes.Buffer
	err := json.NewEncoder(&body).Encode(signup)
	if err != nil {
		return errors.New("failed to encode signup for mailing list")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", config.SignupURI, &body)
	if err != nil {
		return errors.New("failed to encode signup for mailing list")
	}
	req.Header.Add("X-api-key", config.SignupAPIKey)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return errors.New("failed to post to mailing list signup service")
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return errors.New("mailing list signup service returned status " + strconv.Itoa(resp.StatusCode))
	}

	return nil
}
