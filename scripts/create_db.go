package main

import (
	"flag"
	"fmt"

	"github.com/dxe/adb/config"
	"github.com/dxe/adb/model"
)

var noFakeData bool
var devEmail string

func init() {
	noFake := flag.Bool("no-fake-data", false, "Don't pre-populate with data")
	devEmailPtr := flag.String("dev-email", "test-dev@directactioneverywhere.com", "The developer email used to login for local development")
	flag.Parse()

	noFakeData = *noFake
	devEmail = *devEmailPtr
}

func createDevDB(name string) {
	db := model.NewDB(config.DBUser + ":" + config.DBPassword + "@/" + name + "?multiStatements=true")
	defer db.Close()
	model.WipeDatabase(db)
	insertStatement := fmt.Sprintf(`
INSERT INTO activists
  (id, name, email, chapter, phone, location, facebook, activist_level, exclude_from_leaderboard, core_staff, global_team_member, liberation_pledge)
  VALUES
  (1, 'Adam Kol', 'test@directactioneverywhere.com', 'Sf Bay', '7035558484', 'Berkeley, United States', '', 'activist', 0, 0, 1, 1),
  (2, 'Robin Houseman', 'testtest@gmail.com', 'SF Bay', '7035558484', 'United States','', 'activist', 0, 0, 0, 0),
  (3, 'aaa', 'test@comcast.net', 'SF Bay', '7035558484', 'Fairfield, United States', '', 'activist', 0, 0, 0, 0),
  (4, 'bbb', 'test@comcast.net', 'SF Bay', '7035558484', 'Fairfield, United States', '', 'activist', 0, 0, 0, 0),
  (5, 'ccc', 'test@comcast.net', 'SF Bay', '7035558484', 'Fairfield, United States', '', 'activist', 0, 0, 0, 0),
  (100, 'ddd', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
(101, 'eee', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
(102, 'fff', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
(103, 'ggg', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
(104, 'hhh', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
(105, 'iii', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
(106, 'jjj', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
(107, 'lll', 'test.test.test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0),
  (108, 'mmm', 'test@gmail.com', 'SF Bay', '', 'United States', '', 'activist', 0, 0, 0, 0);

INSERT INTO events VALUES
  (1, 'Event One', '2016-07-15', 'Working Group'),
  (2, 'Event Two', '2017-07-16', 'Protest'),
  (3, 'Event Three', '2017-07-17', 'Community'),
  (4, 'Event Four', '2017-07-18', 'Outreach'),
  (5, 'Event Five', '2017-07-19', 'Key Event'),
  (6, 'Event Six', '2017-07-13', 'Key Event');


INSERT INTO event_attendance (activist_id, event_id) VALUES
  (1, 1), (1, 2), (2, 2), (3,3), (4,6), (5,5), (5,1), (5,6);

INSERT INTO adb_users (id, email, admin, disabled) VALUES
  (1, 'nosefrog@gmail.com', 1, 0),
  (2, 'cwbailey20042@gmail.com', 1, 0),
  (3, 'jakehobbs@gmail.com', 1, 0),
  (4, 'samer@directactioneverywhere.com', 1, 0),
  (5, 'jake@directactioneverywhere.com', 1, 0),
  (6, '%s', 1, 0);
`, devEmail)
	if !noFakeData {
		// Insert sample data
		db.MustExec(insertStatement)
	}
}

func main() {
	createDevDB("adb_db")
	createDevDB("adb_test_db")
}
