package config

import (
	"flag"
	"os"
)

var IsProd bool

var (
	DBUser, DBPassword, DBName string

	Port string

	Route0, Route1, Route2 string

	CookieSecret string
)

func mustGetenv(envvar string) string {
	val := os.Getenv(envvar)
	if val == "" {
		panic("Environmental variable " + envvar + " cannot be empty")
	}
	return val
}

func init() {
	prod := flag.Bool("prod", false, "Run in production mode")
	flag.Parse()
	IsProd = *prod

	if IsProd {
		DBUser = mustGetenv("DB_USER")
		DBPassword = mustGetenv("DB_PASSWORD")
		DBName = mustGetenv("DB_NAME")

		Port = mustGetenv("PORT")

		Route0 = mustGetenv("ROUTE_0")
		Route1 = mustGetenv("ROUTE_1")
		Route2 = mustGetenv("ROUTE_2")

		CookieSecret = mustGetenv("COOKIE_SECRET")
	} else {
		DBUser = "adb_user"
		DBPassword = "adbpassword"
		DBName = "adb_db"

		Port = "8080"

		Route0 = "/route0"
		Route1 = "/route1"
		Route2 = "/route2"

		CookieSecret = "some-fake-secret"
	}
}

func DBDataSource() string {
	return DBUser + ":" + DBPassword + "@/" + DBName + "?parseTime=true"
}

func DBTestDataSource() string {
	return DBUser + ":" + DBPassword + "@/adb_test_db?parseTime=true"
}
