package main

import (
	"os"

	"github.com/directactioneverywhere/adb/model"
)

func main() {
	os.Remove("adb.db")
	db := model.NewDB("adb.db")
	model.CreateDatabase(db)

	// Insert sample data
	db.MustExec(`
INSERT INTO activists VALUES
  (1, 'Adam Kol', 'adam@directactioneverywhere.com', 2, '9542635719', 'Berkeley', '', 'United States', '', 0, 0, 1, 1);
INSERT INTO activists VALUES
  (2, 'Robin Houseman', 'testtest@gmail.com', 2, '4398943', '', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO activists VALUES
  (3, 'Jake Hong', 'test@comcast.net', 2, '7077206366', 'Fairfield', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO activists VALUES
  (4, 'Samer Samer', 'test.test.test@gmail.com', 2, '', '', '', 'United States', '', 0, 0, 0, 0);
INSERT INTO activists VALUES
  (5, 'Samer Masterson', 'alexis.l.levitt@gmail.com', 2, '', '', '', 'United States', '', 0, 0, 0, 0);

INSERT INTO events VALUES
  (1, 'Event One', datetime(1092941466, 'unixepoch'), 'Working Group');
INSERT INTO events VALUES
  (2, 'Event Two', datetime(1092941466, 'unixepoch'), 'Protest');

INSERT INTO event_attendance (activist_id, event_id) VALUES
  (1, 1);
INSERT INTO event_attendance (activist_id, event_id) VALUES
  (2, 1);
INSERT INTO event_attendance (activist_id, event_id) VALUES
  (3, 1);
INSERT INTO event_attendance (activist_id, event_id) VALUES
  (4, 2);
INSERT INTO event_attendance (activist_id, event_id) VALUES
  (5, 2);

`)
}
