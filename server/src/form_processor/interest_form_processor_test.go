package form_processor

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
)

type activistBuilder struct {
	id        int
	email     string
	name      string
	chapterID int
}

func newActivistBuilder() *activistBuilder {
	return &activistBuilder{
		id:        0,
		email:     "email1",
		name:      "name1",
		chapterID: model.SFBayChapterIdDevTest,
	}
}

func (b *activistBuilder) withEmail(email string) *activistBuilder {
	b.email = email
	return b
}

func (b *activistBuilder) withName(name string) *activistBuilder {
	b.name = name
	return b
}

func (b *activistBuilder) withChapterID(chapterID int) *activistBuilder {
	b.chapterID = chapterID
	return b
}

func insertActivistForInterestTest(t *testing.T, db *sqlx.DB, b *activistBuilder) {
	_, err := db.Exec(
		`INSERT INTO activists (id, email, name, chapter_id) VALUES (?, ?, ?, ?)`,
		b.id, b.email, b.name, b.chapterID,
	)
	if err != nil {
		t.Fatalf("insertActivistForInterestTest failed: %s", err)
	}
}

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

func (b *interestResponseBuilder) withEmail(email string) *interestResponseBuilder {
	b.email = email
	return b
}

func (b *interestResponseBuilder) withName(name string) *interestResponseBuilder {
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
	insertActivistForInterestTest(t, db, newActivistBuilder().withName("Bob").withEmail("foo@example.org"))
	insertInterestFormResponse(t, db, newInterestResponseBuilder().withName("Bob").withEmail("bar@example.org"))

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
	insertActivistForInterestTest(t, db, newActivistBuilder().withName("Bob").withEmail("match@example.org"))
	insertInterestFormResponse(t, db, newInterestResponseBuilder().withName("Alice").withEmail("match@example.org"))

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
	insertActivistForInterestTest(t, db, newActivistBuilder().withName("Bob").withEmail("match@example.org"))
	insertActivistForInterestTest(t, db, newActivistBuilder().withName("Carl").withEmail("match@example.org"))
	insertInterestFormResponse(t, db, newInterestResponseBuilder().withName("Alice").withEmail("match@example.org"))

	/* Call functionality under test */
	ProcessInterestForms(db)

	verifyActivistCount(t, db, 2)
	verifyFormWasNotMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}

func TestProcessFormInterestForMatchingChapterIdAndName(t *testing.T) {
	/* Set up */
	db := useTestDb()
	defer db.Close()

	insertActivistForInterestTest(t, db, newActivistBuilder().
		withName("Sam").
		withEmail("foo@example.org").
		withChapterID(10))
	insertActivistForInterestTest(t, db, newActivistBuilder().
		withName("Sam").
		withEmail("bar@example.org").
		withChapterID(20))
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		withName("Sam").
		withEmail("baz@example.org").
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

	insertActivistForInterestTest(t, db, newActivistBuilder().
		withName("Alice").
		withEmail("match@example.org").
		withChapterID(10))
	insertActivistForInterestTest(t, db, newActivistBuilder().
		withName("Bob").
		withEmail("match@example.org").
		withChapterID(20))
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		withName("Carl").
		withEmail("match@example.org").
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

	insertActivistForInterestTest(t, db, newActivistBuilder().
		withName("Alice").
		withEmail("match@example.org").
		withChapterID(10))
	insertInterestFormResponse(t, db, newInterestResponseBuilder().
		withName("Alice").
		withEmail("match@example.org").
		withChapterId(20),
	)

	/* Call functionality under test */
	ProcessInterestForms(db)

	/* Verify */
	verifyActivistCount(t, db, 2)
	verifyFormWasMarkedAsProcessed(t, db, interestProcessingStatusQuery)
}
