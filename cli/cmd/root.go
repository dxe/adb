package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "adb",
	Short: "ADB CLI — command-line tools for the Activist Database",
	Long: `ADB CLI is a command-line interface for managing and querying the Activist Database (ADB).

Database connection is configured via environment variables:
  DB_USER
  DB_PASSWORD
  DB_NAME
  DB_PROTOCOL

Set PROD=true to indicate a production environment. Commands that could cause
harm will refuse to run when PROD=true is detected.`,
}

// Execute is the entry point called by main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
