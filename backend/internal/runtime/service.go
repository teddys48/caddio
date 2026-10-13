package runtime

import (
	"context"
	"encoding/json"
	"fmt"

	"caddy-manager/internal/caddy"
	"caddy-manager/internal/configstore"
	"caddy-manager/internal/models"
	"caddy-manager/internal/sites"
)

type Service struct {
	healthChecker *HealthChecker
	logManager    *LogManager
	configService *configstore.Service
}

func NewService(caddyClient caddy.CaddyClient, versionProvider caddy.SystemVersionProvider, sitesService *sites.Service, configService *configstore.Service) *Service {
	return &Service{
		healthChecker: NewHealthChecker(caddyClient, versionProvider, sitesService),
		logManager:    NewLogManager(caddyClient),
		configService: configService,
	}
}

func (s *Service) GetStatus(ctx context.Context) RuntimeStatusResponse {
	logsRes := s.logManager.GetLogs(ctx, RuntimeLogFilters{Limit: 1})
	return s.healthChecker.CheckStatus(ctx, logsRes.Available, logsRes.Source)
}

func (s *Service) GetLogs(ctx context.Context, filters RuntimeLogFilters) RuntimeLogsResponse {
	return s.logManager.GetLogs(ctx, filters)
}

func (s *Service) SetMockLogs(entries []RuntimeLogEntry) {
	s.logManager.SetMockLogs(entries)
}

// EnableLogging safely provisions Caddy stdout JSON logging using the Phase 3 safe-apply pipeline.
func (s *Service) EnableLogging(ctx context.Context, userID int64, username string, req EnableLoggingRequest) (*EnableLoggingResponse, error) {
	if s.configService == nil {
		return nil, fmt.Errorf("config service unavailable")
	}

	// 1. Fetch current active configuration
	activeRes, err := s.configService.GetConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch current caddy config: %w", err)
	}

	// 2. Optimistic Concurrency Check
	if req.BaseHash != "" && req.BaseHash != activeRes.Hash {
		return &EnableLoggingResponse{
			Changed: false,
			Reason:  "CONFIG_CONFLICT",
			Message: "Configuration has been modified since page load",
		}, configstore.ErrConfigConflict
	}

	// 3. Inspect existing JSON structure
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(activeRes.Config), &root); err != nil {
		return nil, fmt.Errorf("invalid caddy json: %w", err)
	}

	// In Caddy v2 JSON layout, "logging" is at root level: root["logging"]
	if loggingRaw, hasLogging := root["logging"]; hasLogging && loggingRaw != nil {
		if loggingMap, ok := loggingRaw.(map[string]interface{}); ok {
			if _, hasLogs := loggingMap["logs"]; hasLogs {
				return &EnableLoggingResponse{
					Changed: false,
					Message: "Caddy logging is already configured",
					Hash:    activeRes.Hash,
				}, nil
			}
		}
		return &EnableLoggingResponse{
			Changed: false,
			Reason:  "existing_logging_configuration",
			Message: "Custom logging configuration already exists",
			Hash:    activeRes.Hash,
		}, nil
	}

	// Case B: No logging configuration -> Inject default root-level stdout JSON logging safely
	root["logging"] = map[string]interface{}{
		"logs": map[string]interface{}{
			"default": map[string]interface{}{
				"writer": map[string]interface{}{
					"output": "stdout",
				},
				"encoder": map[string]interface{}{
					"format": "json",
				},
			},
		},
	}

	proposedBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to format proposed json: %w", err)
	}

	// 4. Delegate to Phase 3 ApplyConfig (snapshot -> validate -> apply -> post-apply verify -> audit log)
	applyReq := models.ApplyConfigRequest{
		BaseHash: activeRes.Hash,
		Config:   string(proposedBytes),
	}

	applyRes, err := s.configService.ApplyConfig(ctx, userID, username, applyReq)
	if err != nil {
		return nil, err
	}

	return &EnableLoggingResponse{
		Changed: true,
		Message: "Caddy stdout JSON logging enabled successfully",
		Hash:    applyRes.Hash,
	}, nil
}
