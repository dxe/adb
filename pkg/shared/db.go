package shared

import "fmt"

// DBConnParams are the MySQL connection parameters shared by all environments.
// Please keep relevant options in sync with the db() function in scripts/shell/adb_functions.bash.
const DBConnParams = "parseTime=true&charset=utf8mb4&clientFoundRows=true"

// Local development database settings, used as fallbacks by the server and CLI
// when the corresponding DB_* environment variable is unset and PROD is not
// true. They match server/dev-db.env, server/scripts/init.sql and the MySQL
// port published by server/compose.yaml. The protocol is the host-side address;
// the devcontainer overrides it with DB_PROTOCOL.
const (
	DevDBUser     = "adb_user"
	DevDBPassword = "adbpassword"
	DevDBName     = "adb_db"
	DevDBProtocol = "tcp(127.0.0.1:3306)"
)

// BuildDBDataSource constructs a MySQL DSN from the given connection components.
func BuildDBDataSource(user, password, protocol, dbName string, isProd bool) string {
	dsn := fmt.Sprintf("%s:%s@%s/%s?"+DBConnParams, user, password, protocol, dbName)
	if isProd {
		dsn += "&tls=true"
	}
	return dsn
}
