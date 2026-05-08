package persistence

import (
	"testing"

	"github.com/dxe/adb/model"
	"github.com/dxe/adb/pkg/shared"
	"github.com/dxe/adb/testdb"
	"github.com/stretchr/testify/require"
)

func TestGetUser_ByID(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	user, err := repo.GetUser(model.DevTestUserId, "")
	require.NoError(t, err)
	require.Equal(t, model.DevTestUserId, user.ID)
	require.Equal(t, model.DevTestUserEmail, user.Email)
	require.Equal(t, model.SFBayChapterIdDevTest, user.ChapterID)
}

func TestGetUser_ByEmail(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	user, err := repo.GetUser(0, model.DevTestUserEmail)
	require.NoError(t, err)
	require.Equal(t, model.DevTestUserId, user.ID)
	require.Equal(t, model.DevTestUserEmail, user.Email)
}

func TestGetUser_NotFound(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	_, err := repo.GetUser(0, "nobody@example.org")
	require.Error(t, err)
}

func TestGetUser_NoArgs(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	_, err := repo.GetUser(0, "")
	require.Error(t, err)
}

func TestGetUsers_All(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	users, err := repo.GetUsers(model.GetUserOptions{})
	require.NoError(t, err)
	found := false
	for _, user := range users {
		if user.ID == model.DevTestUserId {
			found = true
			break
		}
	}
	require.True(t, found)
}

func TestGetUsers_ByID(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	users, err := repo.GetUsers(model.GetUserOptions{ID: model.DevTestUserId})
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Equal(t, model.DevTestUserId, users[0].ID)
}

func TestGetUsers_PopulatesRoles(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	users, err := repo.GetUsers(model.GetUserOptions{ID: model.DevTestUserId, PopulateRoles: true})
	require.NoError(t, err)
	require.Len(t, users, 1)
	require.Contains(t, users[0].Roles, shared.RoleOrganizer)
}

func TestCreateUser(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	created, err := repo.CreateUser(model.ADBUser{
		Email:     "new@example.org",
		Name:      "New User",
		ChapterID: model.SFBayChapterIdDevTest,
		Roles:     []string{shared.RoleOrganizer},
	})
	require.NoError(t, err)
	require.NotZero(t, created.ID)
	require.Equal(t, "new@example.org", created.Email)
	require.Equal(t, "New User", created.Name)
	require.Equal(t, []string{shared.RoleOrganizer}, created.Roles)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	_, err := repo.CreateUser(model.ADBUser{
		Email:     model.DevTestUserEmail,
		Name:      "Dup",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.Error(t, err)
}

func TestCreateUser_NonzeroIDReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	_, err := repo.CreateUser(model.ADBUser{
		ID:        99,
		Email:     "x@example.org",
		Name:      "X",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.Error(t, err)
}

func TestUpdateUser(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	created, err := repo.CreateUser(model.ADBUser{
		Email:     "before@example.org",
		Name:      "Before",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.NoError(t, err)

	updated, err := repo.UpdateUser(model.ADBUser{
		ID:        created.ID,
		Email:     "after@example.org",
		Name:      "After",
		ChapterID: model.SFBayChapterIdDevTest,
		Roles:     []string{shared.RoleAdmin},
	})
	require.NoError(t, err)
	require.Equal(t, "after@example.org", updated.Email)
	require.Equal(t, "After", updated.Name)
	require.Equal(t, []string{shared.RoleAdmin}, updated.Roles)
}

func TestUpdateUser_EmailConflict(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	created, err := repo.CreateUser(model.ADBUser{
		Email:     "user1@example.org",
		Name:      "User One",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.NoError(t, err)

	_, err = repo.UpdateUser(model.ADBUser{
		ID:        created.ID,
		Email:     model.DevTestUserEmail,
		Name:      "User One",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.Error(t, err)
}

func TestUpdateUser_ZeroIDReturnsError(t *testing.T) {
	db := testdb.NewDB()
	defer db.Close()
	repo := NewUserRepository(db)

	_, err := repo.UpdateUser(model.ADBUser{
		ID:        0,
		Email:     "x@example.org",
		Name:      "X",
		ChapterID: model.SFBayChapterIdDevTest,
	})
	require.Error(t, err)
}
