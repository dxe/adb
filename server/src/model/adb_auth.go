package model

import (
	"github.com/dxe/adb/config"
	"github.com/dxe/adb/pkg/shared"
	"github.com/pkg/errors"
)

var (
	attendanceAccessRoles = []string{shared.RoleAdmin, shared.RoleOrganizer, shared.RoleAttendance}
	organizerAccessRoles  = []string{shared.RoleAdmin, shared.RoleOrganizer}
	intlCoordinatorRoles  = []string{shared.RoleAdmin, shared.RoleIntlCoordinator}
)

type ADBUser struct {
	ID       int    `db:"id"`
	Email    string `db:"email"`
	Name     string `db:"name"`
	Disabled bool   `db:"disabled"`
	Roles    []string
	// Chapter ID of the user. When loaded directly from the database or used in
	// API payloads, this is the user's assigned chapter. In the context of a live
	// authenticated ADB session for an admin user, it can be overridden by the
	// chapter selected in the auth session cookie.
	ChapterID   int    `db:"chapter_id"`
	ChapterName string `db:"chapter_name"`
}

type UserRole struct {
	UserID int    `db:"user_id"`
	Role   string `db:"role"`
}

type GetUserOptions struct {
	ID            int
	Name          string
	PopulateRoles bool
}

func ValidateADBUser(user ADBUser) error {
	if user.Email == "" {
		return errors.New("Email cannot be empty")
	}

	if user.Name == "" {
		return errors.New("Name cannot be empty")
	}

	if user.ChapterID == 0 {
		return errors.New("Chapter must not be 0")
	}

	for _, role := range user.Roles {
		if !shared.IsAllowedADBUserRole(role) {
			return errors.Errorf("Invalid role: %s", role)
		}
	}

	return nil
}

func IsSFBayChapterID(chapterID int) bool {
	if !config.IsProd {
		return chapterID == SFBayChapterIdDevTest
	}

	return chapterID == SFBayChapterId
}

func UserHasADBAccess(user ADBUser) bool {
	for _, role := range user.Roles {
		if shared.IsAllowedADBUserRole(role) {
			return true
		}
	}
	return false
}

func UserHasAttendanceAccess(user ADBUser) bool {
	return UserHasAnyRole(attendanceAccessRoles, user)
}

func UserHasOrganizerAccess(user ADBUser) bool {
	return UserHasAnyRole(organizerAccessRoles, user)
}

func UserHasSFBayOrganizerAccess(user ADBUser) bool {
	if UserHasRole(shared.RoleAdmin, user) {
		return true
	}

	return UserHasRole(shared.RoleOrganizer, user) && IsSFBayChapterID(user.ChapterID)
}

func UserHasIntlCoordinatorAccess(user ADBUser) bool {
	return UserHasAnyRole(intlCoordinatorRoles, user)
}

func UserHasAnyRole(roles []string, user ADBUser) bool {
	for i := 0; i < len(roles); i++ {
		if UserHasRole(roles[i], user) {
			return true
		}
	}

	return false
}

func UserHasRole(role string, user ADBUser) bool {
	for _, r := range user.Roles {
		if r == role {
			return true
		}
	}

	return false
}

// Interface for querying and updating users. This avoids a dependency on the persistence package which could create a
// cyclical package reference.
type UserRepository interface {
	GetUser(id int, email string) (ADBUser, error)
	GetUsers(options GetUserOptions) ([]ADBUser, error)
	CreateUser(user ADBUser) (ADBUser, error)
	UpdateUser(user ADBUser) (ADBUser, error)
}

const DevTestUserId = shared.DevTestUserId
const DevTestUserEmail = shared.DevTestUserEmail
