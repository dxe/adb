// Logic for notifying existing organizers that someone submitted the
// international application form.

package international_application_processor

import (
	"fmt"
	"html"
	"log"
	"strings"

	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"
	"github.com/pkg/errors"
)

func sendNotificationEmail(formData model.InternationalFormData, chapter *model.ChapterWithToken) error {
	msg, err := buildNotificationEmail(formData, chapter)
	if err != nil {
		return errors.Wrap(err, "error building int'l application notification email")
	}

	err = mailer.Send(*msg)
	if err != nil {
		return errors.Wrap(err, "error sending int'l application notification email")
	}

	log.Printf("Int'l application notification email sent to %v", formData.Email)

	return err
}

func buildNotificationEmail(formData model.InternationalFormData, chapter *model.ChapterWithToken) (*mailer.Message, error) {
	recipients, err := getNotificationRecipients(chapter, formData.State)
	if err != nil {
		return nil, err
	}

	fullName := sanitizeAndFormatName(formData.FirstName + " " + formData.LastName)

	msg := mailer.Message{
		FromName:    "DxE Join Form",
		FromAddress: "noreply@directactioneverywhere.com",
		ToAddress:   recipients[0],
		CC:          recipients[1:],
		Subject:     fmt.Sprintf("%v signed up to join your chapter", fullName),
	}

	var body strings.Builder
	body.WriteString("<p>Here are the details provided on the international application form:</p>")
	fmt.Fprintf(&body, "<p>Name: %s</p>", html.EscapeString(fullName))
	fmt.Fprintf(&body, "<p>Email: %s</p>", html.EscapeString(formData.Email))
	fmt.Fprintf(&body, "<p>Phone: %s</p>", html.EscapeString(formData.Phone))
	fmt.Fprintf(&body, "<p>City: %s</p>", html.EscapeString(formData.City))
	fmt.Fprintf(&body, "<p>Interest: %s</p>", html.EscapeString(formData.Interest))
	fmt.Fprintf(&body, "<p>Involvement: %s</p>", html.EscapeString(formData.Involvement))
	fmt.Fprintf(&body, "<p>Skills: %s</p>", html.EscapeString(formData.Skills))
	msg.BodyHTML = body.String()

	return &msg, nil
}

func getNotificationRecipients(chapter *model.ChapterWithToken, state string) ([]string, error) {
	if chapter == nil {
		return []string{getChapterEmailFallback(state)}, nil
	}

	if chapter.ChapterID == model.SFBayChapterId {
		return []string{sfBayCoordinator.Address}, nil
	}

	return getChapterEmailsWithFallback(chapter,
		getChapterEmailFallback(state)), nil
}
