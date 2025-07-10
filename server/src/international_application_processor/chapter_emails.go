package international_application_processor

import (
	"github.com/dxe/adb/model"
)

type coordinator struct {
	Name     string
	Role     string
	Address  string
	Pronouns string
}

var (
	sfBayCoordinator = coordinator{
		Name:    "Antonelle Racelis",
		Role:    "Organizer",
		Address: "antonelle@directactioneverywhere.com",
	}
	californiaCoordinator = coordinator{
		Name:     "Almira Tanner",
		Role:     "Lead Organizer",
		Address:  "almira@directactioneverywhere.com",
		Pronouns: "she/her",
	}
	globalCoordinator = coordinator{
		Name:    "Michelle Del Cueto",
		Role:    "International Coordinator",
		Address: "internationalcoordination@directactioneverywhere.com",
	}
	networkMemberProgramCoordinator = coordinator{
		Name:    "Robert Yamada",
		Role:    "International Coordination Team",
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
