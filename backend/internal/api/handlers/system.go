package handlers

import (
	"errors"
	"log"

	"caddy-manager/internal/caddy"

	"github.com/gofiber/fiber/v2"
)

type SystemHandler struct {
	versionProvider caddy.SystemVersionProvider
}

func NewSystemHandler(versionProvider caddy.SystemVersionProvider) *SystemHandler {
	return &SystemHandler{versionProvider: versionProvider}
}

func (h *SystemHandler) GetCaddyVersion(c *fiber.Ctx) error {
	version, err := h.versionProvider.GetBinaryVersion(c.Context())
	if err != nil {
		log.Printf("[ERROR] System Caddy version check failed: %v", err)

		if errors.Is(err, caddy.ErrBinaryNotFound) {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"error": "caddy binary not found",
			})
		}

		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"error": "failed to get Caddy version",
		})
	}

	return c.JSON(fiber.Map{
		"version": version,
	})
}
