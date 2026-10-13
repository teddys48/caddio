package logsource

import (
	"context"
	"encoding/json"
	"time"

	"caddy-manager/internal/caddy"
)

type CaddyAdminLogSource struct {
	caddyClient caddy.CaddyClient
}

func NewCaddyAdminLogSource(caddyClient caddy.CaddyClient) *CaddyAdminLogSource {
	return &CaddyAdminLogSource{
		caddyClient: caddyClient,
	}
}

func (c *CaddyAdminLogSource) Name() string {
	return "caddy_admin"
}

func (c *CaddyAdminLogSource) Available(ctx context.Context) (bool, string) {
	if c.caddyClient == nil {
		return false, "caddy_client_nil"
	}

	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	configBytes, err := c.caddyClient.GetConfig(reqCtx)
	if err != nil {
		return false, "caddy_unreachable"
	}

	if !hasLoggingConfigured(configBytes) {
		return false, "logs_not_configured"
	}

	return true, "ok"
}

func (c *CaddyAdminLogSource) Read(ctx context.Context, filters LogFilters) (LogResult, error) {
	return LogResult{
		Entries:    []LogEntry{},
		HasMore:    false,
		NextCursor: "",
	}, nil
}

func hasLoggingConfigured(configBytes []byte) bool {
	var root map[string]interface{}
	if err := json.Unmarshal(configBytes, &root); err != nil {
		return false
	}
	// In Caddy v2 JSON structure, "logging" is at root level: root["logging"]
	logging, hasLogging := root["logging"]
	if !hasLogging || logging == nil {
		return false
	}
	loggingMap, ok := logging.(map[string]interface{})
	if !ok {
		return false
	}
	_, hasLogs := loggingMap["logs"]
	return hasLogs
}
