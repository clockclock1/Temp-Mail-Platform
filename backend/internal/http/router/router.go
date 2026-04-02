package router

import (
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/config"
	"tempmail/backend/internal/http/handlers"
	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/runtime"
	"tempmail/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(
	cfgManager *config.Manager,
	db *gorm.DB,
	jwtManager *auth.JWTManager,
	addressJWT *auth.AddressJWTManager,
	mailService *service.MailService,
	runtimeController *runtime.Controller,
) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware(cfgManager))

	authHandler := handlers.NewAuthHandler(db, jwtManager)
	domainHandler := handlers.NewDomainHandler(db)
	userHandler := handlers.NewUserHandler(db)
	roleHandler := handlers.NewRoleHandler(db)
	mailboxHandler := handlers.NewMailboxHandler(db, mailService)
	messageHandler := handlers.NewMessageHandler(db)
	statsHandler := handlers.NewStatsHandler(db)
	legacyHandler := handlers.NewLegacyHandler(cfgManager, db, mailService, jwtManager, addressJWT)
	configHandler := handlers.NewConfigHandler(cfgManager, runtimeController)

	r.GET("/healthz", handlers.Health)

	legacyAPI := r.Group("/api")
	{
		legacyAPI.POST("/new_address", legacyHandler.APINewAddress)
		legacyAPI.GET("/mails", legacyHandler.APIMails)
	}
	legacyAdmin := r.Group("/admin")
	{
		legacyAdmin.POST("/new_address", legacyHandler.AdminNewAddress)
		legacyAdmin.GET("/mails", legacyHandler.AdminMails)
		legacyAdmin.DELETE("/mails/:id", legacyHandler.AdminDeleteMail)
		legacyAdmin.DELETE("/delete_address/:id", legacyHandler.AdminDeleteAddress)
		legacyAdmin.DELETE("/clear_inbox/:id", legacyHandler.AdminClearInbox)
		legacyAdmin.DELETE("/clear_sent_items/:id", legacyHandler.AdminClearSentItems)
	}
	legacyUser := r.Group("/user_api")
	{
		legacyUser.POST("/login", legacyHandler.UserLogin)
		legacyUser.POST("/register", legacyHandler.UserRegister)
		legacyUser.GET("/mails", legacyHandler.UserMails)
	}

	api := r.Group("/api/v1")
	{
		api.POST("/auth/login", authHandler.Login)

		secured := api.Group("")
		secured.Use(middleware.AuthRequired(db, jwtManager))
		{
			secured.GET("/auth/me", authHandler.Me)
			secured.GET("/domains/available", domainHandler.Available)

			domains := secured.Group("/domains")
			domains.Use(middleware.RequirePermission(models.PermDomainManage))
			{
				domains.GET("", domainHandler.List)
				domains.POST("", domainHandler.Create)
				domains.PUT("/:id", domainHandler.Update)
				domains.DELETE("/:id", domainHandler.Delete)
			}

			users := secured.Group("/users")
			users.Use(middleware.RequirePermission(models.PermUserManage))
			{
				users.GET("", userHandler.List)
				users.GET("/:id", userHandler.Get)
				users.POST("", userHandler.Create)
				users.PATCH("/:id", userHandler.Update)
				users.POST("/:id/reset-password", userHandler.ResetPassword)
				users.DELETE("/:id", userHandler.Delete)
			}

			roles := secured.Group("/roles")
			roles.Use(middleware.RequirePermission(models.PermRoleManage))
			{
				roles.GET("", roleHandler.List)
				roles.GET("/:id/users", roleHandler.Users)
				roles.POST("", roleHandler.Create)
				roles.PUT("/:id", roleHandler.Update)
				roles.DELETE("/:id", roleHandler.Delete)
			}
			secured.GET("/permissions", middleware.RequirePermission(models.PermRoleManage), roleHandler.Permissions)

			mailboxes := secured.Group("/mailboxes")
			{
				mailboxes.GET("", middleware.RequirePermission(models.PermMailboxRead), mailboxHandler.List)
				mailboxes.POST("", middleware.RequirePermission(models.PermMailboxCreate), mailboxHandler.Create)
				mailboxes.DELETE("/:id", middleware.RequirePermission(models.PermMailboxDelete), mailboxHandler.Delete)
				mailboxes.GET("/:id/messages", middleware.RequirePermission(models.PermMessageRead), mailboxHandler.Messages)
			}

			messages := secured.Group("/messages")
			{
				messages.GET("/:id", middleware.RequirePermission(models.PermMessageRead), messageHandler.Get)
				messages.GET("/:id/raw", middleware.RequirePermission(models.PermMessageRead), messageHandler.Raw)
				messages.DELETE("/:id", middleware.RequirePermission(models.PermMessageDelete), messageHandler.Delete)
			}

			secured.GET("/stats", middleware.RequirePermission(models.PermStatsRead), statsHandler.Get)
			secured.GET("/system/config", middleware.RequirePermission(models.PermConfigManage), configHandler.Get)
			secured.PUT("/system/config", middleware.RequirePermission(models.PermConfigManage), configHandler.Update)
			secured.POST("/system/config/reload", middleware.RequirePermission(models.PermConfigManage), configHandler.Reload)
		}
	}

	registerFrontend(r, cfgManager)

	return r
}

func corsMiddleware(cfgManager *config.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		cfg := cfgManager.Get()
		allowAll := false
		originSet := map[string]struct{}{}
		for _, o := range cfg.CorsOrigins {
			o = strings.TrimSpace(o)
			if o == "*" {
				allowAll = true
			}
			if o != "" {
				originSet[o] = struct{}{}
			}
		}

		origin := c.Request.Header.Get("Origin")
		if allowAll {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		} else if origin != "" {
			if _, ok := originSet[origin]; ok {
				c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
				c.Writer.Header().Set("Vary", "Origin")
			}
		}

		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, x-admin-auth, x-user-token, x-custom-auth")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	}
}

func registerFrontend(r *gin.Engine, cfgManager *config.Manager) {
	r.GET("/", func(c *gin.Context) { serveIndex(c, cfgManager) })
	r.GET("/favicon.ico", func(c *gin.Context) {
		cfg := cfgManager.Get()
		full := filepath.Join(strings.TrimSpace(cfg.WebDir), "favicon.ico")
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(full)
	})
	r.GET("/assets/*filepath", func(c *gin.Context) {
		cfg := cfgManager.Get()
		root := filepath.Join(strings.TrimSpace(cfg.WebDir), "assets")
		reqPath := strings.TrimPrefix(c.Param("filepath"), "/")
		if reqPath == "" {
			c.Status(http.StatusNotFound)
			return
		}
		full, err := safeJoin(root, reqPath)
		if err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		st, err := os.Stat(full)
		if err != nil || st.IsDir() {
			c.Status(http.StatusNotFound)
			return
		}
		c.File(full)
	})

	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path
		if isReservedPath(path) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		serveIndex(c, cfgManager)
	})
}

func serveIndex(c *gin.Context, cfgManager *config.Manager) {
	cfg := cfgManager.Get()
	indexPath := filepath.Join(strings.TrimSpace(cfg.WebDir), "index.html")
	st, err := os.Stat(indexPath)
	if err != nil || st.IsDir() {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "frontend not available"})
		return
	}
	c.File(indexPath)
}

func safeJoin(root, reqPath string) (string, error) {
	base := filepath.Clean(root)
	if base == "." || strings.TrimSpace(base) == "" {
		return "", errors.New("invalid root")
	}
	req := filepath.Clean(reqPath)
	if req == "." || req == "" {
		return "", errors.New("invalid path")
	}
	full := filepath.Clean(filepath.Join(base, req))
	if full == base {
		return full, nil
	}
	prefix := base + string(os.PathSeparator)
	if !strings.HasPrefix(full, prefix) {
		return "", errors.New("path traversal")
	}
	return full, nil
}

func isReservedPath(path string) bool {
	if path == "/healthz" {
		return true
	}
	return path == "/api" || strings.HasPrefix(path, "/api/") ||
		path == "/admin" || strings.HasPrefix(path, "/admin/") ||
		path == "/user_api" || strings.HasPrefix(path, "/user_api/") ||
		path == "/assets" || strings.HasPrefix(path, "/assets/") ||
		path == "/favicon.ico"
}
