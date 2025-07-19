package form_processor

import (
	"github.com/dxe/adb/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Should be run in a goroutine.
func StartFormProcessor(db *sqlx.DB) {
	/* Set default log level */
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Info().Msg("starting go processor; logging to default location")

	/* Set defined log level */
	log.Info().Msgf("setting log level to %d", config.LogLevel)
	zerolog.SetGlobalLevel(zerolog.Level(config.LogLevel))

	/* Start tasks on a scheduule */
	cron := cron.New()
	cron.AddFunc(config.FormProcessorProcessFormsCronExpression, func() {
		ProcessApplicationForms(db)
		ProcessInterestForms(db)
	})
	cron.Run()
}
