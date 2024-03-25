package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/dxe/adb/config"
)

func discordPostRequest(url string, body map[string]string) error {
	requestBody, err := json.Marshal(body)
	if err != nil {
		return errors.New(fmt.Sprintf("ERROR marshalling Discord POST request: %v\n", err.Error()))
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return errors.New(fmt.Sprintf("ERROR creating Discord POST request %v\n", err.Error()))
	}
	defer req.Body.Close()
	req.Header.Add("Auth", config.DiscordSecret)
	req.Header.Add("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.New(fmt.Sprintf("ERROR making Discord POST request to %v %v\n", url, err.Error()))
	}

	if resp.StatusCode != http.StatusOK {
		return errors.New(fmt.Sprintf("ERROR from Discord: status code %v", strconv.Itoa(resp.StatusCode)))
	}

	return nil
}

func GetUserRoles(userID int) map[int]string {
	url := config.DiscordBotBaseUrl + "/roles/get?user=" + strconv.Itoa(userID)

	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("ERROR getting Discord user roles", userID, err)
		return nil
	}
	// read the response & decode the json data
	data := make(map[int]string)
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		log.Println("ERROR getting Discord user roles", userID, err)
		return nil
	}

	return data
}

func AddUserRoles(userID int, roles []string) error {
	if len(roles) == 0 {
		return nil
	}

	for _, r := range roles {
		err := AddUserRole(userID, r)
		if err != nil {
			return err
		}
	}

	return nil
}

func AddUserRole(userID int, role string) error {
	url := config.DiscordBotBaseUrl + "/roles/add"

	body := map[string]string{
		"user": strconv.Itoa(userID),
		"role": role,
	}

	if err := discordPostRequest(url, body); err != nil {
		return errors.New("error adding discord user role: " + err.Error())
	}

	return nil
}

func UpdateNickname(userID int, nickname string) error {
	url := config.DiscordBotBaseUrl + "/update_nickname"

	body := map[string]string{
		"user": strconv.Itoa(userID),
		"name": nickname,
	}

	if err := discordPostRequest(url, body); err != nil {
		return errors.New("error updating discord user nickname: " + err.Error())
	}

	return nil
}

func SendMessage(userID int, role string) error {

	url := config.DiscordBotBaseUrl + "/send_message"

	body := map[string]string{
		"recipient": strconv.Itoa(userID),
		"message":   role,
	}

	if err := discordPostRequest(url, body); err != nil {
		return errors.New("error sending discord message: " + err.Error())
	}

	return nil
}
