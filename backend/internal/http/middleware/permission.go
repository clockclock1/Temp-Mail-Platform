package middleware

import (
	"net/http"

	"tempmail/backend/internal/models"
	"tempmail/backend/internal/util"

	"github.com/gin-gonic/gin"
)

func RequirePermission(permission string) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, ok := CurrentUser(c)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
			c.Abort()
			return
		}
		if user.Role.Name == "admin" {
			c.Next()
			return
		}
		if !util.HasPermission(util.PermissionKeys(user), permission) {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied", "required": permission})
			c.Abort()
			return
		}
		c.Next()
	}
}

func IsAdmin(user models.User) bool {
	return user.Role.Name == "admin"
}
