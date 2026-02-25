package shared

import "fmt"

// BuildDBDataSource constructs a MySQL DSN from the given connection components.
func BuildDBDataSource(user, password, protocol, dbName string, isProd bool) string {
	dsn := fmt.Sprintf("%s:%s@%s/%s?parseTime=true&charset=utf8mb4", user, password, protocol, dbName)
	if isProd {
		dsn += "&tls=true"
	}
	return dsn
}
