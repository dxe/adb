package config

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
)

var (
	DBUser, DBPassword, DBName string

	Port string

	Route0, Route1, Route2 string

	CookieSecret string

	// Path to Google API oauth client_secrets.json file, with
	// access to the following scope:
	// https://www.googleapis.com/auth/admin.directory.group
	// And the "Admin" API enabled. More info:
	//   https://developers.google.com/api-client-library/python/auth/service-accounts
	SyncMailingListsConfigFile string

	// The email for the user that that the oauth account should
	// take action as.
	SyncMailingListsOauthSubject string
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

		SyncMailingListsConfigFile = mustGetenv("SYNC_MAILING_LISTS_CONFIG_FILE")
		SyncMailingListsOauthSubject = mustGetenv("SYNC_MAILING_LISTS_OAUTH_SUBJECT")
	} else {
		DBUser = "adb_user"
		DBPassword = "adbpassword"
		DBName = "adb_db"

		Port = "8080"

		Route0 = "/route0"
		Route1 = "/route1"
		Route2 = "/route2"

		CookieSecret = "some-fake-secret"

		// Empty during development.
		SyncMailingListsConfigFile = ""
		SyncMailingListsOauthSubject = ""
	}
}

func DBDataSource() string {
	return DBUser + ":" + DBPassword + "@/" + DBName + "?parseTime=true"
}

func DBTestDataSource() string {
	return DBUser + ":" + DBPassword + "@/adb_test_db?parseTime=true"
}

var staticResourcesHash string = strconv.Itoa(rand.Int())

// Append static resources that we want to "bust" with every restart
// with this hash. This is a hacky solution to because it's too eager
// -- the best solution would to be to append a content hash to every
// static file -- but that's too complicated and this does the trick
// for now.
func StaticResourcesHash() string {
	return staticResourcesHash
}
