package router

import (
	"strings"

	"tempmail/backend/internal/auth"
	"tempmail/backend/internal/config"
	"tempmail/backend/internal/http/handlers"
	"tempmail/backend/internal/http/middleware"
	"tempmail/backend/internal/models"
	"tempmail/backend/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

func New(cfg config.Config, db *gorm.DB, jwtManager *auth.JWTManager, mailService *service.MailService) *gin.Engine {
	r := gin.Default()
	r.Use(corsMiddleware(cfg.CorsOrigins))

	authHandler := handlers.NewAuthHandler(db, jwtManager)
	domainHandler := handlers.NewDomainHandler(db)
	userHandler := handlers.NewUserHandler(db)
	roleHandler := handlers.NewRoleHandler(db)
	mailboxHandler := handlers.NewMailboxHandler(db, mailService)
	messageHandler := handlers.NewMessageHandler(db)
	statsHandler := handlers.NewStatsHandler(db)
	legacyHandler := handlers.NewLegacyHandler(cfg, db, mailService, jwtManager)

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
				users.POST("", userHandler.Create)
				users.PATCH("/:id", userHandler.Update)
				users.DELETE("/:id", userHandler.Delete)
			}

			roles := secured.Group("/roles")
			roles.Use(middleware.RequirePermission(models.PermRoleManage))
			{
				roles.GET("", roleHandler.List)
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
		}
	}

	return r
}

func corsMiddleware(allowOrigins []string) gin.HandlerFunc {
	allowAll := false
	originSet := map[string]struct{}{}
	for _, o := range allowOrigins {
		o = strings.TrimSpace(o)
		if o == "*" {
			allowAll = true
		}
		if o != "" {
			originSet[o] = struct{}{}
		}
	}

	return func(c *gin.Context) {
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
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
