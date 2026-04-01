package util

import "tempmail/backend/internal/models"

func PermissionKeys(user models.User) []string {
	keys := make([]string, 0, len(user.Role.Permissions))
	for _, p := range user.Role.Permissions {
		keys = append(keys, p.Key)
	}
	return keys
}

func HasPermission(keys []string, target string) bool {
	for _, k := range keys {
		if k == target {
			return true
		}
	}
	return false
}
