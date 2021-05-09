package form_processor

import (
	"database/sql"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

func getResponsesToProcess(db *sqlx.DB, query string) ([]int, bool) {
	rawApplicationIds, getApplicationIdsErr := db.Query(query)
	if getApplicationIdsErr != nil {
		log.Error().Msgf("failde to get applicationIds: %s", getApplicationIdsErr)
		return nil, false
	}
	defer rawApplicationIds.Close()
	var applicationIds []int
	for rawApplicationIds.Next() {
		var applicationId int
		err := rawApplicationIds.Scan(&applicationId)
		if err != nil {
			log.Error().Msgf("failed to scan applicationIds: %s", err)
			return nil, false
		}
		applicationIds = append(applicationIds, applicationId)
	}
	return applicationIds, true
}

func getProcessingStatus(db *sqlx.DB, query string, id int) (bool, bool) {
	var processed bool
	err := db.QueryRow(query, id).Scan(&processed)
	if err == sql.ErrNoRows {
		log.Error().Msgf("failed to find requested ID in requested table")
		return false, false
	}
	if err != nil {
		log.Error().Msgf("failed to check processing status for %d; %s", id, err)
		return false, false
	}
	return processed, true
}

func getEmail(db *sqlx.DB, query string, id int) (string, bool) {
	var email string
	err := db.QueryRow(query, id).Scan(&email)
	if err != nil {
		log.Error().Msg("failed to find requested ID in requested table")
		log.Error().Msgf("failed to get email for %d; %s", id, err)
		return "", false
	}
	return email, true
}

func countActivistsForEmail(db *sqlx.DB, email string) (int, bool) {
	var count int
	// TODO: is only getting the first row accceptable?
	err := db.QueryRow(countActivistsForEmailQuery, email).Scan(&count)
	if err != nil {
		log.Error().Msgf("failed to get email count for %s from activists tables (no match found or error?); %s", email, err)
		return 0, false
	}
	return count, true
}
