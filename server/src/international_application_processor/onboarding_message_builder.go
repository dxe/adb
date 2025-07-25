package international_application_processor

import (
	"fmt"
	"html"
	"strings"

	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"
	"github.com/pkg/errors"
)

type onboardingEmailMessageBuilder struct {
	chapter             *model.ChapterWithToken
	nextEvent           *model.ExternalEvent
	firstName           string // Sanitized (but not HTML-escaped)
	fullName            string // Sanitized (but not HTML-escaped)
	email               string // Format validated
	state               string // Sanitized
	interestUnsanitized string
	formDataUnsanitized model.InternationalFormData
}

// onboardingEmailType represents the email templates used to email the
// responder. Each template is specific to factors such as the responder's
// location, nearby chapter, and level of interest.
type onboardingEmailType int

const (
	nearSFBayChapter onboardingEmailType = iota
	nearNonSFBayChapter
	caOrganizerNotNearAnyChapter
	nonCaOrganizerNotNearAnyChapter
	participantNotNearAnyChapter
)

func buildOnboardingEmailMessage(formData model.InternationalFormData, chapter *model.ChapterWithToken, nextEvent *model.ExternalEvent) (*mailer.Message, error) {
	firstName := sanitizeAndFormatName(formData.FirstName)
	fullName := firstName + " " + sanitizeAndFormatName(formData.LastName)

	// Ensure user-provided email is nothing more than a normal email since this
	// value is injected directly into email headers we send.
	if err := validateEmail(formData.Email); err != nil {
		return nil, errors.Wrapf(err, "invalid email address: %v", formData.Email)
	}
	email := formData.Email

	state := sanitizeAndNormalizeState(formData.State)
	interestUnsanitized := formData.Interest

	builder := onboardingEmailMessageBuilder{
		chapter,
		nextEvent,
		firstName,
		fullName,
		email,
		state,
		interestUnsanitized,
		formData,
	}

	msg, err := builder.build()
	return msg, err
}

func (b *onboardingEmailMessageBuilder) build() (*mailer.Message, error) {
	emailType := b.getOnboardingEmailType()

	var builders = map[onboardingEmailType]func(*onboardingEmailMessageBuilder) (mailer.Message, error){
		nearSFBayChapter:                (*onboardingEmailMessageBuilder).nearSFBayChapter,
		nearNonSFBayChapter:             (*onboardingEmailMessageBuilder).nearNonSFBayChapter,
		caOrganizerNotNearAnyChapter:    (*onboardingEmailMessageBuilder).caOrganizerNotNearAnyChapter,
		nonCaOrganizerNotNearAnyChapter: (*onboardingEmailMessageBuilder).nonCaOrganizerNotNearAnyChapter,
		participantNotNearAnyChapter:    (*onboardingEmailMessageBuilder).participantNotNearAnyChapter,
	}

	builder, found := builders[emailType]
	if !found {
		return nil, errors.Errorf("no builder found for email type %v", emailType)
	}

	msg, err := builder(b)
	if err != nil {
		return nil, err
	}

	// Always BCC the sender so they:
	// * Can follow up if there are no replies
	// * See that the emails are getting sent out successfully
	// * Can report any outdated info
	msg.BCC = append(msg.BCC, msg.FromAddress)

	msg.BodyHTML += buildFormResponsesSummary(b.formDataUnsanitized, b.chapter)

	return &msg, nil
}

func (b *onboardingEmailMessageBuilder) getOnboardingEmailType() onboardingEmailType {
	if b.chapter != nil {
		if b.chapter.ChapterID == model.SFBayChapterId {
			return nearSFBayChapter
		}
		return nearNonSFBayChapter
	}

	if b.interestUnsanitized == "organize" {
		if stateIsCalifornia(b.state) {
			return caOrganizerNotNearAnyChapter
		}
		return nonCaOrganizerNotNearAnyChapter
	}

	return participantNotNearAnyChapter
}

func buildFormResponsesSummary(formDataUnsanitized model.InternationalFormData, chapter *model.ChapterWithToken) string {
	fullName := sanitizeAndFormatName(formDataUnsanitized.FirstName + " " + formDataUnsanitized.LastName)
	chapterName := "none"
	if chapter != nil {
		chapterName = chapter.Name
	}
	cityUnsanitized := formDataUnsanitized.City + ", " + formDataUnsanitized.State + ", " + formDataUnsanitized.Country

	var body strings.Builder
	body.WriteString("<br/><br/>")
	body.WriteString("<div style=\"color:#303030\">")
	body.WriteString("<p>Here are the details provided on the international application form:</p>")
	fmt.Fprintf(&body, "<p>Name: %s</p>", html.EscapeString(fullName))
	fmt.Fprintf(&body, "<p>Email: %s</p>", html.EscapeString(formDataUnsanitized.Email))
	fmt.Fprintf(&body, "<p>Phone: %s</p>", html.EscapeString(formDataUnsanitized.Phone))
	fmt.Fprintf(&body, "<p>City: %s</p>", html.EscapeString(cityUnsanitized))
	fmt.Fprintf(&body, "<p>Interest: %s</p>", html.EscapeString(formDataUnsanitized.Interest))
	fmt.Fprintf(&body, "<p>Involvement: %s</p>", html.EscapeString(formDataUnsanitized.Involvement))
	fmt.Fprintf(&body, "<p>Nearby chapter: %s</p>", chapterName)
	body.WriteString("</div>")

	return body.String()
}
