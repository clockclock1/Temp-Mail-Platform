package runtime

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/config"
	"tempmail/backend/internal/service"
)

type ApplyResult struct {
	Warnings        []string `json:"warnings"`
	RestartRequired bool     `json:"restartRequired"`
}

type Controller struct {
	mailService        *service.MailService
	userJWT            *auth.JWTManager
	addressJWT         *auth.AddressJWTManager
	cleanupIntervalNS  atomic.Int64
	cleanupNowSignal   chan struct{}
	cleanupStopSignal  chan struct{}
	cleanupStoppedChan chan struct{}
}

func NewController(mailService *service.MailService, userJWT *auth.JWTManager, addressJWT *auth.AddressJWTManager) *Controller {
	c := &Controller{
		mailService:        mailService,
		userJWT:            userJWT,
		addressJWT:         addressJWT,
		cleanupNowSignal:   make(chan struct{}, 1),
		cleanupStopSignal:  make(chan struct{}),
		cleanupStoppedChan: make(chan struct{}),
	}
	return c
}

func (c *Controller) StartCleanupLoop() {
	go func() {
		defer close(c.cleanupStoppedChan)
		ticker := time.NewTicker(20 * time.Second)
		defer ticker.Stop()
		last := time.Now().Add(-time.Hour)
		for {
			select {
			case <-c.cleanupStopSignal:
				return
			case <-c.cleanupNowSignal:
				_ = c.mailService.CleanupExpiredMailboxes()
				last = time.Now()
			case <-ticker.C:
				interval := time.Duration(c.cleanupIntervalNS.Load())
				if interval <= 0 {
					interval = 10 * time.Minute
				}
				if time.Since(last) >= interval {
					_ = c.mailService.CleanupExpiredMailboxes()
					last = time.Now()
				}
			}
		}
	}()
}

func (c *Controller) StopCleanupLoop() {
	close(c.cleanupStopSignal)
	<-c.cleanupStoppedChan
}

func (c *Controller) Apply(oldCfg, newCfg config.Config) (ApplyResult, error) {
	result := ApplyResult{Warnings: []string{}}

	if err := os.MkdirAll(newCfg.DataDir, 0o755); err != nil {
		return result, fmt.Errorf("create data_dir: %w", err)
	}

	c.userJWT.Update(newCfg.JWTSecret, newCfg.JWTExpireHours)
	c.addressJWT.Update(newCfg.JWTSecret, newCfg.LegacyAddrExpire)
	c.mailService.UpdateDataDir(newCfg.DataDir)
	c.cleanupIntervalNS.Store(newCfg.CleanupInterval().Nanoseconds())

	select {
	case c.cleanupNowSignal <- struct{}{}:
	default:
	}

	if oldCfg.HTTPAddr != newCfg.HTTPAddr {
		result.RestartRequired = true
		result.Warnings = append(result.Warnings, "http_addr changed, restart required to listen on new address")
	}
	if oldCfg.SMTPAddr != newCfg.SMTPAddr {
		result.RestartRequired = true
		result.Warnings = append(result.Warnings, "smtp_addr changed, restart required to listen on new address")
	}
	if oldCfg.DBPath != newCfg.DBPath {
		result.RestartRequired = true
		result.Warnings = append(result.Warnings, "db_path changed, restart required to use new database file")
	}
	if oldCfg.DefaultAdminUser != newCfg.DefaultAdminUser || oldCfg.DefaultAdminPass != newCfg.DefaultAdminPass {
		result.Warnings = append(result.Warnings, "default admin credentials are only used during bootstrap for missing admin account")
	}
	if strings.TrimSpace(newCfg.LegacyCustomAuth) == "" {
		result.Warnings = append(result.Warnings, "legacy_custom_auth is empty; legacy endpoints rely only on x-admin-auth/x-user-token")
	}

	return result, nil
}
