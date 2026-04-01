package handlers

import (
	"net/http"
	"time"

	"tempmail/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type StatsHandler struct {
	db *gorm.DB
}

func NewStatsHandler(db *gorm.DB) *StatsHandler {
	return &StatsHandler{db: db}
}

func (h *StatsHandler) Get(c *gin.Context) {
	var users, domains, mailboxes, messages int64
	_ = h.db.Model(&models.User{}).Count(&users).Error
	_ = h.db.Model(&models.Domain{}).Count(&domains).Error
	_ = h.db.Model(&models.Mailbox{}).Count(&mailboxes).Error
	_ = h.db.Model(&models.Message{}).Count(&messages).Error

	var messages24h int64
	_ = h.db.Model(&models.Message{}).Where("received_at >= ?", time.Now().Add(-24*time.Hour)).Count(&messages24h).Error

	c.JSON(http.StatusOK, gin.H{
		"users":               users,
		"domains":             domains,
		"mailboxes":           mailboxes,
		"messages":            messages,
		"messagesLast24Hours": messages24h,
	})
}
