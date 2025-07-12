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

const applicationProcessingStatusQuery = "SELECT processed FROM form_application WHERE id = ?"

const processApplicationOnNameQuery = `
# try to match on name
UPDATE
	activists
INNER JOIN
	form_application ON activists.name = form_application.name
SET
	activists.email = IF(activists.email = '', form_application.email, activists.email),
	activists.phone = IF(activists.phone = '', form_application.phone, activists.phone),
	activists.pronouns = IF(activists.pronouns = '', form_application.pronouns, activists.pronouns),
	activists.location = IF(activists.location = '', form_application.zip, activists.location),
	activists.dob = IF(activists.dob = '', form_application.birthday, activists.dob),
	# check proper prospect boxes based on application type
	activists.prospect_organizer = IF(form_application.application_type = 'organizer', 1, (IF((form_application.application_type = 'senior-organizer' and activist_level <> 'organizer'), 1, activists.prospect_organizer))),
	activists.prospect_chapter_member = IF(form_application.application_type = 'chapter-member', 1, (IF((form_application.application_type in ('senior-organizer','organizer') and activist_level in ('supporter','circle member','non-local')), 1, activists.prospect_chapter_member))),
	activists.circle_agreement = IF(form_application.application_type = 'circle-member', 1, activists.circle_agreement),
	activists.circle_interest = IF(activists.id NOT in (select activist_id from working_group_members UNION select activist_id from circle_members), 1, activists.circle_interest),
	# update application date & type
	activists.dev_application_date = form_application.timestamp,
	activists.dev_application_type = form_application.application_type,
	# only update the following columns if the new values are not empty
	activists.dev_interest = CONCAT_WS(', ', IF(LENGTH(dev_interest),dev_interest,NULL), IF(LENGTH(form_application.circle_interest),form_application.circle_interest,NULL), IF(LENGTH(wg_interest),wg_interest,NULL), IF(LENGTH(committee_interest),committee_interest,NULL)),
	activists.referral_friends = IF(LENGTH(form_application.referral_friends), form_application.referral_friends, activists.referral_friends),
	activists.referral_apply = IF(LENGTH(form_application.referral_apply), CONCAT_WS(', ', IF(LENGTH(activists.referral_apply),activists.referral_apply,NULL),form_application.referral_apply), activists.referral_apply),
	activists.referral_outlet = IF(LENGTH(form_application.referral_outlet), form_application.referral_outlet, activists.referral_outlet),
	activists.language = IF(LENGTH(form_application.language), form_application.language, activists.language),
	activists.accessibility = IF(LENGTH(form_application.accessibility), form_application.accessibility, activists.accessibility),
	# mark as processed
	form_application.processed = 1
WHERE
    chapter_id = ` + model.SFBayChapterIdStr + `
	and form_application.id = ?
	and form_application.name <> ''
	and form_application.processed = 0
	and activists.hidden = 0
	and form_application.name <> '';
`

const processApplicationOnEmailQuery = `
# try to match on email
UPDATE
	activists
INNER JOIN
	form_application ON activists.email = form_application.email
SET
	activists.phone = IF(activists.phone = '', form_application.phone, activists.phone),
	activists.pronouns = IF(activists.pronouns = '', form_application.pronouns, activists.pronouns),
	activists.location = IF(activists.location = '', form_application.zip, activists.location),
	activists.dob = IF(activists.dob = '', form_application.birthday, activists.dob),
	# check proper prospect boxes based on application type
	activists.prospect_organizer = IF(form_application.application_type = 'organizer', 1, (IF((form_application.application_type = 'senior-organizer' and activist_level <> 'organizer'), 1, activists.prospect_organizer))),
	activists.prospect_chapter_member = IF(form_application.application_type = 'chapter-member', 1, (IF((form_application.application_type in ('senior-organizer','organizer') and activist_level in ('supporter','circle member','non-local')), 1, activists.prospect_chapter_member))),
	activists.circle_agreement = IF(form_application.application_type = 'circle-member', 1, activists.circle_agreement),
	activists.circle_interest = IF(activists.id NOT in (select activist_id from working_group_members UNION select activist_id from circle_members), 1, activists.circle_interest),
	# update application date & type
	activists.dev_application_date = form_application.timestamp,
	activists.dev_application_type = form_application.application_type,
	# only update the following columns if the new values are not empty
	activists.dev_interest = CONCAT_WS(', ', IF(LENGTH(dev_interest),dev_interest,NULL), IF(LENGTH(form_application.circle_interest),form_application.circle_interest,NULL), IF(LENGTH(wg_interest),wg_interest,NULL), IF(LENGTH(committee_interest),committee_interest,NULL)),
	activists.referral_friends = IF(LENGTH(form_application.referral_friends), form_application.referral_friends, activists.referral_friends),
	activists.referral_apply = IF(LENGTH(form_application.referral_apply), CONCAT_WS(', ', IF(LENGTH(activists.referral_apply),activists.referral_apply,NULL),form_application.referral_apply), activists.referral_apply),
	activists.referral_outlet = IF(LENGTH(form_application.referral_outlet), form_application.referral_outlet, activists.referral_outlet),
	activists.language = IF(LENGTH(form_application.language), form_application.language, activists.language),
	activists.accessibility = IF(LENGTH(form_application.accessibility), form_application.accessibility, activists.accessibility),
	# mark as processed
	form_application.processed = 1
WHERE
    chapter_id = ` + model.SFBayChapterIdStr + `
	and form_application.id = ?
	and form_application.name <> ''
	and form_application.processed = 0
	and activists.hidden = 0
	and form_application.email <> '';
`

const processApplicationByInsertQuery = `
# insert new records
INSERT INTO activists (
	id,
	name,
	email,
	phone,
	pronouns,
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
	dev_application_type,
	study_group,
	study_activator,
	study_conversation,
    chapter_id,
	language,
	accessibility
)
select
        NULL,
        concat(form_application.name,' (inserted by application, check for duplicate)'),
        form_application.email,
        form_application.phone,
        form_application.pronouns,
        form_application.zip,
        '',
        'Supporter',
        '0',
        '',
        'Application Form',
        '0',
        NULL,
        form_application.birthday,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        NULL,
        IF(form_application.application_type = 'organizer', 1,
        IF((form_application.application_type = 'senior-organizer'), 1, 0)),
        IF(form_application.application_type = 'chapter-member', 1, IF((form_application.application_type in ('organizer','senior-organizer')), 1, 0)),
        IF(form_application.application_type = 'circle-member', 1, 0),
        '',
        CONCAT_WS(', ', IF(LENGTH(form_application.circle_interest),form_application.circle_interest,NULL), IF(LENGTH(wg_interest),wg_interest,NULL), IF(LENGTH(committee_interest),committee_interest,NULL)),
        NULL,
        NULL,
        '0',
        NULL,
        '0',
        form_application.timestamp,
        NULL,
        NULL,
        NULL,
        NULL,
        form_application.referral_friends,
        form_application.referral_apply,
        form_application.referral_outlet,
        1,
        NULL,
        0,
        NULL,
        NULL,
        '',
        form_application.application_type,
        '',
        '',
        NULL,
        '` + model.SFBayChapterIdStr + `',
		form_application.language,
		form_application.accessibility
from
	form_application
WHERE
	form_application.id = ?
	and form_application.name <> ''
	and form_application.processed = 0
	and form_application.email not in (select * from (select email from activists where hidden < 1 and email <> '') temp1)
	and concat(form_application.name,' (inserted by application, check for duplicate)') not in (select * from (select name from activists where hidden < 1 and name <> '') temp2);
`

const markApplicationProcessedQuery = `
# mark as processed if application date in activists table matches date in application
update
	form_application
INNER JOIN
	activists on activists.name = concat(form_application.name,' (inserted by application, check for duplicate)')
SET
	form_application.processed = 1
WHERE
	form_application.id = ?
	and activists.dev_application_date = cast(form_application.timestamp as date)
	and form_application.processed = 0
	and activists.hidden < 1
    and activists.chapter_id = ` + model.SFBayChapterIdStr + `;
`

func ProcessApplicationForms(db *sqlx.DB) {
	log.Debug().Msg("processing application forms")

	applicationIds, isSuccess := getResponsesToProcess(db,
		"SELECT id FROM form_application WHERE processed = 0 and name <> ''")
	if !isSuccess {
		log.Error().Msg("failed to get applicationIds; exiting")
		return
	}
	if len(applicationIds) == 0 {
		log.Debug().Msg("no new form_application submissions to process")
	}
	for _, id := range applicationIds {
		err := processApplicationForm(id, db)
		if err != nil {
			log.Error().Msgf("error processing application form; exiting: %v", err)
			return
		}
	}

	log.Debug().Msg("finished processing application forms")
}

func processApplicationForm(id int, db *sqlx.DB) error {
	log.Info().Msgf("Processing Application row %d", id)
	_, err := db.Exec(processApplicationOnNameQuery, id)
	if err != nil {
		return fmt.Errorf("failed to prrocess application on name; %s", err)
	}

	// Return early if previous query updated activist based on name.
	processed, err := getProcessingStatus(db, applicationProcessingStatusQuery, id)
	if err != nil {
		return fmt.Errorf("failed to get processing status: %v", err)
	}
	if processed {
		return nil
	}

	// check how many records are tied to this email address
	email, isSuccess := getEmail(db, "SELECT email FROM form_application WHERE id = ?", id)
	if !isSuccess {
		return errors.New("failed to get email; exiting")
	}
	count, isSuccess := countActivistsForEmail(db, email)
	if !isSuccess {
		return errors.New("failed to count activists for email; exiting")
	}

	switch count {
	case 1:
		err := updateActivistWithApplicationForm(db, id)
		if err != nil {
			return fmt.Errorf("failed to update activist: %w", err)
		}
	case 0:
		err := insertActivistFromApplicationForm(db, id)
		if err != nil {
			return fmt.Errorf("failed to insert activist: %w", err)
		}
	default:
		// email count is > 1, so send email to tech
		log.Error().Msgf(
			"%d non-hidden activists associated with email address %s for Application response %d Please correct.",
			count,
			email,
			id,
		)
	}

	return nil
}

func updateActivistWithApplicationForm(db *sqlx.DB, id int) error {
	_, err := db.Exec(processApplicationOnEmailQuery, id)
	if err != nil {
		return fmt.Errorf("failed to processApplicationOnEmailQuery; %s", err)
	}

	log.Info().Msg("Updated activist with application form")
	return nil
}

func insertActivistFromApplicationForm(db *sqlx.DB, id int) error {
	ctx := context.Background()
	tx, txErr := db.BeginTx(ctx, nil)
	if txErr != nil {
		return fmt.Errorf("failed to start transaction; %s", txErr)
	}
	defer tx.Rollback()

	_, processErr := tx.ExecContext(ctx, processApplicationByInsertQuery, id)
	if processErr != nil {
		return fmt.Errorf("failed to processApplicationByInsertQuery; %s", processErr)
	}

	res, updateErr := tx.ExecContext(ctx, markApplicationProcessedQuery, id)
	if updateErr != nil {
		return fmt.Errorf("failed to processApplicationByInsertUpdateQuery; %s", updateErr)
	}

	count, getRowsAffectedErr := res.RowsAffected()
	if getRowsAffectedErr != nil {
		return fmt.Errorf("failed to get processApplicationByInsertUpdateQuery affected rows; %s",
			getRowsAffectedErr)
	}
	if count != 1 {
		log.Error().Msg("the activist was not updated (application date in activists table does not match the date " +
			"in application?) -- please correct")
	}

	commitErr := tx.Commit()
	if commitErr != nil {
		return fmt.Errorf("failed to commit transaction; %s", commitErr)
	}

	log.Info().Msg("inserted activist from application form")
	return nil
}
