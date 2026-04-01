package seed

import (
	"fmt"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/models"

	"gorm.io/gorm"
)

func Bootstrap(db *gorm.DB, adminUser, adminPass string) error {
	if err := seedPermissions(db); err != nil {
		return err
	}
	if err := seedRoles(db); err != nil {
		return err
	}
	if err := seedAdmin(db, adminUser, adminPass); err != nil {
		return err
	}
	return nil
}

func seedPermissions(db *gorm.DB) error {
	for _, perm := range models.DefaultPermissionCatalog {
		var existing models.Permission
		err := db.Where("key = ?", perm.Key).First(&existing).Error
		if err == nil {
			continue
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			return fmt.Errorf("find permission %s: %w", perm.Key, err)
		}
		if err := db.Create(&perm).Error; err != nil {
			return fmt.Errorf("create permission %s: %w", perm.Key, err)
		}
	}
	return nil
}

func seedRoles(db *gorm.DB) error {
	allPerms := []models.Permission{}
	if err := db.Find(&allPerms).Error; err != nil {
		return fmt.Errorf("list permissions: %w", err)
	}

	var adminRole models.Role
	err := db.Preload("Permissions").Where("name = ?", "admin").First(&adminRole).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("find admin role: %w", err)
		}
		adminRole = models.Role{Name: "admin", Description: "System administrator"}
		if err := db.Create(&adminRole).Error; err != nil {
			return fmt.Errorf("create admin role: %w", err)
		}
	}
	if err := db.Model(&adminRole).Association("Permissions").Replace(&allPerms); err != nil {
		return fmt.Errorf("assign admin permissions: %w", err)
	}

	userKeys := map[string]struct{}{
		models.PermMailboxCreate: {},
		models.PermMailboxRead:   {},
		models.PermMailboxDelete: {},
		models.PermMessageRead:   {},
		models.PermMessageDelete: {},
	}
	userPerms := make([]models.Permission, 0, len(userKeys))
	for _, p := range allPerms {
		if _, ok := userKeys[p.Key]; ok {
			userPerms = append(userPerms, p)
		}
	}

	var userRole models.Role
	err = db.Preload("Permissions").Where("name = ?", "user").First(&userRole).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			return fmt.Errorf("find user role: %w", err)
		}
		userRole = models.Role{Name: "user", Description: "Regular user"}
		if err := db.Create(&userRole).Error; err != nil {
			return fmt.Errorf("create user role: %w", err)
		}
	}
	if err := db.Model(&userRole).Association("Permissions").Replace(&userPerms); err != nil {
		return fmt.Errorf("assign user permissions: %w", err)
	}

	return nil
}

func seedAdmin(db *gorm.DB, username, password string) error {
	var role models.Role
	if err := db.Where("name = ?", "admin").First(&role).Error; err != nil {
		return fmt.Errorf("find admin role for user seed: %w", err)
	}

	var existing models.User
	err := db.Where("username = ?", username).First(&existing).Error
	if err == nil {
		return nil
	}
	if err != nil && err != gorm.ErrRecordNotFound {
		return fmt.Errorf("find admin user: %w", err)
	}

	hash, err := auth.HashPassword(password)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}
	admin := models.User{
		Username:     username,
		PasswordHash: hash,
		DisplayName:  "Administrator",
		RoleID:       role.ID,
		Active:       true,
	}
	if err := db.Create(&admin).Error; err != nil {
		return fmt.Errorf("create admin user: %w", err)
	}
	return nil
}
