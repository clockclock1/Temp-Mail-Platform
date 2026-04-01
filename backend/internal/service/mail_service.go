package service

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	stdmail "net/mail"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"tempmail/backend/internal/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type MailService struct {
	db      *gorm.DB
	mu      sync.RWMutex
	dataDir string
}

func NewMailService(db *gorm.DB, dataDir string) *MailService {
	return &MailService{db: db, dataDir: dataDir}
}

func (s *MailService) DataDir() string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.dataDir
}

func (s *MailService) UpdateDataDir(dir string) {
	s.mu.Lock()
	s.dataDir = strings.TrimSpace(dir)
	s.mu.Unlock()
}

func (s *MailService) DB() *gorm.DB {
	return s.db
}

func (s *MailService) CreateMailbox(ownerID uint, localPart string, domainID uint, description string, ttlHours int) (*models.Mailbox, error) {
	localPart = strings.ToLower(strings.TrimSpace(localPart))
	if localPart == "" {
		return nil, fmt.Errorf("local part cannot be empty")
	}

	var domain models.Domain
	if err := s.db.Where("id = ? AND enabled = ?", domainID, true).First(&domain).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("domain not found or disabled")
		}
		return nil, err
	}

	var expiresAt *time.Time
	if ttlHours > 0 {
		t := time.Now().Add(time.Duration(ttlHours) * time.Hour)
		expiresAt = &t
	}

	mailbox := models.Mailbox{
		LocalPart:   localPart,
		DomainID:    domain.ID,
		OwnerID:     ownerID,
		Description: description,
		Enabled:     true,
		ExpiresAt:   expiresAt,
	}
	if err := s.db.Create(&mailbox).Error; err != nil {
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return nil, fmt.Errorf("mailbox already exists")
		}
		return nil, err
	}
	if err := s.db.Preload("Domain").First(&mailbox, mailbox.ID).Error; err != nil {
		return nil, err
	}
	return &mailbox, nil
}

func (s *MailService) FindMailboxByAddress(localPart, domainName string) (*models.Mailbox, error) {
	localPart = strings.ToLower(strings.TrimSpace(localPart))
	domainName = strings.ToLower(strings.TrimSpace(domainName))

	var mailbox models.Mailbox
	err := s.db.Joins("Domain").
		Where("mailboxes.local_part = ?", localPart).
		Where("Domain.name = ?", domainName).
		Where("mailboxes.enabled = ?", true).
		Where("Domain.enabled = ?", true).
		Preload("Domain").
		First(&mailbox).Error
	if err != nil {
		return nil, err
	}
	if mailbox.ExpiresAt != nil && mailbox.ExpiresAt.Before(time.Now()) {
		return nil, gorm.ErrRecordNotFound
	}
	return &mailbox, nil
}

func (s *MailService) StoreIncomingMessage(mailbox *models.Mailbox, envelopeTo, fallbackFrom string, raw []byte) (*models.Message, error) {
	subject, messageID, fromAddr, textBody, htmlBody := parseRawEmail(raw)
	if fromAddr == "" {
		fromAddr = fallbackFrom
	}

	dayPath := filepath.Join(s.DataDir(), time.Now().Format("20060102"))
	if err := os.MkdirAll(dayPath, 0o755); err != nil {
		return nil, fmt.Errorf("create message dir: %w", err)
	}
	fileName := uuid.NewString() + ".eml"
	fullPath := filepath.Join(dayPath, fileName)
	if err := os.WriteFile(fullPath, raw, 0o644); err != nil {
		return nil, fmt.Errorf("save raw email: %w", err)
	}

	now := time.Now()
	msg := models.Message{
		MailboxID:  mailbox.ID,
		MessageID:  messageID,
		FromAddr:   fromAddr,
		ToAddr:     envelopeTo,
		Subject:    subject,
		RawPath:    fullPath,
		TextBody:   textBody,
		HTMLBody:   htmlBody,
		Size:       int64(len(raw)),
		ReceivedAt: now,
	}
	if err := s.db.Create(&msg).Error; err != nil {
		return nil, err
	}
	_ = s.db.Model(&models.Mailbox{}).Where("id = ?", mailbox.ID).Update("last_received", now).Error
	return &msg, nil
}

func (s *MailService) CleanupExpiredMailboxes() error {
	return s.db.Model(&models.Mailbox{}).
		Where("expires_at IS NOT NULL AND expires_at < ?", time.Now()).
		Updates(map[string]any{"enabled": false}).Error
}

func parseRawEmail(raw []byte) (subject, messageID, fromAddr, textBody, htmlBody string) {
	msg, err := stdmail.ReadMessage(bytes.NewReader(raw))
	if err != nil {
		text := string(raw)
		if len(text) > 32_000 {
			text = text[:32_000]
		}
		return "", "", "", text, ""
	}

	decoder := mime.WordDecoder{}
	rawSubject := msg.Header.Get("Subject")
	if decoded, err := decoder.DecodeHeader(rawSubject); err == nil {
		subject = decoded
	} else {
		subject = rawSubject
	}
	messageID = msg.Header.Get("Message-Id")
	if messageID == "" {
		messageID = msg.Header.Get("Message-ID")
	}
	if fromList, err := stdmail.ParseAddressList(msg.Header.Get("From")); err == nil && len(fromList) > 0 {
		fromAddr = fromList[0].String()
	}

	contentType := msg.Header.Get("Content-Type")
	mediaType, params, _ := mime.ParseMediaType(contentType)

	if strings.HasPrefix(mediaType, "multipart/") && params["boundary"] != "" {
		mr := multipart.NewReader(msg.Body, params["boundary"])
		for {
			part, err := mr.NextPart()
			if err == io.EOF {
				break
			}
			if err != nil {
				break
			}
			ct := part.Header.Get("Content-Type")
			mt, _, _ := mime.ParseMediaType(ct)
			b, _ := io.ReadAll(part)
			body := string(b)
			if strings.HasPrefix(mt, "text/plain") && textBody == "" {
				textBody = body
			}
			if strings.HasPrefix(mt, "text/html") && htmlBody == "" {
				htmlBody = body
			}
		}
	} else {
		b, _ := io.ReadAll(msg.Body)
		body := string(b)
		if strings.HasPrefix(mediaType, "text/html") {
			htmlBody = body
		} else {
			textBody = body
		}
	}

	if textBody == "" {
		textBody = htmlBody
	}
	if len(textBody) > 200_000 {
		textBody = textBody[:200_000]
	}
	if len(htmlBody) > 200_000 {
		htmlBody = htmlBody[:200_000]
	}

	return
}
