package form_processor

import (
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

	/* Get config */
	mainEnv, ok := getMainEnv()
	if !ok {
		log.Error().Msg("failed to get ENV variables; will exit main program")
		return
	}
	/* Set defined log level */
	log.Info().Msgf("setting log level to %d", mainEnv.logLevel)
	zerolog.SetGlobalLevel(zerolog.Level(mainEnv.logLevel))

	/* Start tasks on a scheduule */
	cron := cron.New()
	cron.AddFunc(mainEnv.processFormsCronExpression, func() {
		processApplicationForms(db)
		processInterestForms(db)
	})
	cron.Run()
}
