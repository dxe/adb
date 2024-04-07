package form_processor

import (
	"context"
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
	cron.AddFunc(mainEnv.processFormsCronExpression, func() { processForms(db) })
	cron.Run()
}

func processForms(db *sqlx.DB) {
	log.Debug().Msg("starting processing run")

	/* Get form applications to process */
	applicationIds, isSuccess := getResponsesToProcess(db, applicationResponsesToProcessQuery)
	if !isSuccess {
		log.Error().Msg("failed to get applicationIds; exiting")
		return
	}
	if len(applicationIds) == 0 {
		log.Debug().Msg("no new form_application submissions to process")
	}
	for _, id := range applicationIds {
		log.Info().Msgf("Processing Application row %d", id)
		_, processErr := db.Exec(processApplicationOnNameQuery, id)
		if processErr != nil {
			log.Error().Msgf("failed to prrocess application on name; exiting; %s", processErr)
			return
		}
		log.Info().Msg("executed sql command to process Application based on name")

		processed, isSuccess := getProcessingStatus(db, applicationProcessingStatusQuery, id)
		if !isSuccess {
			log.Error().Msg("failed to get processing status; exiting")
			return
		}

		if !processed {
			// check how many records are tied to this email address
			email, isSuccess := getEmail(db, applicationSelectEmailQuery, id)
			if !isSuccess {
				log.Error().Msg("failed to get email; exiting")
				return
			}
			count, isSuccess := countActivistsForEmail(db, email)
			if !isSuccess {
				log.Error().Msg("failed to count activists for email; exiting")
				return
			}

			switch count {
			case 1:
				// if 1, update record
				// update record based on email
				_, err := db.Exec(processApplicationOnEmailQuery, id)
				if err != nil {
					log.Error().Msgf("failed to processApplicationOnEmailQuery; exiting; %s", err)
					return
				}
				log.Info().Msg("executed sql command to process Application based on email")
			case 0:
				// insert new record
				ctx := context.Background()
				tx, txErr := db.BeginTx(ctx, nil)
				if txErr != nil {
					log.Error().Msgf("failed to BeginTx for processApplicationByInsertQuery; exiting; %s", txErr)
					return
				} else {
					log.Info().Msg("successfully began transaction; proceeding")
				}
				_, processErr := tx.ExecContext(ctx, processApplicationByInsertQuery, id)
				if processErr != nil {
					log.Error().Msgf("failed to processApplicationByInsertQuery; exiting; %s", processErr)
					return
				}
				log.Info().Msg("executed sql command to insert new activist record from Application")
				res, updateErr := tx.ExecContext(ctx, processApplicationByInsertUpdateQuery, id)
				if updateErr != nil {
					log.Error().Msgf("failed to processApplicationByInsertUpdateQuery; exiting; %s", updateErr)
					return
				}
				count, getRowsAffectedErr := res.RowsAffected()
				if getRowsAffectedErr != nil {
					log.Error().Msgf(
						"failed to get the number of rows affected from processApplicationByInsertUpdateQuery;"+
							" exiting; %s",
						getRowsAffectedErr,
					)
					return
				}
				if count != 1 {
					log.Error().Msg("the activist was not updated (application date in activists table does not match the" +
						" date in application?); please correct")
				} else {
					log.Info().Msg("executed sql command to mark Application as processed")
				}
				commitErr := tx.Commit()
				if commitErr != nil {
					log.Error().Msgf("failed to commit transaction; exiting; %s", commitErr)
					return
				} else {
					log.Info().Msg("successfully committed transaction; proceeding")
				}
			default:
				// email count is > 1, so send email to tech
				log.Error().Msgf(
					"%d non-hidden activists associated with email address %s for"+
						" Application response %d Please correct.",
					count,
					email,
					id,
				)
			}
		}
	}

	/* Get form interests to process */
	interestIds, isSuccess := getResponsesToProcess(db, interestResponsesToProcessQuery)
	if !isSuccess {
		log.Error().Msg("failed to get interestIds; exiting")
		return
	}
	if len(interestIds) == 0 {
		log.Debug().Msg("no new form_interest submissions to process")
	}
	for _, id := range interestIds {
		log.Info().Msgf("processing Interest row %d", id)

		_, processErr := db.Exec(processInterestOnNameQuery, id)
		if processErr != nil {
			log.Error().Msgf("failed to process interest on name; exiting; %s", processErr)
			return
		}
		log.Info().Msg("executed sql command to process Interest based on name")

		processed, isSuccess := getProcessingStatus(db, interestProcessingStatusQuery, id)
		if !isSuccess {
			log.Error().Msg("failed to get processing status; exiting")
			return
		}

		if !processed {
			// check how many records are tied to this email address
			email, isSuccess := getEmail(db, interestSelectEmailQuery, id)
			if !isSuccess {
				log.Error().Msg("failed to get email; exiting")
				return
			}
			count, isSuccess := countActivistsForEmail(db, email)
			if !isSuccess {
				log.Error().Msg("failed to count activists for email; exiting")
				return
			}
			switch count {
			case 1:
				// if 1, update record
				// update record based on email
				_, err := db.Exec(processInterestOnEmailQuery, id)
				if err != nil {
					log.Error().Msgf("failed to processInterestOnEmailQuery; exiting; %s", err)
					return
				}
				log.Info().Msg("executed sql command to process Interest based on email")
			case 0:
				// insert new record
				ctx := context.Background()
				tx, txErr := db.BeginTx(ctx, nil)
				if txErr != nil {
					log.Error().Msgf("failed to BeginTx for processInterestByInsertQuery; exiting; %s", txErr)
					return
				} else {
					log.Info().Msg("successfully began transaction; proceeding")
				}
				_, processErr := db.ExecContext(ctx, processInterestByInsertQuery, id)
				if processErr != nil {
					log.Error().Msgf("failed to processInterestByInsertQuery; exiting; %s", processErr)
					return
				}
				log.Info().Msg("executed sql command to insert new activist record from Interest")
				res, updateErr := db.ExecContext(ctx, processInsertByInsertUpdateQuery, id)
				if updateErr != nil {
					log.Error().Msgf("failed to processInsertByInsertUpdateQuery; exiting; %s", updateErr)
					return
				}
				count, getRowsAffectedErr := res.RowsAffected()
				if getRowsAffectedErr != nil {
					log.Error().Msgf(
						"failed to get the number of rows affected from processApplicationByInsertUpdateQuery;"+
							" exiting; %s",
						getRowsAffectedErr,
					)
					return
				}
				if count != 1 {
					log.Error().Msg("the activist was not updated (application date in activists table" +
						" does not match the date in interest?); please correct")
				} else {
					log.Info().Msg("successfully executed sql command to mark Intereest as processed")
				}
				commitErr := tx.Commit()
				if commitErr != nil {
					log.Error().Msgf("failed to commit transaction; exiting; %s", commitErr)
					return
				} else {
					log.Info().Msg("successfully committed transaction; proceeding")
				}
			default:
				// email count is > 1, so send email to tech
				log.Error().Msgf(
					"%d non-hidden activists associated with email address %s for Interest"+
						" response %d Please correct.",
					count,
					email,
					id,
				)
			}
		}
	}

	log.Debug().Msg("finished processing run")
}
