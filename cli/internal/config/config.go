package config

import (
	"fmt"
	"os"
	"strconv"
)

// IsProd returns true when the PROD environment variable is set to a truthy value (same convention as the ADB server).
func IsProd() bool {
	val, _ := strconv.ParseBool(os.Getenv("PROD"))
	return val
}

// DBDataSource builds the MySQL DSN from environment variables.
func DBDataSource() string {
	user := getEnv("DB_USER", "")
	password := getEnv("DB_PASSWORD", "")
	name := getEnv("DB_NAME", "")
	protocol := getEnv("DB_PROTOCOL", "")

	dsn := fmt.Sprintf("%s:%s@%s/%s?parseTime=true&charset=utf8mb4", user, password, protocol, name)
	if IsProd() {
		dsn += "&tls=true"
	}
	return dsn
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
