package model

import (
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"

  "encoding/json"
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

/** Type Definitions */

type ADBUser struct {
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Admin    bool   `db:"admin"`
	Disabled bool   `db:"disabled"`
}

type UserJSON struct {
  ID        int     `json:"id"`
  Email     string  `json:"email"`
  Admin     bool    `json:"admin"`
  Disabled  bool    `json:"disabled"`
}

type GetUserOptions struct {
  ID        int
  Email     string
  Admin     bool
  Disabled  bool
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
		return ADBUser{}, errors.Wrapf(err, "cannot get adb user %d")
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
  return getUsers(db, options)
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

func buildUserJSONArray(users []ADBUser) []UserJSON {
	var usersJSON []UserJSON

	for _, u := range users {
		
		usersJSON = append(usersJSON, UserJSON{
			ID:            u.ID,
			Email:         u.Email,
			Admin:         u.Admin,
			Disabled:      u.Disabled,
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

  if err := checkForDangerousChars(userJSON.Email); err != nil {
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
