package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/dxe/adb/model"
	"github.com/pkg/errors"
)

// ExternalEventsListHandler lists upcoming external events for the authed
// user's active chapter. The chapter always comes from the session (admins
// change it with the chapter switcher), never from the request, so organizers
// can't read other chapters' events.
func ExternalEventsListHandler(w http.ResponseWriter, r *http.Request, user model.ADBUser, repo model.ExternalEventRepository) {
	// Midnight UTC so the page treats "today" the same everywhere.
	now := time.Now().UTC()
	startOfToday := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	events, err := repo.ListUpcomingForChapter(user.ChapterID, startOfToday)
	if err != nil {
		sendErrorMessage(w, http.StatusInternalServerError, err)
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
	finishExternalEventWrite(w, repo.SetFeaturedForChapter(user.ChapterID, body.ID, body.Featured))
}

func ExternalEventCancelHandler(w http.ResponseWriter, r *http.Request, user model.ADBUser, repo model.ExternalEventRepository) {
	var body struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.ID == "" {
		sendErrorMessage(w, http.StatusBadRequest, errors.New("invalid request body"))
		return
	}
	finishExternalEventWrite(w, repo.CancelForChapter(user.ChapterID, body.ID))
}

func finishExternalEventWrite(w http.ResponseWriter, err error) {
	if errors.Is(err, model.ErrExternalEventNotFound) {
		sendErrorMessage(w, http.StatusNotFound, err)
		return
	}
	if err != nil {
		sendErrorMessage(w, http.StatusInternalServerError, err)
		return
	}
	writeJSON(w, map[string]interface{}{"status": "success"})
}
