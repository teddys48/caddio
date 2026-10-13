package runtime

import (
	"context"
	"encoding/json"
	"time"

	"caddy-manager/internal/caddy"
	"caddy-manager/internal/configstore"
	"caddy-manager/internal/sites"
)

type HealthChecker struct {
	caddyClient     caddy.CaddyClient
	versionProvider caddy.SystemVersionProvider
	sitesService    *sites.Service
}

func NewHealthChecker(caddyClient caddy.CaddyClient, versionProvider caddy.SystemVersionProvider, sitesService *sites.Service) *HealthChecker {
	return &HealthChecker{
		caddyClient:     caddyClient,
		versionProvider: versionProvider,
		sitesService:    sitesService,
	}
}

func (h *HealthChecker) CheckStatus(ctx context.Context, logsAvailable bool, logSource string) RuntimeStatusResponse {
	checkedAt := time.Now().UTC()

	// 1. Query Caddy Admin API for active config with 3s context timeout
	reqCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	configBytes, err := h.caddyClient.GetConfig(reqCtx)
	if err != nil {
		return RuntimeStatusResponse{
			Status:        "unreachable",
			Reachable:     false,
			LastChecked:   checkedAt,
			LogsAvailable: false,
			LogSource:     logSource,
		}
	}

	// 2. Reuse Phase 3 configstore.NormalizeJSON for SHA-256 active config hash calculation
	_, hashHex, err := configstore.NormalizeJSON(string(configBytes))
	if err != nil {
		hashHex = ""
	}

	// 3. Count HTTP servers and sites dynamically from active Caddy config (reusing Phase 2 logic)
	serverCount := countServers(configBytes)
	siteCount := 0
	if h.sitesService != nil {
		if siteResult, err := h.sitesService.GetSites(reqCtx); err == nil {
			siteCount = len(siteResult.Sites)
		}
	} else if siteResult, err := caddy.ParseSitesFromConfig(configBytes); err == nil {
		siteCount = len(siteResult.Sites)
	}

	// 4. Determine authoritative Caddy Version
	var caddyVerPtr *string
	caddyVer := ""
	if h.versionProvider != nil {
		if ver, err := h.versionProvider.GetBinaryVersion(reqCtx); err == nil && ver != "" {
			caddyVer = ver
		}
	}
	if caddyVer == "" {
		if ver, err := h.caddyClient.GetVersion(reqCtx); err == nil && ver != "" {
			caddyVer = ver
		}
	}
	if caddyVer != "" {
		caddyVerPtr = &caddyVer
	}

	// Caddy Admin API is reachable and core runtime info is available -> healthy
	status := "healthy"

	return RuntimeStatusResponse{
		Status:           status,
		Reachable:        true,
		ActiveConfigHash: hashHex,
		CaddyVersion:     caddyVerPtr,
		LastChecked:      checkedAt,
		ServerCount:      serverCount,
		SiteCount:        siteCount,
		LogsAvailable:    logsAvailable,
		LogSource:        logSource,
	}
}

func countServers(configBytes []byte) int {
	var root map[string]interface{}
	if err := json.Unmarshal(configBytes, &root); err != nil {
		return 0
	}

	apps, ok := root["apps"].(map[string]interface{})
	if !ok {
		return 0
	}

	httpApp, ok := apps["http"].(map[string]interface{})
	if !ok {
		return 0
	}

	servers, ok := httpApp["servers"].(map[string]interface{})
	if !ok {
		return 0
	}

	return len(servers)
}
