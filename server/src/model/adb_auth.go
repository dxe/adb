package model

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
