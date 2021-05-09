package form_processor

import (
	"github.com/dxe/adb/config"
)

type mainEnv struct {
	logLevel                     int // Use command-line argument value if it exists. Use ENV value otherwise.
	logFilePath                  string
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

func getMainEnv() (mainEnv, bool) {
	return mainEnv{
			logLevel:                     config.LogLevel,
			logFilePath:                  config.FormProcessorLogFilePath,
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
