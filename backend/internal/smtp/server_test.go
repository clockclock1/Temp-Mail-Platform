package smtp

import (
	"path/filepath"
	"strings"
	"testing"

	"tempmail/backend/internal/db"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/service"

	"gorm.io/gorm"
)

func TestSessionStoresMailSentToSubdomainAddress(t *testing.T) {
	mailService, database := newTestSMTPService(t)
	ownerID := createSMTPTestUser(t, database)
	root := createSMTPTestDomain(t, database, models.Domain{
		Name:      "example.com",
		Enabled:   true,
		Level:     2,
		CreatedBy: ownerID,
	})

	mailbox, err := mailService.CreateMailbox(ownerID, "alice", root.ID, "smtp-test", 24)
	if err != nil {
		t.Fatalf("CreateMailbox returned error: %v", err)
	}

	sess := &session{mailService: mailService}
	if err := sess.Mail("sender@example.net", nil); err != nil {
		t.Fatalf("Mail returned error: %v", err)
	}
	if err := sess.Rcpt("alice@mx.mail.example.com", nil); err != nil {
		t.Fatalf("Rcpt returned error: %v", err)
	}

	raw := strings.NewReader("Subject: Hello\r\nFrom: sender@example.net\r\n\r\nbody")
	if err := sess.Data(raw); err != nil {
		t.Fatalf("Data returned error: %v", err)
	}

	var messages []models.Message
	if err := database.Where("mailbox_id = ?", mailbox.ID).Find(&messages).Error; err != nil {
		t.Fatalf("query messages: %v", err)
	}
	if len(messages) != 1 {
		t.Fatalf("expected 1 stored message, got %d", len(messages))
	}
	if messages[0].ToAddr != "alice@mx.mail.example.com" {
		t.Fatalf("expected envelope recipient to be preserved, got %q", messages[0].ToAddr)
	}
}

func newTestSMTPService(t *testing.T) (*service.MailService, *gorm.DB) {
	t.Helper()

	root := t.TempDir()
	database, err := db.Open(filepath.Join(root, "test.db"))
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	sqlDB, err := database.DB()
	if err != nil {
		t.Fatalf("open sql db handle: %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	return service.NewMailService(database, filepath.Join(root, "messages")), database
}

func createSMTPTestUser(t *testing.T, database *gorm.DB) uint {
	t.Helper()

	role := models.Role{Name: "admin", Description: "admin"}
	if err := database.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	user := models.User{
		Username:     "smtp-tester",
		PasswordHash: "hashed",
		DisplayName:  "SMTP Tester",
		Active:       true,
		RoleID:       role.ID,
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user.ID
}

func createSMTPTestDomain(t *testing.T, database *gorm.DB, domain models.Domain) models.Domain {
	t.Helper()

	if err := database.Create(&domain).Error; err != nil {
		t.Fatalf("create domain: %v", err)
	}
	return domain
}
