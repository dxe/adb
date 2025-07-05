package form_processor

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

/* Common queries */
const insertActivistQuery = `
INSERT INTO activists (id, email, name, chapter_id) VALUES (NULL, "email1", ?, ` + model.SFBayChapterIdStr + `);
`

const getActivistsQuery = `SELECT id FROM activists;`

/* Common utils */
type activist struct {
	id int
}

func verifyActivistIsInserted(t *testing.T, db *sqlx.DB) {
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
	if len(activists) > 1 {
		t.Error("found more than 1 activist")
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
