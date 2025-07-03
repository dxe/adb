package form_processor

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

type formResponse struct {
	Id        int `db:"id"`
	ChapterId int `db:"chapter_id"`
}

func getResponsesToProcess(db *sqlx.DB, query string) ([]formResponse, bool) {
	var responses []formResponse
	err := db.Select(&responses, query)
	if err != nil {
		log.Error().Msgf("failed to get responses to process: %s", err)
		return nil, false
	}
	return responses, true
}

func getProcessingStatus(db *sqlx.DB, query string, id int) (bool, error) {
	var processed bool
	err := db.QueryRow(query, id).Scan(&processed)
	if err == sql.ErrNoRows {
		return false, errors.New("failed to find requested ID in requested table")
	}
	if err != nil {
		return false, fmt.Errorf("failed to check processing status for %d; %s", id, err)
	}
	return processed, nil
}

func getEmail(db *sqlx.DB, query string, id int) (string, bool) {
	var email string
	err := db.QueryRow(query, id).Scan(&email)
	if err != nil {
		log.Error().Msgf("failed to get email for %d; (failed to find requested ID in requested table) %s", id, err)
		return "", false
	}
	return email, true
}

func countActivistsForEmail(db *sqlx.DB, email string, chapterId int) (int, bool) {
	var count int
	err := db.Get(&count, `
		SELECT count(*)
		FROM activists
		WHERE hidden = 0 and email = ? and chapter_id = ?`,
		email,
		chapterId)
	if err != nil {
		log.Error().Msgf("failed to get email count for %s from activists tables (no match found or error?); %s", email, err)
		return 0, false
	}
	return count, true
}
