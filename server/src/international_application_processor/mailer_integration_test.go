package international_application_processor

import (
	"flag"
	"log"
	"testing"
	"time"

	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"
	"github.com/dxe/adb/testfixtures"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

// This is a test of the integration of this package with the email service
// and allows viewing the emails in a real email client.

// SMTP_ environment variables must be configured to run this test.

var enableMailerIntegrationTest = flag.Bool("intl-application-mailer-integration-test-enable", false, "Enable the international application form processor mailer integration test")
var mailerIntegrationTestEmail = flag.String("intl-application-mailer-integration-test-email", "", "Email for the international application form processor mailer integration test")

type coords struct {
	lat float64
	lng float64
}

var downtownBerkeley = coords{
	lat: 37.870352730245024,
	lng: -122.26794876651053,
}

func makeFormData(location coords, state string, interest string) model.InternationalFormData {
	return testfixtures.NewInternationalFormDataBuilder().
		WithFirstName("John").
		WithLastName("Doe").
		WithEmail(*mailerIntegrationTestEmail).
		WithLat(location.lat).
		WithLng(location.lng).
		WithState(state).
		WithInterest(interest).
		Build()
}

func makeSfBayChapter() model.ChapterWithToken {
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

func replaceRecipientsWithTestAddress(msg *mailer.Message) {
	// Replace the recipient with the test email address
	msg.ToAddress = *mailerIntegrationTestEmail
	msg.CC = []string{}
	msg.BCC = []string{}
}

func sendTestNotificationEmail(formData model.InternationalFormData, chapter *model.ChapterWithToken) error {
	msg, err := buildNotificationEmail(formData, chapter)
	if err != nil {
		return errors.Wrap(err, "error building int'l application notification email")
	}

	replaceRecipientsWithTestAddress(msg)

	err = mailer.Send(*msg)
	if err != nil {
		return errors.Wrap(err, "error sending int'l application notification email")
	}

	log.Printf("Test int'l application notification email sent to %v", formData.Email)

	return err
}

func sendTestOnboardingEmail(formData model.InternationalFormData, chapter *model.ChapterWithToken) error {
	msg, err := buildOnboardingEmailMessage(formData, chapter, makeEvent())
	if err != nil {
		return errors.Wrap(err, "error building email message")
	}

	replaceRecipientsWithTestAddress(msg)

	if err := mailer.Send(*msg); err != nil {
		return errors.Wrap(err, "error sending email for international form submission")
	}

	log.Printf("Int'l mailer onboarding email sent to %v", msg.ToAddress)

	return nil
}

func TestMailIntegration(t *testing.T) {
	flag.Parse()
	if !*enableMailerIntegrationTest {
		t.Skip()
		return
	}
	if *mailerIntegrationTestEmail == "" {
		t.Skip("No email address provided for mailer integration test")
		return
	}

	t.Run("SendsNotificationEmail", func(t *testing.T) {
		formData := makeFormData(downtownBerkeley, "CA", "participate")
		chapter := makeSfBayChapter()

		err := sendTestNotificationEmail(formData, &chapter)
		assert.NoError(t, err)
	})

	t.Run("OnboardingEmails", func(t *testing.T) {
		t.Run("SendsSFBayEmail", func(t *testing.T) {
			formData := makeFormData(downtownBerkeley, "CA", "participate")
			chapter := makeSfBayChapter()

			err := sendTestOnboardingEmail(formData, &chapter)
			assert.NoError(t, err)
		})

		t.Run("SendsNearbyChapterEmail", func(t *testing.T) {
			formData := makeFormData(coords{0, 0}, "ZZ", "participate")
			chapter := testfixtures.NewChapterBuilder().Build()

			err := sendTestOnboardingEmail(formData, &chapter)
			assert.NoError(t, err)
		})

		t.Run("SendsCAOrganizerEmail", func(t *testing.T) {
			formData := makeFormData(coords{0, 0}, "CA", "organize")

			err := sendTestOnboardingEmail(formData, nil)
			assert.NoError(t, err)
		})

		t.Run("SendsNonCAOrganizerEmail", func(t *testing.T) {
			formData := makeFormData(coords{0, 0}, "ZZ", "organize")

			err := sendTestOnboardingEmail(formData, nil)
			assert.NoError(t, err)
		})

		t.Run("SendsCAParticipatantEmail", func(t *testing.T) {
			formData := makeFormData(coords{0, 0}, "CA", "participate")

			err := sendTestOnboardingEmail(formData, nil)
			assert.NoError(t, err)
		})

		t.Run("SendsNonCAParticipantEmail", func(t *testing.T) {
			formData := makeFormData(coords{0, 0}, "ZZ", "participate")

			err := sendTestOnboardingEmail(formData, nil)
			assert.NoError(t, err)
		})
	})
}
