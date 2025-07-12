package form_processor

import (
	"testing"
)

const insertIntoFormInterestQuery = `
INSERT INTO form_interest (
  id,
  form,
  email,
  name,
  phone,
  zip,
  referral_friends,
  referral_apply,
  referral_outlet,
  comments,
  interests
) VALUES (
  NULL,
  "form1",
  "email1",
  "name1",
  "phone1",
  "zip1",
  "referral_friends1",
  "referral_apply1",
  "referral_outlet1",
  "comments1",
  "interests1"
);
`

func TestProcessFormInterestForNoMatchingActivist(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Exec(insertIntoFormInterestQuery)
	if err != nil {
		t.Fatalf("insertIntoFormInterestQuery failed: %s", err)
	}

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistIsInserted(t, db)
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
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistIsInserted(t, db)
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
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistIsInserted(t, db)
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
	ProcessInterestForms(db)

	// For now, manually check error message "ERROR: 2 non-hidden activists associated"
	verifyFormWasNotMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}
