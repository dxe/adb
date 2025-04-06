package model

import (
	"github.com/dxe/adb/config"
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewDB(dataSourceName string) *sqlx.DB {
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	return db
}

func WipeDatabase(db *sqlx.DB) {
	if config.IsProd {
		panic("Cannot drop tables in prod")
	}

	db.MustExec(`DROP TABLE IF EXISTS activists`)
	db.MustExec(`DROP TABLE IF EXISTS activists_history`)
	db.MustExec(`DROP TABLE IF EXISTS events`)
	db.MustExec(`DROP TABLE IF EXISTS event_attendance`)
	db.MustExec(`DROP TABLE IF EXISTS users_roles`)
	db.MustExec(`DROP TABLE IF EXISTS adb_users`)
	db.MustExec(`DROP TABLE IF EXISTS merged_activist_attendance`)
	db.MustExec(`DROP TABLE IF EXISTS working_groups`)
	db.MustExec(`DROP TABLE IF EXISTS working_group_members`)
	db.MustExec(`DROP TABLE IF EXISTS circles`)
	db.MustExec(`DROP TABLE IF EXISTS circle_members`)
	db.MustExec(`DROP TABLE IF EXISTS fb_pages`)
	db.MustExec(`DROP TABLE IF EXISTS fb_events`)
	db.MustExec(`DROP TABLE IF EXISTS discord_users`)
	db.MustExec(`DROP TABLE IF EXISTS discord_messages`)
	db.MustExec(`DROP TABLE IF EXISTS form_application`)
	db.MustExec(`DROP TABLE IF EXISTS form_interest`)
	db.MustExec(`DROP TABLE IF EXISTS form_international`)
	db.MustExec(`DROP TABLE IF EXISTS form_discord`)
	db.MustExec(`DROP TABLE IF EXISTS form_international_actions`)
	db.MustExec(`DROP TABLE IF EXISTS interactions`)

	db.MustExec(`
CREATE TABLE activists (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  preferred_name VARCHAR(80) NOT NULL DEFAULT '',
  pronouns VARCHAR(20) NOT NULL DEFAULT '',
  email VARCHAR(80) NOT NULL DEFAULT '',
  phone VARCHAR(20) NOT NULL DEFAULT '',
  location VARCHAR(200) DEFAULT '',
  facebook VARCHAR(200) NOT NULL DEFAULT '',
  activist_level VARCHAR(40) NOT NULL DEFAULT 'Supporter',
  hidden TINYINT(1) NOT NULL DEFAULT '0',
  connector VARCHAR(100) NOT NULL DEFAULT '',
  source VARCHAR(255) NOT NULL DEFAULT '',
  hiatus TINYINT(1) NOT NULL DEFAULT '0',
  date_organizer DATE,
  dob TEXT,
  training0 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training1 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training2 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training3 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training4 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training5 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  training6 varchar(20) COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  consent_quiz VARCHAR(20),
  training_protest VARCHAR(20),
  prospect_organizer TINYINT(1) NOT NULL DEFAULT '0',
  prospect_chapter_member TINYINT NOT NULL DEFAULT '0',
  circle_agreement TINYINT NOT NULL DEFAULT '0',
  dev_manager VARCHAR(100) NOT NULL DEFAULT '',
  dev_interest VARCHAR(200) NOT NULL DEFAULT '',
  dev_auth VARCHAR(20),
  dev_email_sent VARCHAR(20),
  dev_vetted TINYINT(1) NOT NULL DEFAULT '0',
  dev_interview VARCHAR(20),
  dev_onboarding TINYINT(1) NOT NULL DEFAULT '0',
  dev_application_date DATE,
  dev_application_type VARCHAR(40) NOT NULL DEFAULT '',
  dev_quiz VARCHAR(20),
  cm_first_email VARCHAR(20),
  cm_approval_email VARCHAR(20),
  cm_warning_email VARCHAR(20),
  cir_first_email VARCHAR(20),
  referral_friends varchar(100) NOT NULL DEFAULT '',
  referral_apply varchar(100) NOT NULL DEFAULT '',
  referral_outlet varchar(100) NOT NULL DEFAULT '',
  circle_interest tinyint(1) NOT NULL DEFAULT '0',
  interest_date VARCHAR(20),
  mpi tinyint(1) NOT NULL DEFAULT '0',
  notes TEXT,
  vision_wall varchar(10) NOT NULL DEFAULT '',
  study_group varchar(40) NOT NULL DEFAULT '',
  study_activator varchar(40) NOT NULL DEFAULT '',
  study_conversation varchar(20),
  survey_completion VARCHAR(20),
  voting_agreement TINYINT(1) NOT NULL DEFAULT '0',
  street_address VARCHAR(200) NOT NULL DEFAULT '',
  city VARCHAR(100) NOT NULL DEFAULT '',
  state VARCHAR(40) NOT NULL DEFAULT '',
  discord_id BIGINT(18) DEFAULT NULL,
  lat FLOAT(10,6) NULL DEFAULT '0',
  lng FLOAT(10,6) NULL DEFAULT '0',
  chapter_id int(11) DEFAULT '0',
  assigned_to int(11) DEFAULT '0',
  followup_date date DEFAULT NULL,
  language varchar(40) NOT NULL DEFAULT '',
  accessibility varchar(300) NOT NULL DEFAULT '',
  UNIQUE (name, chapter_id)
)
`)

	db.MustExec(`
CREATE TABLE activists_history (
  revision INTEGER AUTO_INCREMENT,
  action VARCHAR(20) NOT NULL,
  timestamp TIMESTAMP DEFAULT NOW(),
  user_email VARCHAR(80) NOT NULL,
  activist_id INTEGER NOT NULL,
  name VARCHAR(80) NOT NULL,
  email VARCHAR(80) NOT NULL,
  facebook VARCHAR(200) NOT NULL,
  activist_level VARCHAR(40) NOT NULL,
  PRIMARY KEY (revision),
  INDEX (activist_id, revision)
  )
`)

	db.MustExec(`
CREATE TABLE event_attendance (
  activist_id INTEGER NOT NULL,
  event_id INTEGER NOT NULL,
  UNIQUE (activist_id, event_id),
  UNIQUE (event_id, activist_id)
)
`)

	db.MustExec(`
CREATE TABLE events (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(60) NOT NULL,
  date DATE NOT NULL,
  event_type VARCHAR(60) NOT NULL,
  survey_sent TINYINT(1) NOT NULL DEFAULT '0',
  suppress_survey TINYINT(1) NOT NULL DEFAULT '0',
  circle_id INTEGER NOT NULL DEFAULT '0',
  chapter_id INT(11) DEFAULT '0',
  INDEX (date, name),
  FULLTEXT (name)
)
`)

	db.MustExec(`
CREATE TABLE adb_users (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(60) NOT NULL,
  name VARCHAR(150) NOT NULL DEFAULT '',
  admin TINYINT(1) NOT NULL DEFAULT '0',
  chapter_id int(11) DEFAULT '0',
  disabled TINYINT(1) NOT NULL DEFAULT '0'
)
`)

	db.MustExec(`
CREATE TABLE merged_activist_attendance (
  original_activist_id INTEGER NOT NULL,
  target_activist_id INTEGER NOT NULL,
  event_id INTEGER NOT NULL,
  replaced_with_target_activist TINYINT(1) NOT NULL,
  UNIQUE (original_activist_id, target_activist_id, event_id)
)
`)

	db.MustExec(`
CREATE TABLE working_groups (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(60) NOT NULL,
  group_email VARCHAR(100) NOT NULL,
  visible TINYINT(1) NOT NULL DEFAULT '0',
  description TEXT NOT NULL,
  meeting_time TEXT NOT NULL,
  meeting_location TEXT NOT NULL,
  coords TEXT NOT NULL,
  UNIQUE (name)
)
`)

	db.MustExec(`
CREATE TABLE working_group_members (
  working_group_id INTEGER NOT NULL,
  activist_id INTEGER NOT NULL,
  -- True if the activist is the point person of the working group.
  -- There should be only one point person per working group, but we
  -- don't restrict that on the backend.
  point_person TINYINT NOT NULL DEFAULT '0',
  -- Some activists need to be on the mailing list even though they
  -- aren't in the workin group.
  non_member_on_mailing_list TINYINT NOT NULL DEFAULT '0',
  UNIQUE (working_group_id, activist_id),
  INDEX (activist_id)
)
`)

	db.MustExec(`
CREATE TABLE circles (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(60) NOT NULL,
  type TINYINT(1) NOT NULL,
  group_email VARCHAR(100) NOT NULL DEFAULT '1',
  visible TINYINT(1) NOT NULL DEFAULT '1',
  description TEXT NOT NULL,
  meeting_time TEXT NOT NULL,
  meeting_location TEXT NOT NULL,
  coords TEXT NOT NULL,
  UNIQUE (name)
)
`)

	db.MustExec(`
CREATE TABLE circle_members (
  circle_id INTEGER NOT NULL,
  activist_id INTEGER NOT NULL,
  point_person TINYINT NOT NULL DEFAULT '0',
  non_member_on_mailing_list TINYINT NOT NULL DEFAULT '0',
  UNIQUE (circle_id, activist_id),
  INDEX (activist_id)
)
`)

	db.MustExec(`
CREATE TABLE users_roles (
  user_id INT NOT NULL,
  role VARCHAR(45) NOT NULL,
  PRIMARY KEY (user_id, role),
  CONSTRAINT user_id
    FOREIGN KEY (user_id)
    REFERENCES adb_users (id)
    ON DELETE NO ACTION
    ON UPDATE NO ACTION
)
`)

	db.MustExec(`
CREATE TABLE fb_pages (
  id BIGINT(16) NOT NULL DEFAULT '0',
  chapter_id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(75) NOT NULL,
  region ENUM('','North America','Central & South America','Europe','Middle East & Africa','Asia-Pacific','Online') NOT NULL DEFAULT '',
  flag VARCHAR(2) NOT NULL DEFAULT '',
  lat FLOAT(10,6) NOT NULL DEFAULT '0.000000',
  lng FLOAT(10,6) NOT NULL DEFAULT '0.000000',
  fb_url VARCHAR(100) NOT NULL,
  twitter_url VARCHAR(100) NOT NULL DEFAULT '',
  insta_url VARCHAR(100) NOT NULL DEFAULT '',
  email VARCHAR(100) NOT NULL DEFAULT '',
  token VARCHAR(200) NOT NULL DEFAULT '',
  eventbrite_id VARCHAR(16) NOT NULL DEFAULT '',
  eventbrite_token VARCHAR(32) NOT NULL DEFAULT '',
  ml_type VARCHAR(100) NOT NULL DEFAULT '',
  ml_radius SMALLINT NOT NULL DEFAULT '0',
  ml_id VARCHAR(100) NOT NULL DEFAULT '',
  mentor VARCHAR(100) NOT NULL DEFAULT '',
  country VARCHAR(128) NOT NULL DEFAULT '',
  notes VARCHAR(512) NOT NULL DEFAULT '',
  last_contact VARCHAR(10) DEFAULT '',
  last_action VARCHAR(10) DEFAULT '',
  organizers JSON,
  email_token VARCHAR(64) DEFAULT NULL,
  last_checkin_email_sent TIMESTAMP DEFAULT NULL
)
`)

	db.MustExec(`
CREATE TABLE fb_events (
  id BIGINT NOT NULL,
  page_id BIGINT NOT NULL,
  name VARCHAR(200) NOT NULL,
  description TEXT,
  start_time DATETIME NOT NULL,
  end_time DATETIME NOT NULL,
  location_name VARCHAR(140),
  location_city VARCHAR(200),
  location_country VARCHAR(200),
  location_state VARCHAR(200),
  location_address VARCHAR(200),
  location_zip VARCHAR(20),
  lat FLOAT(10,6) DEFAULT NULL,
  lng FLOAT(10,6) DEFAULT NULL,
  cover VARCHAR(500),
  attending_count MEDIUMINT NOT NULL DEFAULT '0',
  interested_count MEDIUMINT NOT NULL DEFAULT '0',
  is_canceled TINYINT NOT NULL DEFAULT '0',
  last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  eventbrite_id VARCHAR(32) NOT NULL DEFAULT '',
  eventbrite_url VARCHAR(400) NOT NULL DEFAULT '',
  featured TINYINT NOT NULL DEFAULT '0',
  PRIMARY KEY (id, page_id)
)
`)

	db.MustExec(`
CREATE TABLE discord_users (
  id BIGINT(18) PRIMARY KEY,
  email VARCHAR(200) NOT NULL,
  token VARCHAR(64) NOT NULL,
  confirmed TINYINT(1) NOT NULL DEFAULT '0',
  confirm_date DATETIME DEFAULT NULL
)
`)

	db.MustExec(`
CREATE TABLE discord_messages (
  message_name VARCHAR(100) PRIMARY KEY,
  message_text VARCHAR(2000) NOT NULL,
  last_updated datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  updated_by BIGINT(18) NOT NULL
)
`)

	db.MustExec(`
CREATE TABLE form_application (
  id int(11) NOT NULL AUTO_INCREMENT,
  email varchar(200) NOT NULL,
  name varchar(200) NOT NULL,
  phone varchar(20) NOT NULL,
  address varchar(200) NOT NULL,
  city varchar(200) NOT NULL,
  zip varchar(10) NOT NULL,
  birthday varchar(20) NOT NULL,
  pronouns varchar(20) NOT NULL DEFAULT '',
  application_type varchar(50) NOT NULL,
  agree_circle varchar(200) NOT NULL DEFAULT ' ',
  agree_mpp varchar(200) NOT NULL DEFAULT ' ',
  circle_interest varchar(200) NOT NULL DEFAULT ' ',
  wg_interest varchar(200) NOT NULL DEFAULT ' ',
  committee_interest varchar(200) NOT NULL DEFAULT '  ',
  referral_friends varchar(200) NOT NULL DEFAULT ' ',
  referral_apply varchar(200) NOT NULL DEFAULT ' ',
  referral_outlet varchar(200) NOT NULL DEFAULT ' ',
  contact_method varchar(50) NOT NULL DEFAULT '',
  language varchar(40) NOT NULL DEFAULT '',
  accessibility varchar(300) NOT NULL DEFAULT '',
  timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  processed tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (id)
)`)

	db.MustExec(`
CREATE TABLE form_interest (
  id int(11) NOT NULL AUTO_INCREMENT,
  form varchar(255) NOT NULL,
  email varchar(255) NOT NULL,
  name varchar(255) NOT NULL,
  phone varchar(20) NOT NULL,
  zip varchar(10) NOT NULL,
  referral_friends varchar(200) NOT NULL DEFAULT '',
  referral_apply varchar(200) NOT NULL DEFAULT '',
  referral_outlet varchar(200) NOT NULL DEFAULT '',
  comments varchar(200) NOT NULL DEFAULT '',
  interests varchar(400) NOT NULL DEFAULT '',
  discord_id varchar(18) NOT NULL DEFAULT '',
  timestamp datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,
  processed tinyint(1) NOT NULL DEFAULT '0',
  PRIMARY KEY (id)
)
`)

	db.MustExec(`
CREATE TABLE form_international (
  id int(11) NOT NULL AUTO_INCREMENT,
  first_name varchar(64) NOT NULL,
  last_name varchar(64) NOT NULL,
  email varchar(64) NOT NULL,
  phone varchar(64) NOT NULL,
  interest varchar(64) NOT NULL DEFAULT '',
  skills varchar(512) NOT NULL DEFAULT '',
  involvement varchar(512) NOT NULL DEFAULT '',
  city varchar(256) NOT NULL DEFAULT '',
  state varchar(256) NOT NULL DEFAULT '',
  country varchar(256) NOT NULL DEFAULT '',
  lat float(10,6) DEFAULT NULL,
  lng float(10,6) DEFAULT NULL,
  form_submitted timestamp DEFAULT CURRENT_TIMESTAMP,
  email_sent timestamp DEFAULT NULL,
  PRIMARY KEY (id)
)
`)

	db.MustExec(`
CREATE TABLE form_discord (
  id int(11) NOT NULL AUTO_INCREMENT,
  discord_id BIGINT(18) DEFAULT NULL,
  first_name varchar(64) NOT NULL,
  last_name varchar(64) NOT NULL,
  email varchar(64) NOT NULL,
  city varchar(256) NOT NULL DEFAULT '',
  state varchar(256) NOT NULL DEFAULT '',
  country varchar(256) NOT NULL DEFAULT '',
  lat float(10,6) DEFAULT NULL,
  lng float(10,6) DEFAULT NULL,
  PRIMARY KEY (id)
)
`)

	db.MustExec(`
CREATE TABLE form_international_actions (
  id int(11) NOT NULL AUTO_INCREMENT,
  chapter_id int(11) NOT NULL,
  organizer_name varchar(128) NOT NULL,
  submitted_at timestamp DEFAULT CURRENT_TIMESTAMP,
  last_action VARCHAR(10) DEFAULT '',
  needs TEXT,
  processed BOOLEAN DEFAULT FALSE,
  PRIMARY KEY (id)
)
`)

	db.MustExec(`
CREATE TABLE interactions (
  id int(11) NOT NULL AUTO_INCREMENT,
  activist_id int(11) NOT NULL,
  user_id int(11) NOT NULL,
  timestamp timestamp DEFAULT CURRENT_TIMESTAMP,
  method varchar(16) DEFAULT '',
  outcome varchar(32) DEFAULT '',
  notes TEXT,
  PRIMARY KEY (id)
)
`)

}

func newTestDB() *sqlx.DB {
	db := NewDB(config.DBTestDataSource())
	WipeDatabase(db)

	return db
}
