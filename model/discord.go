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

type DiscordMessage struct {
	Name      string `db:"message_name"`
	Text      string `db:"message_text"`
	UpdatedBy int    `db:"updated_by"`
}

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

func GetEmailFromDiscordToken(db *sqlx.DB, token string) (string, error) {
	query := `SELECT email
		FROM discord_users
		WHERE token = ?`
	var users []DiscordUser
	err := db.Select(&users, query, token)
	if err != nil {
		return "", errors.Wrap(err, "failed to select user")
	}
	if len(users) > 1 {
		return "", errors.New("found too many users")
	}
	if len(users) == 1 {
		return users[0].Email, nil
	}
	return "", nil
}

func GetDiscordMessage(db *sqlx.DB, messageName string) (string, error) {
	query := `SELECT message_name, message_text, updated_by
		FROM discord_messages
		WHERE message_name = ?`
	var messages []DiscordMessage
	err := db.Select(&messages, query, messageName)
	if err != nil {
		return "", errors.Wrap(err, "failed to select Discord message")
	}
	if len(messages) == 0 {
		return "", errors.New("could not find Discord message")
	}
	if len(messages) > 1 {
		return "", errors.New("found too many Discord messages")
	}
	if len(messages) == 1 {
		return messages[0].Text, nil
	}
	return "", nil
}

func SetDiscordMessage(db *sqlx.DB, message DiscordMessage) error {
	_, err := db.NamedExec(`REPLACE INTO discord_messages (message_name, message_text, last_updated, updated_by)
			VALUES (:message_name, :message_text, now(), :updated_by)`, message)
	if err != nil {
		return err
	}
	return nil
}
