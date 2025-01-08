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

/* Form application tests */
func TestProcessFormApplicationForNoMatchingActivist(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Query(insertIntoFormApplicationQuery)
	if err != nil {
		t.Fatalf("insertIntoFormApplicationQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	/* Verify */
	verifyFormWasMarkedAsProcessed(t, db, applicationProcessingStatusQuery)
}

func TestProcessFormApplicationForActivistMatchingOnName(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()

	_, err := db.Exec(insertActivistQuery, "name1")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertIntoFormApplicationQuery)
	if err != nil {
		t.Fatalf("insertIntoFormApplicationQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	/* Verify */
	verifyFormWasMarkedAsProcessed(t, db, applicationProcessingStatusQuery)
}

func TestProcessFormApplicationForActivistMatchingOnEmail(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()

	_, err := db.Exec(insertActivistQuery, "non-matching_name")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertIntoFormApplicationQuery)
	if err != nil {
		t.Fatalf("insertIntoFormApplicationQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	/* Verify */
	verifyFormWasMarkedAsProcessed(t, db, applicationProcessingStatusQuery)
}

func TestProcessFormApplicationForMultipleMatchingActivistsOnEmail(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Exec(insertActivistQuery, "non-matching_name1")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertActivistQuery, "non-matching_name2")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertIntoFormApplicationQuery)
	if err != nil {
		t.Fatalf("insertIntoFormApplicationQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	// For now, manually check error message "ERROR: 2 non-hidden activists associated"
}

/* Form interest tests */
func TestProcessFormInterestForNoMatchingActivist(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Exec(insertIntoFormInterestQuery)
	if err != nil {
		t.Fatalf("insertIntoFormInterestQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	/* Verify */
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForActivistMatchingOnName(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Exec(insertActivistQuery, "name1")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertIntoFormInterestQuery)
	if err != nil {
		t.Fatalf("insertIntoFormInterestQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	/* Verify */
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForActivistMatchingOnEmail(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Exec(insertActivistQuery, "non-matching_name")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertIntoFormInterestQuery)
	if err != nil {
		t.Fatalf("insertIntoFormInterestQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	/* Verify */
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForMultipleMatchingActivistsOnEmail(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Exec(insertActivistQuery, "non-matching_name1")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertActivistQuery, "non-matching_name2")
	if err != nil {
		t.Fatalf("insertActivistQuery failed: %s", err)
	}
	_, err = db.Exec(insertIntoFormInterestQuery)
	if err != nil {
		t.Fatalf("insertIntoFormInterestQuery failed: %s", err)
	}

	/* Call functionality under test */
	processForms(db)

	// For now, manually check error message "ERROR: 2 non-hidden activists associated"
}
