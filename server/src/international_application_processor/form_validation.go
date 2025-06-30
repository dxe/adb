package international_application_processor

import (
	"net/mail"
	"strings"
	"unicode"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func sanitizeAndFormatName(name string) string {
	sanitized := selectNumbersLettersAndSpaces(name)

	return strings.TrimSpace(
		cases.Title(language.AmericanEnglish).String(sanitized))
}

func selectNumbersLettersAndSpaces(str string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			return r
		}
		return -1
	}, str)
}

func validateEmail(str string) error {
	_, err := mail.ParseAddress(str)
	return err
}

func sanitizeAndNormalizeState(state string) string {
	return strings.TrimSpace(strings.ToUpper(selectLettersOnly(state)))
}

func selectLettersOnly(state string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		return -1
	}, state)
}
