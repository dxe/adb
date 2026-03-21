package model

import (
	"context"
	"database/sql"
	"fmt"
	"sync"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/pkg/shared"
)

var initializeTestDBSchemaOnce sync.Once

func NewDB(dataSourceName string) *sqlx.DB {
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

func initializeTestDBSchema() {
	initializeTestDBSchemaOnce.Do(func() {
		shared.WipeDatabase(config.DBTestDataSource()+"&multiStatements=true", false)
	})
}

func listResettableTables(ctx context.Context, conn *sql.Conn) ([]string, error) {
	rows, err := conn.QueryContext(ctx, `
SELECT table_name
FROM information_schema.tables
WHERE table_schema = DATABASE()
  AND table_type = 'BASE TABLE'
  AND table_name <> 'schema_migrations'
ORDER BY table_name
`)
	if err != nil {
		return nil, fmt.Errorf("query resettable tables: %w", err)
	}
	defer rows.Close()

	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, fmt.Errorf("scan resettable table: %w", err)
		}
		tables = append(tables, table)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate resettable tables: %w", err)
	}

	return tables, nil
}

func resetTestDB(db *sqlx.DB) error {
	ctx := context.Background()

	conn, err := db.Conn(ctx)
	if err != nil {
		return fmt.Errorf("get test database connection: %w", err)
	}
	defer conn.Close()

	if _, err := conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0"); err != nil {
		return fmt.Errorf("disable foreign key checks: %w", err)
	}
	defer func() {
		_, _ = conn.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1")
	}()

	tables, err := listResettableTables(ctx, conn)
	if err != nil {
		return err
	}

	for _, table := range tables {
		if _, err := conn.ExecContext(ctx, fmt.Sprintf("TRUNCATE TABLE `%s`", table)); err != nil {
			return fmt.Errorf("truncate %s: %w", table, err)
		}
	}

	return nil
}

func newTestDB() *sqlx.DB {
	initializeTestDBSchema()

	db := NewDB(config.DBTestDataSource())
	if err := resetTestDB(db); err != nil {
		db.Close()
		panic(fmt.Errorf("reset test database: %w", err))
	}

	return db
}
