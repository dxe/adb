package shared

import (
	"embed"
	"fmt"
	"log"

	_ "github.com/go-sql-driver/mysql"
	migrate "github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jmoiron/sqlx"
)

//go:embed db-migrations/*
var migrationsFS embed.FS

type migrationLogger struct {
	verboseLogging bool
}

func (l *migrationLogger) Printf(format string, v ...interface{}) {
	log.Printf(format, v...)
}

func (l *migrationLogger) Verbose() bool {
	return l.verboseLogging
}

// ApplyMigrations applies all pending schema migrations to the database using
// the embedded migration files.
func ApplyMigrations(db *sqlx.DB, verboseLogging bool) error {
	d, err := iofs.New(migrationsFS, "db-migrations")
	if err != nil {
		return fmt.Errorf("error loading embedded migrations: %w", err)
	}

	driver, err := migratemysql.WithInstance(db.DB, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("error getting migration driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", d, "mysql", driver)
	if err != nil {
		return fmt.Errorf("error initializing migrations: %w", err)
	}
	m.Log = &migrationLogger{verboseLogging}

	upErr := m.Up()
	if upErr == migrate.ErrNoChange {
		if verboseLogging {
			log.Println("Database schema is already up-to-date.")
		}
		return nil
	}
	if upErr != nil {
		return fmt.Errorf("error applying migrations: %w", upErr)
	}
	if verboseLogging {
		log.Println("Database schema changes applied.")
	}
	return nil
}
