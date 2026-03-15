package shared

import "sort"

const DevTestUserId = 1
const DevTestUserEmail = "test@example.org"

var allowedADBUserRolesSet = map[string]struct{}{
	"admin":            {},
	"organizer":        {},
	"attendance":       {},
	"intl_coordinator": {},
}

func AllowedADBUserRoles() []string {
	roles := make([]string, 0, len(allowedADBUserRolesSet))
	for role := range allowedADBUserRolesSet {
		roles = append(roles, role)
	}
	sort.Strings(roles)
	return roles
}

func IsAllowedADBUserRole(role string) bool {
	_, ok := allowedADBUserRolesSet[role]
	return ok
}
