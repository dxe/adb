// Package db provides a shared MySQL connection helper for the jobs lambdas.
package db

import (
	"fmt"

	"github.com/dxe/adb/pkg/shared"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	dbName = "adb2"
	dbPort = 3306
)

// Open connects to the ADB MySQL database using the same DSN format as the
// server (see config.DBDataSource / shared.BuildDBDataSource). Credentials are
// resolved by the caller (from SSM) rather than read from the environment, so
// they never appear in the deploy command line or the CloudFormation template.
func Open(host, user, password string) (*sqlx.DB, error) {
	if host == "" || user == "" || password == "" {
		return nil, fmt.Errorf("host, user, and password must be set")
	}

	protocol := fmt.Sprintf("tcp(%s:%d)", host, dbPort)
	dsn := shared.BuildDBDataSource(user, password, protocol, dbName, true /* isProd */)
	conn, err := sqlx.Connect("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	return conn, nil
}
