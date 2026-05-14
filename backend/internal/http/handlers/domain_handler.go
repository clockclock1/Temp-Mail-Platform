package handlers

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/util"

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
	Name        string `json:"name" binding:"required"`
	Enabled     *bool  `json:"enabled"`
	Level       *int   `json:"level"`
	RandomLevel *bool  `json:"randomLevel"`
	LevelMin    *int   `json:"levelMin"`
	LevelMax    *int   `json:"levelMax"`
}

type domainPushRequest struct {
	Names       []string `json:"names" binding:"required"`
	Enabled     *bool    `json:"enabled"`
	Level       *int     `json:"level"`
	RandomLevel *bool    `json:"randomLevel"`
	LevelMin    *int     `json:"levelMin"`
	LevelMax    *int     `json:"levelMax"`
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
	if domain.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "domain name cannot be empty"})
		return
	}
	h.applyDomainConfig(&domain, req)

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
		if domain.Name == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "domain name cannot be empty"})
			return
		}
		if req.Level == nil {
			domain.Level = util.DomainLevelFromName(domain.Name)
		}
	}
	h.applyDomainConfig(&domain, req)

	if err := h.db.Save(&domain).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, domain)
}

func (h *DomainHandler) Push(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)

	var req domainPushRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items := make([]models.Domain, 0, len(req.Names))
	created := 0
	updated := 0

	for _, raw := range req.Names {
		name := strings.ToLower(strings.TrimSpace(raw))
		if name == "" {
			continue
		}

		var domain models.Domain
		err := h.db.Where("name = ?", name).First(&domain).Error
		if err == nil {
			h.applyDomainConfig(&domain, domainRequest{
				Name:        name,
				Enabled:     req.Enabled,
				Level:       req.Level,
				RandomLevel: req.RandomLevel,
				LevelMin:    req.LevelMin,
				LevelMax:    req.LevelMax,
			})
			if err := h.db.Save(&domain).Error; err != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
				return
			}
			updated++
			items = append(items, domain)
			continue
		}
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		domain = models.Domain{
			Name:      name,
			Enabled:   true,
			CreatedBy: user.ID,
		}
		h.applyDomainConfig(&domain, domainRequest{
			Name:        name,
			Enabled:     req.Enabled,
			Level:       req.Level,
			RandomLevel: req.RandomLevel,
			LevelMin:    req.LevelMin,
			LevelMax:    req.LevelMax,
		})
		if err := h.db.Create(&domain).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		created++
		items = append(items, domain)
	}
	if len(items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "no valid domains to push"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items":   items,
		"created": created,
		"updated": updated,
	})
}

func (h *DomainHandler) applyDomainConfig(domain *models.Domain, req domainRequest) {
	if req.Enabled != nil {
		domain.Enabled = *req.Enabled
	}
	if req.RandomLevel != nil {
		domain.RandomLevel = *req.RandomLevel
	}

	if req.Level != nil {
		domain.Level = util.NormalizeDomainLevel(*req.Level)
	} else if domain.Level <= 0 {
		domain.Level = util.DomainLevelFromName(domain.Name)
	}

	if domain.RandomLevel {
		minLevel := domain.LevelMin
		maxLevel := domain.LevelMax
		if req.LevelMin != nil {
			minLevel = *req.LevelMin
		}
		if req.LevelMax != nil {
			maxLevel = *req.LevelMax
		}
		domain.LevelMin, domain.LevelMax = util.NormalizeRandomDomainLevelRange(minLevel, maxLevel)
	} else {
		domain.LevelMin = 0
		domain.LevelMax = 0
	}
}

func (h *DomainHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.db.Delete(&models.Domain{}, id).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}
