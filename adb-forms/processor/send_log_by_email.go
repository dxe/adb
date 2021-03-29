package processor

import (
	"bytes"
	"fmt"
	"github.com/dxe/adb/mailer"
	"github.com/rs/zerolog/log"
	"os"
)

func sendLogByEmail() {
	log.Info().Msg("Sending the log by email")
	sendLogByEmailEnv, ok := getSendLogByEmailEnv()
	if !ok {
		log.Error().Msg("failed to get ENV variables; will not send email")
		return
	}

	/* Open the log file */
	logFile, openLogFileErr := os.OpenFile(
		sendLogByEmailEnv.logFilePath,
		os.O_RDWR|os.O_CREATE|os.O_APPEND,
		0666,
	)
	if openLogFileErr != nil {
		log.Error().Msgf("error opening log file; exiting; %s", openLogFileErr)
		return
	}
	buffer := new(bytes.Buffer)
	buffer.ReadFrom(logFile)
	logFileContents := buffer.String()

	/* Send the email */
	sendErr := mailer.Send(mailer.Message{
		FromName:    "DxE Tech Server",
		FromAddress: "tech-noreply@directactioneverywhere.com",
		ToEmail:     sendLogByEmailEnv.toAddress,
		Subject:     "Form processor log (from ADB)",
		BodyHTML:    fmt.Sprintf("<div style='white-space: pre-line'>%s</div>", logFileContents),
	})

	if sendErr != nil {
		log.Error().Msgf("failed to send email %s", sendErr)
	} else {
		log.Info().Msg("Successfully sent the log by email")
		truncateErr := logFile.Truncate(0)
		if truncateErr != nil {
			log.Error().Msg("failed to truncate log file")
		}
		log.Info().Msg("Successfully truncated the log file")
	}
}
