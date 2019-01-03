package model

import (
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"

	"encoding/json"
	"fmt"
	"io"
)

/** Constant and Variable Definitions */

const selectUserBaseQuery string = `
SELECT
  id,
  email,
  admin,
  disabled
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
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Admin    bool   `db:"admin"`
	Disabled bool   `db:"disabled"`
	Roles    []UserRole
}

type UserJSON struct {
	ID       int      `json:"id"`
	Email    string   `json:"email"`
	Admin    bool     `json:"admin"`
	Disabled bool     `json:"disabled"`
	Roles    []string `json:"roles"`
}

type GetUserOptions struct {
	ID       int
	Email    string
	Admin    bool
	Disabled bool
}

type GetUsersRolesOptions struct {
	Users []ADBUser
	Roles []string
}

var DevTestUser = ADBUser{
	ID:       1,
	Email:    "test@test.com",
	Admin:    true,
	Disabled: false,
}

type UserRole struct {
	UserID int    `db:"user_id"`
	Role   string `db:"role"`
}

type UserRoleJSON struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
}

/** Functions and Methods */

func GetADBUser(db *sqlx.DB, id int, email string) (ADBUser, error) {
	query := `
SELECT
  id,
  email,
  admin,
  disabled
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

	return *adbUser, nil
}

func GetUsersJSON(db *sqlx.DB, options GetUserOptions) ([]UserJSON, error) {
	if options.ID != 0 {
		return nil, errors.New("GetUsersJSON: Cannot include ID in options")
	}

	return getUsersJSON(db, options)
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
		i := userIDToIndex[r.UserID]
		users[i].Roles = append(users[i].Roles, r)
	}

	return users, nil
}

func getUsers(db *sqlx.DB, options GetUserOptions) ([]ADBUser, error) {
	query := selectUserBaseQuery

	var queryArgs []interface{}

	if options.ID != 0 {
		query += " WHERE id = ? "
		queryArgs = append(queryArgs, options.ID)
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
	/*
	  var queryArgs []interface{}
	  var whereClause []string

	  if len(options.Users) != 0 {
	    var userIds = []int{4, 6, 7}
	    //for _, user := range options.Users {
	    //  userIds = append(userIds, strconv.Itoa(user.ID))
	    //}
	    whereClause = append(whereClause, "ur.user_id IN (?)")
	    queryArgs = append(queryArgs, userIds)
	  }

	  if len(options.Roles) != 0 {
	    whereClause = append(whereClause, "ur.role IN (?)")
	    queryArgs = append(queryArgs, options.Roles)
	  }

	  if len(whereClause) != 0 {
	    query += ` WHERE ` + strings.Join(whereClause, " AND ")
	  }
	*/
	fmt.Println(query)
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
			ID:       u.ID,
			Email:    u.Email,
			Admin:    u.Admin,
			Disabled: u.Disabled,
			Roles:    roles,
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
		ID:       userJSON.ID,
		Email:    userJSON.Email,
		Admin:    userJSON.Admin,
		Disabled: userJSON.Disabled,
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

	result, err := db.NamedExec(`
INSERT INTO adb_users (
  email,
  admin,
  disabled
) VALUES (
  :email,
  :admin,
  :disabled
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

	_, err := db.NamedExec(`UPDATE adb_users
SET
  email = :email,
  admin = :admin,
  disabled = :disabled
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
  if (userRole.UserID == 0) {
    return 0, errors.New("Invalid User ID")
  }

  tx, err := db.Beginx()

  if err != nil {
    return userRole.UserID, errors.Wrapf(err, "Could not start transaction for User %d", userRole.UserID)
  }

  query := `
DELETE FROM users_roles
WHERE user_id = ? AND role = ?
`

  _, err = tx.Exec(query, userRole.UserID, userRole.Role)

  if err != nil {
    tx.Rollback()
    return userRole.UserID, errors.Wrapf(err, "Failed to delete User %d", userRole.UserID)
  }

  if err := tx.Commit(); err != nil {
    tx.Rollback()
    return userRole.UserID, errors.Wrapf(err, "Failed to commit delete transaction. Transaction rolled back for User %d", userRole.UserID)
  }

  return userRole.UserID, nil
}
