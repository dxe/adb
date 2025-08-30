package model

import (
	"encoding/json"
	"io"
	"log"

	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

/** Constant and Variable Definitions */

const selectUserBaseQuery string = `
SELECT
  id,
  email,
  name,
  admin,
  disabled,
  chapter_id
FROM adb_users
`

const selectUsersRolesBaseQuery string = `
SELECT
  ur.user_id,
  ur.role
FROM users_roles ur
`

/** Type Definitions */

type ADBUser struct {
	ID          int    `db:"id"`
	Email       string `db:"email"`
	Name        string `db:"name"`
	Admin       bool   `db:"admin"`
	Disabled    bool   `db:"disabled"`
	Roles       []UserRole
	ChapterID   int    `db:"chapter_id"`
	ChapterName string `db:"chapter_name"`
}

type UserJSON struct {
	ID        int      `json:"id"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Admin     bool     `json:"admin"`
	Disabled  bool     `json:"disabled"`
	Roles     []string `json:"roles"`
	ChapterID int      `json:"chapter_id"`
}

type GetUserOptions struct {
	ID   int
	Name string
}

type GetUsersRolesOptions struct {
	Users []ADBUser
	Roles []string
}

type UserRole struct {
	UserID int    `db:"user_id"`
	Role   string `db:"role"`
}

type UserRoleJSON struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

const DevTestUserId = 1
const DevTestUserEmail = "test@example.org"

/** Functions and Methods */

func GetADBUser(db *sqlx.DB, id int, email string) (ADBUser, error) {
	query := `
SELECT
  id,
  email,
  name,
  admin,
  disabled,
  chapter_id,
  @chapter_name := IFNULL((
    SELECT name
    FROM fb_pages
    WHERE fb_pages.chapter_id = adb_users.chapter_id
  ),"") AS chapter_name
FROM adb_users
`
	var queryArgs []interface{}
	if id != 0 {
		query += " WHERE id = ? "
		queryArgs = append(queryArgs, id)
	} else if email != "" {
		query += " WHERE email = ? "
		queryArgs = append(queryArgs, email)
	} else {
		return ADBUser{}, errors.New("Must supply id or email")
	}

	adbUser := &ADBUser{}
	if err := db.Get(adbUser, query, queryArgs...); err != nil {
		return ADBUser{}, errors.Wrapf(err, "cannot get adb user %d", id)
	}

	usersRolesOptions := GetUsersRolesOptions{
		Users: []ADBUser{*adbUser},
	}

	usersRoles, err := getUsersRoles(db, usersRolesOptions)

	// We don't want non-SF Bay users to have access to any of the other roles, so just replace it.
	if adbUser.ChapterName != SFBayChapterName {
		usersRoles = []UserRole{{
			UserID: adbUser.ID,
			Role:   "non-sfbay",
		}}
	}

	if err != nil || len(usersRoles) == 0 {
		return *adbUser, nil
	}

	for _, r := range usersRoles {
		if r.UserID == adbUser.ID {
			adbUser.Roles = append(adbUser.Roles, r)
		}
	}

	log.Println("[User access]", adbUser.Name, "-", adbUser.Email)

	return *adbUser, nil
}

func GetUsersJSON(db *sqlx.DB) ([]UserJSON, error) {
	return getUsersJSON(db, GetUserOptions{})
}

func getUsersJSON(db *sqlx.DB, options GetUserOptions) ([]UserJSON, error) {
	users, err := GetUsers(db, options)

	if err != nil {
		return nil, err
	}

	return buildUserJSONArray(users), nil
}

func GetUsers(db *sqlx.DB, options GetUserOptions) ([]ADBUser, error) {
	users, err := getUsers(db, options)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return users, nil
	}

	usersRolesOptions := GetUsersRolesOptions{
		Users: users,
	}

	usersRoles, err := getUsersRoles(db, usersRolesOptions)

	if err != nil {
		return nil, err
	}

	if len(usersRoles) == 0 {
		return users, nil
	}

	userIDToIndex := map[int]int{}
	for i, user := range users {
		userIDToIndex[user.ID] = i
	}

	for _, r := range usersRoles {
		if a, ok := userIDToIndex[r.UserID]; ok {
			users[a].Roles = append(users[a].Roles, r)
		}
	}

	return users, nil
}

func getUsers(db *sqlx.DB, options GetUserOptions) ([]ADBUser, error) {
	query := selectUserBaseQuery

	var queryArgs []interface{}

	if options.ID != 0 && options.Name != "" {
		return nil, errors.New("You may provide ID or Name but not both.")
	}

	if options.ID != 0 {
		query += " WHERE id = ? "
		queryArgs = append(queryArgs, options.ID)
	}

	if options.Name != "" {
		query += " WHERE name = ? "
		queryArgs = append(queryArgs, options.Name)
	}

	query += " ORDER BY email "

	var users []ADBUser
	if err := db.Select(&users, query, queryArgs...); err != nil {
		return nil, errors.Wrapf(err, "failed to get users")
	}

	return users, nil
}

func getUsersRoles(db *sqlx.DB, options GetUsersRolesOptions) ([]UserRole, error) {
	query := selectUsersRolesBaseQuery

	var userRoles []UserRole
	err := db.Select(&userRoles, query)

	if err != nil {
		return nil, errors.Wrap(err, "failed to select UserRoles")
	}

	if len(userRoles) == 0 {
		return nil, nil
	}

	return userRoles, nil
}

func buildUserJSONArray(users []ADBUser) []UserJSON {
	var usersJSON []UserJSON

	for _, u := range users {
		var roles []string
		for _, userRole := range u.Roles {
			roles = append(roles, userRole.Role)
		}

		usersJSON = append(usersJSON, UserJSON{
			ID:        u.ID,
			Email:     u.Email,
			Name:      u.Name,
			Admin:     u.Admin,
			Disabled:  u.Disabled,
			Roles:     roles,
			ChapterID: u.ChapterID,
		})
	}

	return usersJSON
}

func CleanUserData(body io.Reader) (ADBUser, error) {
	var userJSON UserJSON
	err := json.NewDecoder(body).Decode(&userJSON)

	if err != nil {
		return ADBUser{}, err
	}

	user := ADBUser{
		ID:        userJSON.ID,
		Email:     userJSON.Email,
		Name:      userJSON.Name,
		Admin:     userJSON.Admin,
		Disabled:  userJSON.Disabled,
		ChapterID: userJSON.ChapterID,
	}

	return user, nil
}

func GetUserJSON(db *sqlx.DB, options GetUserOptions) (UserJSON, error) {
	if options.ID == 0 {
		return UserJSON{}, errors.New("GetUserJSON: Must include ID in options")
	}

	users, err := getUsersJSON(db, options)
	if err != nil {
		return UserJSON{}, err
	} else if len(users) == 0 {
		return UserJSON{}, errors.New("Could not find any users")
	} else if len(users) > 1 {
		return UserJSON{}, errors.New("Found too many users")
	}

	return users[0], nil
}

func CreateUser(db *sqlx.DB, user ADBUser) (int, error) {
	if user.ID != 0 {
		return 0, errors.New("User ID must be 0")
	}

	if user.Email == "" {
		return 0, errors.New("User Email cannot be empty")
	}

	if user.Name == "" {
		return 0, errors.New("User Name cannot be empty")
	}

	result, err := db.NamedExec(`
INSERT INTO adb_users (
  email,
  name,
  admin,
  disabled,
  chapter_id
) VALUES (
  :email,
  :name,
  :admin,
  :disabled,
  :chapter_id
)`, user)

	if err != nil {
		return 0, errors.Wrapf(err, "Could not create user: %s", user.Email)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrapf(err, "Could not get LastInsertId for %s", user.Email)
	}

	return int(id), nil
}

func UpdateUser(db *sqlx.DB, user ADBUser) (int, error) {
	if user.ID == 0 {
		return 0, errors.New("User ID cannot be 0")
	}

	if user.Email == "" {
		return 0, errors.New("User Email cannot be empty")
	}

	if user.Name == "" {
		return 0, errors.New("User Name cannot be empty")
	}

	_, err := db.NamedExec(`UPDATE adb_users
SET
  email = :email,
  name  = :name,
  admin = :admin,
  disabled = :disabled,
  chapter_id = :chapter_id
WHERE
id = :id`, user)

	if err != nil {
		return 0, errors.Wrap(err, "failed to update user data")
	}

	return user.ID, nil
}

func RemoveUser(db *sqlx.DB, userID int) (int, error) {
	if userID == 0 {
		return 0, errors.New("User ID not provided")
	}

	// Using a transaction here will allow us to easily
	// extend this feature in the future. The adb_user model
	// might become more complicated with relationships to other models

	tx, err := db.Beginx()

	if err != nil {
		return 0, errors.Wrap(err, "failed to create transaction")
	}

	query := `
    DELETE FROM adb_users
    WHERE id = ?
  `

	_, err = tx.Exec(query, userID)

	if err != nil {
		tx.Rollback()
		return 0, errors.Wrapf(err, "failed to delete user %d", userID)
	}

	if err := tx.Commit(); err != nil {
		tx.Rollback()
		return 0, errors.Wrapf(err, "failed to commit delete transaction for user %d", userID)
	}

	return userID, nil
}

func CreateUserRole(db *sqlx.DB, userRole UserRole) (int, error) {
	if userRole.UserID == 0 {
		return 0, errors.New("Invalid User ID")
	}

	if userRole.Role == "" {
		return userRole.UserID, errors.New("Role cannot be empty")
	}

	_, err := db.Exec(`
INSERT INTO users_roles (user_id, role)
VALUES (?, ?)
`, userRole.UserID, userRole.Role)

	if err != nil {
		return userRole.UserID, errors.Wrapf(err, "Could not add User Role for User %d", userRole.UserID)
	}

	return userRole.UserID, nil
}

func RemoveUserRole(db *sqlx.DB, userRole UserRole) (int, error) {
	if userRole.UserID == 0 {
		return 0, errors.New("Invalid User ID")
	}

	query := `
DELETE FROM users_roles
WHERE user_id = ? AND role = ?
`

	_, err := db.Exec(query, userRole.UserID, userRole.Role)

	if err != nil {
		return userRole.UserID, errors.Wrapf(err, "Failed to delete User %d", userRole.UserID)
	}

	return userRole.UserID, nil
}
