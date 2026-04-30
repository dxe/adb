package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/persistence"
	"github.com/dxe/adb/testdb"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestActivistPatchHandler_RejectsChapterIDField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/api/activists/123", strings.NewReader(`{"chapter_id": 999}`))
	req = mux.SetURLVars(req, map[string]string{"id": "123"})
	rec := httptest.NewRecorder()

	// Just pass nil for db/repo/userRepo because the handler rejects the
	// unknown JSON field before attempting to access these deps.
	ActivistPatchHandler(rec, req, model.ADBUser{}, nil, nil, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "unknown field")
	require.Contains(t, rec.Body.String(), "chapter_id")
}

func TestActivistPatchHandler_PatchesAndReturnsActivist(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	repo := persistence.NewActivistRepository(db)
	userRepo := persistence.NewUserRepository(db)

	devUser, err := userRepo.GetUser(model.DevTestUserId, "")
	require.NoError(t, err)

	activistID, err := model.CreateActivist(db, model.ActivistExtra{
		Activist: model.Activist{
			Name:      "Initial Name",
			ChapterID: model.SFBayChapterIdDevTest,
		},
	})
	require.NoError(t, err)

	body := `{
		"name": "Patched Name",
		"preferred_name": "Patchy",
		"phone": "555-0100",
		"hiatus": true,
		"notes": "patched notes"
	}`

	idStr := strconv.Itoa(activistID)
	req := httptest.NewRequest(http.MethodPatch, "/api/activists/"+idStr, strings.NewReader(body))
	req = mux.SetURLVars(req, map[string]string{"id": idStr})
	rec := httptest.NewRecorder()

	ActivistPatchHandler(rec, req, devUser, db, repo, userRepo)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Activist model.ActivistJSON `json:"activist"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))
	require.Equal(t, "Patched Name", resp.Activist.Name)
	require.Equal(t, "Patchy", resp.Activist.PreferredName)
	require.Equal(t, "555-0100", resp.Activist.Phone)
	require.True(t, resp.Activist.Hiatus)
	require.Equal(t, "patched notes", resp.Activist.Notes)
}

func TestActivistPatchHandler_NotFound(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()

	repo := persistence.NewActivistRepository(db)
	userRepo := persistence.NewUserRepository(db)

	devUser, err := userRepo.GetUser(model.DevTestUserId, "")
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPatch, "/api/activists/9999", strings.NewReader(`{"phone": "555-0100"}`))
	req = mux.SetURLVars(req, map[string]string{"id": "9999"})
	rec := httptest.NewRecorder()

	ActivistPatchHandler(rec, req, devUser, db, repo, userRepo)

	require.Equal(t, http.StatusNotFound, rec.Code)
}
