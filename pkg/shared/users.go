package shared

const DevTestUserId = 1
const DevTestUserEmail = "test@example.org"

var allowedADBUserRoles = map[string]struct{}{
	"admin":            {},
	"organizer":        {},
	"attendance":       {},
	"intl_coordinator": {},
}

func IsAllowedADBUserRole(role string) bool {
	_, ok := allowedADBUserRoles[role]
	return ok
}
