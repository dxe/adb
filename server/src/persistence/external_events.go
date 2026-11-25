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
