package main

import "github.com/directactioneverywhere/adb/model"

func main() {
	db := model.NewDB("adb_user:adbpassword@/adb_db")
	db.MustExec(`DROP TABLE IF EXISTS activists`)
	db.MustExec(`DROP TABLE IF EXISTS events`)
	db.MustExec(`DROP TABLE IF EXISTS event_attendance`)
	model.CreateDatabase(db)

	// Insert sample data
	db.MustExec(`
INSERT INTO activists VALUES
  (1, 'Adam Kol', 'adam@directactioneverywhere.com', 2, '9542635719', 'Berkeley, United States', '', 0, 0, 1, 1),
  (2, 'Robin Houseman', 'testtest@gmail.com', 2, '4398943', 'United States', '', 0, 0, 0, 0),
  (3, 'Jake Hong', 'test@comcast.net', 2, '7077206366', 'Fairfield, United States', '', 0, 0, 0, 0)
`)
	db.MustExec(`
INSERT INTO activists VALUES
  (4, 'Samer Samer', 'test.test.test@gmail.com', 2, '', 'United States', '', 0, 0, 0, 0),
  (5, 'Samer Masterson', 'alexis.l.levitt@gmail.com', 2, '', 'United States', '', 0, 0, 0, 0)
`)

	db.MustExec(`
INSERT INTO events VALUES
  (1, 'Event One', '2017-02-15', 'Working Group'),
  (2, 'Event Two', '2017-02-16', 'Protest')
`)

	db.MustExec(`
INSERT INTO event_attendance (activist_id, event_id) VALUES
  (1, 1),
  (2, 1),
  (3, 1),
  (4, 2),
  (5, 2)
`)
}
