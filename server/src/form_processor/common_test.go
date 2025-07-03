package form_processor

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

/* Common utils */
type activist struct {
	id int
}

func verifyActivistCount(t *testing.T, db *sqlx.DB, count int) {
	var actual int
	err := db.Get(&actual, `SELECT count(*) FROM activists;`)
	if err != nil {
		t.Fatalf("failed to count activists: %s", err)
	}
	if actual != count {
		t.Errorf("found %v activists, expected %v", actual, count)
	}
}

func verifyFormWasMarkedAsProcessed(t *testing.T, db *sqlx.DB, query string) {
	isProcessed, err := getProcessingStatus(db, query, 1)
	if err != nil {
		t.Errorf("failed to get form processing status: %v", err)
	}
	if !isProcessed {
		t.Error("form was not marked as processed")
	}
}

func verifyFormWasNotMarkedAsProcessed(t *testing.T, db *sqlx.DB, query string) {
	isProcessed, err := getProcessingStatus(db, query, 1)
	if err != nil {
		t.Errorf("failed to get form processing status: %v", err)
	}
	if isProcessed {
		t.Error("form was marked as processed")
	}
}
