package config

import (
	"io/ioutil"
	"math/rand"
	"os"
	"strconv"
	"time"
)

var (
	DBUser     = mustGetenv("DB_USER", "adb_user", true)
	DBPassword = mustGetenv("DB_PASSWORD", "adbpassword", true)
	DBName     = mustGetenv("DB_NAME", "adb_db", true)
	DBProtocol = mustGetenv("DB_PROTOCOL", "", true)

	Port    = mustGetenv("PORT", "8080", true)
	UrlPath = mustGetenv("ADB_URL_PATH", "http://localhost:"+Port, true)

	// Cluster role is used to assign a role to each running instance of the app.
	// Possible values: standalone (run everything), webserver (process incoming requests), background (run background tasks)
	ClusterRole = mustGetenv("CLUSTER_ROLE", "standalone", true)

	// Deprecated
	Route0 = mustGetenv("ROUTE_0", "/route0", true)

	CookieSecret = mustGetenv("COOKIE_SECRET", "some-fake-secret", true)
	CsrfAuthKey  = mustGetenv("CSRF_AUTH_KEY", "", true)

	// Path to Google API oauth client_secrets.json file, with
	// access to the following scope:
	// https://www.googleapis.com/auth/admin.directory.group
	// And the "Admin" API enabled. More info:
	//   https://developers.google.com/api-client-library/python/auth/service-accounts
	SyncMailingListsConfigFile = mustGetenv("SYNC_MAILING_LISTS_CONFIG_FILE", "", false)

	// The email for the user that that the oauth account should
	// take action as.
	SyncMailingListsOauthSubject = mustGetenv("SYNC_MAILING_LISTS_OAUTH_SUBJECT", "", false)

	// For sending surveys
	AWSAccessKey       = mustGetenv("AWS_ACCESS_KEY_ID", "", false)
	AWSSecretKey       = mustGetenv("AWS_SECRET_KEY", "", false)
	AWSSESEndpoint     = mustGetenv("AWS_SES_ENDPOINT", "", false)
	SurveyMissingEmail = mustGetenv("SURVEY_MISSING_EMAIL", "", false)
	SurveyFromEmail    = mustGetenv("SURVEY_FROM_EMAIL", "", false)

	// for IP geolocation
	IPGeolocationKey = mustGetenv("IPGEOLOCATION_KEY", "", false)

	// For members.dxesf.org
	MembersClientID     = mustGetenv("MEMBERS_CLIENT_ID", "", false)
	MembersClientSecret = mustGetenv("MEMBERS_CLIENT_SECRET", "", false)

	// For Discord bot
	DiscordSecret         = mustGetenv("DISCORD_SECRET", "some-fake-secret", false)
	DiscordBotBaseUrl     = mustGetenv("DISCORD_BOT_BASE_URL", "http://localhost:6070", false)
	DiscordFromEmail      = mustGetenv("DISCORD_FROM_EMAIL", "", false)
	DiscordModeratorEmail = mustGetenv("DISCORD_MODERATOR_EMAIL", "", false)
	SupportEmail          = mustGetenv("SUPPORT_EMAIL", "tech@dxe.io", false)
)

func mustGetenv(key, fallback string, mandatory bool) string {
	val := os.Getenv(key)
	if val != "" {
		return val
	}
	if !mandatory || !IsProd {
		return fallback
	}

	panic("Environment variable " + key + " cannot be empty")
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

func DBDataSource() string {
	connectionString := DBUser + ":" + DBPassword + "@" + DBProtocol + "/" + DBName + "?parseTime=true&charset=utf8mb4"
	if IsProd {
		return connectionString + "&tls=true"
	}
	return connectionString
}

func DBTestDataSource() string {
	return DBUser + ":" + DBPassword + "@/adb_test_db?parseTime=true"
}

var staticResourcesHash = strconv.FormatInt(rand.NewSource(time.Now().UnixNano()).Int63(), 10)

// Append static resources that we want to "bust" with every restart
// with this hash. This is a hacky solution to because it's too eager
// -- the best solution would to be to append a content hash to every
// static file -- but that's too complicated and this does the trick
// for now.
func StaticResourcesHash() string {
	return staticResourcesHash
}
