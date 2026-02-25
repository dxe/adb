package cmd

import (
	"fmt"

	"github.com/dxe/adb/cli/internal/db"
	"github.com/dxe/adb/pkg/shared"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(resetDevUserCmd)
}

var usersCmd = &cobra.Command{
	Use:   "users",
	Short: "Commands for working with ADB users",
}

var resetDevUserCmd = &cobra.Command{
	Use:   "reset-dev-user",
	Short: "Delete and recreate the dev test user (non-production only)",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireNotProd(); err != nil {
			return err
		}

		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		tx, err := conn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		if _, err := tx.Exec(`DELETE FROM users_roles WHERE user_id = ?`, shared.DevTestUserId); err != nil {
			return fmt.Errorf("failed to delete user roles: %w", err)
		}

		if _, err := tx.Exec(`DELETE FROM adb_users WHERE id = ?`, shared.DevTestUserId); err != nil {
			return fmt.Errorf("failed to delete user: %w", err)
		}

		if _, err := tx.Exec(
			`INSERT INTO adb_users (id, email, name, disabled, chapter_id) VALUES (?, ?, 'Test User', 0, ?)`,
			shared.DevTestUserId, shared.DevTestUserEmail, shared.SFBayChapterIdDevTest,
		); err != nil {
			return fmt.Errorf("failed to create user: %w", err)
		}

		if _, err := tx.Exec(
			`INSERT INTO users_roles (user_id, role) VALUES (?, 'admin')`,
			shared.DevTestUserId,
		); err != nil {
			return fmt.Errorf("failed to assign admin role: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf("Dev test user reset (id=%d, email=%s, role=admin)\n", shared.DevTestUserId, shared.DevTestUserEmail)
		return nil
	},
}
