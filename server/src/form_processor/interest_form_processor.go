package form_processor

import (
	"context"
	"errors"
	"fmt"

	"github.com/dxe/adb/model"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"github.com/rs/zerolog/log"
)

const interestProcessingStatusQuery = "SELECT processed FROM form_interest WHERE id = ?"

const processInterestOnNameQuery = `
# try to match on name
UPDATE
	activists
INNER JOIN
	form_interest ON activists.name = form_interest.name
SET
	activists.email = IF(activists.email = '', form_interest.email, activists.email),
	activists.phone = IF(activists.phone = '', form_interest.phone, activists.phone),
	activists.location = IF(activists.location = '', form_interest.zip, activists.location),
	# check proper prospect boxes based on application type
	activists.circle_interest = IF(form_interest.form = 'Circle Interest Form', 1, activists.circle_interest),
	# update interest date only if it's currently null
	activists.interest_date = COALESCE(activists.interest_date, form_interest.timestamp),
	# only update the following columns if the new values are not empty
	activists.dev_interest = IFNULL(CONCAT_WS(', ', IF(LENGTH(dev_interest),dev_interest,NULL), IF(LENGTH(form_interest.interests),form_interest.interests,NULL)),''),
	activists.referral_friends = IF(LENGTH(form_interest.referral_friends), form_interest.referral_friends, activists.referral_friends),
	activists.referral_apply = IF(LENGTH(form_interest.referral_apply), CONCAT_WS(', ', IF(LENGTH(activists.referral_apply),activists.referral_apply,NULL),form_interest.referral_apply), activists.referral_apply),
	activists.referral_outlet = IF(LENGTH(form_interest.referral_outlet), form_interest.referral_outlet, activists.referral_outlet),
	# only update source if source is currently empty
	activists.source = IF(LENGTH(activists.source), activists.source, form_interest.form),
	activists.discord_id = IF(LENGTH(activists.discord_id), activists.discord_id, IF(LENGTH(form_interest.discord_id), form_interest.discord_id, NULL)),
	# mark as processed
	form_interest.processed = 1

WHERE
    chapter_id = ` + model.SFBayChapterIdStr + `
	and form_interest.id = ?
	and form_interest.processed = 0
	and activists.hidden = 0
	and form_interest.name <> '';
`

const processInterestOnEmailQuery = `
# try to match on email
UPDATE
	activists
INNER JOIN
	form_interest ON activists.email = form_interest.email
SET
	activists.phone = IF(activists.phone = '', form_interest.phone, activists.phone),
	activists.location = IF(activists.location = '', form_interest.zip, activists.location),
	# check proper prospect boxes based on application type
	activists.circle_interest = IF(form_interest.form = 'Circle Interest Form', 1, activists.circle_interest),
	# update interest date only if it's currently null
	activists.interest_date = COALESCE(activists.interest_date, form_interest.timestamp),
	# only update the following columns if the new values are not empty
	activists.dev_interest = IFNULL(CONCAT_WS(', ', IF(LENGTH(dev_interest),dev_interest,NULL), IF(LENGTH(form_interest.interests),form_interest.interests,NULL)),''),
	activists.referral_friends = IF(LENGTH(form_interest.referral_friends), form_interest.referral_friends, activists.referral_friends),
	activists.referral_apply = IF(LENGTH(form_interest.referral_apply), CONCAT_WS(', ', IF(LENGTH(activists.referral_apply),activists.referral_apply,NULL),form_interest.referral_apply), activists.referral_apply),
	activists.referral_outlet = IF(LENGTH(form_interest.referral_outlet), form_interest.referral_outlet, activists.referral_outlet),
	# only update source if source is currently empty
	activists.source = IF(LENGTH(activists.source), activists.source, form_interest.form),
	activists.discord_id = IF(LENGTH(activists.discord_id), activists.discord_id, IF(LENGTH(form_interest.discord_id), form_interest.discord_id, NULL)),
	# mark as processed
	form_interest.processed = 1
WHERE
    chapter_id = ` + model.SFBayChapterIdStr + `
	and form_interest.id = ?
	AND form_interest.processed = 0
	AND activists.hidden = 0
	AND form_interest.email <> ''
	AND form_interest.name <> '';
`

const processInterestByInsertQuery = `
# insert new records
INSERT INTO activists (
	id,
	name,
	email,
	phone,
	location,
    facebook,
    activist_level,
    hidden,
    connector,
    source,
    hiatus,
    date_organizer,
    dob,
    training0,
    training1,
    training2,
    training3,
    training4,
    training5,
    training6,
    prospect_organizer,
    prospect_chapter_member,
    circle_agreement,
    dev_manager,
    dev_interest,
    dev_auth,
    dev_email_sent,
    dev_vetted,
    dev_interview,
    dev_onboarding,
    dev_application_date,
    cm_first_email,
    cm_approval_email,
    cm_warning_email,
    cir_first_email,
    referral_friends,
    referral_apply,
    referral_outlet,
    circle_interest,
    interest_date,
    mpi,
    notes,
    dev_quiz,
    vision_wall,
    study_group,
    study_activator,
    study_conversation,
	discord_id,
    chapter_id
)
SELECT
    NULL,
    form_interest.name,
    form_interest.email,
    form_interest.phone,
    form_interest.zip,
    '',
    'Supporter',
    '0',
    '',
    form_interest.form,
    '0',
     NULL,
     NULL,
     NULL,
     NULL,
     NULL,
     NULL,
     NULL,
     NULL,
     NULL,
     0,
     0,
     0,
     '',
    IF(LENGTH(form_interest.interests),form_interest.interests,''),
    NULL,
    NULL,
    '0',
    NULL,
    '0',
    NULL,
    NULL,
    NULL,
    NULL,
    NULL,
    form_interest.referral_friends,
    form_interest.referral_apply,
    form_interest.referral_outlet,
    IF(form_interest.form = 'Circle Interest Form', 1, 0),
    form_interest.timestamp,
    0,
    NULL,
    NULL,
    '',
    '',
    '',
    NULL,
	IF(LENGTH(form_interest.discord_id),form_interest.discord_id,NULL),
    '` + model.SFBayChapterIdStr + `'
FROM
	form_interest
WHERE
	form_interest.id = ?
	and form_interest.processed = 0
	and form_interest.email not in (select * from (select email from activists where hidden < 1 and email <> '') temp1)
	and form_interest.name not in (select * from (select name from activists where hidden < 1 and name <> '') temp2)
	and form_interest.name <> '';
`

const markInterestProcessedQuery = `
UPDATE
	form_interest
INNER JOIN
	activists on activists.name = form_interest.name
SET
	form_interest.processed = 1
WHERE
    activists.chapter_id = ` + model.SFBayChapterIdStr + `
	AND form_interest.id = ?
	AND form_interest.processed = 0
	AND activists.hidden < 1;
`

func ProcessInterestForms(db *sqlx.DB) {
	log.Debug().Msg("processing interest forms")

	interestIds, isSuccess := getResponsesToProcess(db,
		"SELECT id FROM form_interest WHERE processed = 0 and name <> ''")
	if !isSuccess {
		log.Error().Msg("failed to get interestIds; exiting")
		return
	}
	if len(interestIds) == 0 {
		log.Debug().Msg("no new form_interest submissions to process")
	}
	for _, id := range interestIds {
		err := processInterestForm(id, db)
		if err != nil {
			log.Error().Msgf("error processing interest form; exiting: %v", err)
			return
		}
	}

	log.Debug().Msg("finished processing interest forms")
}

func processInterestForm(id int, db *sqlx.DB) error {
	log.Info().Msgf("processing Interest row %d", id)
	_, err := db.Exec(processInterestOnNameQuery, id)
	if err != nil {
		return fmt.Errorf("failed to process interest on name; %s", err)
	}

	// Return early if previous query updated activist based on name.
	processed, err := getProcessingStatus(db, interestProcessingStatusQuery, id)
	if err != nil {
		return fmt.Errorf("failed to get processing status: %v", err)
	}
	if processed {
		log.Info().Msg("Updated activist with interest form based on name")
		return nil
	}

	// check how many records are tied to this email address
	email, isSuccess := getEmail(db, "SELECT email FROM form_interest WHERE id = ?", id)
	if !isSuccess {
		return errors.New("failed to get email")
	}
	count, isSuccess := countActivistsForEmail(db, email)
	if !isSuccess {
		return errors.New("failed to count activists for email")
	}

	switch count {
	case 1:
		err := updateActivistWithInterestFormBasedOnEmail(db, id)
		if err != nil {
			return fmt.Errorf("failed to update activist: %w", err)
		}
	case 0:
		err := insertActivistFromInterestForm(db, id)
		if err != nil {
			return fmt.Errorf("failed to insert activist: %w", err)
		}
	default:
		// email count is > 1, so send email to tech
		log.Error().Msgf(
			"%d non-hidden activists associated with email address %s for Interest response %d Please correct.",
			count,
			email,
			id,
		)
	}

	return nil
}

func updateActivistWithInterestFormBasedOnEmail(db *sqlx.DB, id int) error {
	res, err := db.Exec(processInterestOnEmailQuery, id)
	if err != nil {
		return fmt.Errorf("failed to processInterestOnEmailQuery; %s", err)
	}

	count, getRowsAffectedErr := res.RowsAffected()
	if getRowsAffectedErr != nil {
		return fmt.Errorf("failed to get processInterestOnEmailQuery affected rows; %s",
			getRowsAffectedErr)
	}
	if count != 1 {
		return fmt.Errorf("no rows updated on processInterestOnEmailQuery")
	}

	log.Info().Msg("Updated activist with interest form based on email")
	return nil
}

func insertActivistFromInterestForm(db *sqlx.DB, id int) error {
	ctx := context.Background()
	tx, txErr := db.BeginTx(ctx, nil)
	if txErr != nil {
		return fmt.Errorf("failed to start transaction; %s", txErr)
	}
	defer tx.Rollback()

	insertResult, processErr := db.ExecContext(ctx, processInterestByInsertQuery, id)
	if processErr != nil {
		return fmt.Errorf("failed to processInterestByInsertQuery; %s", processErr)
	}
	insertCount, getRowsAffectedErr := insertResult.RowsAffected()
	if getRowsAffectedErr != nil {
		return fmt.Errorf("failed to get processInterestByInsertQuery affected rows; %s",
			getRowsAffectedErr)
	}
	if insertCount != 1 {
		return fmt.Errorf("no rows updated on processInterestByInsertQuery")
	}

	markResult, updateErr := db.ExecContext(ctx, markInterestProcessedQuery, id)
	if updateErr != nil {
		return fmt.Errorf("failed to markInterestProcessedQuery; %s", updateErr)
	}

	markCount, getRowsAffectedErr := markResult.RowsAffected()
	if getRowsAffectedErr != nil {
		return fmt.Errorf("failed to get markInterestProcessedQuery affected rows; %s",
			getRowsAffectedErr)
	}
	if markCount != 1 {
		log.Error().Msg("interest form was processed but not marked as such")
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction; %s", commitErr)
	}

	log.Info().Msg("inserted activist from interest form")
	return nil
}
