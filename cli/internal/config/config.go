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

// DBDataSource builds the MySQL DSN from environment variables.
func DBDataSource() string {
	return shared.BuildDBDataSource(
		getEnv("DB_USER", ""),
		getEnv("DB_PASSWORD", ""),
		getEnv("DB_PROTOCOL", ""),
		getEnv("DB_NAME", ""),
		IsProd(),
	)
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
