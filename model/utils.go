package model

import (
	"strings"

	"github.com/pkg/errors"
)

/** Constant and Global Variable Definitions */

const DangerousCharacters = "<>&"

/** Functions and Methods */

func checkForDangerousChars(data string) error {
	if strings.ContainsAny(data, DangerousCharacters) {
		return errors.New("User input cannot include <, >, or &.")
	}
	return nil
}
