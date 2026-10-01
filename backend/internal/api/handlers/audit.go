package handlers

import (
	"caddy-manager/internal/audit"
	"caddy-manager/internal/models"

	"github.com/gofiber/fiber/v2"
)

type AuditHandler struct {
	auditLogger *audit.Logger
}

func NewAuditHandler(auditLogger *audit.Logger) *AuditHandler {
	return &AuditHandler{auditLogger: auditLogger}
}

func (h *AuditHandler) GetAuditLogs(c *fiber.Ctx) error {
	limit := c.QueryInt("limit", 50)
	logs, err := h.auditLogger.GetRecentLogs(limit)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to retrieve audit logs",
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: logs,
	})
}
