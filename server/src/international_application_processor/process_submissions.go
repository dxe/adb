package international_application_processor

import (
	"log"
	"time"

	"github.com/dxe/adb/model"
	"github.com/pkg/errors"

	"github.com/jmoiron/sqlx"
)

// Sends emails every 60 minutes.
// Should be run in a goroutine.
func RunProcessor(db *sqlx.DB) {
	for {
		log.Println("Starting international mailer")
		ProcessFormSubmissions(db)
		log.Println("Finished international mailer")

		time.Sleep(60 * time.Minute)
	}
}

func ProcessFormSubmissions(db *sqlx.DB) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered from panic in international mailer:", r)
		}
	}()

	records, err := model.GetInternationalFormSubmissionsToEmail(db)
	if err != nil {
		log.Println("Failed to get int'l form submissions to email:", err)
	}
	log.Printf("Int'l form mailer found %d records to process", len(records))

	for _, rec := range records {
		if err := processFormSubmission(db, rec); err != nil {
			log.Printf("Error processing int'l form submission with ID %v: %v", rec.ID, err)
		}
	}
}

func processFormSubmission(db *sqlx.DB, formDataUnsanitized model.InternationalFormData) error {
	chapter, err := getNearestChapterOrNil(db, formDataUnsanitized)
	if err != nil {
		return errors.Wrap(err, "error getting nearest chapter")
	}

	// Email the person who submitted the form.
	err = sendOnboardingEmail(db, formDataUnsanitized, chapter)
	if err != nil {
		return errors.Wrap(err, "failed to send onboarding email")
	}

	// Mark submission as procesed.
	err = model.UpdateInternationalFormSubmissionEmailStatus(db, formDataUnsanitized.ID)
	if err != nil {
		return errors.Wrap(err, "error updating status: %w")
	}

	return nil
}

func getNearestChapterOrNil(db *sqlx.DB, formData model.InternationalFormData) (*model.ChapterWithToken, error) {
	nearestChapters, err := model.FindNearestChaptersSortedByDistanceDeprecated(db, formData.Lat, formData.Lng)
	if err != nil {
		return nil, errors.Wrap(err, "error fetching nearby chapters: %w")
	}

	nearestChapter := pickNearestChapterOrNil(nearestChapters, formData.Country)

	if nearestChapter != nil {
		// `GetChapterByID` returns more details about the chapter than `FindNearestChapters`.
		*nearestChapter, err = model.GetChapterWithTokenById(db, nearestChapter.ChapterID)
		if err != nil {
			return nil, errors.Wrap(err, "error fetching chapter: %w")
		}
	}

	return nearestChapter, nil
}

// pickNearestChapterOrNil picks the first chapter that is within 100 miles and in
// the same country.
//
// `nearestChapters` MUST be sorted by distance in ascending order.
func pickNearestChapterOrNil(nearestChapters []model.ChapterWithToken, country string) *model.ChapterWithToken {
	for _, chapter := range nearestChapters {
		if chapter.Distance > 100 {
			break
		}
		if chapter.Country != country {
			continue
		}

		return &chapter
	}

	return nil
}
