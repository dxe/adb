package form_processor

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

// Do not use same DB as used by `models` package because go tests in different
// packages run in parallel.
const dbName = "adb_test_forms_db"

// createTempDb creates a database for testing. Please call `dropTempDb` when
// tests finish.
func createTempDb(name string) {
	db, err := sql.Open("mysql", config.DataSourceBase+"/")
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Create the database
	createDBQuery := "CREATE DATABASE IF NOT EXISTS " + name + " CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci"
	_, err = db.Exec(createDBQuery)
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}
}
func dropTempDb(name string) {
	db, err := sql.Open("mysql", config.DataSourceBase+"/")
	if err != nil {
		log.Fatalf("Error connecting to database: %v", err)
	}
	defer db.Close()

	// Create the database
	createDBQuery := "DROP DATABASE " + name
	_, err = db.Exec(createDBQuery)
	if err != nil {
		log.Fatalf("Error creating database: %v", err)
	}
}
func useTestDb() *sqlx.DB {
	dataSource := config.DataSourceBase + "/" + dbName + "?parseTime=true"
	model.WipeDatabase(dataSource+"&multiStatements=true", config.DBMigrationsLocation())
	return model.NewDB(dataSource)
}

func TestMain(m *testing.M) {
	// Use a dedicated database for this package's tests. Tests in this package
	// must run in serial, but may run in parallel with tests in other packages.
	createTempDb(dbName)
	defer dropTempDb(dbName)

	os.Exit(m.Run())
}
