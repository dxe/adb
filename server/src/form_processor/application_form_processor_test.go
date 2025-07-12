package form_processor

import (
	"testing"
)

const insertIntoFormApplicationQuery = `
INSERT INTO form_application (
  id,
  email,
  name,
  phone,
  address,
  city,
  zip,
  birthday,
  pronouns,
  application_type,
  agree_circle,
  agree_mpp,
  circle_interest,
  wg_interest,
  committee_interest,
  referral_friends,
  referral_apply,
  referral_outlet,
  contact_method,
  processed
) VALUES (
  NULL,
  "email1",
  "name1",
  "phone1",
  "address1",
  "city1",
  "zip1",
  "birthday1",
  "pronouns1",
  "application_type1",
  "agree_circle1",
  "agree_mpp1",
  "circle_interest1",
  "wg_interest1",
  "committee_interest1",
  "referral_friends1",
  "referral_apply1",
  "referral_outlet1",
  "contact_method1",
  false
);
`

func TestProcessFormApplicationForNoMatchingActivist(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	_, err := db.Query(insertIntoFormApplicationQuery)
	if err != nil {
		t.Fatalf("insertIntoFormApplicationQuery failed: %s", err)
	}

	/* Call functionality under test */
	ProcessApplicationForms(db)

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
	ProcessApplicationForms(db)

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
	ProcessApplicationForms(db)

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
	ProcessApplicationForms(db)

	// For now, manually check error message "ERROR: 2 non-hidden activists associated"
	verifyFormWasNotMarkedAsProcessed(t, db, applicationProcessingStatusQuery)
}
