package processor

// NOTE(AlexisDeschamps): the `CREATE TABLE` queries were obtained by calling
// `SHOW CREATE TABLE <table name>` and then modifying `AUTO_INCREMENT=<>` to be
// `AUTO_INCREMENT=1`.

/* CREATE TABLE queries */
const createTableFormApplicationQuery = `
CREATE TABLE form_application (
  id int(11) NOT NULL AUTO_INCREMENT,
  email text COLLATE utf8mb4_unicode_ci NOT NULL,
  name text COLLATE utf8mb4_unicode_ci NOT NULL,
  phone text COLLATE utf8mb4_unicode_ci NOT NULL,
  address text COLLATE utf8mb4_unicode_ci NOT NULL,
  city text COLLATE utf8mb4_unicode_ci NOT NULL,
  zip text COLLATE utf8mb4_unicode_ci NOT NULL,
  birthday text COLLATE utf8mb4_unicode_ci NOT NULL,
  pronouns text COLLATE utf8mb4_unicode_ci NOT NULL,
  application_type text COLLATE utf8mb4_unicode_ci NOT NULL,
  agree_circle text COLLATE utf8mb4_unicode_ci NOT NULL,
  agree_mpp text COLLATE utf8mb4_unicode_ci NOT NULL,
  circle_interest text COLLATE utf8mb4_unicode_ci NOT NULL,
  wg_interest text COLLATE utf8mb4_unicode_ci NOT NULL,
  committee_interest text COLLATE utf8mb4_unicode_ci NOT NULL,
  referral_friends text COLLATE utf8mb4_unicode_ci NOT NULL,
  referral_apply text COLLATE utf8mb4_unicode_ci NOT NULL,
  referral_outlet text COLLATE utf8mb4_unicode_ci NOT NULL,
  contact_method varchar(50) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  processed tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (id)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createTableWorkingGroupMembersQuery = `
CREATE TABLE working_group_members (
  working_group_id int(11) NOT NULL,
  activist_id int(11) NOT NULL,
  point_person tinyint(4) NOT NULL DEFAULT '0',
  non_member_on_mailing_list tinyint(4) NOT NULL DEFAULT '0',
  UNIQUE KEY working_group_member_ukey (working_group_id,activist_id),
  KEY activist_id (activist_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createTableCircleMembersQuery = `
CREATE TABLE circle_members (
  circle_id int(11) NOT NULL,
  activist_id int(11) NOT NULL,
  point_person tinyint(4) NOT NULL DEFAULT '0',
  non_member_on_mailing_list tinyint(4) NOT NULL DEFAULT '0',
  UNIQUE KEY working_group_member_ukey (circle_id,activist_id),
  KEY activist_id (activist_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createTableActivistsQuery = `
CREATE TABLE activists (
  id int(11) NOT NULL AUTO_INCREMENT,
  name varchar(80) COLLATE utf8mb4_unicode_ci NOT NULL,
  preferred_name varchar(80) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  email varchar(80) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  phone varchar(20) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  location varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT '',
  facebook varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  activist_level varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT 'Supporter',
  hidden tinyint(1) NOT NULL DEFAULT '0',
  connector varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  source varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  hiatus tinyint(1) NOT NULL DEFAULT '0',
  date_organizer date DEFAULT NULL,
  date_senior_organizer date DEFAULT NULL,
  dob text COLLATE utf8mb4_unicode_ci,
  training0 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training1 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training2 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training3 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training4 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training5 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training6 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  consent_quiz varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training_protest varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  prospect_organizer tinyint(1) NOT NULL DEFAULT '0',
  prospect_chapter_member tinyint(4) NOT NULL DEFAULT '0',
  circle_agreement tinyint(1) NOT NULL DEFAULT '0',
  dev_manager varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  dev_interest varchar(1000) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  dev_auth varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  dev_email_sent varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  dev_vetted tinyint(1) NOT NULL DEFAULT '0',
  dev_interview varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  dev_onboarding tinyint(1) NOT NULL DEFAULT '0',
  dev_application_date date DEFAULT NULL,
  dev_application_type varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  dev_quiz varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  cm_first_email varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  cm_approval_email varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  cm_warning_email varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  cir_first_email varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  referral_friends varchar(500) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  referral_apply varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  referral_outlet varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  circle_interest tinyint(1) NOT NULL DEFAULT '0',
  interest_date varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  mpi tinyint(1) NOT NULL DEFAULT '0',
  notes text COLLATE utf8mb4_unicode_ci,
  vision_wall varchar(10) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  study_group varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  study_activator varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  study_conversation varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  survey_completion varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  voting_agreement tinyint(1) NOT NULL DEFAULT '0',
  street_address varchar(200) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  city varchar(100) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  state varchar(40) COLLATE utf8mb4_unicode_ci NOT NULL DEFAULT '',
  discord_id bigint(18) DEFAULT NULL,
  PRIMARY KEY (id),
  UNIQUE KEY name_ukey (name),
  KEY activists_email (email),
  KEY activist_level (activist_level),
  KEY hidden (hidden),
  KEY source (source),
  KEY mpi (mpi)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

const createTableFormInterestQuery = `
CREATE TABLE form_interest (
  id int(11) NOT NULL AUTO_INCREMENT,
  form varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  email varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  name varchar(255) COLLATE utf8mb4_unicode_ci NOT NULL,
  phone text COLLATE utf8mb4_unicode_ci NOT NULL,
  zip text COLLATE utf8mb4_unicode_ci NOT NULL,
  referral_friends text COLLATE utf8mb4_unicode_ci NOT NULL,
  referral_apply text COLLATE utf8mb4_unicode_ci NOT NULL,
  referral_outlet text COLLATE utf8mb4_unicode_ci NOT NULL,
  comments text COLLATE utf8mb4_unicode_ci NOT NULL,
  interests text COLLATE utf8mb4_unicode_ci NOT NULL,
  timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  processed tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (id),
  UNIQUE KEY uidx (form,name,email)
) ENGINE=InnoDB AUTO_INCREMENT=1 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
`

/* Common queries */
const insertActivistQuery = `
INSERT INTO activists (id, email, name) VALUES (NULL, "email1", ?);
`

const getActivistsQuery = `SELECT id FROM activists;`

/* Form application queries */
const insertIntoFormApplicationQuery = `
INSERT INTO form_application (
  id,
  email,
  name,
  phone,
  address,
  city,
  zip,
  birthday,
  pronouns,
  application_type,
  agree_circle,
  agree_mpp,
  circle_interest,
  wg_interest,
  committee_interest,
  referral_friends,
  referral_apply,
  referral_outlet,
  contact_method,
  processed
) VALUES (
  NULL,
  "email1",
  "name1",
  "phone1",
  "address1",
  "city1",
  "zip1",
  "birthday1",
  "pronouns1",
  "application_type1",
  "agree_circle1",
  "agree_mpp1",
  "circle_interest1",
  "wg_interest1",
  "committee_interest1",
  "referral_friends1",
  "referral_apply1",
  "referral_outlet1",
  "contact_method1",
  false
);
`

/* Form interest queries */
const insertIntoFormInterestQuery = `
INSERT INTO form_interest (
  id,
  form,
  email,
  name,
  phone,
  zip,
  referral_friends,
  referral_apply,
  referral_outlet,
  comments,
  interests
) VALUES (
  NULL,
  "form1",
  "email1",
  "name1",
  "phone1",
  "zip1",
  "referral_friends1",
  "referral_apply1",
  "referral_outlet1",
  "comments1",
  "interests1"
);
`
