package handlers

import (
	"errors"
	"net/http"

	"caddy-manager/internal/configstore"
	"caddy-manager/internal/models"

	"github.com/gofiber/fiber/v2"
)

type ConfigHandler struct {
	configService *configstore.Service
}

func NewConfigHandler(configService *configstore.Service) *ConfigHandler {
	return &ConfigHandler{
		configService: configService,
	}
}

// GetConfig fetches the active Caddy configuration, hash, and version
func (h *ConfigHandler) GetConfig(c *fiber.Ctx) error {
	res, err := h.configService.GetConfig(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "CADDY_CONFIG_ERROR",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: res,
	})
}

// ValidateConfig dry-runs validation for proposed configuration
func (h *ConfigHandler) ValidateConfig(c *fiber.Ctx) error {
	var req models.ValidateConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_BODY",
				Message: "Invalid JSON request body",
			},
		})
	}

	res, currentHash, hasConflict, err := h.configService.ValidateConfig(c.Context(), req)
	if err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "VALIDATION_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: fiber.Map{
			"valid":        res.Valid,
			"message":      res.Message,
			"details":      res.Details,
			"current_hash": currentHash,
			"has_conflict": hasConflict,
		},
	})
}

// DiffConfig returns backend computed line diff comparing active vs proposed configuration
func (h *ConfigHandler) DiffConfig(c *fiber.Ctx) error {
	var req models.DiffConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_BODY",
				Message: "Invalid JSON request body",
			},
		})
	}

	res, err := h.configService.DiffConfig(c.Context(), req)
	if err != nil {
		if errors.Is(err, configstore.ErrConfigConflict) {
			return c.Status(http.StatusConflict).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "CONFIG_CONFLICT",
					Message: err.Error(),
				},
			})
		}
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "DIFF_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: res,
	})
}

// ApplyConfig validates, snapshots, checks optimistic concurrency, and applies proposed configuration to Caddy
func (h *ConfigHandler) ApplyConfig(c *fiber.Ctx) error {
	var req models.ApplyConfigRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_BODY",
				Message: "Invalid JSON request body",
			},
		})
	}

	userID, username := getUserInfoFromLocals(c)
	res, err := h.configService.ApplyConfig(c.Context(), userID, username, req)
	if err != nil {
		if errors.Is(err, configstore.ErrConfigConflict) {
			return c.Status(http.StatusConflict).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "CONFIG_CONFLICT",
					Message: err.Error(),
				},
			})
		}
		if errors.Is(err, configstore.ErrInvalidJSON) || errors.Is(err, configstore.ErrConfigTooLarge) {
			return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "INVALID_CONFIG",
					Message: err.Error(),
				},
			})
		}
		if errors.Is(err, configstore.ErrApplyVerificationFailed) {
			return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "APPLY_VERIFICATION_FAILED",
					Message: err.Error(),
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

// GetHistory returns snapshot metadata history list
func (h *ConfigHandler) GetHistory(c *fiber.Ctx) error {
	items, err := h.configService.GetHistory(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "HISTORY_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: items,
	})
}

// GetSnapshot returns full details and raw JSON of a single snapshot by ID
func (h *ConfigHandler) GetSnapshot(c *fiber.Ctx) error {
	id := c.Params("id")
	snap, err := h.configService.GetSnapshot(c.Context(), id)
	if err != nil {
		if errors.Is(err, configstore.ErrSnapshotNotFound) {
			return c.Status(http.StatusNotFound).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "SNAPSHOT_NOT_FOUND",
					Message: "Snapshot not found",
				},
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "SNAPSHOT_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: snap,
	})
}

// RollbackConfig restores historical configuration snapshot
func (h *ConfigHandler) RollbackConfig(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.RollbackRequest
	_ = c.BodyParser(&req)

	userID, username := getUserInfoFromLocals(c)
	res, err := h.configService.RollbackConfig(c.Context(), userID, username, id, req)
	if err != nil {
		if errors.Is(err, configstore.ErrSnapshotNotFound) {
			return c.Status(http.StatusNotFound).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "SNAPSHOT_NOT_FOUND",
					Message: "Snapshot not found",
				},
			})
		}
		if errors.Is(err, configstore.ErrConfigConflict) {
			return c.Status(http.StatusConflict).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "CONFIG_CONFLICT",
					Message: err.Error(),
				},
			})
		}
		if errors.Is(err, configstore.ErrApplyVerificationFailed) {
			return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "APPLY_VERIFICATION_FAILED",
					Message: err.Error(),
				},
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "ROLLBACK_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: res,
	})
}

func getUserInfoFromLocals(c *fiber.Ctx) (int64, string) {
	u := c.Locals("user")
	if u == nil {
		return 0, "system"
	}
	if user, ok := u.(*models.User); ok && user != nil {
		return user.ID, user.Username
	}
	return 0, "system"
}
