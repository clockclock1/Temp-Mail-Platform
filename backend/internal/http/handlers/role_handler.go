package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"tempmail/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type RoleHandler struct {
	db *gorm.DB
}

func NewRoleHandler(db *gorm.DB) *RoleHandler {
	return &RoleHandler{db: db}
}

type roleRequest struct {
	Name           string   `json:"name" binding:"required"`
	Description    string   `json:"description"`
	PermissionKeys []string `json:"permissionKeys"`
}

func (h *RoleHandler) List(c *gin.Context) {
	var roles []models.Role
	if err := h.db.Preload("Permissions").Order("id asc").Find(&roles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	type roleSummary struct {
		models.Role
		UserCount int64 `json:"userCount"`
	}
	items := make([]roleSummary, 0, len(roles))
	for _, role := range roles {
		var count int64
		if err := h.db.Model(&models.User{}).Where("role_id = ?", role.ID).Count(&count).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		items = append(items, roleSummary{Role: role, UserCount: count})
	}

	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *RoleHandler) Create(c *gin.Context) {
	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	perms, err := h.findPermissionsByKeys(req.PermissionKeys)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role := models.Role{Name: strings.TrimSpace(req.Name), Description: strings.TrimSpace(req.Description)}
	if err := h.db.Create(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Model(&role).Association("Permissions").Replace(&perms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Preload("Permissions").First(&role, role.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, role)
}

func (h *RoleHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var role models.Role
	if err := h.db.Preload("Permissions").First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.Name == "admin" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "admin role cannot be modified"})
		return
	}

	var req roleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	perms, err := h.findPermissionsByKeys(req.PermissionKeys)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	role.Name = strings.TrimSpace(req.Name)
	role.Description = strings.TrimSpace(req.Description)
	if err := h.db.Save(&role).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Model(&role).Association("Permissions").Replace(&perms); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Preload("Permissions").First(&role, role.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, role)
}

func (h *RoleHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	var role models.Role
	if err := h.db.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}
	if role.Name == "admin" || role.Name == "user" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "built-in role cannot be deleted"})
		return
	}

	var count int64
	if err := h.db.Model(&models.User{}).Where("role_id = ?", role.ID).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if count > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "role is used by users"})
		return
	}

	if err := h.db.Delete(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *RoleHandler) Permissions(c *gin.Context) {
	var perms []models.Permission
	if err := h.db.Order("id asc").Find(&perms).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": perms})
}

func (h *RoleHandler) Users(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var role models.Role
	if err := h.db.First(&role, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "role not found"})
		return
	}

	type item struct {
		ID          uint   `json:"id"`
		Username    string `json:"username"`
		DisplayName string `json:"displayName"`
		Active      bool   `json:"active"`
	}
	var users []item
	if err := h.db.Model(&models.User{}).
		Where("role_id = ?", role.ID).
		Select("id, username, display_name, active").
		Order("id desc").
		Find(&users).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": users})
}

func (h *RoleHandler) findPermissionsByKeys(keys []string) ([]models.Permission, error) {
	if len(keys) == 0 {
		return []models.Permission{}, nil
	}

	trimmed := make([]string, 0, len(keys))
	for _, k := range keys {
		k = strings.TrimSpace(k)
		if k != "" {
			trimmed = append(trimmed, k)
		}
	}
	if len(trimmed) == 0 {
		return []models.Permission{}, nil
	}

	var perms []models.Permission
	if err := h.db.Where("key IN ?", trimmed).Find(&perms).Error; err != nil {
		return nil, err
	}
	if len(perms) != len(trimmed) {
		return nil, fmt.Errorf("some permission keys are invalid")
	}
	return perms, nil
}
