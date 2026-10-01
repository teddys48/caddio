package caddy

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"caddy-manager/internal/models"
)

const CaddioSiteIDPrefix = "caddio_site_"

var (
	ErrUnmanagedRoute    = errors.New("cannot modify or delete unmanaged Caddy routes")
	ErrSiteNotFound      = errors.New("site not found")
	ErrCaddyApplyFailed  = errors.New("failed to apply configuration to Caddy")
	ErrInvalidSiteConfig = errors.New("invalid site configuration")
)

type ConfigStore interface {
	GetConfig(ctx context.Context) ([]byte, error)
	LoadConfig(ctx context.Context, configBytes []byte) error
}

// GenerateSiteID creates a deterministic Caddio-managed site ID from domain name
func GenerateSiteID(domain string) string {
	cleanDomain := strings.ToLower(strings.TrimSpace(domain))
	hash := sha256.Sum256([]byte(cleanDomain))
	slug := strings.ReplaceAll(cleanDomain, ".", "_")
	slug = strings.ReplaceAll(slug, "-", "_")

	// Max 32 chars slug + hash suffix
	if len(slug) > 32 {
		slug = slug[:32]
	}
	return fmt.Sprintf("%s%s_%s", CaddioSiteIDPrefix, slug, hex.EncodeToString(hash[:4]))
}

func IsCaddioManagedID(id string) bool {
	return strings.HasPrefix(id, CaddioSiteIDPrefix)
}

func (c *Client) GetSites(ctx context.Context) (models.SiteDiscoveryResult, error) {
	configBytes, err := c.GetConfig(ctx)
	if err != nil {
		return models.SiteDiscoveryResult{}, fmt.Errorf("failed to fetch caddy config: %w", err)
	}

	return ParseSitesFromConfig(configBytes)
}

func ParseSitesFromConfig(configBytes []byte) (models.SiteDiscoveryResult, error) {
	result := models.SiteDiscoveryResult{
		Sites:          []models.Site{},
		UnmanagedCount: 0,
	}

	if len(configBytes) == 0 {
		return result, nil
	}

	var root map[string]interface{}
	if err := json.Unmarshal(configBytes, &root); err != nil {
		return result, fmt.Errorf("invalid caddy json: %w", err)
	}

	apps, ok := root["apps"].(map[string]interface{})
	if !ok {
		return result, nil
	}

	httpApp, ok := apps["http"].(map[string]interface{})
	if !ok {
		return result, nil
	}

	servers, ok := httpApp["servers"].(map[string]interface{})
	if !ok {
		return result, nil
	}

	// Dynamically iterate across ALL server blocks (no hardcoded srv0)
	for serverKey, serverVal := range servers {
		serverMap, ok := serverVal.(map[string]interface{})
		if !ok {
			continue
		}

		routesRaw, ok := serverMap["routes"].([]interface{})
		if !ok {
			continue
		}

		serverListens := []string{}
		if listensRaw, ok := serverMap["listen"].([]interface{}); ok {
			for _, l := range listensRaw {
				if s, ok := l.(string); ok {
					serverListens = append(serverListens, s)
				}
			}
		}

		isHTTPS := checkHTTPSFromListens(serverListens)

		for idx, routeRaw := range routesRaw {
			routeMap, ok := routeRaw.(map[string]interface{})
			if !ok {
				result.UnmanagedCount++
				continue
			}

			site, supported := extractSiteFromRoute(routeMap, serverKey, idx, isHTTPS)
			if !supported {
				result.UnmanagedCount++
				continue
			}

			result.Sites = append(result.Sites, site)
		}
	}

	return result, nil
}

func checkHTTPSFromListens(listens []string) bool {
	for _, l := range listens {
		if strings.Contains(l, "443") || strings.HasPrefix(l, ":443") {
			return true
		}
	}
	// Default Caddy behavior is automatic HTTPS
	return true
}

func extractSiteFromRoute(routeMap map[string]interface{}, serverKey string, routeIdx int, defaultHTTPS bool) (models.Site, bool) {
	routeID, _ := routeMap["@id"].(string)
	isManaged := IsCaddioManagedID(routeID)

	domain := extractDomainFromRoute(routeMap)
	if domain == "" {
		return models.Site{}, false
	}

	upstreams := extractUpstreamsFromRoute(routeMap)
	if len(upstreams) == 0 {
		return models.Site{}, false
	}

	if routeID == "" {
		routeID = fmt.Sprintf("ext_%s_%d", serverKey, routeIdx)
	}

	return models.Site{
		ID:        routeID,
		Name:      domain,
		Domain:    domain,
		Upstreams: upstreams,
		Enabled:   true,
		HTTPS:     defaultHTTPS,
		IsManaged: isManaged,
	}, true
}

func extractDomainFromRoute(routeMap map[string]interface{}) string {
	// 1. Direct matchers in routeMap
	if matchers, ok := routeMap["match"].([]interface{}); ok {
		for _, mRaw := range matchers {
			if mMap, ok := mRaw.(map[string]interface{}); ok {
				if hostsRaw, ok := mMap["host"].([]interface{}); ok && len(hostsRaw) > 0 {
					if h, ok := hostsRaw[0].(string); ok && h != "" {
						return h
					}
				}
			}
		}
	}

	// 2. Check nested subroutes in handle
	if handlers, ok := routeMap["handle"].([]interface{}); ok {
		for _, hRaw := range handlers {
			if hMap, ok := hRaw.(map[string]interface{}); ok {
				if handlerType, _ := hMap["handler"].(string); handlerType == "subroute" {
					if subRoutes, ok := hMap["routes"].([]interface{}); ok {
						for _, srRaw := range subRoutes {
							if srMap, ok := srRaw.(map[string]interface{}); ok {
								if dom := extractDomainFromRoute(srMap); dom != "" {
									return dom
								}
							}
						}
					}
				}
			}
		}
	}

	return ""
}

func extractUpstreamsFromRoute(routeMap map[string]interface{}) []models.Upstream {
	var upstreams []models.Upstream

	handlers, ok := routeMap["handle"].([]interface{})
	if !ok || len(handlers) == 0 {
		return upstreams
	}

	for _, hRaw := range handlers {
		hMap, ok := hRaw.(map[string]interface{})
		if !ok {
			continue
		}
		handlerType, _ := hMap["handler"].(string)
		if handlerType == "reverse_proxy" {
			if upstreamsRaw, ok := hMap["upstreams"].([]interface{}); ok {
				for _, uRaw := range upstreamsRaw {
					if uMap, ok := uRaw.(map[string]interface{}); ok {
						if dial, ok := uMap["dial"].(string); ok && dial != "" {
							upstreams = append(upstreams, models.Upstream{Address: dial})
						}
					}
				}
			}
		} else if handlerType == "subroute" {
			if subRoutes, ok := hMap["routes"].([]interface{}); ok {
				for _, srRaw := range subRoutes {
					if srMap, ok := srRaw.(map[string]interface{}); ok {
						subUpstreams := extractUpstreamsFromRoute(srMap)
						upstreams = append(upstreams, subUpstreams...)
					}
				}
			}
		}
	}

	return upstreams
}

// BuildCaddyRouteMap constructs the exact Caddy route map for a Caddio-managed site
func BuildCaddyRouteMap(site models.Site) map[string]interface{} {
	if site.ID == "" || !IsCaddioManagedID(site.ID) {
		site.ID = GenerateSiteID(site.Domain)
	}

	upstreamsList := []map[string]interface{}{}
	for _, u := range site.Upstreams {
		addr := strings.TrimSpace(u.Address)
		// Strip http:// or https:// if provided in dial address
		addr = strings.TrimPrefix(addr, "http://")
		addr = strings.TrimPrefix(addr, "https://")
		upstreamsList = append(upstreamsList, map[string]interface{}{
			"dial": addr,
		})
	}

	return map[string]interface{}{
		"@id": site.ID,
		"match": []map[string]interface{}{
			{
				"host": []string{site.Domain},
			},
		},
		"handle": []map[string]interface{}{
			{
				"handler":   "reverse_proxy",
				"upstreams": upstreamsList,
			},
		},
	}
}

func (c *Client) ApplySite(ctx context.Context, site models.Site) error {
	return ApplySiteToConfigStore(ctx, c, site)
}

func ApplySiteOnMock(store ConfigStore, ctx context.Context, site models.Site) error {
	return ApplySiteToConfigStore(ctx, store, site)
}

func ApplySiteToConfigStore(ctx context.Context, store ConfigStore, site models.Site) error {
	if site.ID != "" && !IsCaddioManagedID(site.ID) {
		return ErrUnmanagedRoute
	}

	site.ID = GenerateSiteID(site.Domain)
	site.IsManaged = true

	configBytes, err := store.GetConfig(ctx)
	if err != nil {
		configBytes = []byte(`{"apps":{"http":{"servers":{"srv0":{"listen":[":443",":80"],"routes":[]}}}}}`)
	}

	var root map[string]interface{}
	if err := json.Unmarshal(configBytes, &root); err != nil {
		root = make(map[string]interface{})
	}

	// Dynamically locate or initialize apps.http.servers
	apps, ok := root["apps"].(map[string]interface{})
	if !ok {
		apps = make(map[string]interface{})
		root["apps"] = apps
	}

	httpApp, ok := apps["http"].(map[string]interface{})
	if !ok {
		httpApp = make(map[string]interface{})
		apps["http"] = httpApp
	}

	servers, ok := httpApp["servers"].(map[string]interface{})
	if !ok || len(servers) == 0 {
		servers = make(map[string]interface{})
		httpApp["servers"] = servers
	}

	// Dynamically find existing server block or pick/create target server block
	targetServerKey := ""
	for sKey, sVal := range servers {
		if sMap, ok := sVal.(map[string]interface{}); ok {
			if _, ok := sMap["routes"]; ok {
				targetServerKey = sKey
				break
			}
		}
	}

	if targetServerKey == "" {
		targetServerKey = "srv0"
		servers["srv0"] = map[string]interface{}{
			"listen": []string{":443", ":80"},
			"routes": []interface{}{},
		}
	}

	targetServer := servers[targetServerKey].(map[string]interface{})
	routesRaw, _ := targetServer["routes"].([]interface{})

	newRoute := BuildCaddyRouteMap(site)
	updatedRoutes := []interface{}{}
	replaced := false

	for _, rRaw := range routesRaw {
		rMap, ok := rRaw.(map[string]interface{})
		if ok {
			if id, ok := rMap["@id"].(string); ok && id == site.ID {
				updatedRoutes = append(updatedRoutes, newRoute)
				replaced = true
				continue
			}
		}
		// Preserve ALL unrelated routes 100% intact!
		updatedRoutes = append(updatedRoutes, rRaw)
	}

	if !replaced {
		updatedRoutes = append(updatedRoutes, newRoute)
	}

	targetServer["routes"] = updatedRoutes

	newConfigBytes, err := json.Marshal(root)
	if err != nil {
		return fmt.Errorf("failed to marshal updated caddy config: %w", err)
	}

	return store.LoadConfig(ctx, newConfigBytes)
}

func (c *Client) DeleteSite(ctx context.Context, siteID string) error {
	return DeleteSiteFromConfigStore(ctx, c, siteID)
}

func DeleteSiteOnMock(store ConfigStore, ctx context.Context, siteID string) error {
	return DeleteSiteFromConfigStore(ctx, store, siteID)
}

func DeleteSiteFromConfigStore(ctx context.Context, store ConfigStore, siteID string) error {
	if !IsCaddioManagedID(siteID) {
		return ErrUnmanagedRoute
	}

	configBytes, err := store.GetConfig(ctx)
	if err != nil {
		return fmt.Errorf("failed to fetch caddy config: %w", err)
	}

	var root map[string]interface{}
	if err := json.Unmarshal(configBytes, &root); err != nil {
		return fmt.Errorf("invalid caddy config: %w", err)
	}

	apps, ok := root["apps"].(map[string]interface{})
	if !ok {
		return ErrSiteNotFound
	}
	httpApp, ok := apps["http"].(map[string]interface{})
	if !ok {
		return ErrSiteNotFound
	}
	servers, ok := httpApp["servers"].(map[string]interface{})
	if !ok {
		return ErrSiteNotFound
	}

	found := false
	for _, sVal := range servers {
		sMap, ok := sVal.(map[string]interface{})
		if !ok {
			continue
		}
		routesRaw, ok := sMap["routes"].([]interface{})
		if !ok {
			continue
		}

		filteredRoutes := []interface{}{}
		for _, rRaw := range routesRaw {
			rMap, ok := rRaw.(map[string]interface{})
			if ok {
				if id, ok := rMap["@id"].(string); ok && id == siteID {
					found = true
					continue
				}
			}
			// Keep unrelated routes intact
			filteredRoutes = append(filteredRoutes, rRaw)
		}
		sMap["routes"] = filteredRoutes
	}

	if !found {
		return ErrSiteNotFound
	}

	newConfigBytes, err := json.Marshal(root)
	if err != nil {
		return fmt.Errorf("failed to marshal caddy config: %w", err)
	}

	return store.LoadConfig(ctx, newConfigBytes)
}

func (c *Client) LoadConfig(ctx context.Context, configBytes []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.adminURL+"/load", strings.NewReader(string(configBytes)))
	if err != nil {
		return fmt.Errorf("failed to create caddy load request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to communicate with caddy admin api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("%w: status %d - %s", ErrCaddyApplyFailed, resp.StatusCode, string(body))
	}

	return nil
}
