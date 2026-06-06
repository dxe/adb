// Package db provides a shared MySQL connection helper for the jobs lambdas.
package db

import (
	"fmt"
	"os"

	"github.com/dxe/adb/pkg/shared"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	dbName = "adb2"
	dbPort = 3306
)

// Open connects to the ADB MySQL database using the same DSN format as the
// server (see config.DBDataSource / shared.BuildDBDataSource).
func Open() (*sqlx.DB, error) {
	host := os.Getenv("DB_HOST")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	if host == "" || user == "" || password == "" {
		return nil, fmt.Errorf("DB_HOST, DB_USER, and DB_PASSWORD must be set")
	}

	protocol := fmt.Sprintf("tcp(%s:%d)", host, dbPort)
	dsn := shared.BuildDBDataSource(user, password, protocol, dbName, true /* isProd */)
	conn, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	return conn, nil
}
