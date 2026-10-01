package handlers

import (
	"errors"
	"strings"

	"caddy-manager/internal/api/middleware"
	"caddy-manager/internal/audit"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/models"

	"github.com/gofiber/fiber/v2"
)

type AuthHandler struct {
	authService *auth.Service
	auditLogger *audit.Logger
}

func NewAuthHandler(authService *auth.Service, auditLogger *audit.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		auditLogger: auditLogger,
	}
}

func (h *AuthHandler) Setup(c *fiber.Ctx) error {
	var req models.SetupRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid setup payload format",
			},
		})
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || len(req.Password) < 8 {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Username is required and password must be at least 8 characters long",
			},
		})
	}

	if req.Password != req.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Passwords do not match",
			},
		})
	}

	user, err := h.authService.CreateInitialAdmin(req.Username, req.Password)
	if err != nil {
		if errors.Is(err, auth.ErrSetupAlreadyCompleted) {
			return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "SETUP_ALREADY_COMPLETED",
					Message: "Initial administrator setup has already been completed",
				},
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: "Failed to create administrator account",
			},
		})
	}

	_ = h.auditLogger.Log(&user.ID, "SETUP_COMPLETED", "system", "initial_setup", "Initial administrator account created")

	// Automatically create session for newly created admin
	session, err := h.authService.CreateSession(user.ID)
	if err == nil {
		c.Cookie(&fiber.Cookie{
			Name:     middleware.SessionCookieName,
			Value:    session.ID,
			Expires:  session.ExpiresAt,
			HTTPOnly: true,
			SameSite: "Lax",
			Path:     "/",
		})
	}

	return c.Status(fiber.StatusCreated).JSON(models.APIResponse{
		Data: models.UserResponse{
			ID:        user.ID,
			Username:  user.Username,
			CreatedAt: user.CreatedAt,
		},
	})
}

func (h *AuthHandler) Login(c *fiber.Ctx) error {
	var req models.LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_REQUEST",
				Message: "Invalid login payload format",
			},
		})
	}

	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" || req.Password == "" {
		return c.Status(fiber.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "VALIDATION_ERROR",
				Message: "Username and password are required",
			},
		})
	}

	user, err := h.authService.Authenticate(req.Username, req.Password)
	if err != nil {
		return c.Status(fiber.StatusUnauthorized).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_CREDENTIALS",
				Message: "Invalid username or password",
			},
		})
	}

	session, err := h.authService.CreateSession(user.ID)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "SESSION_ERROR",
				Message: "Failed to establish session",
			},
		})
	}

	c.Cookie(&fiber.Cookie{
		Name:     middleware.SessionCookieName,
		Value:    session.ID,
		Expires:  session.ExpiresAt,
		HTTPOnly: true,
		SameSite: "Lax",
		Path:     "/",
	})

	_ = h.auditLogger.Log(&user.ID, "LOGIN", "auth", user.Username, "User logged in successfully")

	return c.JSON(models.APIResponse{
		Data: models.UserResponse{
			ID:          user.ID,
			Username:    user.Username,
			CreatedAt:   user.CreatedAt,
			LastLoginAt: user.LastLoginAt,
		},
	})
}

func (h *AuthHandler) Logout(c *fiber.Ctx) error {
	sessionID := c.Cookies(middleware.SessionCookieName)
	if sessionID != "" {
		if user, ok := c.Locals("user").(*models.User); ok && user != nil {
			_ = h.auditLogger.Log(&user.ID, "LOGOUT", "auth", user.Username, "User logged out")
		}
		_ = h.authService.DeleteSession(sessionID)
	}

	c.ClearCookie(middleware.SessionCookieName)

	return c.JSON(models.APIResponse{
		Data: fiber.Map{
			"message": "Logged out successfully",
		},
	})
}

func (h *AuthHandler) Me(c *fiber.Ctx) error {
	hasUsers, err := h.authService.HasAnyUser()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "DATABASE_ERROR",
				Message: "Failed to check setup state",
			},
		})
	}

	if !hasUsers {
		return c.JSON(models.APIResponse{
			Data: fiber.Map{
				"authenticated":  false,
				"setup_required": true,
			},
		})
	}

	sessionID := c.Cookies(middleware.SessionCookieName)
	if sessionID == "" {
		return c.JSON(models.APIResponse{
			Data: fiber.Map{
				"authenticated":  false,
				"setup_required": false,
			},
		})
	}

	user, err := h.authService.GetSessionUser(sessionID)
	if err != nil {
		c.ClearCookie(middleware.SessionCookieName)
		return c.JSON(models.APIResponse{
			Data: fiber.Map{
				"authenticated":  false,
				"setup_required": false,
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: fiber.Map{
			"authenticated":  true,
			"setup_required": false,
			"user": models.UserResponse{
				ID:          user.ID,
				Username:    user.Username,
				CreatedAt:   user.CreatedAt,
				LastLoginAt: user.LastLoginAt,
			},
		},
	})
}
