package caddy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"caddy-manager/internal/models"
)

type CaddyClient interface {
	GetVersion(ctx context.Context) (string, error)
	GetConfig(ctx context.Context) ([]byte, error)
	Health(ctx context.Context) error
	GetSites(ctx context.Context) (models.SiteDiscoveryResult, error)
	ApplySite(ctx context.Context, site models.Site) error
	DeleteSite(ctx context.Context, siteID string) error
	ApplyConfig(ctx context.Context, configBytes []byte) error
}

type Client struct {
	adminURL   string
	httpClient *http.Client
}

func NewClient(adminURL string) *Client {
	adminURL = strings.TrimRight(adminURL, "/")
	return &Client{
		adminURL: adminURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

func (c *Client) Health(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.adminURL+"/config/", nil)
	if err != nil {
		return fmt.Errorf("failed to create caddy health request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("unable to connect to caddy admin api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("caddy admin api returned status code %d", resp.StatusCode)
	}

	return nil
}

func (c *Client) GetVersion(ctx context.Context) (string, error) {
	// Strategy 1: Check Server or Caddy-Version response headers from /config/
	if ver := c.tryVersionFromHeaders(ctx, c.adminURL+"/config/"); ver != "" {
		return ver, nil
	}

	// Strategy 2: Check /metrics endpoint for Prometheus caddy_build_info metric
	if ver := c.tryVersionFromMetrics(ctx); ver != "" {
		return ver, nil
	}

	// Strategy 3: Check Server response headers from root / endpoint
	if ver := c.tryVersionFromHeaders(ctx, c.adminURL+"/"); ver != "" {
		return ver, nil
	}

	// Strategy 4: Check Server response headers from /reverse_proxy/upstreams
	if ver := c.tryVersionFromHeaders(ctx, c.adminURL+"/reverse_proxy/upstreams"); ver != "" {
		return ver, nil
	}

	return "", nil
}

func (c *Client) tryVersionFromHeaders(ctx context.Context, endpoint string) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ""
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	// Check Server header (e.g. Server: Caddy/v2.8.4 or Server: Caddy/2.8.4)
	serverHeader := resp.Header.Get("Server")
	if ver := extractVersionFromHeader(serverHeader); ver != "" {
		return ver
	}

	// Check Caddy-Version or X-Caddy-Version headers
	if ver := resp.Header.Get("Caddy-Version"); ver != "" {
		return normalizeVersion(ver)
	}
	if ver := resp.Header.Get("X-Caddy-Version"); ver != "" {
		return normalizeVersion(ver)
	}

	return ""
}

func (c *Client) tryVersionFromMetrics(ctx context.Context) string {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.adminURL+"/metrics", nil)
	if err != nil {
		return ""
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return ""
	}

	// Read limited body (max 64KB)
	limitedReader := io.LimitReader(resp.Body, 64*1024)
	bodyBytes, err := io.ReadAll(limitedReader)
	if err != nil {
		return ""
	}

	body := string(bodyBytes)
	return extractVersionFromMetrics(body)
}

func extractVersionFromHeader(serverHeader string) string {
	if serverHeader == "" {
		return ""
	}

	// Matches Server: Caddy/v2.8.4, Caddy/2.8.4, Caddy/v2.9.0-rc.1, etc.
	if idx := strings.Index(serverHeader, "Caddy/"); idx != -1 {
		ver := strings.TrimSpace(serverHeader[idx+len("Caddy/"):])
		if ver != "" {
			return normalizeVersion(ver)
		}
	}

	return ""
}

var metricsVersionRegex = regexp.MustCompile(`caddy_build_info\{[^}]*version="([^"]+)"`)

func extractVersionFromMetrics(body string) string {
	matches := metricsVersionRegex.FindStringSubmatch(body)
	if len(matches) > 1 {
		ver := strings.TrimSpace(matches[1])
		if ver != "" && ver != "unknown" {
			return normalizeVersion(ver)
		}
	}
	return ""
}

func normalizeVersion(ver string) string {
	ver = strings.TrimSpace(ver)
	if ver == "" {
		return ""
	}
	if !strings.HasPrefix(ver, "v") && len(ver) > 0 && ver[0] >= '0' && ver[0] <= '9' {
		return "v" + ver
	}
	return ver
}

func (c *Client) GetConfig(ctx context.Context) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.adminURL+"/config/", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("caddy admin api returned error code %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var js json.RawMessage
	if err := json.Unmarshal(body, &js); err != nil {
		return nil, fmt.Errorf("invalid json config from caddy: %w", err)
	}

	return body, nil
}

func (c *Client) ApplyConfig(ctx context.Context, configBytes []byte) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.adminURL+"/load", bytes.NewReader(configBytes))
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
		return fmt.Errorf("caddy apply config failed: status %d - %s", resp.StatusCode, string(body))
	}

	return nil
}
