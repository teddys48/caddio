package sites

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"strings"

	"caddy-manager/internal/audit"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/models"
)

type Service struct {
	caddyClient caddy.CaddyClient
	auditLogger *audit.Logger
}

func NewService(caddyClient caddy.CaddyClient, auditLogger *audit.Logger) *Service {
	return &Service{
		caddyClient: caddyClient,
		auditLogger: auditLogger,
	}
}

// GetSites returns all discovered sites and unmanaged count
func (s *Service) GetSites(ctx context.Context) (models.SiteDiscoveryResult, error) {
	return s.caddyClient.GetSites(ctx)
}

// GetSiteByID retrieves a specific site by ID
func (s *Service) GetSiteByID(ctx context.Context, id string) (models.Site, error) {
	res, err := s.caddyClient.GetSites(ctx)
	if err != nil {
		return models.Site{}, err
	}

	for _, site := range res.Sites {
		if site.ID == id {
			return site, nil
		}
	}

	return models.Site{}, caddy.ErrSiteNotFound
}

// ValidateSite performs dry-run validation on site configuration
func (s *Service) ValidateSite(site models.SaveSiteRequest) models.ValidationResult {
	var details []string

	domain := strings.TrimSpace(site.Domain)
	if domain == "" {
		details = append(details, "Domain cannot be empty")
	} else {
		if strings.HasPrefix(domain, "http://") || strings.HasPrefix(domain, "https://") {
			details = append(details, "Domain should not include http:// or https:// scheme")
		}
		if strings.Contains(domain, " ") || strings.Contains(domain, "/") {
			details = append(details, "Domain contains invalid characters (spaces or slashes)")
		}
	}

	if len(site.Upstreams) == 0 {
		details = append(details, "At least one backend upstream address is required")
	} else {
		for i, u := range site.Upstreams {
			addr := strings.TrimSpace(u.Address)
			if addr == "" {
				details = append(details, fmt.Sprintf("Upstream #%d address cannot be empty", i+1))
				continue
			}
			// Strip scheme if present
			cleanAddr := strings.TrimPrefix(addr, "http://")
			cleanAddr = strings.TrimPrefix(cleanAddr, "https://")
			if strings.Contains(cleanAddr, "/") {
				cleanAddr = strings.Split(cleanAddr, "/")[0]
			}

			host, _, _ := net.SplitHostPort(cleanAddr)
			if host == "" {
				host = cleanAddr
			}
			if host == "" {
				details = append(details, fmt.Sprintf("Upstream #%d address is invalid (%s)", i+1, addr))
			}
		}
	}

	if len(details) > 0 {
		return models.ValidationResult{
			Valid:   false,
			Message: "Validation failed with errors",
			Details: details,
		}
	}

	return models.ValidationResult{
		Valid:   true,
		Message: "Site configuration is valid",
	}
}

// SaveAndApplySite validates, scopes ownership, applies to Caddy, and logs audit event
func (s *Service) SaveAndApplySite(ctx context.Context, userID int64, req models.SaveSiteRequest) (models.Site, error) {
	valRes := s.ValidateSite(req)
	if !valRes.Valid {
		return models.Site{}, fmt.Errorf("%w: %s", caddy.ErrInvalidSiteConfig, strings.Join(valRes.Details, "; "))
	}

	domain := strings.TrimSpace(req.Domain)
	domain = strings.TrimPrefix(domain, "http://")
	domain = strings.TrimPrefix(domain, "https://")

	var siteID string
	isUpdate := false

	if req.ID != "" {
		if !caddy.IsCaddioManagedID(req.ID) {
			return models.Site{}, caddy.ErrUnmanagedRoute
		}
		siteID = req.ID
		isUpdate = true
	} else {
		siteID = caddy.GenerateSiteID(domain)
	}

	cleanUpstreams := make([]models.Upstream, len(req.Upstreams))
	for i, u := range req.Upstreams {
		addr := strings.TrimSpace(u.Address)
		addr = strings.TrimPrefix(addr, "http://")
		addr = strings.TrimPrefix(addr, "https://")
		cleanUpstreams[i] = models.Upstream{Address: addr}
	}

	site := models.Site{
		ID:        siteID,
		Name:      req.Name,
		Domain:    domain,
		Upstreams: cleanUpstreams,
		Enabled:   req.Enabled,
		HTTPS:     true, // Caddy defaults to HTTPS
		IsManaged: true,
	}
	if site.Name == "" {
		site.Name = domain
	}

	if err := s.caddyClient.ApplySite(ctx, site); err != nil {
		return models.Site{}, fmt.Errorf("failed to apply site: %w", err)
	}

	// Audit log
	action := "site.created"
	if isUpdate {
		action = "site.updated"
	}

	metaMap := map[string]interface{}{
		"domain":    domain,
		"upstreams": cleanUpstreams,
	}
	metaBytes, _ := json.Marshal(metaMap)

	var uID *int64
	if userID > 0 {
		uID = &userID
	}
	_ = s.auditLogger.Log(uID, action, "site", siteID, string(metaBytes))

	return site, nil
}

// DeleteSite removes a site if managed by Caddio and logs audit event
func (s *Service) DeleteSite(ctx context.Context, userID int64, siteID string) error {
	if !caddy.IsCaddioManagedID(siteID) {
		return caddy.ErrUnmanagedRoute
	}

	existingSite, err := s.GetSiteByID(ctx, siteID)
	domain := ""
	if err == nil {
		domain = existingSite.Domain
	}

	if err := s.caddyClient.DeleteSite(ctx, siteID); err != nil {
		return err
	}

	metaMap := map[string]interface{}{
		"domain": domain,
	}
	metaBytes, _ := json.Marshal(metaMap)

	var uID *int64
	if userID > 0 {
		uID = &userID
	}
	_ = s.auditLogger.Log(uID, "site.deleted", "site", siteID, string(metaBytes))

	return nil
}

// GetSummaryMetrics computes summary stats for dashboard
func (s *Service) GetSummaryMetrics(ctx context.Context) (models.SiteSummaryMetrics, error) {
	discovery, err := s.caddyClient.GetSites(ctx)
	if err != nil {
		return models.SiteSummaryMetrics{}, err
	}

	metrics := models.SiteSummaryMetrics{
		TotalSites:   len(discovery.Sites),
		EnabledSites: 0,
		HTTPSSites:   0,
		ManagedSites: 0,
	}

	for _, site := range discovery.Sites {
		if site.Enabled {
			metrics.EnabledSites++
		}
		if site.HTTPS {
			metrics.HTTPSSites++
		}
		if site.IsManaged {
			metrics.ManagedSites++
		}
	}

	return metrics, nil
}
