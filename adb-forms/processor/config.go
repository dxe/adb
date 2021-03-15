package processor

import (
	"flag"
	"github.com/joho/godotenv"
	"github.com/rs/zerolog/log"
	"strconv"
)

type mainEnv struct {
	logLevel                     int // Use command-line argument value if it exists. Use ENV value otherwise.
	logFilePath                  string
	sendLogByEmailCronExpression string
	processFormsCronExpression   string
}

type processEnv struct {
	logFilePath           string
	lockFilePath          string
	mysqlConnectionString string
}

type sendLogByEmailEnv struct {
	logFilePath   string
	emailTo       string
	emailHost     string
	emailPort     int
	emailUsername string
	emailPassword string
}

func getMainEnv(logLevel int) (mainEnv, bool) {
	var emptyMainEnv mainEnv
	env, ok := getEnv()
	if !ok {
		return emptyMainEnv, false
	}
	if !isFlagPassed("logLevel") {
		envLogLevel, ok := parseUint64(env["LOG_LEVEL"])
		if !ok {
			return emptyMainEnv, false
		}
		logLevel = int(envLogLevel)
	}
	return mainEnv{
			logLevel:                     logLevel,
			logFilePath:                  env["LOG_FILE_PATH"],
			sendLogByEmailCronExpression: env["SEND_LOG_BY_EMAIL_CRON_EXPRESSION"],
			processFormsCronExpression:   env["PROCESS_FORMS_CRON_EXPRESSION"],
		},
		true

}

func getProcessEnv() (processEnv, bool) {
	var emptyProcessEnv processEnv
	env, ok := getEnv()
	if !ok {
		return emptyProcessEnv, false
	}
	return processEnv{
			logFilePath:           env["LOG_FILE_PATH"],
			lockFilePath:          env["LOCK_FILE_PATH"],
			mysqlConnectionString: env["MYSQL_CONNECTION_STRING"],
		},
		true
}

func getSendLogByEmailEnv() (sendLogByEmailEnv, bool) {
	var emptySendLogByEmailEnv sendLogByEmailEnv
	env, ok := getEnv()
	if !ok {
		return emptySendLogByEmailEnv, false
	}
	emailPort, ok := parseUint64(env["EMAIL_PORT"])
	if !ok {
		return emptySendLogByEmailEnv, false
	}
	return sendLogByEmailEnv{
			logFilePath:   env["LOG_FILE_PATH"],
			emailTo:       env["EMAIL_TO"],
			emailHost:     env["EMAIL_HOST"],
			emailPort:     int(emailPort),
			emailUsername: env["EMAIL_USERNAME"],
			emailPassword: env["EMAIL_PASSWORD"],
		},
		true
}

func getEnv() (map[string]string, bool) {
	/* Refresh .env */
	err := godotenv.Overload()
	if err != nil {
		log.Error().Msgf("did not find a .env file; exiting; err: %s", err)
		return nil, false
	}

	/* Get the required ENV variables */
	env, err := godotenv.Read()
	if err != nil {
		log.Error().Msgf("failed to load env variables; %s", err)
		return nil, false
	}
	return env, true
}

func parseUint64(value string) (uint64, bool) {
	parsed, parseErr := strconv.ParseUint(value, 10, 64)
	if parseErr != nil {
		log.Error().Msgf("failed to parse '%s' env variable; %s", value, parseErr)
		return 0, false
	}
	return parsed, true
}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
