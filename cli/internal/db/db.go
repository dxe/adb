package db

import (
	"github.com/dxe/adb/cli/internal/config"
	"github.com/jmoiron/sqlx"

	// MySQL driver — imported for its side-effect of registering the driver.
	_ "github.com/go-sql-driver/mysql"
)

// Connect opens a connection to the ADB database.
func Connect() (*sqlx.DB, error) {
	return sqlx.Connect("mysql", config.DBDataSource())
}
