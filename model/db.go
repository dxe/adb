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
	db.MustExec(`DROP TABLE IF EXISTS activists`)
	db.MustExec(`DROP TABLE IF EXISTS events`)
	db.MustExec(`DROP TABLE IF EXISTS event_attendance`)
	db.MustExec(`DROP TABLE IF EXISTS adb_users`)

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
  suspended TINYINT(1) NOT NULL DEFAULT '0',
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
}

func newTestDB() *sqlx.DB {
	db := NewDB(config.DBTestDataSource())
	WipeDatabase(db)

	return db
}
