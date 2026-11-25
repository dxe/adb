package transport

import (
	"net/http"
	"strconv"
	"time"

	"github.com/dxe/adb/model"
	"github.com/pkg/errors"
)

// ExternalEventsListHandler lists upcoming external events for one chapter.
// Query params: chapter_id (required).
func ExternalEventsListHandler(w http.ResponseWriter, r *http.Request, repo model.ExternalEventRepository) {
	chapterID, err := strconv.Atoi(r.URL.Query().Get("chapter_id"))
	if err != nil || chapterID <= 0 {
		sendErrorMessage(w, http.StatusBadRequest, errors.New("chapter_id is required"))
		return
	}

	// Midnight UTC so the page treats "today" the same everywhere.
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	events, err := repo.ListUpcomingForChapter(chapterID, startOfToday)
	if err != nil {
		sendErrorMessage(w, http.StatusInternalServerError, err)
		return
	}
	if events == nil {
		events = []model.ExternalEvent{}
	}
	writeJSON(w, map[string]interface{}{"events": events})
}
