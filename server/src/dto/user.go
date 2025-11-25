package dto

import "github.com/dxe/adb/model"

// User is the transport-level representation used in HTTP responses.
type User struct {
	ID        int      `json:"id"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Admin     bool     `json:"admin"`
	Disabled  bool     `json:"disabled"`
	Roles     []string `json:"roles"`
	ChapterID int      `json:"chapter_id"`
}

func FromModel(u model.ADBUser) User {
	var roles []string
	for _, r := range u.Roles {
		roles = append(roles, r.Role)
	}

	return User{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Admin:     u.Admin,
		Disabled:  u.Disabled,
		Roles:     roles,
		ChapterID: u.ChapterID,
	}
}

func FromModels(users []model.ADBUser) []User {
	out := make([]User, 0, len(users))
	for _, u := range users {
		out = append(out, FromModel(u))
	}
	return out
}
