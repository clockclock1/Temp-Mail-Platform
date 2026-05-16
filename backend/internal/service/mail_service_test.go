package service

import (
	"path/filepath"
	"strings"
	"testing"

	"tempmail/backend/internal/db"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/util"

	"gorm.io/gorm"
)

func TestCreateMailboxGeneratesConfiguredSubdomain(t *testing.T) {
	svc, database := newTestMailService(t)
	ownerID := createTestUser(t, database)
	root := createTestDomain(t, database, models.Domain{
		Name:      "example.com",
		Enabled:   true,
		Level:     4,
		CreatedBy: ownerID,
	})

	mailbox, err := svc.CreateMailbox(ownerID, "alice", root.ID, "test", 24)
	if err != nil {
		t.Fatalf("CreateMailbox returned error: %v", err)
	}

	if mailbox.Domain.Name == root.Name {
		t.Fatalf("expected generated subdomain, got root domain %q", mailbox.Domain.Name)
	}
	if util.DomainDepth(mailbox.Domain.Name) != 4 {
		t.Fatalf("expected depth 4, got %d for %q", util.DomainDepth(mailbox.Domain.Name), mailbox.Domain.Name)
	}
	if !strings.HasSuffix(mailbox.Domain.Name, "."+root.Name) {
		t.Fatalf("expected %q to end with %q", mailbox.Domain.Name, root.Name)
	}
}

func TestFindMailboxByAddressMatchesRootMailboxForSubdomain(t *testing.T) {
	svc, database := newTestMailService(t)
	ownerID := createTestUser(t, database)
	root := createTestDomain(t, database, models.Domain{
		Name:      "example.com",
		Enabled:   true,
		Level:     2,
		CreatedBy: ownerID,
	})

	mailbox, err := svc.CreateMailbox(ownerID, "alice", root.ID, "test", 24)
	if err != nil {
		t.Fatalf("CreateMailbox returned error: %v", err)
	}

	found, err := svc.FindMailboxByAddress("alice", "mx.mail.example.com")
	if err != nil {
		t.Fatalf("FindMailboxByAddress returned error: %v", err)
	}
	if found.ID != mailbox.ID {
		t.Fatalf("expected mailbox %d, got %d", mailbox.ID, found.ID)
	}
}

func newTestMailService(t *testing.T) (*MailService, *gorm.DB) {
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

	return NewMailService(database, filepath.Join(root, "messages")), database
}

func createTestUser(t *testing.T, database *gorm.DB) uint {
	t.Helper()

	role := models.Role{Name: "admin", Description: "admin"}
	if err := database.Create(&role).Error; err != nil {
		t.Fatalf("create role: %v", err)
	}

	user := models.User{
		Username:     "tester",
		PasswordHash: "hashed",
		DisplayName:  "Tester",
		Active:       true,
		RoleID:       role.ID,
	}
	if err := database.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}

	return user.ID
}

func createTestDomain(t *testing.T, database *gorm.DB, domain models.Domain) models.Domain {
	t.Helper()

	if err := database.Create(&domain).Error; err != nil {
		t.Fatalf("create domain: %v", err)
	}
	return domain
}
