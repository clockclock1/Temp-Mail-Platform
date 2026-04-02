package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UserHandler struct {
	db *gorm.DB
}

func NewUserHandler(db *gorm.DB) *UserHandler {
	return &UserHandler{db: db}
}

type createUserRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required,min=8"`
	DisplayName string `json:"displayName"`
	RoleID      uint   `json:"roleId" binding:"required"`
}

type updateUserRequest struct {
	DisplayName *string `json:"displayName"`
	Password    *string `json:"password"`
	RoleID      *uint   `json:"roleId"`
	Active      *bool   `json:"active"`
}

func (h *UserHandler) List(c *gin.Context) {
	page, pageSize := parsePagination(c, 1, 20, 100)
	q := strings.TrimSpace(c.Query("q"))
	activeQuery := strings.TrimSpace(c.Query("active"))
	roleIDQuery := strings.TrimSpace(c.Query("roleId"))

	query := h.db.Model(&models.User{})
	if q != "" {
		like := "%" + q + "%"
		query = query.Where("username LIKE ? OR display_name LIKE ?", like, like)
	}
	if activeQuery != "" {
		if active, err := strconv.ParseBool(activeQuery); err == nil {
			query = query.Where("active = ?", active)
		}
	}
	if roleIDQuery != "" {
		if roleID, err := strconv.Atoi(roleIDQuery); err == nil {
			query = query.Where("role_id = ?", roleID)
		}
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var users []models.User
	if err := query.Preload("Role.Permissions").
		Order("id desc").
		Limit(pageSize).
		Offset((page - 1) * pageSize).
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"items":    users,
		"total":    total,
		"page":     page,
		"pageSize": pageSize,
	})
}

func (h *UserHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user models.User
	if err := h.db.Preload("Role.Permissions").First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"item": user})
}

func (h *UserHandler) Create(c *gin.Context) {
	var req createUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var role models.Role
	if err := h.db.First(&role, req.RoleID).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roleId"})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	user := models.User{
		Username:     strings.TrimSpace(req.Username),
		PasswordHash: hash,
		DisplayName:  strings.TrimSpace(req.DisplayName),
		Active:       true,
		RoleID:       req.RoleID,
	}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Preload("Role.Permissions").First(&user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (h *UserHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	currentUser, _ := middleware.CurrentUser(c)
	if currentUser.ID == user.ID {
		if active := c.Query("force"); active != "true" {
			var req updateUserRequest
			if err := c.ShouldBindJSON(&req); err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			if req.Active != nil && !*req.Active {
				c.JSON(http.StatusBadRequest, gin.H{"error": "cannot disable yourself without force=true"})
				return
			}
			if req.RoleID != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": "cannot change your own role without force=true"})
				return
			}
			if req.DisplayName != nil {
				user.DisplayName = strings.TrimSpace(*req.DisplayName)
			}
			if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
				hash, err := auth.HashPassword(*req.Password)
				if err != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
					return
				}
				user.PasswordHash = hash
			}
			if err := h.db.Save(&user).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			if err := h.db.Preload("Role.Permissions").First(&user, user.ID).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
				return
			}
			c.JSON(http.StatusOK, user)
			return
		}
	}

	var req updateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*req.DisplayName)
	}
	if req.Password != nil && strings.TrimSpace(*req.Password) != "" {
		hash, err := auth.HashPassword(*req.Password)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
			return
		}
		user.PasswordHash = hash
	}
	if req.Active != nil {
		user.Active = *req.Active
	}
	if req.RoleID != nil {
		var role models.Role
		if err := h.db.First(&role, *req.RoleID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid roleId"})
			return
		}
		user.RoleID = *req.RoleID
	}

	if err := h.db.Save(&user).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Preload("Role.Permissions").First(&user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (h *UserHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	current, _ := middleware.CurrentUser(c)
	if uint(id) == current.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "cannot delete yourself"})
		return
	}
	if err := h.db.Delete(&models.User{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

type resetPasswordRequest struct {
	Password string `json:"password"`
}

func (h *UserHandler) ResetPassword(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var user models.User
	if err := h.db.First(&user, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil && !strings.Contains(err.Error(), "EOF") {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	password := strings.TrimSpace(req.Password)
	autoGenerated := false
	if password == "" {
		password = util.RandomAlphaNum(14)
		autoGenerated = true
	}
	if len(password) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 chars"})
		return
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to hash password"})
		return
	}

	if err := h.db.Model(&user).Update("password_hash", hash).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp := gin.H{
		"success":       true,
		"userId":        user.ID,
		"username":      user.Username,
		"autoGenerated": autoGenerated,
	}
	if autoGenerated {
		resp["password"] = password
	}
	c.JSON(http.StatusOK, resp)
}

func parsePagination(c *gin.Context, defaultPage, defaultPageSize, maxPageSize int) (int, int) {
	page := defaultPage
	pageSize := defaultPageSize

	if v := strings.TrimSpace(c.Query("page")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			page = parsed
		}
	}
	if v := strings.TrimSpace(c.Query("pageSize")); v != "" {
		if parsed, err := strconv.Atoi(v); err == nil && parsed > 0 {
			pageSize = parsed
		}
	}

	if pageSize > maxPageSize {
		pageSize = maxPageSize
	}
	return page, pageSize
}
