package config

import (
	"flag"
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var (
	DBUser         = mustGetenv("DB_USER", "adb_user", true)
	DBPassword     = mustGetenv("DB_PASSWORD", "adbpassword", true)
	DBName         = mustGetenv("DB_NAME", "adb_db", true)
	DBProtocol     = mustGetenv("DB_PROTOCOL", "", true)
	DataSourceBase = DBUser + ":" + DBPassword + "@" + DBProtocol

	Port        = mustGetenv("PORT", "8080", true)
	MembersPort = mustGetenv("MEMBERS_PORT", "8081", true)
	UrlPath     = mustGetenv("ADB_URL_PATH", "http://localhost:"+Port, true)

	IsProd            = mustGetenvAsBool("PROD")
	IsDocker          = IsProd || mustGetenvAsBool("ADB_IN_DOCKER")
	RunBackgroundJobs = mustGetenvAsBool("RUN_BACKGROUND_JOBS")
	LogLevel          = 1

	CookieSecret = mustGetenv("COOKIE_SECRET", "some-fake-secret", true)
	CsrfAuthKey  = mustGetenv("CSRF_AUTH_KEY", "", true)

	TemplatesDirectory = mustGetenv("TEMPLATES_DIRECTORY", "./templates", false)
	StaticDirectory    = mustGetenv("STATIC_DIRECTORY", "./static", false)
	DistDirectory      = mustGetenv("DIST_DIRECTORY", "./dist", false)
	NextJsProxyUrl     = mustGetenv("NEXT_JS_PROXY_URL", "", false)

	// Path to Google API oauth client_secrets.json file, with
	// access to the following scope:
	// https://www.googleapis.com/auth/admin.directory.group
	// And the "Admin" API enabled. More info:
	//   https://developers.google.com/api-client-library/python/auth/service-accounts
	SyncMailingListsConfigFile = mustGetenv("SYNC_MAILING_LISTS_CONFIG_FILE", "", false)

	// The email for the user that that the oauth account should
	// take action as.
	SyncMailingListsOauthSubject = mustGetenv("SYNC_MAILING_LISTS_OAUTH_SUBJECT", "", false)

	// For sending email
	SMTPHost     = mustGetenv("SMTP_HOST", "", false)
	SMTPPort     = mustGetenv("SMTP_PORT", "", false)
	SMTPUser     = mustGetenv("SMTP_USER", "", false)
	SMTPPassword = mustGetenv("SMTP_PASSWORD", "", false)

	// For surveys
	SurveyMissingEmail = mustGetenv("SURVEY_MISSING_EMAIL", "", false)
	SurveyFromEmail    = mustGetenv("SURVEY_FROM_EMAIL", "", false)

	// For IP geolocation
	IPGeolocationKey = mustGetenv("IPGEOLOCATION_KEY", "", false)

	// For members.dxesf.org
	MembersClientID     = mustGetenv("MEMBERS_CLIENT_ID", "", false)
	MembersClientSecret = mustGetenv("MEMBERS_CLIENT_SECRET", "", false)

	// For Discord bot
	DiscordSecret         = mustGetenv("DISCORD_SECRET", "some-fake-secret", false)
	DiscordBotBaseUrl     = mustGetenv("DISCORD_BOT_BASE_URL", "http://localhost:6070", false)
	DiscordFromEmail      = mustGetenv("DISCORD_FROM_EMAIL", "", false)
	DiscordModeratorEmail = mustGetenv("DISCORD_MODERATOR_EMAIL", "", false)

	// For mailing list signups
	SignupURI    = mustGetenv("SIGNUP_ENDPOINT", "", false)
	SignupAPIKey = mustGetenv("SIGNUP_KEY", "", false)

	// For location picker on International form
	GooglePlacesAPIKey        = mustGetenv("GOOGLE_PLACES_API_KEY", "", false)
	GooglePlacesBackendAPIKey = mustGetenv("GOOGLE_PLACES_API_KEY_BACKEND", "", false)

	// For form processor
	FormProcessorProcessFormsCronExpression = mustGetenv(
		"FORM_PROCESSOR_PROCESS_FORMS_CRON_EXPRESSION",
		"@every 10s",
		false,
	)
	FormProcessorLockFilePath = mustGetenv(
		"FORM_PROCESSOR_LOCK_FILE_PATH",
		"output/FORM_PROCESSOR_ROCESSOR_RUNNING",
		false,
	)
	FormProcessorLogFilePath = mustGetenv(
		"FORM_PROCESSOR_LOG_FILE_PATH",
		"output/FORM_PROCESSOR_LOG_FILE",
		false,
	)

	// Path to `server` directory which contains `src` and `scripts`
	DevServerDir string
)

func init() {
	_, b, _, _ := runtime.Caller(0)
	DevServerDir = filepath.Join(filepath.Dir(b), "../../")
}

func SetCommandLineFlags(isProdArgument bool, logLevel int) {
	if IsFlagPassed("prod") {
		IsProd = isProdArgument
	}

	if IsFlagPassed("logLevel") {
		LogLevel = logLevel
	} else {
		parsedLogLevel, ok := parseUint64(os.Getenv("LOG_LEVEL"))
		if ok {
			LogLevel = int(parsedLogLevel)
		} else {
			LogLevel = 1
		}
	}
}

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

// mustGetenvAsBool always defaults to false, so it should only be used to enable extra features.
func mustGetenvAsBool(key string) bool {
	val := os.Getenv(key)
	if val, err := strconv.ParseBool(val); err == nil {
		return val
	}
	return false
}

func DBDataSource() string {
	connectionString := DataSourceBase + "/" + DBName + "?parseTime=true&charset=utf8mb4"
	if IsProd {
		return connectionString + "&tls=true"
	}
	return connectionString
}

func DBTestDataSource() string {
	return DataSourceBase + "/adb_test_db?parseTime=true"
}

func DBMigrationsLocation() string {
	if !IsDocker {
		// Use `DevServerDir` to reliably locate the db-migrations directory, even when this package is invoked from
		// another go module such as `create_db_wrapper`, or from a test in a dev or CI environment.
		return "file://" + DevServerDir + "/scripts/db-migrations"
	}

	return "file://db-migrations"
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

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func parseUint64(value string) (uint64, bool) {
	parsed, parseErr := strconv.ParseUint(value, 10, 64)
	if parseErr != nil {
		return 0, false
	}
	return parsed, true
}
