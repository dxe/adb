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

	db.MustExec(`
CREATE TABLE IF NOT EXISTS activists (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  email VARCHAR(80) NOT NULL DEFAULT '',
  chapter VARCHAR(80) NOT NULL DEFAULT '',
  phone varchar(20) NOT NULL DEFAULT '',
  location TEXT,
  facebook VARCHAR(80) NOT NULL DEFAULT '',
  activist_level VARCHAR(40) NOT NULL DEFAULT 'activist',
  exclude_from_leaderboard TINYINT(1) NOT NULL DEFAULT '0',
  core_staff TINYINT(1) NOT NULL DEFAULT '0',
  global_team_member TINYINT(1) NOT NULL DEFAULT '0',
  liberation_pledge TINYINT(1) NOT NULL DEFAULT '0',
  hidden TINYINT(1) NOT NULL DEFAULT '0',
  connector VARCHAR(100) NOT NULL DEFAULT '',
  contacted_date VARCHAR(20) NOT NULL DEFAULT '',
  interested VARCHAR(5) NOT NULL DEFAULT '',
  meeting_date VARCHAR(20) NOT NULL DEFAULT '',
  escalation VARCHAR(5) NOT NULL DEFAULT '',
  core_training TINYINT(1) NOT NULL DEFAULT '0',
  eligible_senior_organizer TINYINT(1) NOT NULL DEFAULT '0',
  eligible_organizer TINYINT(1) NOT NULL DEFAULT '0',
  source VARCHAR(255) NOT NULL DEFAULT '',
  interview_organizer VARCHAR(20) NOT NULL DEFAULT '',
  interview_senior_organizer VARCHAR(20) NOT NULL DEFAULT '',
  action_team_focus VARCHAR(40) NOT NULL DEFAULT '',
  doing_work TINYINT(1) NOT NULL DEFAULT '0',

  CONSTRAINT name_ukey UNIQUE (name)
)`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS event_attendance (
  activist_id INTEGER NOT NULL,
  event_id INTEGER NOT NULL,
  CONSTRAINT activist_event_ukey UNIQUE (activist_id, event_id)
)`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(60) NOT NULL,
  date DATE NOT NULL,
  event_type VARCHAR(60) NOT NULL,
  FULLTEXT INDEX name_idx (name)
)`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS adb_users (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  email VARCHAR(60) NOT NULL,
  admin TINYINT(1) NOT NULL DEFAULT '0',
  disabled TINYINT(1) NOT NULL DEFAULT '0'
)
`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS merged_activist_attendance (
  original_activist_id INTEGER NOT NULL,
  target_activist_id INTEGER NOT NULL,
  event_id INTEGER NOT NULL,
  replaced_with_target_activist TINYINT(1) NOT NULL,
  CONSTRAINT merged_activist_attendance_ukey UNIQUE (original_activist_id, target_activist_id, event_id)
)
`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS working_groups (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(60) NOT NULL,
  type TINYINT(1) NOT NULL,
  group_email VARCHAR(100) NOT NULL,
  CONSTRAINT working_groups_name_ukey UNIQUE (name)
)
`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS working_group_members (
  working_group_id INTEGER NOT NULL,
  activist_id INTEGER NOT NULL,
  -- True if the activist is the point person of the working group.
  -- There should be only one point person per working group, but we
  -- don't restrict that on the backend.
  point_person TINYINT NOT NULL DEFAULT '0',
  -- Some activists need to be on the mailing list even though they
  -- aren't in the workin group.
  non_member_on_mailing_list TINYINT NOT NULL DEFAULT '0',
  CONSTRAINT working_group_member_ukey UNIQUE (working_group_id, activist_id)
)
`)

}

func newTestDB() *sqlx.DB {
	db := NewDB(config.DBTestDataSource())
	WipeDatabase(db)

	return db
}
