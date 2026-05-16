package config

import (
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	AppName                string   `yaml:"app_name" json:"appName"`
	HTTPAddr               string   `yaml:"http_addr" json:"httpAddr"`
	SMTPAddr               string   `yaml:"smtp_addr" json:"smtpAddr"`
	WebDir                 string   `yaml:"web_dir" json:"webDir"`
	JWTSecret              string   `yaml:"jwt_secret" json:"jwtSecret"`
	JWTExpireHours         int      `yaml:"jwt_expire_hours" json:"jwtExpireHours"`
	DBPath                 string   `yaml:"db_path" json:"dbPath"`
	DataDir                string   `yaml:"data_dir" json:"dataDir"`
	CorsOrigins            []string `yaml:"cors_origins" json:"corsOrigins"`
	DNSResolvers           []string `yaml:"dns_resolvers" json:"dnsResolvers"`
	DefaultAdminUser       string   `yaml:"default_admin_user" json:"defaultAdminUser"`
	DefaultAdminPass       string   `yaml:"default_admin_pass" json:"defaultAdminPass"`
	CleanupIntervalMinutes int      `yaml:"cleanup_interval_minutes" json:"cleanupIntervalMinutes"`
}

func Default() Config {
	return Config{
		AppName:                "Temp Mail Service",
		HTTPAddr:               ":8080",
		SMTPAddr:               ":2525",
		WebDir:                 "./web",
		JWTSecret:              "change-me-in-production",
		JWTExpireHours:         24,
		DBPath:                 "./data/tempmail.db",
		DataDir:                "./data/messages",
		CorsOrigins:            []string{"http://localhost:5173", "http://localhost:8080"},
		DNSResolvers:           []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"},
		DefaultAdminUser:       "admin",
		DefaultAdminPass:       "admin123456",
		CleanupIntervalMinutes: 10,
	}
}

func (c Config) CleanupInterval() time.Duration {
	minutes := c.CleanupIntervalMinutes
	if minutes <= 0 {
		minutes = 10
	}
	return time.Duration(minutes) * time.Minute
}

func (c Config) Clone() Config {
	out := c
	out.CorsOrigins = append([]string(nil), c.CorsOrigins...)
	out.DNSResolvers = append([]string(nil), c.DNSResolvers...)
	return out
}

func LoadFromFile(path string) (Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Config{}, err
	}
	cfg := Default()
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return Config{}, fmt.Errorf("parse config yaml: %w", err)
	}
	Normalize(&cfg)
	if err := Validate(cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func SaveToFile(path string, cfg Config) error {
	Normalize(&cfg)
	if err := Validate(cfg); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshal config yaml: %w", err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("write config file: %w", err)
	}
	return nil
}

func Normalize(cfg *Config) {
	cfg.AppName = strings.TrimSpace(cfg.AppName)
	cfg.HTTPAddr = strings.TrimSpace(cfg.HTTPAddr)
	cfg.SMTPAddr = strings.TrimSpace(cfg.SMTPAddr)
	cfg.WebDir = strings.TrimSpace(cfg.WebDir)
	cfg.JWTSecret = strings.TrimSpace(cfg.JWTSecret)
	cfg.DBPath = strings.TrimSpace(cfg.DBPath)
	cfg.DataDir = strings.TrimSpace(cfg.DataDir)
	cfg.DefaultAdminUser = strings.TrimSpace(cfg.DefaultAdminUser)
	cfg.DefaultAdminPass = strings.TrimSpace(cfg.DefaultAdminPass)

	origins := make([]string, 0, len(cfg.CorsOrigins))
	for _, o := range cfg.CorsOrigins {
		o = strings.TrimSpace(o)
		if o != "" {
			origins = append(origins, o)
		}
	}
	if len(origins) == 0 {
		origins = []string{"*"}
	}
	cfg.CorsOrigins = origins
	cfg.DNSResolvers = normalizeResolverList(cfg.DNSResolvers)

	if cfg.CleanupIntervalMinutes <= 0 {
		cfg.CleanupIntervalMinutes = 10
	}
	if cfg.JWTExpireHours <= 0 {
		cfg.JWTExpireHours = 24
	}
}

func Validate(cfg Config) error {
	if strings.TrimSpace(cfg.HTTPAddr) == "" {
		return errors.New("http_addr cannot be empty")
	}
	if strings.TrimSpace(cfg.SMTPAddr) == "" {
		return errors.New("smtp_addr cannot be empty")
	}
	if strings.TrimSpace(cfg.JWTSecret) == "" {
		return errors.New("jwt_secret cannot be empty")
	}
	if strings.TrimSpace(cfg.DBPath) == "" {
		return errors.New("db_path cannot be empty")
	}
	if strings.TrimSpace(cfg.DataDir) == "" {
		return errors.New("data_dir cannot be empty")
	}
	if strings.TrimSpace(cfg.DefaultAdminUser) == "" {
		return errors.New("default_admin_user cannot be empty")
	}
	if strings.TrimSpace(cfg.DefaultAdminPass) == "" {
		return errors.New("default_admin_pass cannot be empty")
	}
	if cfg.JWTExpireHours <= 0 {
		return errors.New("jwt_expire_hours must be > 0")
	}
	if cfg.CleanupIntervalMinutes <= 0 {
		return errors.New("cleanup_interval_minutes must be > 0")
	}
	return nil
}

func normalizeResolverList(resolvers []string) []string {
	if len(resolvers) == 0 {
		return []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"}
	}

	out := make([]string, 0, len(resolvers))
	seen := map[string]struct{}{}
	for _, resolver := range resolvers {
		resolver = strings.TrimSpace(resolver)
		if resolver == "" {
			continue
		}
		if _, _, err := net.SplitHostPort(resolver); err != nil {
			resolver = net.JoinHostPort(resolver, "53")
		}
		if _, ok := seen[resolver]; ok {
			continue
		}
		seen[resolver] = struct{}{}
		out = append(out, resolver)
	}
	if len(out) == 0 {
		return []string{"1.1.1.1:53", "8.8.8.8:53", "9.9.9.9:53"}
	}
	return out
}

type Manager struct {
	path string
	mu   sync.RWMutex
	cfg  Config
}

func NewManager(path string) (*Manager, error) {
	if strings.TrimSpace(path) == "" {
		path = "./config.yaml"
	}
	path = filepath.Clean(path)

	cfg, err := LoadFromFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			cfg = Default()
			if err := SaveToFile(path, cfg); err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	}

	return &Manager{path: path, cfg: cfg}, nil
}

func (m *Manager) Path() string {
	return m.path
}

func (m *Manager) Get() Config {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.cfg.Clone()
}

func (m *Manager) Update(newCfg Config) error {
	Normalize(&newCfg)
	if err := Validate(newCfg); err != nil {
		return err
	}
	if err := SaveToFile(m.path, newCfg); err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = newCfg.Clone()
	m.mu.Unlock()
	return nil
}

func (m *Manager) Reload() error {
	cfg, err := LoadFromFile(m.path)
	if err != nil {
		return err
	}
	m.mu.Lock()
	m.cfg = cfg
	m.mu.Unlock()
	return nil
}
