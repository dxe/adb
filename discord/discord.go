package discord

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/dxe/adb/config"
	"log"
	"net/http"
	"strconv"
)

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

func AddUserRole(userID int, role string) error {

	url := config.DiscordBotBaseUrl + "/roles/add"

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

	url := config.DiscordBotBaseUrl + "/update_nickname"

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

	url := config.DiscordBotBaseUrl + "/send_message"

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
