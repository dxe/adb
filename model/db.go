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
	db.MustExec(`DROP TABLE IF EXISTS events`)
	db.MustExec(`DROP TABLE IF EXISTS event_attendance`)
	db.MustExec(`DROP TABLE IF EXISTS adb_users`)
	db.MustExec(`DROP TABLE IF EXISTS merged_activist_attendance`)
	db.MustExec(`DROP TABLE IF EXISTS working_groups`)
	db.MustExec(`DROP TABLE IF EXISTS working_group_members`)
	db.MustExec(`DROP TABLE IF EXISTS circles`)
	db.MustExec(`DROP TABLE IF EXISTS circle_members`)
  db.MustExec(`DROP TABLE IF EXISTS users_roles`)

	db.MustExec(`
CREATE TABLE activists (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  email VARCHAR(80) NOT NULL DEFAULT '',
  phone VARCHAR(20) NOT NULL DEFAULT '',
  location VARCHAR(200) DEFAULT '',
  facebook VARCHAR(200) NOT NULL DEFAULT '',
  activist_level VARCHAR(40) NOT NULL DEFAULT 'Supporter',
  hidden TINYINT(1) NOT NULL DEFAULT '0',
  connector VARCHAR(100) NOT NULL DEFAULT '',
  contacted_date VARCHAR(20) NOT NULL DEFAULT '',
  interested VARCHAR(5) NOT NULL DEFAULT '',
  meeting_date VARCHAR(20) NOT NULL DEFAULT '',
  escalation VARCHAR(5) NOT NULL DEFAULT '',
  source VARCHAR(255) NOT NULL DEFAULT '',
  date_organizer DATE,
  date_senior_organizer DATE,
  dob TEXT,
  training0 VARCHAR(20),
  training1 VARCHAR(20),
  training2 VARCHAR(20),
  training3 VARCHAR(20),
  training4 VARCHAR(20),
  training5 VARCHAR(20),
  training6 VARCHAR(20),
  prospect_organizer TINYINT(1) NOT NULL DEFAULT '0',
  prospect_chapter_member TINYINT NOT NULL DEFAULT '0',
  prospect_circle_member TINYINT NOT NULL DEFAULT '0',
  dev_manager VARCHAR(100) NOT NULL DEFAULT '',
  dev_interest VARCHAR(200) NOT NULL DEFAULT '',
  dev_auth VARCHAR(20),
  dev_email_sent VARCHAR(20),
  dev_vetted TINYINT(1) NOT NULL DEFAULT '0',
  dev_interview VARCHAR(20),
  dev_onboarding TINYINT(1) NOT NULL DEFAULT '0',
  dev_application_date DATE,
  cm_first_email VARCHAR(20),
  cm_approval_email VARCHAR(20),
  cm_warning_email VARCHAR(20),
  cir_first_email VARCHAR(20),
  UNIQUE (name)
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
  INDEX (date, name),
  FULLTEXT (name)
)
`)

	db.MustExec(`
CREATE TABLE adb_users (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(60) NOT NULL,
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
    ON UPDATE NO ACTION);
`)

}

func newTestDB() *sqlx.DB {
	db := NewDB(config.DBTestDataSource())
	WipeDatabase(db)

	return db
}
