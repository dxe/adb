package shared

import "fmt"

// DBConnParams are the MySQL connection parameters shared by all environments.
// Please keep relevant options in sync with the db() function in .devcontainer/.bash_adb_functions.
const DBConnParams = "parseTime=true&charset=utf8mb4&clientFoundRows=true"

// BuildDBDataSource constructs a MySQL DSN from the given connection components.
func BuildDBDataSource(user, password, protocol, dbName string, isProd bool) string {
	dsn := fmt.Sprintf("%s:%s@%s/%s?"+DBConnParams, user, password, protocol, dbName)
	if isProd {
		dsn += "&tls=true"
	}
	return dsn
}
