package form_processor

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

type interestResponseBuilder struct {
	id        int
	chapterId int
	form      string
	email     string
	name      string
	phone     string
}

func newInterestResponseBuilder() *interestResponseBuilder {
	return &interestResponseBuilder{
		id:        0,
		chapterId: model.SFBayChapterIdDevTest,
		form:      "form1",
		email:     "email1",
		name:      "name1",
		phone:     "phone1",
	}
}

func (b *interestResponseBuilder) withChapterId(chapterID int) *interestResponseBuilder {
	b.chapterId = chapterID
	return b
}

func (b *interestResponseBuilder) WithEmail(email string) *interestResponseBuilder {
	b.email = email
	return b
}

func (b *interestResponseBuilder) WithName(name string) *interestResponseBuilder {
	b.name = name
	return b
}

func insertInterestFormResponse(t *testing.T, db *sqlx.DB, b *interestResponseBuilder) {
	_, err := db.Exec(`
INSERT INTO form_interest (
  id,
  chapter_id,
  form,
  email,
  name,
  phone,
  zip,
  referral_friends,
  referral_apply,
  referral_outlet,
  comments,
  interests
) VALUES (
  ?, ?, ?, ?, ?, ?,
  "zip1",
  "referral_friends1",
  "referral_apply1",
  "referral_outlet1",
  "comments1",
  "interests1"
);
`,
		b.id, b.chapterId, b.form, b.email, b.name, b.phone,
	)

	if err != nil {
		t.Fatalf("insertInterestFormResponse failed: %s", err)
	}
}

func TestProcessFormInterestForNoMatchingActivist(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	insertInterestFormResponse(t, db, newInterestResponseBuilder())

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistCount(t, db, 1)
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForActivistMatchingOnName(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Bob").
		WithEmail("foo@example.org").
		Build())
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		WithName("Bob").
		WithEmail("bar@example.org"))

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistCount(t, db, 1)
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForActivistMatchingOnEmail(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Bob").
		WithEmail("match@example.org").
		Build())
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		WithName("Alice").
		WithEmail("match@example.org"))

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistCount(t, db, 1)
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForMultipleMatchingActivistsOnEmail(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()
	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Bob").
		WithEmail("match@example.org").
		Build())
	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Carl").
		WithEmail("match@example.org").
		Build())
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		WithName("Alice").
		WithEmail("match@example.org"))

	/* Call functionality under test */
	ProcessInterestForms(db)

	verifyActivistCount(t, db, 2)
	verifyFormWasNotMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForMatchingChapterIdAndName(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()

	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Sam").
		WithEmail("foo@example.org").
		WithChapterID(10).
		Build())
	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Sam").
		WithEmail("bar@example.org").
		WithChapterID(20).
		Build())
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		WithName("Sam").
		WithEmail("baz@example.org").
		withChapterId(10),
	)

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistCount(t, db, 2)
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForMatchingChapterIdAndEmail(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()

	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Alice").
		WithEmail("match@example.org").
		WithChapterID(10).
		Build())
	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Bob").
		WithEmail("match@example.org").
		WithChapterID(20).
		Build())
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		WithName("Carl").
		WithEmail("match@example.org").
		withChapterId(10),
	)

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistCount(t, db, 2)
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForNonMatchingChapter(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()

	model.MustInsertActivist(t, db, model.NewActivistBuilder().
		WithName("Alice").
		WithEmail("match@example.org").
		WithChapterID(10).Build())
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		WithName("Alice").
		WithEmail("match@example.org").
		withChapterId(20),
	)

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistCount(t, db, 2)
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}
