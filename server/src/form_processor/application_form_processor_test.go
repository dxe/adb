package form_processor

import (
	"testing"
)

func TestProcessFormApplicationForNoMatchingActivist(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Query(insertIntoFormApplicationQuery)
	if err != nil {
		t.Fatalf("insertIntoFormApplicationQuery failed: %s", err)
	}

	/* Call functionality under test */
	processApplicationForms(db)

	/* Verify */
	verifyActivistIsInserted(t, db)
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
	processApplicationForms(db)

	/* Verify */
	verifyActivistIsInserted(t, db)
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
	processApplicationForms(db)

	/* Verify */
	verifyActivistIsInserted(t, db)
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
	processApplicationForms(db)

	// For now, manually check error message "ERROR: 2 non-hidden activists associated"
	verifyFormWasNotMarkedAsProcessed(t, db, applicationProcessingStatusQuery)
}
