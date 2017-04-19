package model

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

func NewDB(dataSourceName string) *sqlx.DB {
	db, err := sqlx.Open("mysql", dataSourceName)
	if err != nil {
		panic(err)
	}
	CreateDatabase(db)
	return db
}

func CreateDatabase(db *sqlx.DB) {
	db.MustExec(`
CREATE TABLE IF NOT EXISTS activists (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(80) NOT NULL,
  email VARCHAR(80) NOT NULL DEFAULT '',
  chapter_id INTEGER DEFAULT NULL,
  phone varchar(20) NOT NULL DEFAULT '',
  location TEXT,
  facebook VARCHAR(80) NOT NULL DEFAULT '',
  exclude_from_leaderboard TINYINT(1) NOT NULL DEFAULT '0',
  core_staff TINYINT(1) NOT NULL DEFAULT '0',
  global_team_member TINYINT(1) NOT NULL DEFAULT '0',
  liberation_pledge TINYINT(1) DEFAULT NULL
)`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS event_attendance (
  activist_id INTEGER NOT NULL,
  event_id INTEGER NOT NULL
)`)

	db.MustExec(`
CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(60) NOT NULL,
  date DATE NOT NULL,
  event_type VARCHAR(60) NOT NULL
)`)

  db.MustExec(`
CREATE TABLE IF NOT EXISTS chapters (
  id INTEGER PRIMARY KEY AUTO_INCREMENT,
  name VARCHAR(60) NOT NULL
)`)


}
