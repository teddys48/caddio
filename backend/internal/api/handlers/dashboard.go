package handlers

import (
	"context"
	"time"

	"caddy-manager/internal/audit"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
	"caddy-manager/internal/sites"

	"github.com/gofiber/fiber/v2"
)

type DashboardHandler struct {
	db           *database.DB
	caddyClient  caddy.CaddyClient
	authService  *auth.Service
	auditLogger  *audit.Logger
	sitesService *sites.Service
}

func NewDashboardHandler(db *database.DB, caddyClient caddy.CaddyClient, authService *auth.Service, auditLogger *audit.Logger, sitesService *sites.Service) *DashboardHandler {
	return &DashboardHandler{
		db:           db,
		caddyClient:  caddyClient,
		authService:  authService,
		auditLogger:  auditLogger,
		sitesService: sitesService,
	}
}

func (h *DashboardHandler) GetDashboard(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()

	// Check Caddy connectivity & version
	caddyStatus := "Connected"
	caddyVersion := ""
	if err := h.caddyClient.Health(ctx); err != nil {
		caddyStatus = "Disconnected"
	} else {
		ver, err := h.caddyClient.GetVersion(ctx)
		if err == nil && ver != "" {
			caddyVersion = ver
		}
	}

	// Check database connection
	dbStatus := "connected"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "disconnected"
	}

	// App health overall
	appStatus := "healthy"
	if dbStatus != "connected" {
		appStatus = "unhealthy"
	}

	// Setup status
	hasUsers, _ := h.authService.HasAnyUser()

	// Get recent audit logs
	logs, err := h.auditLogger.GetRecentLogs(10)
	if err != nil {
		logs = []models.AuditLog{}
	}

	// Get site summary metrics if site service is provided
	siteSummary := models.SiteSummaryMetrics{}
	if h.sitesService != nil {
		if summary, err := h.sitesService.GetSummaryMetrics(ctx); err == nil {
			siteSummary = summary
		}
	}

	response := models.DashboardResponse{
		Caddy: models.CaddyStatusResponse{
			Status:  caddyStatus,
			Version: caddyVersion,
		},
		Health: models.AppHealthResponse{
			Status:        appStatus,
			Database:      dbStatus,
			CaddyStatus:   caddyStatus,
			SetupRequired: !hasUsers,
		},
		RecentLogs:  logs,
		HasUser:     hasUsers,
		SiteSummary: siteSummary,
	}

	return c.JSON(models.APIResponse{
		Data: response,
	})
}
