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

func sendTestNotificationEmail(formData model.InternationalFormData, chapter *model.ChapterWithToken) error {
	msg, err := buildNotificationEmail(formData, chapter)
	if err != nil {
		return errors.Wrap(err, "error building int'l application notification email")
	}

	err = mailer.Send(*msg)
	if err != nil {
		return errors.Wrap(err, "error sending int'l application notification email")
	}

	log.Printf("Test int'l application notification email sent to %v", formData.Email)

	return err
}

func sendTestOnboardingEmail(formData model.InternationalFormData, chapter *model.ChapterWithToken) error {
	nextEvent := &model.ExternalEvent{
		ID:        "555500",
		Name:      "Test Event",
		StartTime: time.Now().Add(48 * time.Hour),
	}

	msg, err := buildOnboardingEmailMessage(formData, chapter, nextEvent)
	if err != nil {
		return errors.Wrap(err, "error building email message")
	}

	if err := mailer.Send(*msg); err != nil {
		return errors.Wrap(err, "error sending email for international form submission")
	}

	log.Printf("Int'l mailer onboarding email sent to %v", msg.ToAddress)

	return nil
}

func TestFunctionWithOptionalE2EE(t *testing.T) {
	flag.Parse()
	if !*enableMailerIntegrationTest {
		t.Skip()
		return
	}
	if *mailerIntegrationTestEmail == "" {
		t.Skip("No email address provided for mailer integration test")
		return
	}

	downtownBerkeley := coords{
		lat: 37.870352730245024,
		lng: -122.26794876651053,
	}

	formData := testfixtures.NewInternationalFormDataBuilder().
		WithFirstName("John").
		WithLastName("Doe").
		WithEmail(*mailerIntegrationTestEmail).
		WithLat(downtownBerkeley.lat).
		WithLng(downtownBerkeley.lng).
		WithState("CA").
		Build()

	chapter := testfixtures.NewChapterBuilder().
		WithChapterID(model.SFBayChapterId).
		WithName("SF Bay").
		WithFbURL("https://facebook.com/test-chapter").
		WithInstaURL("https://instagram.com/test-chapter").
		WithTwitterURL("https://twitter.com/test-chapter").
		WithEmail("chapter-email@example.org").
		Build()

	// Email the person who submitted the form.
	err := sendTestOnboardingEmail(formData, &chapter)
	assert.NoError(t, err)

	// Email relevant existing organizers to notify them of the form submission.
	err = sendTestNotificationEmail(formData, &chapter)
	assert.NoError(t, err)
}
