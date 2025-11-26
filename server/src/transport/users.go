package transport

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/persistence"
	"github.com/gorilla/mux"
	"github.com/jmoiron/sqlx"
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

func UserJsonFromModel(u model.ADBUser) UserJson {
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

func UserJsonFromModels(users []model.ADBUser) []UserJson {
	out := make([]UserJson, 0, len(users))
	for _, u := range users {
		out = append(out, UserJsonFromModel(u))
	}
	return out
}

func UsersListHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	users, err := persistence.GetUsers(db, persistence.GetUserOptions{PopulateRoles: true})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"users": UserJsonFromModels(users),
	})
}

func UserGetHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	userID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, errors.Wrapf(err, "invalid user id %s", rawID))
		return
	}

	users, err := persistence.GetUsers(db, persistence.GetUserOptions{ID: userID, PopulateRoles: true})
	if err != nil {
		sendErrorMessage(w, err)
		return
	}
	if len(users) == 0 {
		sendErrorMessage(w, errors.Errorf("no user found with ID %d", userID))
		return
	}

	userJSON := UserJsonFromModel(users[0])

	writeJSON(w, map[string]interface{}{
		"user": userJSON,
	})
}

func UserCreateHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	input, err := CleanUserWithRolesData(r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	createdUser, err := persistence.CreateUserWithRoles(db, input)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status": "success",
		"user":   UserJsonFromModel(createdUser),
	})
}

func UserUpdateHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	vars := mux.Vars(r)
	rawID := vars["id"]
	userID, err := strconv.Atoi(rawID)
	if err != nil {
		sendErrorMessage(w, errors.Wrapf(err, "invalid user id %s", rawID))
		return
	}

	input, err := CleanUserWithRolesData(r.Body)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	if input.ID != 0 && input.ID != userID {
		sendErrorMessage(w, errors.Errorf("mismatched user ids %d and %d", input.ID, userID))
		return
	}

	input.ID = userID

	updatedUser, err := persistence.UpdateUserWithRoles(db, input)
	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, map[string]interface{}{
		"status": "success",
		"user":   UserJsonFromModel(updatedUser),
	})
}

func UserListHandler(w http.ResponseWriter, r *http.Request, db *sqlx.DB) {
	users, err := persistence.GetUsers(db, persistence.GetUserOptions{PopulateRoles: true})

	if err != nil {
		sendErrorMessage(w, err)
		return
	}

	writeJSON(w, UserJsonFromModels(users))
}

func CleanUserWithRolesData(body io.Reader) (model.ADBUser, error) {
	var payload UserJson

	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return model.ADBUser{}, err
	}

	user := model.ADBUser{
		ID:        payload.ID,
		Email:     strings.TrimSpace(payload.Email),
		Name:      strings.TrimSpace(payload.Name),
		Disabled:  payload.Disabled,
		ChapterID: payload.ChapterID,
	}

	return user, nil
}
