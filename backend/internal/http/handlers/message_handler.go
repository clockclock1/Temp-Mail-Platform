package handlers

import (
	"net/http"
	"os"
	"strconv"

	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MessageHandler struct {
	db *gorm.DB
}

func NewMessageHandler(db *gorm.DB) *MessageHandler {
	return &MessageHandler{db: db}
}

func (h *MessageHandler) Get(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, _ := middleware.CurrentUser(c)

	var msg models.Message
	if err := h.db.Preload("Mailbox.Domain").First(&msg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	if !middleware.IsAdmin(user) && msg.Mailbox.OwnerID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item": msg})
}

func (h *MessageHandler) Delete(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, _ := middleware.CurrentUser(c)

	var msg models.Message
	if err := h.db.Preload("Mailbox").First(&msg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	if !middleware.IsAdmin(user) && msg.Mailbox.OwnerID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	_ = os.Remove(msg.RawPath)
	if err := h.db.Delete(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func (h *MessageHandler) Raw(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	user, _ := middleware.CurrentUser(c)

	var msg models.Message
	if err := h.db.Preload("Mailbox").First(&msg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "message not found"})
		return
	}
	if !middleware.IsAdmin(user) && msg.Mailbox.OwnerID != user.ID {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}

	data, err := os.ReadFile(msg.RawPath)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read raw email"})
		return
	}
	c.Data(http.StatusOK, "message/rfc822", data)
}
