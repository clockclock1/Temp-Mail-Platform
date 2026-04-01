package main

import (
	"context"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/config"
	"tempmail/backend/internal/db"
	httprouter "tempmail/backend/internal/http/router"
	"tempmail/backend/internal/seed"
	"tempmail/backend/internal/service"
	smtpsrv "tempmail/backend/internal/smtp"
)

func main() {
	cfg := config.Load()

	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		log.Fatalf("failed to create data directory: %v", err)
	}
	if err := seed.Bootstrap(database, cfg.DefaultAdminUser, cfg.DefaultAdminPass); err != nil {
		log.Fatalf("failed to bootstrap data: %v", err)
	}

	jwtManager := auth.NewJWTManager(cfg.JWTSecret, cfg.JWTExpireHours)
	mailService := service.NewMailService(database, cfg.DataDir)
	r := httprouter.New(cfg, database, jwtManager, mailService)

	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
	}

	smtpServer := smtpsrv.New(cfg.SMTPAddr, mailService)

	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()
	go startCleanupTicker(cleanupCtx, mailService, cfg.CleanupInterval)

	go func() {
		log.Printf("http server listening on %s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server failed: %v", err)
		}
	}()

	go func() {
		if err := smtpServer.Start(); err != nil {
			if isExpectedShutdownErr(err) {
				return
			}
			log.Fatalf("smtp server failed: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cleanupCancel()

	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("http shutdown error: %v", err)
	}
	if err := smtpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("smtp shutdown error: %v", err)
	}

	log.Println("service stopped")
}

func startCleanupTicker(ctx context.Context, mailService *service.MailService, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := mailService.CleanupExpiredMailboxes(); err != nil {
				log.Printf("cleanup expired mailboxes error: %v", err)
			}
		}
	}
}

func isExpectedShutdownErr(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, net.ErrClosed) {
		return true
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "server closed") || strings.Contains(msg, "closed network connection")
}
