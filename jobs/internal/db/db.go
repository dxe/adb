// Package db provides a shared MySQL connection helper for the jobs lambdas.
package db

import (
	"context"
	"fmt"

	"github.com/dxe/adb/pkg/shared"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

const (
	dbName = "adb2"
	dbPort = 3306
)

func Open(ctx context.Context, host, user, password string) (*sqlx.DB, error) {
	if host == "" || user == "" || password == "" {
		return nil, fmt.Errorf("host, user, and password must be set")
	}

	protocol := fmt.Sprintf("tcp(%s:%d)", host, dbPort)
	dsn := shared.BuildDBDataSource(user, password, protocol, dbName, true /* isProd */)
	conn, err := sqlx.ConnectContext(ctx, "mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("connecting to database: %w", err)
	}
	return conn, nil
}
