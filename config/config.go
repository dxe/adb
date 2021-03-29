package config

import (
	"flag"
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

	IsProd            = getIsProd()
	RunBackgroundJobs = mustGetenvAsBool("RUN_BACKGROUND_JOBS")

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
	GooglePlacesAPIKey = mustGetenv("GOOGLE_PLACES_API_KEY", "", false)

	// For form processor
	FormProcessorLogLevel = mustGetenv(
		"FORM_PROCESSOR_LOG_LEVEL",
		"1",
		false,
	)
	FormProcessorProcessFormsCronExpression = mustGetenv(
		"FORM_PROCESSOR_PROCESS_FORMS_CRON_EXPRESSION",
		"@every 10s",
		false,
	)
	FormProcessorSendLogByEmailCronExpression = mustGetenv(
		"FORM_PROCESSOR_SEND_LOG_BY_EMAIL_CRON_EXPRESSION",
		"@daily",
		false,
	)
	FormProcessorLockFilePath = mustGetenv(
		"FORM_PROCESSOR_LOCK_FILE_PATH",
		"./adb-forms/output/PROCESSOR_RUNNING",
		false,
	)
	FormProcessorLogFilePath = mustGetenv(
		"FORM_PROCESSOR_LOG_FILE_PATH",
		"./adb-forms/output/LOG_FILE",
		false,
	)
	FormProcessorLogEmailToAddress = mustGetenv(
		"FORM_PROCESSOR_LOG_EMAIL_TO_ADDRESS",
		"",
		false,
	)
)

func getIsProd() bool {
	var isProd = flag.Bool("prod", false, "whether to run as production")
	if !IsFlagPassed("prod") {
		*isProd = mustGetenvAsBool("PROD")
	}
	return *isProd
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

func IsFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
