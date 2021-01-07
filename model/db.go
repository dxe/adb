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

	db.MustExec(`
CREATE TABLE activists (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  preferred_name VARCHAR(80) NOT NULL DEFAULT '',
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
  training0 VARCHAR(20),
  training1 VARCHAR(20),
  training4 VARCHAR(20),
  training5 VARCHAR(20),
  training6 VARCHAR(20),
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
  UNIQUE (name)
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
  PRIMARY KEY (activist_id, revision)
  ) ENGINE=MyISAM
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
  type TINYINT(1) NOT NULL,
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
  eventbrite_token VARCHAR(32) NOT NULL DEFAULT ''
)
`)

	db.MustExec(`
CREATE TABLE fb_events (
  id BIGINT NOT NULL,
  page_id BIGINT NOT NULL,
  name VARCHAR(64) NOT NULL,
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
  cover VARCHAR(300),
  attending_count MEDIUMINT NOT NULL DEFAULT '0',
  interested_count MEDIUMINT NOT NULL DEFAULT '0',
  is_canceled TINYINT NOT NULL DEFAULT '0',
  last_update TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  PRIMARY KEY (id, page_id)
)
`)

	db.MustExec(`
CREATE TABLE discord_users (
  id BIGINT(18) PRIMARY KEY,
  email VARCHAR(200) NOT NULL,
  token VARCHAR(64) NOT NULL,
  confirmed TINYINT(1) NOT NULL DEFAULT '0'
)
`)

}

func newTestDB() *sqlx.DB {
	db := NewDB(config.DBTestDataSource())
	WipeDatabase(db)

	return db
}
