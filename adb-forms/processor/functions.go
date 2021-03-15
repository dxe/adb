package processor

import (
	"database/sql"
	"github.com/rs/zerolog/log"
)

func getResponsesToProcess(db *sql.DB, query string) ([]int, bool) {
	rawApplicationIds, getApplicationIdsErr := db.Query(query)
	defer rawApplicationIds.Close()
	if getApplicationIdsErr != nil {
		log.Error().Msgf("error getting applicationIds: %s", getApplicationIdsErr)
		return nil, false
	}
	var applicationIds []int
	for rawApplicationIds.Next() {
		var applicationId int
		err := rawApplicationIds.Scan(&applicationId)
		if err != nil {
			log.Error().Msgf("error scanning applicationIds: %s", err)
			return nil, false
		}
		applicationIds = append(applicationIds, applicationId)
	}
	return applicationIds, true
}

func getProcessingStatus(db *sql.DB, query string, id int) (bool, bool) {
	var processed bool
	err := db.QueryRow(query, id).Scan(&processed)
	if err == sql.ErrNoRows {
		log.Error().Msgf("Could not find requested ID in requested table")
		return false, false
	}
	if err != nil {
		log.Error().Msgf("ERROR checking processing status for %d; %s", id, err)
		return false, false
	}
	return processed, true
}

func getEmail(db *sql.DB, query string, id int) (string, bool) {
	var email string
	err := db.QueryRow(query, id).Scan(&email)
	if err != nil {
		log.Error().Msg("Could not find requested ID in requested table")
		log.Error().Msgf("ERROR getting email for %d; %s", id, err)
		return "", false
	}
	return email, true
}

func countActivistsForEmail(db *sql.DB, email string) (int, bool) {
	var count int
	// TODO: is only getting the first row accceptable?
	err := db.QueryRow(countActivistsForEmailQuery, email).Scan(&count)
	if err != nil {
		log.Error().Msg("No match found or error?")
		log.Error().Msgf("ERROR getting email count for %s from activists tables; %s", email, err)
		return 0, false
	}
	return count, true
}
