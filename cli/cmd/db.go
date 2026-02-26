package cmd

import (
	"fmt"
	"strings"
	"time"

	"github.com/dxe/adb/cli/internal/config"
	"github.com/dxe/adb/pkg/shared"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"

	// MySQL driver — imported for its side-effect of registering the driver.
	_ "github.com/go-sql-driver/mysql"
)

var dbCreateNoFakeData bool
var dbCreateDevEmail string

func init() {
	rootCmd.AddCommand(dbCmd)
	dbCmd.AddCommand(dbCreateCmd)
	dbCreateCmd.Flags().BoolVar(&dbCreateNoFakeData, "no-fake-data", false, "Skip inserting fake dev data")
	dbCreateCmd.Flags().StringVar(&dbCreateDevEmail, "dev-email", "test-dev@directactioneverywhere.com", "Developer email used to log in for local development")
}

var dbCmd = &cobra.Command{
	Use:   "db",
	Short: "Commands for managing the ADB database",
}

var dbCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create and migrate adb_db and adb_test_db (non-production only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireNotProd(); err != nil {
			return err
		}

		for _, name := range []string{"adb_db", "adb_test_db"} {
			fmt.Printf("Creating %s\n", name)
			dsn := config.DBDataSourceForDB(name) + "&multiStatements=true"
			shared.WipeDatabase(dsn, false)

			if !dbCreateNoFakeData {
				conn, err := sqlx.Connect("mysql", config.DBDataSourceForDB(name)+"&multiStatements=true")
				if err != nil {
					return fmt.Errorf("failed to connect to %s: %w", name, err)
				}
				conn.MustExec(buildFakeDataSQL(dbCreateDevEmail))
				conn.Close()
			}
		}

		return nil
	},
}

func currentDateString(day int) string {
	t := time.Now().AddDate(0, -1, 0)
	return fmt.Sprintf("%02d-%02d-%02d", t.Year(), t.Month(), day)
}

func buildEventsSQL() string {
	days := []int{15, 16, 17, 18, 19, 13}
	formats := []string{
		"(1, 'Event One', '%s', 'Working Group', '0', '0', '0', '1'),",
		"(2, 'Event Two', '%s', 'Protest', '0', '0', '0', '1'),",
		"(3, 'Event Three', '%s', 'Community', '0', '0', '0', '1'),",
		"(4, 'Event Four', '%s', 'Outreach', '0', '0', '0', '1'),",
		"(5, 'Event Five', '%s', 'Key Event', '0', '0', '0', '1'),",
		"(6, 'Event Six', '%s', 'Key Event', '0', '0', '0', '1');",
	}
	rows := make([]string, len(days))
	for i, d := range days {
		rows[i] = fmt.Sprintf(formats[i], currentDateString(d))
	}
	return strings.Join(rows, "\n  ")
}

func buildFakeDataSQL(devEmail string) string {
	ch := shared.SFBayChapterIdDevTestStr
	return fmt.Sprintf(`
INSERT INTO activists
  (chapter_id, id, name, email, phone, location, activist_level)
  VALUES
  (`+ch+`, 1, 'Adam Kol', 'test@directactioneverywhere.com', '7035558484', 'Berkeley, United States', 'Supporter'),
  (`+ch+`, 2, 'Robin Houseman', 'testtest@gmail.com', '7035558484', 'United States', 'Supporter'),
  (`+ch+`, 3, 'aaa', 'test@comcast.net', '7035558484', 'Fairfield, United States', 'Supporter'),
  (`+ch+`, 4, 'bbb', '', '7035558484', 'Fairfield, United States', 'Supporter'),
  (`+ch+`, 5, 'ccc', 'test@comcast.net', '7035558484', 'Fairfield, United States', 'Supporter'),
  (`+ch+`, 100, 'ddd', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 101, 'eee', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 102, 'fff', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 103, 'ggg', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 104, 'hhh', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 105, 'iii', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 106, 'jjj', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 107, 'lll', 'test.test.test@gmail.com', '', 'United States', 'Supporter'),
  (`+ch+`, 108, 'mmm', 'test@gmail.com', '', 'United States', 'Supporter');

INSERT INTO activists
  (chapter_id, id, name, email, phone, location, activist_level, source, interest_date)
  VALUES
  (`+ch+`, 1000, 'nnn', 'test2@gmail.com', '', 'United States', 'Supporter', 'Petition: no-more-bad-things', '`+currentDateString(1)+`');

INSERT INTO events VALUES
  %s

INSERT INTO event_attendance (activist_id, event_id) VALUES
  (1, 1), (1, 2), (2, 2), (3,3), (4,6), (5,5), (5,1), (5,6);

INSERT INTO adb_users (id, email, name, disabled, chapter_id) VALUES
  (`+shared.SFBayChapterIdDevTestStr+`, '`+shared.DevTestUserEmail+`', 'Test User', 0, `+shared.SFBayChapterIdDevTestStr+`),
  (2, 'cwbailey20042@gmail.com', 'Cameron Bailey', 0, 1),
  (3, 'jakehobbs@gmail.com', 'Jake Hobbs', 0, 1),
  (4, 'samer@directactioneverywhere.com', 'The Real Samer', 0, 0),
  (5, 'jake@directactioneverywhere.com', 'The Real Jake Hobbs', 0, 3),
  (6, '%s', 'Dev User', 0, 1);

INSERT INTO users_roles (user_id, role)
SELECT id, 'admin' FROM adb_users WHERE id IN(`+shared.SFBayChapterIdDevTestStr+`, 2, 3, 4, 5, 6);

INSERT INTO fb_pages (id, name, flag, fb_url, twitter_url, insta_url, email, region, lat, lng, token, organizers) VALUES
(1, '`+shared.SFBayChapterName+`', '🇺🇸', 'facebook.com/a', 'twitter.com/a', 'instagram.com/a', 'a@dxe.io', 'North America', '1.000', '2.000', 'xyz', "[]"),
(2, 'Chapter B', '🇺🇸', 'facebook.com/b', 'twitter.com/b', 'instagram.com/b', 'b@dxe.io', 'North America', '3.000', '2.000', '', "[]"),
(3, 'Chapter C', '🇺🇸', 'facebook.com/c', 'twitter.com/c', 'instagram.com/c', 'c@dxe.io', 'North America', '7.000', '1.000', '', "[]");
`, buildEventsSQL(), devEmail)
}
