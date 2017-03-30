package model

import "github.com/jmoiron/sqlx"

func NewDB(databaseName string) *sqlx.DB {
	db, err := sqlx.Open("sqlite3", databaseName)
	if err != nil {
		panic(err)
	}
	CreateDatabase(db)
	return db
}

func CreateDatabase(db *sqlx.DB) {
	db.MustExec(`
CREATE TABLE IF NOT EXISTS activists (
  id INTEGER PRIMARY KEY,
  name varchar(80) NOT NULL,
  email varchar(80) NOT NULL DEFAULT '',
  chapter_id int(3) DEFAULT NULL,
  phone varchar(20) NOT NULL DEFAULT '',
  city varchar(40) NOT NULL DEFAULT '',
  zipcode varchar(15) NOT NULL DEFAULT '',
  country varchar(80) NOT NULL DEFAULT '',
  facebook varchar(80) NOT NULL DEFAULT '',
  exclude_from_leaderboard tinyint(1) NOT NULL DEFAULT '0',
  core_staff tinyint(1) NOT NULL DEFAULT '0',
  global_team_member tinyint(1) NOT NULL DEFAULT '0',
  liberation_pledge tinyint(1) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS event_attendance (
  activist_id int(8) NOT NULL,
  event_id int(8) NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY,
  name varchar(60) NOT NULL,
  date date NOT NULL,
  type varchar(60) NOT NULL
);

`)
}
