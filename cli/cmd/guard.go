package cmd

import (
	"errors"

	"github.com/dxe/adb/cli/internal/config"
)

// requireNotProd returns an error if the CLI is running against a production database. Call this at the top of any
// command that mutates or deletes data to prevent accidental production changes.
func requireNotProd() error {
	if config.IsProd() {
		return errors.New("this command is not allowed in production")
	}
	return nil
}
