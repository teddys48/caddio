package handlers

import (
	"errors"
	"net/http"

	"caddy-manager/internal/caddy"
	"caddy-manager/internal/models"
	"caddy-manager/internal/sites"

	"github.com/gofiber/fiber/v2"
)

type SitesHandler struct {
	sitesService *sites.Service
}

func NewSitesHandler(sitesService *sites.Service) *SitesHandler {
	return &SitesHandler{
		sitesService: sitesService,
	}
}

// GetSites returns all discovered sites and unmanaged count
func (h *SitesHandler) GetSites(c *fiber.Ctx) error {
	result, err := h.sitesService.GetSites(c.Context())
	if err != nil {
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "CADDY_CONFIG_ERROR",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: result,
	})
}

// GetSiteByID returns details of a single site by ID
func (h *SitesHandler) GetSiteByID(c *fiber.Ctx) error {
	id := c.Params("id")
	site, err := h.sitesService.GetSiteByID(c.Context(), id)
	if err != nil {
		if errors.Is(err, caddy.ErrSiteNotFound) {
			return c.Status(http.StatusNotFound).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "SITE_NOT_FOUND",
					Message: "Site not found",
				},
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INTERNAL_ERROR",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: site,
	})
}

// ValidateSite performs dry-run validation on site configuration
func (h *SitesHandler) ValidateSite(c *fiber.Ctx) error {
	var req models.SaveSiteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_BODY",
				Message: "Invalid JSON request body",
			},
		})
	}

	res := h.sitesService.ValidateSite(req)
	return c.JSON(models.APIResponse{
		Data: res,
	})
}

// CreateSite validates, applies, and logs new site
func (h *SitesHandler) CreateSite(c *fiber.Ctx) error {
	var req models.SaveSiteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_BODY",
				Message: "Invalid JSON request body",
			},
		})
	}

	req.ID = "" // Force auto-generated ID for new creation

	userID := getUserIDFromLocals(c)
	site, err := h.sitesService.SaveAndApplySite(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, caddy.ErrInvalidSiteConfig) {
			return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "INVALID_CONFIG",
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

	return c.Status(http.StatusCreated).JSON(models.APIResponse{
		Data: site,
	})
}

// UpdateSite validates, updates, applies, and logs managed site
func (h *SitesHandler) UpdateSite(c *fiber.Ctx) error {
	id := c.Params("id")
	var req models.SaveSiteRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "INVALID_BODY",
				Message: "Invalid JSON request body",
			},
		})
	}

	req.ID = id

	userID := getUserIDFromLocals(c)
	site, err := h.sitesService.SaveAndApplySite(c.Context(), userID, req)
	if err != nil {
		if errors.Is(err, caddy.ErrUnmanagedRoute) {
			return c.Status(http.StatusForbidden).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "UNMANAGED_ROUTE",
					Message: "Cannot modify route not managed by Caddio",
				},
			})
		}
		if errors.Is(err, caddy.ErrInvalidSiteConfig) {
			return c.Status(http.StatusBadRequest).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "INVALID_CONFIG",
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
		Data: site,
	})
}

// DeleteSite removes managed site from Caddy
func (h *SitesHandler) DeleteSite(c *fiber.Ctx) error {
	id := c.Params("id")
	userID := getUserIDFromLocals(c)

	if err := h.sitesService.DeleteSite(c.Context(), userID, id); err != nil {
		if errors.Is(err, caddy.ErrUnmanagedRoute) {
			return c.Status(http.StatusForbidden).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "UNMANAGED_ROUTE",
					Message: "Cannot delete route not managed by Caddio",
				},
			})
		}
		if errors.Is(err, caddy.ErrSiteNotFound) {
			return c.Status(http.StatusNotFound).JSON(models.APIResponse{
				Error: &models.APIError{
					Code:    "SITE_NOT_FOUND",
					Message: "Site not found",
				},
			})
		}
		return c.Status(http.StatusInternalServerError).JSON(models.APIResponse{
			Error: &models.APIError{
				Code:    "DELETE_FAILED",
				Message: err.Error(),
			},
		})
	}

	return c.JSON(models.APIResponse{
		Data: fiber.Map{
			"message": "Site deleted successfully",
		},
	})
}

func getUserIDFromLocals(c *fiber.Ctx) int64 {
	u := c.Locals("user")
	if u == nil {
		return 0
	}
	if user, ok := u.(*models.User); ok && user != nil {
		return user.ID
	}
	return 0
}
