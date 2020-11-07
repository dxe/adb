package model

import (
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DiscordUser struct {
	ID        int    `db:"id"`
	Email     string `db:"email"`
	Token     string `db:"token"`
	Confirmed bool   `db:"confirmed"`
}

type DiscordUserStatus string

const (
	NotFound  DiscordUserStatus = "not found"
	Pending   DiscordUserStatus = "pending"
	Confirmed DiscordUserStatus = "confirmed"
)

func InsertOrUpdateDiscordUser(db *sqlx.DB, user DiscordUser) error {
	_, err := db.NamedExec(`REPLACE INTO discord_users ( id, email, token )
		VALUES ( :id, :email, :token )`, user)
	if err != nil {
		return errors.Wrapf(err, "failed to insert or update user %d", user.ID)
	}
	return nil
}

func GetDiscordUserStatus(db *sqlx.DB, id int) (DiscordUserStatus, error) {
	query := `SELECT id, email, token, confirmed
		FROM discord_users
		WHERE id = ?`
	var users []DiscordUser
	err := db.Select(&users, query, id)
	if err != nil {
		return NotFound, errors.Wrap(err, "failed to select user")
	}
	if len(users) > 1 {
		return NotFound, errors.New("found too many users")
	}
	if len(users) == 1 {
		if users[0].Confirmed {
			return Confirmed, nil
		} else {
			return Pending, nil
		}
	}
	return NotFound, nil
}

func ConfirmDiscordUser(db *sqlx.DB, user DiscordUser) error {
	_, err := db.NamedExec(`UPDATE discord_users SET confirmed = 1 WHERE id = :id AND token = :token`, user)
	if err != nil {
		return err
	}
	return nil
}
