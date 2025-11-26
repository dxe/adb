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

const DevTestUserId = 1
const DevTestUserEmail = "test@example.org"

func roleListHas(roles []string, target string) bool {
	for _, r := range roles {
		if r == target {
			return true
		}
	}
	return false
}
