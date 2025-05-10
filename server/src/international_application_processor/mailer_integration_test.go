package international_application_processor

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testfixtures"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// cd international_application_processor && go test . --intl-application-mailer-integration-test-enable
var enableMailerIntegrationTest = flag.Bool("intl-application-mailer-integration-test-enable", false, "Enable the international application form processor mailer integration test")

var verificationStrategy = writeEmailMsgToFile

func makeFormData(state string, interest string) model.InternationalFormData {
	return testfixtures.NewInternationalFormDataBuilder().
		WithFirstName("John").
		WithLastName("Doe").
		WithState(state).
		WithInterest(interest).
		Build()
}

func makeSfBayChapter() *model.ChapterWithToken {
	return testfixtures.NewChapterBuilder().
		WithChapterID(model.SFBayChapterId).
		WithName("SF Bay").
		WithFbURL("https://facebook.com/test-chapter").
		WithInstaURL("https://instagram.com/test-chapter").
		WithTwitterURL("https://twitter.com/test-chapter").
		WithEmail("chapter-email@example.org").
		Build()
}

func makeEvent() *model.ExternalEvent {
	return &model.ExternalEvent{
		ID:        "555500",
		Name:      "Test Event",
		StartTime: time.Now().Add(48 * time.Hour),
	}
}

// writeEmailMsgToFile writes the given email message to an HTML file named after the
// test. The file includes the message headers in a table followed by the HTML
// body, allowing for easy verification of the HTML in a web browser.
func writeEmailMsgToFile(msg *mailer.Message, t *testing.T) error {
	var sb strings.Builder
	sb.WriteString("<table border='1'><tr><th>Header</th><th>Value</th></tr>")
	sb.WriteString(fmt.Sprintf("<tr><td>From</td><td>%s (%s)</td></tr>", msg.FromAddress, msg.FromName))
	sb.WriteString(fmt.Sprintf("<tr><td>To</td><td>%s (%s)</td></tr>", msg.ToAddress, msg.ToName))
	sb.WriteString(fmt.Sprintf("<tr><td>CC</td><td>%s</td></tr>", strings.Join(msg.CC, ", ")))
	sb.WriteString(fmt.Sprintf("<tr><td>BCC</td><td>%s</td></tr>", strings.Join(msg.BCC, ", ")))
	sb.WriteString(fmt.Sprintf("<tr><td>Subject</td><td>%s</td></tr>", msg.Subject))
	sb.WriteString("</table><hr>")
	sb.WriteString(msg.BodyHTML)

	fileName := t.Name() + ".html"
	const dir = "test-output/"
	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "error creating test-output directory")
	}
	err := os.WriteFile(
		dir+strings.ReplaceAll(fileName, "/", "__"),
		[]byte(sb.String()), 0644)
	if err != nil {
		return errors.Wrap(err, "error writing email to file")
	}

	log.Printf("Email written to file: %s", fileName)
	return nil
}

func TestMailIntegration(t *testing.T) {
	flag.Parse()
	if !*enableMailerIntegrationTest {
		t.Skip()
		return
	}

	t.Run("SendsNotificationEmail", func(t *testing.T) {
		msg, err := buildNotificationEmail(makeFormData("CA", "participate"), makeSfBayChapter())
		assert.NoError(t, err)

		err = verificationStrategy(msg, t)
		assert.NoError(t, err)
	})

	t.Run("OnboardingEmails", func(t *testing.T) {
		t.Run("SendsSFBayEmail", func(t *testing.T) {
			msg, err := buildOnboardingEmailMessage(
				makeFormData("CA", "participate"),
				makeSfBayChapter(),
				makeEvent())
			assert.NoError(t, err)

			err = verificationStrategy(msg, t)
			assert.NoError(t, err)
		})

		t.Run("SendsNearbyChapterEmail", func(t *testing.T) {
			msg, err := buildOnboardingEmailMessage(
				makeFormData("ZZ", "participate"),
				testfixtures.NewChapterBuilder().Build(),
				makeEvent())
			assert.NoError(t, err)

			err = verificationStrategy(msg, t)
			assert.NoError(t, err)
		})

		t.Run("SendsCAOrganizerEmail", func(t *testing.T) {
			msg, err := buildOnboardingEmailMessage(
				makeFormData("CA", "organize"), nil, nil)
			assert.NoError(t, err)

			err = verificationStrategy(msg, t)
			assert.NoError(t, err)
		})

		t.Run("SendsNonCAOrganizerEmail", func(t *testing.T) {
			msg, err := buildOnboardingEmailMessage(
				makeFormData("ZZ", "organize"), nil, nil)
			assert.NoError(t, err)

			err = verificationStrategy(msg, t)
			assert.NoError(t, err)
		})

		t.Run("SendsCAParticipatantEmail", func(t *testing.T) {
			msg, err := buildOnboardingEmailMessage(
				makeFormData("CA", "participate"), nil, nil)
			assert.NoError(t, err)

			err = verificationStrategy(msg, t)
			assert.NoError(t, err)
		})

		t.Run("SendsNonCAParticipantEmail", func(t *testing.T) {
			msg, err := buildOnboardingEmailMessage(
				makeFormData("ZZ", "participate"), nil, nil)
			assert.NoError(t, err)

			err = verificationStrategy(msg, t)
			assert.NoError(t, err)
		})
	})
}
