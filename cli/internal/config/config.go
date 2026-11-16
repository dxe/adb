package config

import (
	"os"
	"strconv"

	"github.com/dxe/adb/pkg/shared"
)

// IsProd returns true when the PROD environment variable is set to a truthy value (same convention as the ADB server).
func IsProd() bool {
	val, _ := strconv.ParseBool(os.Getenv("PROD"))
	return val
}

// DBDataSource builds the MySQL DSN from environment variables. Unset
// variables fall back to local development values unless PROD is true.
func DBDataSource() string {
	return DBDataSourceForDB(devEnv("DB_NAME", shared.DevDBName))
}

// DBDataSourceForDB builds a MySQL DSN for a specific database name.
func DBDataSourceForDB(dbName string) string {
	return shared.BuildDBDataSource(
		devEnv("DB_USER", shared.DevDBUser),
		devEnv("DB_PASSWORD", shared.DevDBPassword),
		devEnv("DB_PROTOCOL", shared.DevDBProtocol),
		dbName,
		IsProd(),
	)
}

// devEnv returns the environment variable, or the local development default
// when it is unset and PROD is not true. In prod there is no fallback, so a
// missing value fails to connect rather than silently using dev settings.
func devEnv(key, devDefault string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	if IsProd() {
		return ""
	}
	return devDefault
}
