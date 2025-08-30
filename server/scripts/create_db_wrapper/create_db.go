package main

import (
	"flag"
	"fmt"
	"strings"
	"time"

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

func createCurrentDateString(day int) string {
	currentTime := time.Now().AddDate(0, -1, 0)
	currentYear := currentTime.Year()
	currentMonth := currentTime.Month()

	return fmt.Sprintf("%d-%d-%d", currentYear, currentMonth, day)
}

func createEventsDevDB() string {
	days := []int{15, 16, 17, 18, 19, 13}
	eventFormatStrings := []string{
		"(1, 'Event One', '%s', 'Working Group', '0', '0', '0', '1'),",
		"(2, 'Event Two', '%s', 'Protest', '0', '0', '0', '1'),",
		"(3, 'Event Three', '%s', 'Community', '0', '0', '0', '1'),",
		"(4, 'Event Four', '%s', 'Outreach', '0', '0', '0', '1'),",
		"(5, 'Event Five', '%s', 'Key Event', '0', '0', '0', '1'),",
		"(6, 'Event Six', '%s', 'Key Event', '0', '0', '0', '1');"}

	//assert
	if len(days) != len(eventFormatStrings) {
		panic("Fake data lengths do not match")
	}

	var eventStrings []string
	for i, d := range days {
		eventStrings = append(eventStrings, fmt.Sprintf(eventFormatStrings[i], createCurrentDateString(d)))
	}

	return strings.Join(eventStrings, "\n  ")
}

func createDevDB(name string) {
	db := model.NewDB(config.DataSourceBase + "/" + name + "?multiStatements=true")
	defer db.Close()
	model.WipeDatabaseWithDb(db, config.DBMigrationsLocation())
	insertStatement := fmt.Sprintf(`
INSERT INTO activists
  (id, name, email, phone, location, activist_level)
  VALUES
  (1, 'Adam Kol', 'test@directactioneverywhere.com', '7035558484', 'Berkeley, United States', 'Supporter'),
  (2, 'Robin Houseman', 'testtest@gmail.com', '7035558484', 'United States', 'Supporter'),
  (3, 'aaa', 'test@comcast.net', '7035558484', 'Fairfield, United States', 'Supporter'),
  (4, 'bbb', '', '7035558484', 'Fairfield, United States', 'Supporter'),
  (5, 'ccc', 'test@comcast.net', '7035558484', 'Fairfield, United States', 'Supporter'),
  (100, 'ddd', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (101, 'eee', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (102, 'fff', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (103, 'ggg', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (104, 'hhh', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (105, 'iii', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (106, 'jjj', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (107, 'lll', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (108, 'mmm', 'test@gmail.com', '', 'United States', 'Supporter');

INSERT INTO events VALUES
  %s


INSERT INTO event_attendance (activist_id, event_id) VALUES
  (1, 1), (1, 2), (2, 2), (3,3), (4,6), (5,5), (5,1), (5,6);

INSERT INTO adb_users (id, email, name, admin, disabled, chapter_id) VALUES
  (1, '`+model.DevTestUserEmail+`', 'Test User', 1, 0, `+model.SFBayChapterIdDevTestStr+`),
  (2, 'cwbailey20042@gmail.com', 'Cameron Bailey', 1, 0, 1),
  (3, 'jakehobbs@gmail.com', 'Jake Hobbs', 1, 0, 1),
  (4, 'samer@directactioneverywhere.com', 'The Real Samer', 1, 0, 0),
  (5, 'jake@directactioneverywhere.com', 'The Real Jake Hobbs', 1, 0, 3),
  (6, '%s', 'Dev User', 1, 0, 1);

INSERT INTO users_roles (user_id, role)
SELECT id, 'admin' FROM adb_users WHERE id IN(1, 2, 3, 4, 5, 6);

INSERT INTO fb_pages (id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng, token, organizers) VALUES
(1, '`+model.SFBayChapterName+`', 'z', 'facebook.com/a', 'twitter.com/a', 'instagram.com/a', 'a@dxe.io', 'North America', '1.000', '2.000', 'xyz', "[]"),
(2, 'Chapter B', 'z', 'facebook.com/b', 'twitter.com/b', 'instagram.com/b', 'b@dxe.io', 'North America', '3.000', '2.000', '', "[]"),
(3, 'Chapter C', 'z', 'facebook.com/c', 'twitter.com/c', 'instagram.com/c', 'c@dxe.io', 'North America', '7.000', '1.000', '', "[]");

`, createEventsDevDB(), devEmail)
	if !noFakeData {
		// Insert sample data
		db.MustExec(insertStatement)
	}
}

func main() {
	fmt.Println("Creating adb_db")
	createDevDB("adb_db")
	fmt.Println("Creating adb_test_db")
	createDevDB("adb_test_db")
}
