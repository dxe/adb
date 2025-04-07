package international_application_processor

import (
	"net/mail"
	"strings"
	"unicode"

	"github.com/dxe/adb/model"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var (
	sfBayCoordinator = mail.Address{
		Name:    "Antonelle Racelis",
		Address: "antonelle@directactioneverywhere.com",
	}
	californiaCoordinator = mail.Address{
		Name:    "Almira Tanner",
		Address: "almira@directactioneverywhere.com",
	}
	globalCoordinator = mail.Address{
		Name:    "Michelle Del Cueto",
		Address: "internationalcoordination@directactioneverywhere.com",
	}
)

func stateIsCalifornia(state string) bool {
	return state == "CA"
}

func getChapterEmailFallback(state string) string {
	if stateIsCalifornia(state) {
		return californiaCoordinator.Address
	} else {
		return globalCoordinator.Address
	}
}

func getChapterEmailsWithFallback(chapter *model.ChapterWithToken, fallback string) []string {
	emails := getChapterEmails(chapter)
	if len(emails) == 0 {
		return []string{fallback}
	}
	return emails
}

func getChapterEmails(chapter *model.ChapterWithToken) []string {
	var emails []string

	if chapter.Email != "" {
		emails = append(emails, chapter.Email)
	}

	emails = append(emails, getChapterOrganizerEmails(chapter)...)

	return emails
}

func getChapterOrganizerEmails(chapter *model.ChapterWithToken) []string {
	organizers := chapter.Organizers

	emails := make([]string, 0, len(organizers))
	if len(organizers) > 0 {
		for _, o := range organizers {
			if o.Email != "" {
				emails = append(emails, o.Email)
			}
		}
	}

	return emails
}

func sanitizeAndFormatName(name string) string {
	sanitized := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' {
			return r
		}
		return -1
	}, name)

	return strings.TrimSpace(
		cases.Title(language.AmericanEnglish).String(sanitized))
}

func validateEmail(str string) error {
	_, err := mail.ParseAddress(str)
	return err
}

func sanitizeState(state string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) {
			return r
		}
		return -1
	}, state)
}
