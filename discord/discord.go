package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strconv"
)

// TODO: move url & port to config
const DISCORD_BOT_BASE_URL = "http://localhost:6070"

func GetUserRoles(userID int) map[int]string {

	url := DISCORD_BOT_BASE_URL + "/roles/get?user=" + strconv.Itoa(userID)

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

func AddUserRole(userID int, role string) error {

	url := DISCORD_BOT_BASE_URL + "/roles/add"

	requestBody, err := json.Marshal(map[string]string{
		"user": strconv.Itoa(userID),
		"role": role,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("ERROR adding Discord role", userID, err)
		return errors.New("ERROR adding Discord role.")
	}
	return nil
}

func UpdateNickname(userID int, nickname string) error {

	url := DISCORD_BOT_BASE_URL + "/update_nickname"

	requestBody, err := json.Marshal(map[string]string{
		"user": strconv.Itoa(userID),
		"name": nickname,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("ERROR updating Discord nickname", userID, err)
		return errors.New("Error updating Discord nickname.")
	}
	return nil
}

func SendMessage(userID int, role string) error {

	url := DISCORD_BOT_BASE_URL + "/send_message"

	requestBody, err := json.Marshal(map[string]string{
		"recipient": strconv.Itoa(userID),
		"message":   role,
	})
	if err != nil {
		return err
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("ERROR adding Discord role", userID, err)
		return errors.New("ERROR adding Discord role.")
	}
	return nil
}