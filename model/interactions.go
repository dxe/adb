package model

import (
	"time"

	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
)

type Interaction struct {
	ID         int       `db:"id" json:"id"`
	ActivistID int       `db:"activist_id" json:"activist_id"`
	UserID     int       `db:"user_id" json:"user_id"`
	UserName   string    `db:"user_name" json:"user_name"`
	Timestamp  time.Time `db:"timestamp" json:"timestamp"` // TODO: ensure this type works fine for our needs here
	Method     string    `db:"method" json:"method"`
	Outcome    string    `db:"outcome" json:"outcome"`
	Notes      string    `db:"notes" json:"notes"`
}

func ListActivistInteractions(db *sqlx.DB, activistID int) ([]Interaction, error) {
	if activistID == 0 {
		return nil, errors.New("Activist ID must be supplied")
	}

	query := `SELECT interactions.id, activist_id, user_id, IFNULL(adb_users.name, '') as user_name, timestamp, method, outcome, notes
		FROM interactions
		LEFT JOIN adb_users on interactions.user_id = adb_users.id
		WHERE activist_id = ?
		ORDER BY timestamp desc`
	var interactions []Interaction
	err := db.Select(&interactions, query, activistID)
	if err != nil {
		return nil, errors.Wrap(err, "failed to select interactions")
	}

	if len(interactions) == 0 {
		return []Interaction{}, nil
	}

	return interactions, nil
}

func SaveInteraction(db *sqlx.DB, interaction Interaction) error {
	// if id == 0, insert new interaction
	if interaction.ID == 0 {
		_, err := db.NamedExec(`INSERT INTO interactions (activist_id, user_id, method, outcome, notes)
			VALUES (:activist_id, :user_id, :method, :outcome, :notes)
		`, interaction)
		if err != nil {
			return errors.Wrapf(err, "failed to insert interaction")
		}
		return nil
	}

	// otherwise, update existing interaction
	_, err := db.NamedExec(`UPDATE interactions
		SET 
		    activist_id = :activist_id,
    		user_id = :user_id,
    		timestamp = :timestamp,
			method = :method,
			outcome = :outcome,
			notes = :notes
		WHERE id = :id`, interaction)
	if err != nil {
		return errors.Wrapf(err, "failed to update interaction")
	}
	return nil
}

func DeleteInteraction(db *sqlx.DB, interactionID int) error {
	if interactionID == 0 {
		return errors.New("Interaction ID must be provided")
	}
	_, err := db.Exec(`DELETE FROM interactions
		WHERE id = ?`, interactionID)
	if err != nil {
		return errors.Wrapf(err, "failed to delete interaction %d", interactionID)
	}
	return nil
}
