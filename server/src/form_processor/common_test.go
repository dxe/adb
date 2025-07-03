package form_processor

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

/* Common utils */
type activist struct {
	id int
}

func verifyFormWasMarkedAsProcessed(t *testing.T, db *sqlx.DB, query string) {
	rawActivists, err := db.Query(getActivistsQuery)
	if err != nil {
		t.Fatalf("getActivistsQuery failed: %s", err)
	}
	defer rawActivists.Close()

	var activists []activist
	for rawActivists.Next() {
		var activist activist
		err := rawActivists.Scan(&activist.id)
		if err != nil {
			t.Error("error scanning activists: ", err)
		}
		activists = append(activists, activist)
	}
	if activists[0].id != 1 {
		t.Error("new activist was not inserted")
	}
	isProcessed, isSuccess := getProcessingStatus(db, query, 1)
	if !isSuccess || !isProcessed {
		t.Error("form was not marked as processed")
	}
}
