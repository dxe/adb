package persistence

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/dxe/adb/model"
	"github.com/jmoiron/sqlx"
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
		return nil, fmt.Errorf("failed to look up chapter's page: %w", err)
	}
	if pageID == 0 {
		return nil, nil
	}
	return model.GetExternalEvents(r.db, pageID, since, time.Time{})
}

func (r *DBExternalEventRepository) GetChapterID(eventID string) (int, error) {
	var chapterID int
	err := r.db.Get(&chapterID, `SELECT p.chapter_id FROM fb_events e
		JOIN fb_pages p ON p.id = e.page_id
		WHERE e.id = ?`, eventID)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w: external event %s not found", model.ErrNotFound, eventID)
	}
	if err != nil {
		return 0, fmt.Errorf("failed to look up event's chapter: %w", err)
	}
	return chapterID, nil
}

func (r *DBExternalEventRepository) SetFeatured(chapterID int, eventID string, featured bool) error {
	return r.updateForChapter(chapterID, eventID, "featured = ?", featured)
}

func (r *DBExternalEventRepository) Cancel(chapterID int, eventID string) error {
	return r.updateForChapter(chapterID, eventID, "is_canceled = 1")
}

// updateForChapter applies `set` to the event only if it's on chapterID's page,
// returning ErrNotFound otherwise.
func (r *DBExternalEventRepository) updateForChapter(chapterID int, eventID string, set string, setArgs ...interface{}) error {
	args := append(setArgs, eventID, chapterID)
	res, err := r.db.Exec(`UPDATE fb_events SET `+set+` WHERE id = ?
		AND page_id IN (SELECT id FROM fb_pages WHERE chapter_id = ?)`, args...)
	if err != nil {
		return fmt.Errorf("failed to update external event: %w", err)
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to update external event: %w", err)
	}
	if rows > 0 {
		return nil
	}
	// MySQL reports 0 rows when the values were already set, so tell that apart
	// from the event not belonging to the chapter.
	var found int
	err = r.db.Get(&found, `SELECT 1 FROM fb_events WHERE id = ?
		AND page_id IN (SELECT id FROM fb_pages WHERE chapter_id = ?)`, eventID, chapterID)
	if errors.Is(err, sql.ErrNoRows) {
		return fmt.Errorf("%w: external event %s not found", model.ErrNotFound, eventID)
	}
	if err != nil {
		return fmt.Errorf("failed to check external event: %w", err)
	}
	return nil
}
