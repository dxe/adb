package config

import (
	"io/ioutil"
	"os"
)

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

func isEC2() bool {
	// see http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/identify_ec2_instances.html
	data, err := ioutil.ReadFile("/sys/hypervisor/uuid")
	if err != nil {
		// The file must exist on EC2
		return false
	}
	return string(data[:3]) == "ec2"
}

// Always run as IsProd in EC2. This means you can't develop on EC2,
// but we'll cross that bridge when we get there.
var IsProd bool = isEC2()

func init() {
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
