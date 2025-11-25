package model

import (
	"database/sql"
	"encoding/json"
	"io"
	"log"
	"strings"

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

type GetUserOptions struct {
	ID   int
	Name string
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

	usersRoles, err := getUsersRoles(db)

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

func GetUsers(db *sqlx.DB, options GetUserOptions) ([]ADBUser, error) {
	users, err := getUsers(db, options)

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return users, nil
	}

	usersRoles, err := getUsersRoles(db)

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

func getUsersRoles(db *sqlx.DB) ([]UserRole, error) {
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

func CleanUserWithRolesData(body io.Reader) (ADBUser, error) {
	// Temporary until UserRole is replaced with string.
	var payload struct {
		ID        int      `json:"id"`
		Email     string   `json:"email"`
		Name      string   `json:"name"`
		Disabled  bool     `json:"disabled"`
		Roles     []string `json:"roles"`
		ChapterID int      `json:"chapter_id"`
	}

	if err := json.NewDecoder(body).Decode(&payload); err != nil {
		return ADBUser{}, err
	}

	// Temporary until UserRole is replaced with string.
	roles := makeRolesStructArray(payload.ID, payload.Roles)

	user := ADBUser{
		ID:        payload.ID,
		Email:     strings.TrimSpace(payload.Email),
		Name:      strings.TrimSpace(payload.Name),
		Disabled:  payload.Disabled,
		ChapterID: payload.ChapterID,
		Roles:     roles,
	}

	user.Admin = roleListHas(payload.Roles, "admin")

	return user, nil
}

func makeRolesStructArray(id int, strings []string) []UserRole {
	structs := make([]UserRole, 0, len(strings))
	for _, r := range strings {
		structs = append(structs, UserRole{
			UserID: id,
			Role:   r,
		})
	}
	return structs
}
func makeRolesStringArray(structs []UserRole) []string {
	strings := make([]string, 0, len(structs))
	for _, r := range structs {
		strings = append(strings, r.Role)
	}
	return strings
}

func roleListHas(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}

func syncUserRolesTx(tx *sqlx.Tx, userID int, roles []string) error {
	existingRoles, err := getUserRolesTx(tx, userID)
	if err != nil {
		return err
	}

	existingSet := map[string]struct{}{}
	for _, r := range existingRoles {
		existingSet[r] = struct{}{}
	}

	desiredSet := map[string]struct{}{}
	for _, r := range roles {
		desiredSet[r] = struct{}{}
	}

	for role := range existingSet {
		if _, ok := desiredSet[role]; ok {
			continue
		}
		if _, err := tx.Exec(`DELETE FROM users_roles WHERE user_id = ? AND role = ?`, userID, role); err != nil {
			return errors.Wrapf(err, "failed to remove role %s for user %d", role, userID)
		}
	}

	for role := range desiredSet {
		if _, ok := existingSet[role]; ok {
			continue
		}
		if _, err := tx.Exec(`INSERT INTO users_roles (user_id, role) VALUES (?, ?)`, userID, role); err != nil {
			return errors.Wrapf(err, "failed to add role %s for user %d", role, userID)
		}
	}

	return nil
}

func getUserRolesTx(tx *sqlx.Tx, userID int) ([]string, error) {
	var roles []string
	if err := tx.Select(&roles, `SELECT role FROM users_roles WHERE user_id = ?`, userID); err != nil {
		return nil, errors.Wrapf(err, "failed to fetch roles for user %d", userID)
	}
	return roles, nil
}

func CreateUserWithRoles(db *sqlx.DB, user ADBUser) (ADBUser, error) {
	if user.ID != 0 {
		return ADBUser{}, errors.New("User ID must be 0 when creating a user")
	}

	user.Email = strings.TrimSpace(user.Email)
	user.Name = strings.TrimSpace(user.Name)
	user.Admin = roleListHas(makeRolesStringArray(user.Roles), "admin")

	tx, err := db.Beginx()
	if err != nil {
		return ADBUser{}, errors.Wrap(err, "failed to start create user transaction")
	}
	defer tx.Rollback()

	var existing struct {
		ID       int  `db:"id"`
		Disabled bool `db:"disabled"`
	}

	err = tx.Get(&existing, `SELECT id, disabled FROM adb_users WHERE email = ?`, user.Email)
	if err != nil && err != sql.ErrNoRows {
		return ADBUser{}, errors.Wrapf(err, "failed to check existing user %s", user.Email)
	}
	if err == nil {
		if existing.Disabled {
			return ADBUser{}, errors.Errorf("user with email %s already exists and is suspended", user.Email)
		}
		return ADBUser{}, errors.Errorf("user with email %s already exists", user.Email)
	}

	result, err := tx.NamedExec(`
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
		return ADBUser{}, errors.Wrapf(err, "Could not create user: %s", user.Email)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return ADBUser{}, errors.Wrapf(err, "Could not get LastInsertId for %s", user.Email)
	}

	userID := int(id)

	if err := syncUserRolesTx(tx, userID, makeRolesStringArray(user.Roles)); err != nil {
		return ADBUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return ADBUser{}, errors.Wrap(err, "failed to commit create user transaction")
	}

	users, err := GetUsers(db, GetUserOptions{ID: userID})
	if err != nil {
		return ADBUser{}, err
	}
	if len(users) == 0 {
		return ADBUser{}, errors.Errorf("no user found with ID %d after create", userID)
	}
	return users[0], nil
}

func UpdateUserWithRoles(db *sqlx.DB, user ADBUser) (ADBUser, error) {
	if user.ID == 0 {
		return ADBUser{}, errors.New("User ID cannot be 0 when updating a user")
	}

	user.Email = strings.TrimSpace(user.Email)
	user.Name = strings.TrimSpace(user.Name)
	user.Admin = roleListHas(makeRolesStringArray(user.Roles), "admin")

	tx, err := db.Beginx()
	if err != nil {
		return ADBUser{}, errors.Wrap(err, "failed to start update user transaction")
	}
	defer tx.Rollback()

	var existing struct {
		ID       int  `db:"id"`
		Disabled bool `db:"disabled"`
	}

	err = tx.Get(&existing, `SELECT id, disabled FROM adb_users WHERE email = ?`, user.Email)
	if err != nil && err != sql.ErrNoRows {
		return ADBUser{}, errors.Wrapf(err, "failed to check existing user %s", user.Email)
	}
	if err == nil && existing.ID != user.ID {
		if existing.Disabled {
			return ADBUser{}, errors.Errorf("user with email %s already exists and is suspended", user.Email)
		}
		return ADBUser{}, errors.Errorf("user with email %s already exists", user.Email)
	}

	result, err := tx.NamedExec(`UPDATE adb_users
SET
  email = :email,
  name  = :name,
  admin = :admin,
  disabled = :disabled,
  chapter_id = :chapter_id
	WHERE
	id = :id`, user)

	if err != nil {
		return ADBUser{}, errors.Wrap(err, "failed to update user data")
	}

	rowsUpdated, err := result.RowsAffected()
	if err != nil {
		return ADBUser{}, errors.Wrap(err, "failed to read rows affected when updating user")
	}

	if rowsUpdated == 0 {
		return ADBUser{}, errors.Errorf("no user found with ID %d", user.ID)
	}

	if err := syncUserRolesTx(tx, user.ID, makeRolesStringArray(user.Roles)); err != nil {
		return ADBUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return ADBUser{}, errors.Wrap(err, "failed to commit update user transaction")
	}

	users, err := GetUsers(db, GetUserOptions{ID: user.ID})
	if err != nil {
		return ADBUser{}, err
	}
	if len(users) == 0 {
		return ADBUser{}, errors.Errorf("no user found with ID %d after update", user.ID)
	}
	return users[0], nil
}
