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
  name VARCHAR(80) NOT NULL,
  email VARCHAR(80) NOT NULL DEFAULT '',
  chapter_id INTEGER DEFAULT NULL,
  phone varchar(20) NOT NULL DEFAULT '',
  location TEXT NOT NULL DEFAULT '',
  facebook VARCHAR(80) NOT NULL DEFAULT '',
  exclude_from_leaderboard TINYINT(1) NOT NULL DEFAULT '0',
  core_staff TINYINT(1) NOT NULL DEFAULT '0',
  global_team_member TINYINT(1) NOT NULL DEFAULT '0',
  liberation_pledge TINYINT(1) DEFAULT NULL
);

CREATE TABLE IF NOT EXISTS event_attendance (
  activist_id INTEGER NOT NULL,
  event_id INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS events (
  id INTEGER PRIMARY KEY,
  name VARCHAR(60) NOT NULL,
  date DATE NOT NULL,
  event_type VARCHAR(60) NOT NULL
);
`)
}
