package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/pkg/shared"
	"github.com/stretchr/testify/require"
)

// Authorization lives in the model layer (see model/external_events_authz_test.go);
// these tests cover request parsing and how its errors map to HTTP statuses.
type fakeExternalEventRepo struct{}

func (fakeExternalEventRepo) ListUpcomingForChapter(int, time.Time) ([]model.ExternalEvent, error) {
	return nil, nil
}
func (fakeExternalEventRepo) GetChapterID(id string) (int, error) {
	if id == "1" {
		return 10, nil
	}
	return 0, model.ErrNotFound
}
func (fakeExternalEventRepo) SetFeatured(int, string, bool) error { return nil }
func (fakeExternalEventRepo) Cancel(int, string) error            { return nil }

func TestExternalEventHandlers_StatusCodes(t *testing.T) {
	organizer := model.ADBUser{Roles: []string{shared.RoleOrganizer}, ChapterID: 10}
	otherOrganizer := model.ADBUser{Roles: []string{shared.RoleOrganizer}, ChapterID: 20}
	repo := fakeExternalEventRepo{}

	list := func(user model.ADBUser, query string) int {
		rec := httptest.NewRecorder()
		ExternalEventsListHandler(rec, httptest.NewRequest(http.MethodGet, "/api/external-events"+query, nil), user, repo)
		return rec.Code
	}
	require.Equal(t, http.StatusOK, list(organizer, "?chapter_id=10"))
	require.Equal(t, http.StatusBadRequest, list(organizer, ""))
	require.Equal(t, http.StatusBadRequest, list(organizer, "?chapter_id=10x"))
	require.Equal(t, http.StatusBadRequest, list(otherOrganizer, "?chapter_id=10"), "model validation error")

	cancel := func(user model.ADBUser, body string) int {
		rec := httptest.NewRecorder()
		ExternalEventCancelHandler(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), user, repo)
		return rec.Code
	}
	require.Equal(t, http.StatusOK, cancel(organizer, `{"id":"1"}`))
	require.Equal(t, http.StatusNotFound, cancel(organizer, `{"id":"2"}`))
	require.Equal(t, http.StatusBadRequest, cancel(otherOrganizer, `{"id":"1"}`))
	require.Equal(t, http.StatusBadRequest, cancel(organizer, `{}`))

	feature := func(user model.ADBUser, body string) int {
		rec := httptest.NewRecorder()
		ExternalEventFeatureHandler(rec, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), user, repo)
		return rec.Code
	}
	require.Equal(t, http.StatusOK, feature(organizer, `{"id":"1","featured":true}`))
	require.Equal(t, http.StatusBadRequest, feature(otherOrganizer, `{"id":"1","featured":true}`))
}
