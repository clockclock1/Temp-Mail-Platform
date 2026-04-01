package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppName          string
	HTTPAddr         string
	SMTPAddr         string
	WebDir           string
	JWTSecret        string
	JWTExpireHours   int
	LegacyAdminAuth  string
	LegacyCustomAuth string
	LegacyAddrExpire int
	DBPath           string
	DataDir          string
	CorsOrigins      []string
	DefaultAdminUser string
	DefaultAdminPass string
	CleanupInterval  time.Duration
}

func Load() Config {
	cleanupMinutes := envInt("CLEANUP_INTERVAL_MINUTES", 10)
	defaultAdminPass := env("DEFAULT_ADMIN_PASS", "admin123456")
	origins := strings.Split(env("CORS_ORIGINS", "http://localhost:5173"), ",")
	for i := range origins {
		origins[i] = strings.TrimSpace(origins[i])
	}

	return Config{
		AppName:          env("APP_NAME", "Temp Mail Service"),
		HTTPAddr:         env("HTTP_ADDR", ":8080"),
		SMTPAddr:         env("SMTP_ADDR", ":2525"),
		WebDir:           env("WEB_DIR", "./web"),
		JWTSecret:        env("JWT_SECRET", "change-me-in-production"),
		JWTExpireHours:   envInt("JWT_EXPIRE_HOURS", 24),
		LegacyAdminAuth:  env("LEGACY_ADMIN_AUTH", defaultAdminPass),
		LegacyCustomAuth: env("LEGACY_CUSTOM_AUTH", ""),
		LegacyAddrExpire: envInt("LEGACY_ADDRESS_JWT_EXPIRE_HOURS", 24*30),
		DBPath:           env("DB_PATH", "./data/tempmail.db"),
		DataDir:          env("DATA_DIR", "./data/messages"),
		CorsOrigins:      origins,
		DefaultAdminUser: env("DEFAULT_ADMIN_USER", "admin"),
		DefaultAdminPass: defaultAdminPass,
		CleanupInterval:  time.Duration(cleanupMinutes) * time.Minute,
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		n, err := strconv.Atoi(v)
		if err == nil {
			return n
		}
	}
	return fallback
}
