package middleware

import (
	"caddy-manager/internal/auth"
	"caddy-manager/internal/models"

	"github.com/gofiber/fiber/v2"
)

const SessionCookieName = "caddio_session"

func RequireAuth(authService *auth.Service) fiber.Handler {
	return func(c *fiber.Ctx) error {
		sessionID := c.Cookies(SessionCookieName)
		if sessionID == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "Authentication required",
				},
			})
		}

		user, err := authService.GetSessionUser(sessionID)
		if err != nil {
			c.ClearCookie(SessionCookieName)
			return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "UNAUTHORIZED",
					Message: "Session expired or invalid. Please log in again.",
				},
			})
		}

		c.Locals("user", user)
		return c.Next()
	}
}
