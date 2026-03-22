package transport

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/dxe/adb/model"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestActivistPatchHandler_RejectsChapterIDField(t *testing.T) {
	req := httptest.NewRequest(http.MethodPatch, "/api/activists/123", strings.NewReader(`{"chapter_id": 999}`))
	req = mux.SetURLVars(req, map[string]string{"id": "123"})
	rec := httptest.NewRecorder()

	ActivistPatchHandler(rec, req, model.ADBUser{}, nil, nil, nil)

	require.Equal(t, http.StatusBadRequest, rec.Code)
	require.Contains(t, rec.Body.String(), "unknown field")
	require.Contains(t, rec.Body.String(), "chapter_id")
}
