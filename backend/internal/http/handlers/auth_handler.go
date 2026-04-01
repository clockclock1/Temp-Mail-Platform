package handlers

import (
	"net/http"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db         *gorm.DB
	jwtManager *auth.JWTManager
}

func NewAuthHandler(db *gorm.DB, jwtManager *auth.JWTManager) *AuthHandler {
	return &AuthHandler{db: db, jwtManager: jwtManager}
}

type loginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var user models.User
	if err := h.db.Preload("Role.Permissions").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !user.Active {
		c.JSON(http.StatusForbidden, gin.H{"error": "account disabled"})
		return
	}
	if !auth.VerifyPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	perms := util.PermissionKeys(user)
	token, err := h.jwtManager.GenerateToken(user.ID, user.Username, user.Role.Name, perms)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user":  user,
		"perms": perms,
	})
}

func (h *AuthHandler) Me(c *gin.Context) {
	user, ok := middleware.CurrentUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"user": user, "perms": util.PermissionKeys(user)})
}
