package processor

/* Form application queries */
const applicationResponsesToProcessQuery = "SELECT id FROM form_application WHERE processed = 0 and name <> ''"

const processApplicationOnNameQuery = `
# try to match on name
UPDATE
	activists
INNER JOIN
	form_application ON activists.name = form_application.name
SET
	activists.email = IF(activists.email = '', form_application.email, activists.email),
	activists.phone = IF(activists.phone = '', form_application.phone, activists.phone),
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
	# mark as processed
	form_application.processed = 1
WHERE
	form_application.id = ?
	and form_application.name <> ''
	and form_application.processed = 0
	and activists.hidden = 0
	and form_application.name <> '';
`

const applicationProcessingStatusQuery = "SELECT processed FROM form_application WHERE id = ?"

const applicationSelectEmailQuery = "SELECT email FROM form_application WHERE id = ?"

const processApplicationOnEmailQuery = `
# try to match on email
UPDATE
	activists
INNER JOIN
	form_application ON activists.email = form_application.email
SET
	activists.phone = IF(activists.phone = '', form_application.phone, activists.phone),
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
	# mark as processed
	form_application.processed = 1
WHERE
	form_application.id = ?
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
	study_conversation
)
select
        NULL,
        concat(form_application.name,' (inserted by application, check for duplicate)'),
        form_application.email,
        form_application.phone,
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
        NULL
from
	form_application
WHERE
	form_application.id = ?
	and form_application.name <> ''
	and form_application.processed = 0
	and form_application.email not in (select * from (select email from activists where hidden < 1 and email <> '') temp1)
	and concat(form_application.name,' (inserted by application, check for duplicate)') not in (select * from (select name from activists where hidden < 1 and name <> '') temp2);
`

const processApplicationByInsertUpdateQuery = `
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
	and activists.hidden < 1;
`

/* Form interest query */
const interestResponsesToProcessQuery = "SELECT id FROM form_interest WHERE processed = 0 and name <> ''"

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
	# update interest date
	activists.interest_date = form_interest.timestamp,
	# only update the following columns if the new values are not empty
	activists.dev_interest = IFNULL(CONCAT_WS(', ', IF(LENGTH(dev_interest),dev_interest,NULL), IF(LENGTH(form_interest.interests),form_interest.interests,NULL)),''),
	activists.referral_friends = IF(LENGTH(form_interest.referral_friends), form_interest.referral_friends, activists.referral_friends),
	activists.referral_apply = IF(LENGTH(form_interest.referral_apply), CONCAT_WS(', ', IF(LENGTH(activists.referral_apply),activists.referral_apply,NULL),form_interest.referral_apply), activists.referral_apply),
	activists.referral_outlet = IF(LENGTH(form_interest.referral_outlet), form_interest.referral_outlet, activists.referral_outlet),
	# only update source if source is currently empty
	activists.source = IF(LENGTH(activists.source), activists.source, form_interest.form),
	# mark as processed
	form_interest.processed = 1

WHERE
	form_interest.id = ?
	and form_interest.processed = 0
	and activists.hidden = 0
	and form_interest.name <> '';
`

const interestProcessingStatusQuery = "SELECT processed FROM form_interest WHERE id = ?"

const interestSelectEmailQuery = "SELECT email FROM form_interest WHERE id = ?"

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
	# update interest date
	activists.interest_date = form_interest.timestamp,
	# only update the following columns if the new values are not empty
	activists.dev_interest = IFNULL(CONCAT_WS(', ', IF(LENGTH(dev_interest),dev_interest,NULL), IF(LENGTH(form_interest.interests),form_interest.interests,NULL)),''),
	activists.referral_friends = IF(LENGTH(form_interest.referral_friends), form_interest.referral_friends, activists.referral_friends),
	activists.referral_apply = IF(LENGTH(form_interest.referral_apply), CONCAT_WS(', ', IF(LENGTH(activists.referral_apply),activists.referral_apply,NULL),form_interest.referral_apply), activists.referral_apply),
	activists.referral_outlet = IF(LENGTH(form_interest.referral_outlet), form_interest.referral_outlet, activists.referral_outlet),
	# only update source if source is currently empty
	activists.source = IF(LENGTH(activists.source), activists.source, form_interest.form),
	# mark as processed
	form_interest.processed = 1
WHERE
	form_interest.id = ?
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
    study_conversation
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
    NULL
FROM
	form_interest
WHERE
	form_interest.id = ?
	and form_interest.processed = 0
	and form_interest.email not in (select * from (select email from activists where hidden < 1 and email <> '') temp1)
	and form_interest.name not in (select * from (select name from activists where hidden < 1 and name <> '') temp2)
	and form_interest.name <> '';
`

const processInsertByInsertUpdateQuery = `
# mark as processed if application date in activists table matches date in application
UPDATE
	form_interest
INNER JOIN
	activists on activists.name = form_interest.name
SET
	form_interest.processed = 1
WHERE
	form_interest.id = ?
	AND activists.interest_date = timestamp
	AND form_interest.processed = 0
	AND activists.hidden < 1;
`

/* Common queries */
const countActivistsForEmailQuery = "SELECT count(id) AS amount FROM activists WHERE hidden = 0 and email = ?"
