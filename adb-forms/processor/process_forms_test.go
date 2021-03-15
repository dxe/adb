package processor

import (
	"database/sql"
	"github.com/lestrrat-go/test-mysqld"
	"testing"
)

/* Common utils */
type activist struct {
	id int
}

func createTables(t *testing.T) (*mysqltest.TestMysqld, *sql.DB) {
	/* Set up MySQL */
	mysqld, err := mysqltest.NewMysqld(nil)
	if err != nil {
		t.Fatalf("failed to start mysqld: %s", err)
	}
	db, err := sql.Open("mysql", mysqld.Datasource("test", "", "", 0))
	if err != nil {
		t.Fatalf("failed to open MySQL connection: %s", err)
	}

	/* Crate tables */
	_, err = db.Exec(createTableFormApplicationQuery)
	if err != nil {
		t.Fatalf("createTableFormApplicationQuery failed: %s", err)
	}
	_, err = db.Exec(createTableActivistsQuery)
	if err != nil {
		t.Fatalf("createTableActivistsQuery failed: %s", err)
	}
	_, err = db.Exec(createTableWorkingGroupMembersQuery)
	if err != nil {
		t.Fatalf("createTableWorkingGroupMembersQuery failed: %s", err)
	}
	_, err = db.Exec(createTableCircleMembersQuery)
	if err != nil {
		t.Fatalf("createTableCircleMembersQuery failed: %s", err)
	}
	_, err = db.Exec(createTableFormInterestQuery)
	if err != nil {
		t.Fatalf("createTableFormInterestQuery failed: %s", err)
	}
	return mysqld, db
}

func verifyFormWasMarkedAsProcessed(t *testing.T, db *sql.DB, query string) {
	rawActivists, err := db.Query(getActivistsQuery)
	defer rawActivists.Close()
	if err != nil {
		t.Fatalf("getActivistsQuery failed: %s", err)
	}
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
	mysqld, db := createTables(t)
	defer mysqld.Stop()
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
