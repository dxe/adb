package form_processor

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func getResponsesToProcess(db *sqlx.DB, query string) ([]int, bool) {
	rawIds, getIdsErr := db.Query(query)
	if getIdsErr != nil {
		log.Error().Msgf("failed to get responses to process: %s", getIdsErr)
		return nil, false
	}
	defer rawIds.Close()
	var ids []int
	for rawIds.Next() {
		var id int
		err := rawIds.Scan(&id)
		if err != nil {
			log.Error().Msgf("failed to scan responses to process: %s", err)
			return nil, false
		}
		ids = append(ids, id)
	}
	return ids, true
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

func countActivistsForEmail(db *sqlx.DB, email string) (int, bool) {
	const countActivistsForEmailQuery = `
		SELECT count(id) AS amount
		FROM activists
		WHERE hidden = 0 and email = ? and chapter_id = ` + model.SFBayChapterIdStr
	var count int
	// TODO: is only getting the first row accceptable?
	err := db.QueryRow(countActivistsForEmailQuery, email).Scan(&count)
	if err != nil {
		log.Error().Msgf("failed to get email count for %s from activists tables (no match found or error?); %s", email, err)
		return 0, false
	}
	return count, true
}
