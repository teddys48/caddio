package routes

import (
	"caddy-manager/internal/api/handlers"
	"caddy-manager/internal/api/middleware"
	"caddy-manager/internal/audit"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/database"
	"caddy-manager/internal/sites"
	"caddy-manager/internal/configstore"

	"github.com/gofiber/fiber/v2"
)

func SetupRoutes(
	app *fiber.App,
	db *database.DB,
	authService *auth.Service,
	caddyClient caddy.CaddyClient,
	auditLogger *audit.Logger,
	versionProvider caddy.SystemVersionProvider,
	sitesService *sites.Service,
	configService *configstore.Service,
) {
	healthHandler := handlers.NewHealthHandler(db, caddyClient)
	authHandler := handlers.NewAuthHandler(authService, auditLogger)
	dashboardHandler := handlers.NewDashboardHandler(db, caddyClient, authService, auditLogger, sitesService)
	auditHandler := handlers.NewAuditHandler(auditLogger)
	systemHandler := handlers.NewSystemHandler(versionProvider)
	sitesHandler := handlers.NewSitesHandler(sitesService)
	configHandler := handlers.NewConfigHandler(configService)

	// Liveness and Readiness
	app.Get("/health", healthHandler.Health)
	app.Get("/ready", healthHandler.Ready)

	api := app.Group("/api")

	// Public Auth endpoints
	authGroup := api.Group("/auth")
	authGroup.Post("/setup", authHandler.Setup)
	authGroup.Post("/login", authHandler.Login)
	authGroup.Post("/logout", authHandler.Logout)
	authGroup.Get("/me", authHandler.Me)

	// System Caddy Version endpoint (GET /api/system/caddy/version)
	systemGroup := api.Group("/system")
	systemGroup.Get("/caddy/version", systemHandler.GetCaddyVersion)

	// Protected endpoints
	protected := api.Group("/", middleware.RequireAuth(authService))
	protected.Get("/dashboard", dashboardHandler.GetDashboard)
	protected.Get("/audit-logs", auditHandler.GetAuditLogs)

	// Sites API
	sitesGroup := protected.Group("/sites")
	sitesGroup.Get("/", sitesHandler.GetSites)
	sitesGroup.Post("/validate", sitesHandler.ValidateSite)
	sitesGroup.Get("/:id", sitesHandler.GetSiteByID)
	sitesGroup.Post("/", sitesHandler.CreateSite)
	sitesGroup.Put("/:id", sitesHandler.UpdateSite)
	sitesGroup.Delete("/:id", sitesHandler.DeleteSite)

	// Config Workspace API
	configGroup := protected.Group("/config")
	configGroup.Get("/", configHandler.GetConfig)
	configGroup.Post("/validate", configHandler.ValidateConfig)
	configGroup.Post("/diff", configHandler.DiffConfig)
	configGroup.Post("/apply", configHandler.ApplyConfig)
	configGroup.Get("/history", configHandler.GetHistory)
	configGroup.Get("/history/:id", configHandler.GetSnapshot)
	configGroup.Post("/history/:id/rollback", configHandler.RollbackConfig)
}
