package middleware

import (
	"net/http"
	"strings"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const userContextKey = "auth_user"

func AuthRequired(db *gorm.DB, jwtManager *auth.JWTManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing authorization header"})
			c.Abort()
			return
		}
		parts := strings.SplitN(header, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid authorization header"})
			c.Abort()
			return
		}

		claims, err := jwtManager.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
			c.Abort()
			return
		}

		var user models.User
		if err := db.Preload("Role.Permissions").Where("id = ? AND active = ?", claims.UserID, true).First(&user).Error; err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "user not found or inactive"})
			c.Abort()
			return
		}

		c.Set(userContextKey, user)
		c.Next()
	}
}

func CurrentUser(c *gin.Context) (models.User, bool) {
	user, ok := c.Get(userContextKey)
	if !ok {
		return models.User{}, false
	}
	u, ok := user.(models.User)
	return u, ok
}
