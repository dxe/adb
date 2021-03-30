package processor

import (
	"github.com/dxe/adb/config"
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
	logFilePath  string
	lockFilePath string
}

type sendLogByEmailEnv struct {
	logFilePath string
	toAddress   string
}

func getMainEnv(logLevel int) (mainEnv, bool) {
	var emptyMainEnv mainEnv
	if !config.IsFlagPassed("logLevel") {
		envLogLevel, ok := parseUint64(config.FormProcessorLogLevel)
		if !ok {
			return emptyMainEnv, false
		}
		logLevel = int(envLogLevel)
	}
	return mainEnv{
			logLevel:                     logLevel,
			logFilePath:                  config.FormProcessorLogFilePath,
			sendLogByEmailCronExpression: config.FormProcessorSendLogByEmailCronExpression,
			processFormsCronExpression:   config.FormProcessorProcessFormsCronExpression,
		},
		true

}

func getProcessEnv() (processEnv, bool) {
	return processEnv{
			logFilePath:  config.FormProcessorLogFilePath,
			lockFilePath: config.FormProcessorLockFilePath,
		},
		true
}

func getSendLogByEmailEnv() (sendLogByEmailEnv, bool) {
	var emptySendLogByEmailEnv sendLogByEmailEnv
	if config.FormProcessorLogEmailToAddress == "" {
		return emptySendLogByEmailEnv, false
	}
	return sendLogByEmailEnv{
			logFilePath: config.FormProcessorLogFilePath,
			toAddress:   config.FormProcessorLogEmailToAddress,
		},
		true
}

func parseUint64(value string) (uint64, bool) {
	parsed, parseErr := strconv.ParseUint(value, 10, 64)
	if parseErr != nil {
		log.Error().Msgf("failed to parse '%s' env variable; %s", value, parseErr)
		return 0, false
	}
	return parsed, true
}
