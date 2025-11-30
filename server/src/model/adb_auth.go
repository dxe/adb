package model

import (
	"github.com/pkg/errors"
)

type ADBUser struct {
	ID          int    `db:"id"`
	Email       string `db:"email"`
	Name        string `db:"name"`
	Disabled    bool   `db:"disabled"`
	Roles       []string
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

var allowedUserRoles = map[string]struct{}{
	"admin":      {},
	"organizer":  {},
	"attendance": {},
	"non-sfbay":  {},
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
		if _, ok := allowedUserRoles[role]; !ok {
			return errors.Errorf("Invalid role: %s", role)
		}
	}

	return nil
}

// Interface for querying and updating users. This avoids a dependency on the persistence package which could create a
// cyclical package reference.
type UserRepository interface {
	GetUser(id int, email string) (ADBUser, error)
	GetUsers(options GetUserOptions) ([]ADBUser, error)
	CreateUser(user ADBUser) (ADBUser, error)
	UpdateUser(user ADBUser) (ADBUser, error)
}

const DevTestUserId = 1
const DevTestUserEmail = "test@example.org"
