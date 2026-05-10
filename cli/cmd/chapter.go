package cmd

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"

	"github.com/dxe/adb/cli/internal/db"
	"github.com/jmoiron/sqlx"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(chapterCmd)
	chapterCmd.AddCommand(chapterListCmd)
	chapterCmd.AddCommand(chapterGetFacebookCmd)
	chapterCmd.AddCommand(chapterSetFacebookCmd)
	chapterCmd.AddCommand(chapterGetEventbriteCmd)
	chapterCmd.AddCommand(chapterSetEventbriteCmd)
	chapterCmd.AddCommand(chapterGetMailingListCmd)
	chapterCmd.AddCommand(chapterSetMailingListCmd)
	chapterCmd.AddCommand(chapterGetMailingListRadiusCmd)
	chapterCmd.AddCommand(chapterSetMailingListRadiusCmd)
}

func promptField(prompt string) (string, error) {
	fmt.Fprint(os.Stderr, prompt)
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return "", fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(line), nil
}

func chapterName(conn *sqlx.DB, chapterID int) (string, error) {
	var name string
	if err := conn.Get(&name, `SELECT name FROM fb_pages WHERE chapter_id = ?`, chapterID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", fmt.Errorf("chapter %d not found", chapterID)
		}
		return "", fmt.Errorf("failed to query chapter %d: %w", chapterID, err)
	}
	return name, nil
}

// confirmChanges prints the chapter, the proposed field changes, and asks y/N.
func confirmChanges(chapterID int, name string, pairs [][2]string) (bool, error) {
	fmt.Printf("Chapter: %s (id=%d)\n", name, chapterID)
	for _, p := range pairs {
		fmt.Printf("  %s = %s\n", p[0], p[1])
	}
	fmt.Print("Apply changes? [y/N] ")
	r := bufio.NewReader(os.Stdin)
	line, err := r.ReadString('\n')
	if err != nil {
		return false, fmt.Errorf("failed to read input: %w", err)
	}
	return strings.TrimSpace(strings.ToLower(line)) == "y", nil
}

func parseChapterID(s string) (int, error) {
	id, err := strconv.Atoi(s)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid chapter ID %q: must be a positive integer", s)
	}
	return id, nil
}

var chapterCmd = &cobra.Command{
	Use:   "chapters",
	Short: "Commands for working with chapters",
}

var chapterListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all chapters with their IDs and names",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		var chapters []struct {
			ChapterID int    `db:"chapter_id"`
			Name      string `db:"name"`
		}
		if err := conn.Select(&chapters, `SELECT chapter_id, name FROM fb_pages ORDER BY name`); err != nil {
			return fmt.Errorf("failed to query chapters: %w", err)
		}
		for _, c := range chapters {
			fmt.Printf("%d\t%s\n", c.ChapterID, c.Name)
		}
		return nil
	},
}

var chapterGetFacebookCmd = &cobra.Command{
	Use:   "get-facebook <chapter_id>",
	Short: "Print the Facebook page ID and token for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}
		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		var result struct {
			ID    int64  `db:"id"`
			Token string `db:"token"`
		}
		if err := conn.Get(&result, `SELECT id, token FROM fb_pages WHERE chapter_id = ?`, chapterID); err != nil {
			return fmt.Errorf("failed to query chapter %d: %w", chapterID, err)
		}
		fmt.Printf("facebook_id=%d\ntoken=%s\n", result.ID, result.Token)
		return nil
	},
}

var chapterSetFacebookCmd = &cobra.Command{
	Use:   "set-facebook <chapter_id>",
	Short: "Set the Facebook page ID and token for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}

		fbIDStr, err := promptField("Facebook page ID: ")
		if err != nil {
			return err
		}
		fbID, err := strconv.ParseInt(fbIDStr, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid facebook_id %q: must be an integer", fbIDStr)
		}
		token, err := promptField("Facebook token: ")
		if err != nil {
			return err
		}

		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		name, err := chapterName(conn, chapterID)
		if err != nil {
			return err
		}
		ok, err := confirmChanges(chapterID, name, [][2]string{
			{"facebook_id", fmt.Sprintf("%d", fbID)},
			{"token", token},
		})
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}

		if _, err := conn.Exec(
			`UPDATE fb_pages SET id = ?, token = ? WHERE chapter_id = ?`,
			fbID, token, chapterID,
		); err != nil {
			return fmt.Errorf("failed to update chapter %d: %w", chapterID, err)
		}
		fmt.Printf("chapter %d updated: facebook_id=%d\n", chapterID, fbID)
		return nil
	},
}

var chapterGetEventbriteCmd = &cobra.Command{
	Use:   "get-eventbrite <chapter_id>",
	Short: "Print the Eventbrite ID and token for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}
		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		var result struct {
			EventbriteID    string `db:"eventbrite_id"`
			EventbriteToken string `db:"eventbrite_token"`
		}
		if err := conn.Get(&result, `SELECT eventbrite_id, eventbrite_token FROM fb_pages WHERE chapter_id = ?`, chapterID); err != nil {
			return fmt.Errorf("failed to query chapter %d: %w", chapterID, err)
		}
		fmt.Printf("eventbrite_id=%s\neventbrite_token=%s\n", result.EventbriteID, result.EventbriteToken)
		return nil
	},
}

var chapterSetEventbriteCmd = &cobra.Command{
	Use:   "set-eventbrite <chapter_id>",
	Short: "Set the Eventbrite ID and token for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}

		eventbriteID, err := promptField("Eventbrite ID: ")
		if err != nil {
			return err
		}
		token, err := promptField("Eventbrite token: ")
		if err != nil {
			return err
		}

		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		name, err := chapterName(conn, chapterID)
		if err != nil {
			return err
		}
		ok, err := confirmChanges(chapterID, name, [][2]string{
			{"eventbrite_id", eventbriteID},
			{"eventbrite_token", token},
		})
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}

		if _, err := conn.Exec(
			`UPDATE fb_pages SET eventbrite_id = ?, eventbrite_token = ? WHERE chapter_id = ?`,
			eventbriteID, token, chapterID,
		); err != nil {
			return fmt.Errorf("failed to update chapter %d: %w", chapterID, err)
		}
		fmt.Printf("chapter %d updated: eventbrite_id=%s\n", chapterID, eventbriteID)
		return nil
	},
}

var chapterGetMailingListCmd = &cobra.Command{
	Use:   "get-mailing-list <chapter_id>",
	Short: "Print the mailing list type and ID for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}
		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		var result struct {
			MLType string `db:"ml_type"`
			MLID   string `db:"ml_id"`
		}
		if err := conn.Get(&result, `SELECT ml_type, ml_id FROM fb_pages WHERE chapter_id = ?`, chapterID); err != nil {
			return fmt.Errorf("failed to query chapter %d: %w", chapterID, err)
		}
		fmt.Printf("ml_type=%s\nml_id=%s\n", result.MLType, result.MLID)
		return nil
	},
}

var chapterSetMailingListCmd = &cobra.Command{
	Use:   "set-mailing-list <chapter_id>",
	Short: "Set the mailing list type and ID for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}

		mlType, err := promptField(`Mailing list type ("Sendy", "SendGrid", "Google Groups", ""): `)
		if err != nil {
			return err
		}
		if !slices.Contains([]string{"", "Sendy", "SendGrid", "Google Groups"}, mlType) {
			return fmt.Errorf("invalid mailing list type %q", mlType)
		}
		mlID, err := promptField("Mailing list ID: ")
		if err != nil {
			return err
		}

		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		name, err := chapterName(conn, chapterID)
		if err != nil {
			return err
		}
		ok, err := confirmChanges(chapterID, name, [][2]string{
			{"ml_type", mlType},
			{"ml_id", mlID},
		})
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}

		if _, err := conn.Exec(
			`UPDATE fb_pages SET ml_type = ?, ml_id = ? WHERE chapter_id = ?`,
			mlType, mlID, chapterID,
		); err != nil {
			return fmt.Errorf("failed to update chapter %d: %w", chapterID, err)
		}
		fmt.Printf("chapter %d updated: ml_type=%s ml_id=%s\n", chapterID, mlType, mlID)
		return nil
	},
}

var chapterGetMailingListRadiusCmd = &cobra.Command{
	Use:   "get-mailing-list-radius <chapter_id>",
	Short: "Print the mailing list radius for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}
		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		var radius int
		if err := conn.Get(&radius, `SELECT ml_radius FROM fb_pages WHERE chapter_id = ?`, chapterID); err != nil {
			return fmt.Errorf("failed to query chapter %d: %w", chapterID, err)
		}
		fmt.Printf("ml_radius=%d\n", radius)
		return nil
	},
}

var chapterSetMailingListRadiusCmd = &cobra.Command{
	Use:   "set-mailing-list-radius <chapter_id>",
	Short: "Set the mailing list radius for a chapter",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		chapterID, err := parseChapterID(args[0])
		if err != nil {
			return err
		}

		radiusStr, err := promptField("Mailing list radius: ")
		if err != nil {
			return err
		}
		radius, err := strconv.Atoi(radiusStr)
		if err != nil || radius < 0 {
			return fmt.Errorf("invalid radius %q: must be a non-negative integer", radiusStr)
		}

		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		name, err := chapterName(conn, chapterID)
		if err != nil {
			return err
		}
		ok, err := confirmChanges(chapterID, name, [][2]string{
			{"ml_radius", fmt.Sprintf("%d", radius)},
		})
		if err != nil {
			return err
		}
		if !ok {
			fmt.Println("Cancelled.")
			return nil
		}

		if _, err := conn.Exec(
			`UPDATE fb_pages SET ml_radius = ? WHERE chapter_id = ?`,
			radius, chapterID,
		); err != nil {
			return fmt.Errorf("failed to update chapter %d: %w", chapterID, err)
		}
		fmt.Printf("chapter %d updated: ml_radius=%d\n", chapterID, radius)
		return nil
	},
}
