package persistence

import (
	"database/sql"
	"time"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
)

type DBExternalEventRepository struct {
	db *sqlx.DB
}

var _ model.ExternalEventRepository = (*DBExternalEventRepository)(nil)

func NewExternalEventRepository(db *sqlx.DB) *DBExternalEventRepository {
	return &DBExternalEventRepository{db: db}
}

func (r *DBExternalEventRepository) ListUpcomingForChapter(chapterID int, since time.Time) ([]model.ExternalEvent, error) {
	var pageID int
	err := r.db.Get(&pageID, `SELECT id FROM fb_pages WHERE chapter_id = ?`, chapterID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(err, "failed to look up chapter's page")
	}
	if pageID == 0 {
		return nil, nil
	}
	return model.GetExternalEvents(r.db, pageID, since, time.Time{})
}

func (r *DBExternalEventRepository) ownedByChapter(chapterID int, eventID string) error {
	var found int
	err := r.db.Get(&found, `SELECT 1 FROM fb_events e
		JOIN fb_pages p ON p.id = e.page_id
		WHERE e.id = ? AND p.chapter_id = ?`, eventID, chapterID)
	if errors.Is(err, sql.ErrNoRows) {
		return model.ErrExternalEventNotFound
	}
	return errors.Wrap(err, "failed to check event's chapter")
}

func (r *DBExternalEventRepository) SetFeaturedForChapter(chapterID int, eventID string, featured bool) error {
	if err := r.ownedByChapter(chapterID, eventID); err != nil {
		return err
	}
	_, err := r.db.Exec(`UPDATE fb_events SET featured = ? WHERE id = ?`, featured, eventID)
	return errors.Wrap(err, "failed to feature event")
}

func (r *DBExternalEventRepository) CancelForChapter(chapterID int, eventID string) error {
	if err := r.ownedByChapter(chapterID, eventID); err != nil {
		return err
	}
	_, err := r.db.Exec(`UPDATE fb_events SET is_canceled = 1 WHERE id = ?`, eventID)
	return errors.Wrap(err, "failed to cancel event")
}
