package shared

import (
	"context"
	"log"
	"strings"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// WipeDatabase drops all tables and reapplies all migrations.
// The DSN must include multiStatements=true. isProd must be false.
func WipeDatabase(dsn string, isProd bool) {
	db := sqlx.MustConnect("mysql", dsn)
	defer func() { _ = db.Close() }()
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

	var tables []string
	err := db.Select(&tables, `
		SELECT table_name FROM information_schema.tables
		WHERE table_schema = DATABASE() AND table_type = 'BASE TABLE'`)
	if err != nil {
		log.Panicf("error listing tables: %v", err)
	}
	if len(tables) == 0 {
		return
	}

	// FOREIGN_KEY_CHECKS is per-connection, so pin one connection for the
	// whole operation and restore the setting before returning it to the pool.
	ctx := context.Background()
	conn, err := db.Connx(ctx)
	if err != nil {
		log.Panicf("error getting connection: %v", err)
	}
	defer func() { _ = conn.Close() }()

	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		log.Panicf("error disabling foreign key checks: %v", err)
	}
	defer func() { _, _ = conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1") }()

	quoted := make([]string, len(tables))
	for i, t := range tables {
		quoted[i] = "`" + strings.ReplaceAll(t, "`", "``") + "`"
	}
	if _, err := conn.ExecContext(ctx, "DROP TABLE IF EXISTS "+strings.Join(quoted, ", ")); err != nil {
		log.Panicf("error dropping tables: %v", err)
	}
}
