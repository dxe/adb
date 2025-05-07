// Logic for emailing the applicant / responder after they submit the
// international application form.

package international_application_processor

import (
	"log"
	"time"

	"github.com/dxe/adb/mailer"
	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

func sendOnboardingEmail(db *sqlx.DB, formData model.InternationalFormData, chapter *model.ChapterWithToken) error {
	nextEvent, err := getNextEventOrNil(db, chapter)
	if err != nil {
		log.Println("error getting next event:", err)
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

// getNextEventOrNil retrieves the next event for the given chapter within a
// time range. If no events are found, or if `chapter` is nil, it returns nil.
func getNextEventOrNil(db *sqlx.DB, chapter *model.ChapterWithToken) (*model.ExternalEvent, error) {
	if chapter == nil {
		return nil, nil
	}

	if chapter.ID == 0 {
		return nil, errors.New("chapter ID cannot be 0")
	}

	startTime := time.Now()
	endTime := time.Now().Add(60 * 24 * time.Hour)
	events, err := model.GetExternalEvents(db, chapter.ID, startTime, endTime)
	if err != nil {
		return nil, err
	}
	if len(events) == 0 {
		return nil, nil
	}

	return &events[0], nil
}
