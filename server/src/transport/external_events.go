package transport

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/dxe/adb/model"
)

// ExternalEventsListHandler lists upcoming external events for a chapter.
// Query params: chapter_id (required). Authorization is enforced in the model.
func ExternalEventsListHandler(w http.ResponseWriter, r *http.Request, user model.ADBUser, repo model.ExternalEventRepository) {
	chapterID, err := strconv.Atoi(r.URL.Query().Get("chapter_id"))
	if err != nil || chapterID <= 0 {
		sendErrorMessage(w, http.StatusBadRequest, errors.New("chapter_id is required"))
		return
	}

	// Midnight UTC so the page treats "today" the same everywhere.
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	events, err := model.ListExternalEventsForChapter(repo, user, chapterID, startOfToday)
	if err != nil {
		sendExternalEventError(w, err)
		return
	}
	if events == nil {
		events = []model.ExternalEvent{}
	}
	writeJSON(w, map[string]interface{}{"events": events})
}

func ExternalEventFeatureHandler(w http.ResponseWriter, r *http.Request, user model.ADBUser, repo model.ExternalEventRepository) {
	var body struct {
		ID       string `json:"id"`
		Featured bool   `json:"featured"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		sendErrorMessage(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	if err := model.FeatureExternalEvent(repo, user, body.ID, body.Featured); err != nil {
		sendExternalEventError(w, err)
		return
	}
	writeJSON(w, map[string]interface{}{"status": "success"})
}

func ExternalEventCancelHandler(w http.ResponseWriter, r *http.Request, user model.ADBUser, repo model.ExternalEventRepository) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		sendErrorMessage(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	if err := model.CancelExternalEvent(repo, user, body.ID); err != nil {
		sendExternalEventError(w, err)
		return
	}
	writeJSON(w, map[string]interface{}{"status": "success"})
}

func sendExternalEventError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, model.ErrValidation):
		sendErrorMessage(w, http.StatusBadRequest, err)
	case errors.Is(err, model.ErrNotFound):
		sendErrorMessage(w, http.StatusNotFound, err)
	default:
		sendErrorMessage(w, http.StatusInternalServerError, err)
	}
}
