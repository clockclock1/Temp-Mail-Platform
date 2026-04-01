package handlers

import (
	"crypto/subtle"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/config"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/service"
	"tempmail/backend/internal/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type LegacyHandler struct {
	db      *gorm.DB
	mailSvc *service.MailService
	userJWT *auth.JWTManager
	addrJWT *auth.AddressJWTManager
	cfg     config.Config
}

func NewLegacyHandler(cfg config.Config, db *gorm.DB, mailSvc *service.MailService, userJWT *auth.JWTManager) *LegacyHandler {
	return &LegacyHandler{
		db:      db,
		mailSvc: mailSvc,
		userJWT: userJWT,
		addrJWT: auth.NewAddressJWTManager(cfg.JWTSecret, cfg.LegacyAddrExpire),
		cfg:     cfg,
	}
}

type legacyNewAddressRequest struct {
	EnablePrefix bool   `json:"enablePrefix"`
	Name         string `json:"name" binding:"required"`
	Domain       string `json:"domain" binding:"required"`
	TTLHours     int    `json:"ttlHours"`
}

type legacyUserAuthRequest struct {
	Username    string `json:"username" binding:"required"`
	Password    string `json:"password" binding:"required"`
	DisplayName string `json:"displayName"`
}

func (h *LegacyHandler) AdminNewAddress(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	owner, ok := h.ensureAdmin(c)
	if !ok {
		return
	}
	var req legacyNewAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	h.createAddress(c, req, owner, true)
}

func (h *LegacyHandler) APINewAddress(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	var req legacyNewAddressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if owner, ok := h.tryAdmin(c); ok {
		h.createAddress(c, req, owner, true)
		return
	}

	user, ok := h.ensureUser(c)
	if !ok {
		return
	}
	if !util.HasPermission(util.PermissionKeys(user), models.PermMailboxCreate) && user.Role.Name != "admin" {
		c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
		return
	}
	h.createAddress(c, req, user, false)
}

func (h *LegacyHandler) APIMails(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	claims, ok := h.ensureAddressToken(c)
	if !ok {
		return
	}

	var mailbox models.Mailbox
	if err := h.db.Preload("Domain").Where("id = ?", claims.MailboxID).First(&mailbox).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid mailbox token"})
		return
	}

	limit, offset := parsePage(c)
	items, total, err := h.listMessages(
		h.db.Where("messages.mailbox_id = ?", mailbox.ID),
		limit,
		offset,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"data":   items,
		"limit":  limit,
		"offset": offset,
		"total":  total,
	})
}

func (h *LegacyHandler) AdminMails(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	if _, ok := h.ensureAdmin(c); !ok {
		return
	}

	limit, offset := parsePage(c)
	query := h.db.Model(&models.Message{})
	if address := strings.TrimSpace(c.Query("address")); address != "" {
		mailbox, err := h.findMailboxByAddress(address)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"data": []any{}, "limit": limit, "offset": offset, "total": 0})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address"})
			return
		}
		query = query.Where("messages.mailbox_id = ?", mailbox.ID)
	}

	items, total, err := h.listMessages(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "limit": limit, "offset": offset, "total": total})
}

func (h *LegacyHandler) UserMails(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	user, ok := h.ensureUser(c)
	if !ok {
		return
	}

	limit, offset := parsePage(c)
	query := h.db.Model(&models.Message{}).Joins("JOIN mailboxes ON mailboxes.id = messages.mailbox_id")
	if user.Role.Name != "admin" {
		query = query.Where("mailboxes.owner_id = ?", user.ID)
	}

	if address := strings.TrimSpace(c.Query("address")); address != "" {
		mailbox, err := h.findMailboxByAddress(address)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusOK, gin.H{"data": []any{}, "limit": limit, "offset": offset, "total": 0})
				return
			}
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid address"})
			return
		}
		if user.Role.Name != "admin" && mailbox.OwnerID != user.ID {
			c.JSON(http.StatusForbidden, gin.H{"error": "permission denied"})
			return
		}
		query = query.Where("messages.mailbox_id = ?", mailbox.ID)
	}

	items, total, err := h.listMessages(query, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": items, "limit": limit, "offset": offset, "total": total})
}

func (h *LegacyHandler) AdminDeleteMail(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	if _, ok := h.ensureAdmin(c); !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var msg models.Message
	if err := h.db.First(&msg, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "mail not found"})
		return
	}
	_ = os.Remove(msg.RawPath)
	if err := h.db.Delete(&msg).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "id": msg.ID})
}

func (h *LegacyHandler) AdminDeleteAddress(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	if _, ok := h.ensureAdmin(c); !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	var mailbox models.Mailbox
	if err := h.db.First(&mailbox, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}
	deleted, err := h.deleteMailboxMessages(mailbox.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Delete(&mailbox).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "deletedMails": deleted, "id": mailbox.ID})
}

func (h *LegacyHandler) AdminClearInbox(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	if _, ok := h.ensureAdmin(c); !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.ensureMailboxExists(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}
	deleted, err := h.deleteMailboxMessages(uint(id))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "deletedMails": deleted, "id": id})
}

func (h *LegacyHandler) AdminClearSentItems(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	if _, ok := h.ensureAdmin(c); !ok {
		return
	}
	id, _ := strconv.Atoi(c.Param("id"))
	if err := h.ensureMailboxExists(uint(id)); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "address not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "deletedMails": 0, "id": id, "note": "sent items are not stored by this service"})
}

func (h *LegacyHandler) UserLogin(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	var req legacyUserAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user models.User
	if err := h.db.Preload("Role.Permissions").Where("username = ?", req.Username).First(&user).Error; err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	if !user.Active || !auth.VerifyPassword(user.PasswordHash, req.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}
	token, err := h.userJWT.GenerateToken(user.ID, user.Username, user.Role.Name, util.PermissionKeys(user))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jwt": token, "token": token, "user": user, "perms": util.PermissionKeys(user)})
}

func (h *LegacyHandler) UserRegister(c *gin.Context) {
	if !h.ensureCustomAuth(c) {
		return
	}
	var req legacyUserAuthRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if len(strings.TrimSpace(req.Password)) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "password must be at least 8 chars"})
		return
	}

	var role models.Role
	if err := h.db.Preload("Permissions").Where("name = ?", "user").First(&role).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "default user role missing"})
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
		RoleID:       role.ID,
		Role:         role,
	}
	if err := h.db.Create(&user).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := h.db.Preload("Role.Permissions").First(&user, user.ID).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	token, err := h.userJWT.GenerateToken(user.ID, user.Username, user.Role.Name, util.PermissionKeys(user))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"jwt": token, "token": token, "user": user, "perms": util.PermissionKeys(user)})
}

func (h *LegacyHandler) createAddress(c *gin.Context, req legacyNewAddressRequest, owner models.User, allowCreateDomain bool) {
	localPart := strings.ToLower(strings.TrimSpace(req.Name))
	if localPart == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "name is required"})
		return
	}
	if req.EnablePrefix {
		localPart = util.RandomAlphaNum(5) + "-" + localPart
	}
	req.TTLHours = normalizeTTL(req.TTLHours)

	domainName := strings.ToLower(strings.TrimSpace(req.Domain))
	domain, err := h.resolveDomain(domainName, allowCreateDomain, owner.ID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	mailbox, err := h.mailSvc.CreateMailbox(owner.ID, localPart, domain.ID, "legacy-api", req.TTLHours)
	if err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "exists") {
			localPart = fmt.Sprintf("%s-%s", localPart, util.RandomAlphaNum(4))
			mailbox, err = h.mailSvc.CreateMailbox(owner.ID, localPart, domain.ID, "legacy-api", req.TTLHours)
		}
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	}

	address := mailbox.LocalPart + "@" + mailbox.Domain.Name
	token, err := h.addrJWT.Generate(mailbox.ID, address)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue mailbox token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"id":      mailbox.ID,
		"address": address,
		"jwt":     token,
		"token":   token,
		"mailbox": mailbox,
	})
}

func (h *LegacyHandler) listMessages(base *gorm.DB, limit, offset int) ([]gin.H, int64, error) {
	query := base.Preload("Mailbox.Domain")
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var rows []models.Message
	if err := query.Order("messages.received_at desc").Offset(offset).Limit(limit).Find(&rows).Error; err != nil {
		return nil, 0, err
	}

	items := make([]gin.H, 0, len(rows))
	for _, msg := range rows {
		addr := msg.ToAddr
		if msg.Mailbox.LocalPart != "" && msg.Mailbox.Domain.Name != "" {
			addr = msg.Mailbox.LocalPart + "@" + msg.Mailbox.Domain.Name
		}
		raw := readRaw(msg.RawPath)
		items = append(items, gin.H{
			"id":         msg.ID,
			"mailboxId":  msg.MailboxID,
			"address":    addr,
			"from":       msg.FromAddr,
			"to":         msg.ToAddr,
			"subject":    msg.Subject,
			"text":       msg.TextBody,
			"html":       msg.HTMLBody,
			"raw":        raw,
			"source":     raw,
			"size":       msg.Size,
			"createdAt":  msg.ReceivedAt.Unix(),
			"receivedAt": msg.ReceivedAt,
		})
	}
	return items, total, nil
}

func (h *LegacyHandler) resolveDomain(name string, allowCreate bool, creatorID uint) (*models.Domain, error) {
	var domain models.Domain
	err := h.db.Where("name = ?", name).First(&domain).Error
	if err == nil {
		if !domain.Enabled {
			return nil, fmt.Errorf("domain disabled")
		}
		return &domain, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}
	if !allowCreate {
		return nil, fmt.Errorf("domain not found")
	}
	domain = models.Domain{Name: name, Enabled: true, CreatedBy: creatorID}
	if err := h.db.Create(&domain).Error; err != nil {
		return nil, err
	}
	return &domain, nil
}

func (h *LegacyHandler) ensureCustomAuth(c *gin.Context) bool {
	if strings.TrimSpace(h.cfg.LegacyCustomAuth) == "" {
		return true
	}
	v := c.GetHeader("x-custom-auth")
	if subtle.ConstantTimeCompare([]byte(v), []byte(h.cfg.LegacyCustomAuth)) != 1 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid x-custom-auth"})
		return false
	}
	return true
}

func (h *LegacyHandler) ensureAdmin(c *gin.Context) (models.User, bool) {
	if u, ok := h.tryAdmin(c); ok {
		return u, true
	}
	c.JSON(http.StatusUnauthorized, gin.H{"error": "admin auth required"})
	return models.User{}, false
}

func (h *LegacyHandler) tryAdmin(c *gin.Context) (models.User, bool) {
	adminHeader := c.GetHeader("x-admin-auth")
	if adminHeader != "" && subtle.ConstantTimeCompare([]byte(adminHeader), []byte(h.cfg.LegacyAdminAuth)) == 1 {
		u, err := h.firstAdminUser()
		if err == nil {
			return u, true
		}
	}
	if tok := bearerToken(c.GetHeader("Authorization")); tok != "" {
		if u, err := h.userByToken(tok); err == nil && u.Role.Name == "admin" {
			return u, true
		}
	}
	return models.User{}, false
}

func (h *LegacyHandler) ensureUser(c *gin.Context) (models.User, bool) {
	tok := strings.TrimSpace(c.GetHeader("x-user-token"))
	if tok == "" {
		tok = bearerToken(c.GetHeader("Authorization"))
	}
	if tok == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing user token"})
		return models.User{}, false
	}
	u, err := h.userByToken(tok)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid user token"})
		return models.User{}, false
	}
	return u, true
}

func (h *LegacyHandler) ensureAddressToken(c *gin.Context) (*auth.AddressClaims, bool) {
	tok := bearerToken(c.GetHeader("Authorization"))
	if tok == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing address token"})
		return nil, false
	}
	claims, err := h.addrJWT.Parse(tok)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid address token"})
		return nil, false
	}
	return claims, true
}

func (h *LegacyHandler) userByToken(token string) (models.User, error) {
	claims, err := h.userJWT.ParseToken(token)
	if err != nil {
		return models.User{}, err
	}
	var user models.User
	if err := h.db.Preload("Role.Permissions").Where("id = ? AND active = ?", claims.UserID, true).First(&user).Error; err != nil {
		return models.User{}, err
	}
	return user, nil
}

func (h *LegacyHandler) firstAdminUser() (models.User, error) {
	var user models.User
	err := h.db.Preload("Role.Permissions").Joins("Role").Where("Role.name = ?", "admin").Where("users.active = ?", true).First(&user).Error
	return user, err
}

func (h *LegacyHandler) findMailboxByAddress(address string) (*models.Mailbox, error) {
	local, domain, err := util.SplitEmailAddress(address)
	if err != nil {
		return nil, err
	}
	var mailbox models.Mailbox
	if err := h.db.Joins("Domain").Where("mailboxes.local_part = ?", local).Where("Domain.name = ?", domain).Preload("Domain").First(&mailbox).Error; err != nil {
		return nil, err
	}
	return &mailbox, nil
}

func (h *LegacyHandler) deleteMailboxMessages(mailboxID uint) (int64, error) {
	var msgs []models.Message
	if err := h.db.Where("mailbox_id = ?", mailboxID).Find(&msgs).Error; err != nil {
		return 0, err
	}
	for _, m := range msgs {
		_ = os.Remove(m.RawPath)
	}
	if err := h.db.Delete(&models.Message{}, "mailbox_id = ?", mailboxID).Error; err != nil {
		return 0, err
	}
	return int64(len(msgs)), nil
}

func (h *LegacyHandler) ensureMailboxExists(id uint) error {
	var mailbox models.Mailbox
	return h.db.First(&mailbox, id).Error
}

func parsePage(c *gin.Context) (int, int) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}
	return limit, offset
}

func bearerToken(header string) string {
	parts := strings.SplitN(strings.TrimSpace(header), " ", 2)
	if len(parts) == 2 && strings.EqualFold(parts[0], "Bearer") {
		return strings.TrimSpace(parts[1])
	}
	return ""
}

func readRaw(path string) string {
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	const maxRaw = 2 * 1024 * 1024
	if len(b) > maxRaw {
		b = b[:maxRaw]
	}
	return string(b)
}

func normalizeTTL(ttl int) int {
	if ttl <= 0 {
		return 24
	}
	if ttl > 24*30 {
		return 24 * 30
	}
	return ttl
}
