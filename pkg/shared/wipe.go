package shared

import (
	"log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// WipeDatabase drops all tables and reapplies all migrations.
// The DSN must include multiStatements=true. isProd must be false.
func WipeDatabase(dsn string, isProd bool) {
	db := sqlx.MustConnect("mysql", dsn)
	defer db.Close()
	WipeDatabaseWithDb(db, isProd)
}

// WipeDatabaseWithDb drops all tables and reapplies all migrations.
// The DB connection must have been established with multiStatements=true. isProd must be false.
func WipeDatabaseWithDb(db *sqlx.DB, isProd bool) {
	mustDropAllTables(db, isProd)
	if err := ApplyMigrations(db, false); err != nil {
		log.Panicf("error applying migrations: %v", err)
	}
}

func mustDropAllTables(db *sqlx.DB, isProd bool) {
	if isProd {
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
	if err != nil {
		log.Panicf("error dropping tables: %v", err)
	}
}
