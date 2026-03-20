package cmd

import (
	"fmt"
	"strings"

	"github.com/dxe/adb/cli/internal/db"
	"github.com/dxe/adb/pkg/shared"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(usersCmd)
	usersCmd.AddCommand(resetDevUserCmd)
	usersCmd.AddCommand(devUserSetRolesCmd)
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
			`INSERT INTO users_roles (user_id, role) VALUES (?, ?)`,
			shared.DevTestUserId,
			shared.RoleAdmin,
		); err != nil {
			return fmt.Errorf("failed to assign admin role: %w", err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf(
			"Dev test user reset (id=%d, email=%s, role=%s)\n",
			shared.DevTestUserId,
			shared.DevTestUserEmail,
			shared.RoleAdmin,
		)
		return nil
	},
}

var devUserSetRolesCmd = &cobra.Command{
	Use:   "dev-user-set-roles <roles>",
	Short: "Replace the dev test user's roles (non-production only)",
	Long: `Replace the dev test user's roles.

Roles may be provided as separate arguments, a comma-separated list, or a mix of both.`,
	Args: cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireNotProd(); err != nil {
			return err
		}

		roles, err := parseDevUserRoles(args)
		if err != nil {
			return err
		}

		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		tx, err := conn.Beginx()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		var userCount int
		if err := tx.Get(&userCount, `SELECT COUNT(*) FROM adb_users WHERE id = ?`, shared.DevTestUserId); err != nil {
			return fmt.Errorf("failed to look up dev test user: %w", err)
		}
		if userCount == 0 {
			return fmt.Errorf(
				"dev test user does not exist (id=%d, email=%s); run `adb users reset-dev-user` first",
				shared.DevTestUserId,
				shared.DevTestUserEmail,
			)
		}

		if _, err := tx.Exec(`DELETE FROM users_roles WHERE user_id = ?`, shared.DevTestUserId); err != nil {
			return fmt.Errorf("failed to delete existing user roles: %w", err)
		}

		for _, role := range roles {
			if _, err := tx.Exec(`INSERT INTO users_roles (user_id, role) VALUES (?, ?)`, shared.DevTestUserId, role); err != nil {
				return fmt.Errorf("failed to assign role %q: %w", role, err)
			}
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		fmt.Printf(
			"Dev test user roles set (id=%d, email=%s, roles=%s)\n",
			shared.DevTestUserId,
			shared.DevTestUserEmail,
			strings.Join(roles, ","),
		)
		return nil
	},
}

func parseDevUserRoles(args []string) ([]string, error) {
	roles := make([]string, 0, len(args))
	seen := make(map[string]struct{}, len(args))

	for _, arg := range args {
		for _, part := range strings.Split(arg, ",") {
			role := strings.TrimSpace(part)
			if role == "" {
				continue
			}
			if !shared.IsAllowedADBUserRole(role) {
				return nil, fmt.Errorf("invalid role %q (allowed: %s)", role, strings.Join(shared.AllowedADBUserRoles(), ", "))
			}
			if _, ok := seen[role]; ok {
				continue
			}
			seen[role] = struct{}{}
			roles = append(roles, role)
		}
	}

	if len(roles) == 0 {
		return nil, fmt.Errorf("no roles provided")
	}

	return roles, nil
}
