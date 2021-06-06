package form_processor

import (
	"github.com/dxe/adb/config"
)

type mainEnv struct {
	logLevel                   int // Use command-line argument value if it exists. Use ENV value otherwise.
	processFormsCronExpression string
}

func getMainEnv() (mainEnv, bool) {
	return mainEnv{
			logLevel:                   config.LogLevel,
			processFormsCronExpression: config.FormProcessorProcessFormsCronExpression,
		},
		true

}
