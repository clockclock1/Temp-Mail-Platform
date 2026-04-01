package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DomainHandler struct {
	db *gorm.DB
}

func NewDomainHandler(db *gorm.DB) *DomainHandler {
	return &DomainHandler{db: db}
}

type domainRequest struct {
	Name    string `json:"name" binding:"required"`
	Enabled *bool  `json:"enabled"`
}

func (h *DomainHandler) List(c *gin.Context) {
	var domains []models.Domain
	if err := h.db.Order("id desc").Find(&domains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": domains})
}

func (h *DomainHandler) Available(c *gin.Context) {
	var domains []models.Domain
	if err := h.db.Where("enabled = ?", true).Order("id desc").Find(&domains).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": domains})
}

func (h *DomainHandler) Create(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)

	var req domainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	domain := models.Domain{
		Name:      strings.ToLower(strings.TrimSpace(req.Name)),
		Enabled:   true,
		CreatedBy: user.ID,
	}
	if req.Enabled != nil {
		domain.Enabled = *req.Enabled
	}

	if err := h.db.Create(&domain).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, domain)
}

func (h *DomainHandler) Update(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))

	var domain models.Domain
	if err := h.db.First(&domain, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "domain not found"})
		return
	}

	var req domainRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Name != "" {
		domain.Name = strings.ToLower(strings.TrimSpace(req.Name))
	}
	if req.Enabled != nil {
		domain.Enabled = *req.Enabled
	}

	if err := h.db.Save(&domain).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, domain)
}

func (h *DomainHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Delete(&models.Domain{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
