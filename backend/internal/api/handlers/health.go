package handlers

import (
	"context"
	"time"

	"caddy-manager/internal/caddy"
	"caddy-manager/internal/database"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct {
	db          *database.DB
	caddyClient caddy.CaddyClient
}

func NewHealthHandler(db *database.DB, caddyClient caddy.CaddyClient) *HealthHandler {
	return &HealthHandler{
		db:          db,
		caddyClient: caddyClient,
	}
}

// Health is the liveness endpoint (/health). Returns 200 if the application process is running.
func (h *HealthHandler) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status": "ok",
		"app":    "caddy-manager",
	})
}

// Ready is the readiness endpoint (/ready). Returns 200 if SQLite DB is reachable.
func (h *HealthHandler) Ready(c *fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
	defer cancel()

	dbStatus := "ok"
	if err := h.db.PingContext(ctx); err != nil {
		dbStatus = "error"
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":   "unhealthy",
			"database": dbStatus,
			"error":    err.Error(),
		})
	}

	caddyStatus := "connected"
	if err := h.caddyClient.Health(ctx); err != nil {
		caddyStatus = "disconnected"
	}

	return c.JSON(fiber.Map{
		"status":       "ready",
		"database":     dbStatus,
		"caddy_status": caddyStatus,
	})
}
