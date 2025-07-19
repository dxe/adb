package model

import (
	"fmt"
	"log"

	"github.com/dxe/adb/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jmoiron/sqlx"
)

func NewDB(dataSourceName string) *sqlx.DB {
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

type migrationLogger struct {
	verboseLogging bool
}

func (l *migrationLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *migrationLogger) Verbose() bool {
	return l.verboseLogging
}

func ApplyAllMigrations(db *sqlx.DB, sourceURL string, verboseLogging bool) error {
	driver, err := mysql.WithInstance(db.DB, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("error getting driver: %v", err)
	}

	m, err := migrate.NewWithDatabaseInstance(sourceURL, "mysql", driver)
	if err != nil {
		return fmt.Errorf("error initializing migrations: %v", err)
	}

	m.Log = &migrationLogger{verboseLogging}

	err = m.Up()
	if err != nil {
		if err == migrate.ErrNoChange {
			if verboseLogging {
				log.Printf("Database schema is already up-to-date.")
			}
			return nil
		} else {
			return fmt.Errorf("error applying migrations: %v", err)
		}
	}

	if verboseLogging {
		log.Printf("Database schema changes applied.")
	}
	return nil
}

func mustDropAllTables(db *sqlx.DB) error {
	if config.IsProd {
		panic("Cannot drop tables in prod")
	}

	_, err := db.Exec(`
		DROP TABLE IF EXISTS schema_migrations;

		DROP TABLE IF EXISTS activists;
		DROP TABLE IF EXISTS activists_history;
		DROP TABLE IF EXISTS events;
		DROP TABLE IF EXISTS event_attendance;
		DROP TABLE IF EXISTS users_roles;
		DROP TABLE IF EXISTS adb_users;
		DROP TABLE IF EXISTS merged_activist_attendance;
		DROP TABLE IF EXISTS working_groups;
		DROP TABLE IF EXISTS working_group_members;
		DROP TABLE IF EXISTS circles;
		DROP TABLE IF EXISTS circle_members;
		DROP TABLE IF EXISTS fb_pages;
		DROP TABLE IF EXISTS fb_events;
		DROP TABLE IF EXISTS discord_users;
		DROP TABLE IF EXISTS discord_messages;
		DROP TABLE IF EXISTS form_application;
		DROP TABLE IF EXISTS form_interest;
		DROP TABLE IF EXISTS form_international;
		DROP TABLE IF EXISTS form_discord;
		DROP TABLE IF EXISTS form_international_actions;
		DROP TABLE IF EXISTS interactions;
	`)

	return err
}

func newTestDB() *sqlx.DB {
	WipeDatabase(config.DBTestDataSource()+"&multiStatements=true", config.DBMigrationsLocation())
	return NewDB(config.DBTestDataSource())
}

// WipeDatabase wipes the given database, dropping all tables and recreating them.
// Requires that the database connection was established with `multiStatements=true`.
func WipeDatabase(dataSource string, sourceURL string) {
	db := NewDB(dataSource)
	defer db.Close()

	WipeDatabaseWithDb(db, sourceURL)
}

// WipeDatabaseWithDb wipes the given database, dropping all tables and recreating them.
// Requires that the database connection was established with `multiStatements=true`.
func WipeDatabaseWithDb(db *sqlx.DB, sourceURL string) {
	err := mustDropAllTables(db)
	if err != nil {
		log.Panicf("error dropping tables: %v", err)
	}

	// `db` must already have parameter multiStatements=true, according to https://github.com/golang-migrate/migrate/tree/master/database/mysql
	err = ApplyAllMigrations(db, sourceURL, false)
	if err != nil {
		log.Panicf("error applying database schema migrations: %v", err)
	}
}
