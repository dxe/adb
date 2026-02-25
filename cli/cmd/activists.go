package cmd

import (
	"fmt"

	"github.com/dxe/adb/cli/internal/db"
	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(activistsCmd)
	activistsCmd.AddCommand(activistsCountCmd)
}

var activistsCmd = &cobra.Command{
	Use:   "activists",
	Short: "Commands for working with activists",
}

var activistsCountCmd = &cobra.Command{
	Use:   "count",
	Short: "Print the total number of activists in the database",
	RunE: func(cmd *cobra.Command, args []string) error {
		conn, err := db.Connect()
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		defer conn.Close()

		var count int
		if err := conn.Get(&count, "SELECT COUNT(*) FROM activists"); err != nil {
			return fmt.Errorf("failed to query activists: %w", err)
		}

		fmt.Printf("Activist count: %d\n", count)
		return nil
	},
}
