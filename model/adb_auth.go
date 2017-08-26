package model

import (
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
)

/** Type Definitions */

type ADBUser struct {
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Admin    bool   `db:"admin"`
	Disabled bool   `db:"disabled"`
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
