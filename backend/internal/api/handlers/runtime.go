package handlers

import (
	"net/http"
	"strconv"

	"caddy-manager/internal/configstore"
	"caddy-manager/internal/models"
	"caddy-manager/internal/runtime"

	"github.com/gofiber/fiber/v2"
)

type RuntimeHandler struct {
	runtimeService *runtime.Service
}

func NewRuntimeHandler(runtimeService *runtime.Service) *RuntimeHandler {
	return &RuntimeHandler{
		runtimeService: runtimeService,
	}
}

// GetStatus returns the operational health and reachability of Caddy
func (h *RuntimeHandler) GetStatus(c *fiber.Ctx) error {
	res := h.runtimeService.GetStatus(c.Context())
	return c.JSON(models.APIResponse{
		Data: res,
	})
}

// GetLogs returns normalized runtime logs filtered by level, search text, and limit
func (h *RuntimeHandler) GetLogs(c *fiber.Ctx) error {
	level := c.Query("level", "all")
	search := c.Query("search", "")
	limitStr := c.Query("limit", "100")
	cursor := c.Query("cursor", "")

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit <= 0 {
		limit = 100
	}

	filters := runtime.RuntimeLogFilters{
		Level:  level,
		Search: search,
		Limit:  limit,
		Cursor: cursor,
	}

	res := h.runtimeService.GetLogs(c.Context(), filters)
	if !res.Available && res.Reason == "caddy_unreachable" {
		return c.Status(http.StatusServiceUnavailable).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "CADDY_UNREACHABLE",
				Message: "Unable to connect to Caddy Admin API",
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: res,
	})
}

// EnableLogging handles POST /api/runtime/enable-logging
func (h *RuntimeHandler) EnableLogging(c *fiber.Ctx) error {
	var req runtime.EnableLoggingRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid JSON request body",
			},
		})
	}

	var userID int64
	if id, ok := c.Locals("userID").(int64); ok {
		userID = id
	}
	username, _ := c.Locals("username").(string)
	if username == "" {
		username = "admin"
	}

	res, err := h.runtimeService.EnableLogging(c.Context(), userID, username, req)
	if err != nil {
		if err == configstore.ErrConfigConflict {
			return c.Status(http.StatusConflict).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "CONFIG_CONFLICT",
					Message: "Configuration has been modified since page load. Please refresh and try again.",
				},
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "APPLY_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: res,
	})
}
