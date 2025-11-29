package transport

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/dxe/adb/model"
	"github.com/gorilla/mux"
	"github.com/pkg/errors"
)

type UserJson struct {
	ID        int      `json:"id"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Disabled  bool     `json:"disabled"`
	Roles     []string `json:"roles"`
	ChapterID int      `json:"chapter_id"`
}

type UserRoleJSON struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

func userFromJson(json UserJson) (model.ADBUser, error) {
	user := model.ADBUser{
		ID:        json.ID,
		Email:     strings.TrimSpace(json.Email),
		Name:      strings.TrimSpace(json.Name),
		Disabled:  json.Disabled,
		Roles:     json.Roles,
		ChapterID: json.ChapterID,
	}
	if err := model.ValidateADBUser(user); err != nil {
		return model.ADBUser{}, err
	}

	return user, nil
}

func userToJson(u model.ADBUser) UserJson {
	var roles []string
	roles = append(roles, u.Roles...)

	return UserJson{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Disabled:  u.Disabled,
		Roles:     roles,
		ChapterID: u.ChapterID,
	}
}

func usersToJson(users []model.ADBUser) []UserJson {
	out := make([]UserJson, 0, len(users))
	for _, u := range users {
		out = append(out, userToJson(u))
	}
	return out
}

func UsersListHandler(w http.ResponseWriter, r *http.Request, repo model.UserRepository) {
	users, err := repo.GetUsers(model.GetUserOptions{PopulateRoles: true})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"users": usersToJson(users),
	})
}

func UserGetHandler(w http.ResponseWriter, r *http.Request, repo model.UserRepository) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	userID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, errors.Wrapf(err, "invalid user id %s", rawID))
		return
	}

	users, err := repo.GetUsers(model.GetUserOptions{ID: userID, PopulateRoles: true})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}
	if len(users) == 0 {
		sendErrorMessage(w, errors.Errorf("no user found with ID %d", userID))
		return
	}

	userJSON := userToJson(users[0])

	writeJSON(w, map[string]interface{}{
		"user": userJSON,
	})
}

func UserCreateHandler(w http.ResponseWriter, r *http.Request, repo model.UserRepository) {
	var payload UserJson
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		sendErrorMessage(w, err)
		return
	}

	input, err := userFromJson(payload)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	createdUser, err := repo.CreateUser(input)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status": "success",
		"user":   userToJson(createdUser),
	})
}

func UserUpdateHandler(w http.ResponseWriter, r *http.Request, repo model.UserRepository) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	userID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, errors.Wrapf(err, "invalid user id %s", rawID))
		return
	}

	var payload UserJson
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		sendErrorMessage(w, err)
		return
	}

	input, err := userFromJson(payload)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	if input.ID != 0 && input.ID != userID {
		sendErrorMessage(w, errors.Errorf("mismatched user ids %d and %d", input.ID, userID))
		return
	}

	input.ID = userID

	updatedUser, err := repo.UpdateUser(input)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status": "success",
		"user":   userToJson(updatedUser),
	})
}

func UserListHandler(w http.ResponseWriter, r *http.Request, repo model.UserRepository) {
	users, err := repo.GetUsers(model.GetUserOptions{PopulateRoles: true})

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, usersToJson(users))
}
