package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/persistence"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

func TestUsersAssignableListHandler_FiltersDisabledAndSortsByName(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	userRepo := persistence.NewUserRepository(db)

	// testdb.NewDB() already seeds a non-disabled "Dev User" (model.DevTestUserId).
	_, err := userRepo.CreateUser(model.ADBUser{
		Email:     "zoe@example.org",
		Name:      "Zoe Zimmerman",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.NoError(t, err)

	_, err = userRepo.CreateUser(model.ADBUser{
		Email:     "alice@example.org",
		Name:      "Alice Adams",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.NoError(t, err)

	_, err = userRepo.CreateUser(model.ADBUser{
		Email:     "disabled@example.org",
		Name:      "Debbie Disabled",
		ChapterID: model.SFBayChapterIdDevTest,
		Disabled:  true,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/users/assignable", nil)
	rec := httptest.NewRecorder()

	UsersAssignableListHandler(rec, req, userRepo)

	require.Equal(t, http.StatusOK, rec.Code)

	var resp struct {
		Users []AssignableUserJson `json:"users"`
	}
	require.NoError(t, json.NewDecoder(rec.Body).Decode(&resp))

	var names []string
	for _, u := range resp.Users {
		names = append(names, u.Name)
	}

	// The disabled user must be excluded, and the rest sorted by name ascending.
	require.Equal(t, []string{"Alice Adams", "Dev User", "Zoe Zimmerman"}, names)
}

func TestUsersAssignableListHandler_OnlyExposesIDAndName(t *testing.T) {
	db := testdb.NewDB()
	defer func() { _ = db.Close() }()

	userRepo := persistence.NewUserRepository(db)

	created, err := userRepo.CreateUser(model.ADBUser{
		Email:     "someone@example.org",
		Name:      "Someone Else",
		ChapterID: model.SFBayChapterIdDevTest,
		Roles:     []string{"organizer"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodGet, "/api/users/assignable", nil)
	rec := httptest.NewRecorder()

	UsersAssignableListHandler(rec, req, userRepo)

	require.Equal(t, http.StatusOK, rec.Code)

	body := rec.Body.String()
	require.NotContains(t, body, `"email"`)
	require.NotContains(t, body, `"roles"`)
	require.NotContains(t, body, `"disabled"`)
	require.NotContains(t, body, `"chapter_id"`)

	var resp struct {
		Users []map[string]interface{} `json:"users"`
	}
	require.NoError(t, json.Unmarshal([]byte(body), &resp))

	found := false
	for _, u := range resp.Users {
		if int(u["id"].(float64)) == created.ID {
			found = true
			require.ElementsMatch(t, []string{"id", "name"}, keysOf(u))
		}
	}
	require.True(t, found, "expected created user %d to be present in assignable list", created.ID)
}

func keysOf(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}
