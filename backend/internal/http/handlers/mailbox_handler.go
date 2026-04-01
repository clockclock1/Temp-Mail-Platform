package handlers

import (
	"net/http"
	"strconv"
	"strings"
	"time"

	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MailboxHandler struct {
	db          *gorm.DB
	mailService *service.MailService
}

func NewMailboxHandler(db *gorm.DB, mailService *service.MailService) *MailboxHandler {
	return &MailboxHandler{db: db, mailService: mailService}
}

type createMailboxRequest struct {
	LocalPart   string `json:"localPart" binding:"required"`
	DomainID    uint   `json:"domainId" binding:"required"`
	Description string `json:"description"`
	TTLHours    int    `json:"ttlHours"`
}

func (h *MailboxHandler) List(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)
	query := h.db.Preload("Domain").Preload("Owner.Role")

	if !middleware.IsAdmin(user) {
		query = query.Where("owner_id = ?", user.ID)
	}
	if v := strings.TrimSpace(c.Query("ownerId")); v != "" && middleware.IsAdmin(user) {
		if ownerID, err := strconv.Atoi(v); err == nil {
			query = query.Where("owner_id = ?", ownerID)
		}
	}
	if v := strings.TrimSpace(c.Query("active")); v != "" {
		if active, err := strconv.ParseBool(v); err == nil {
			query = query.Where("enabled = ?", active)
		}
	}

	var items []models.Mailbox
	if err := query.Order("id desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": h.decorateMailboxes(items)})
}

func (h *MailboxHandler) Create(c *gin.Context) {
	user, _ := middleware.CurrentUser(c)

	var req createMailboxRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.TTLHours <= 0 {
		req.TTLHours = 24
	}
	if req.TTLHours > 24*30 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ttlHours too large"})
		return
	}

	mailbox, err := h.mailService.CreateMailbox(user.ID, req.LocalPart, req.DomainID, req.Description, req.TTLHours)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"item": h.decorateMailbox(*mailbox)})
}

func (h *MailboxHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, _ := middleware.CurrentUser(c)

	var mailbox models.Mailbox
	if err := h.db.Preload("Domain").First(&mailbox, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}
	if !middleware.IsAdmin(user) && mailbox.OwnerID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	if err := h.db.Delete(&models.Message{}, "mailbox_id = ?", mailbox.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Delete(&mailbox).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusNoContent)
}

func (h *MailboxHandler) Messages(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, _ := middleware.CurrentUser(c)

	var mailbox models.Mailbox
	if err := h.db.Preload("Domain").First(&mailbox, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mailbox not found"})
		return
	}
	if !middleware.IsAdmin(user) && mailbox.OwnerID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	var items []models.Message
	if err := h.db.Where("mailbox_id = ?", mailbox.ID).Order("received_at desc").Find(&items).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}

func (h *MailboxHandler) decorateMailbox(m models.Mailbox) gin.H {
	address := m.LocalPart
	if m.Domain.Name != "" {
		address += "@" + m.Domain.Name
	}
	remaining := int64(-1)
	if m.ExpiresAt != nil {
		remaining = int64(m.ExpiresAt.Sub(time.Now()).Seconds())
	}
	return gin.H{
		"id":               m.ID,
		"localPart":        m.LocalPart,
		"domainId":         m.DomainID,
		"domain":           m.Domain,
		"ownerId":          m.OwnerID,
		"description":      m.Description,
		"enabled":          m.Enabled,
		"expiresAt":        m.ExpiresAt,
		"createdAt":        m.CreatedAt,
		"updatedAt":        m.UpdatedAt,
		"lastReceived":     m.LastReceived,
		"address":          address,
		"remainingSeconds": remaining,
	}
}

func (h *MailboxHandler) decorateMailboxes(items []models.Mailbox) []gin.H {
	out := make([]gin.H, 0, len(items))
	for _, item := range items {
		out = append(out, h.decorateMailbox(item))
	}
	return out
}
