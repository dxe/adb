package form_processor

import (
	"testing"
)

func TestProcessFormInterestForNoMatchingActivist(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Exec(insertIntoFormInterestQuery)
	if err != nil {
		t.Fatalf("insertIntoFormInterestQuery failed: %s", err)
	}

	/* Call functionality under test */
	processInterestForms(db)

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
	processInterestForms(db)

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
	processInterestForms(db)

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
	processInterestForms(db)

	// For now, manually check error message "ERROR: 2 non-hidden activists associated"
}
