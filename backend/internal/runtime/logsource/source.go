package logsource

import (
	"context"
	"os"
	"strings"
	"time"

	"caddy-manager/internal/caddy"
)

type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Logger    string         `json:"logger,omitempty"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

type LogFilters struct {
	Level  string `json:"level"`
	Search string `json:"search"`
	Limit  int    `json:"limit"`
	Cursor string `json:"cursor"`
}

type LogResult struct {
	Entries    []LogEntry `json:"entries"`
	HasMore    bool       `json:"has_more"`
	NextCursor string     `json:"next_cursor,omitempty"`
}

// LogSource represents a pluggable backend for retrieving Caddy runtime logs.
type LogSource interface {
	Name() string
	Available(ctx context.Context) (bool, string)
	Read(ctx context.Context, filters LogFilters) (LogResult, error)
}

// LogSourceManager selects and manages active log sources.
type LogSourceManager struct {
	sources map[string]LogSource
	mode    string
}

func NewLogSourceManager(caddyClient caddy.CaddyClient) *LogSourceManager {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("CADDIO_CADDY_LOG_SOURCE")))
	if mode == "" {
		mode = "auto"
	}

	mgr := &LogSourceManager{
		sources: make(map[string]LogSource),
		mode:    mode,
	}

	// Register log sources
	systemdSource := NewSystemdLogSource()
	mgr.sources["systemd"] = systemdSource

	if caddyClient != nil {
		caddyAdminSource := NewCaddyAdminLogSource(caddyClient)
		mgr.sources["caddy_admin"] = caddyAdminSource
	}

	return mgr
}

func (m *LogSourceManager) RegisterSource(source LogSource) {
	if source != nil {
		m.sources[source.Name()] = source
	}
}

// GetActiveSource selects the appropriate LogSource based on configuration and availability.
func (m *LogSourceManager) GetActiveSource(ctx context.Context) (LogSource, bool, string) {
	if m.mode == "none" {
		return nil, false, "logging_disabled"
	}

	if m.mode != "auto" {
		if src, ok := m.sources[m.mode]; ok {
			avail, reason := src.Available(ctx)
			return src, avail, reason
		}
		return nil, false, "unsupported_log_source"
	}

	// Mode: auto -> 1. Try systemd journald (host deployment)
	if src, ok := m.sources["systemd"]; ok {
		avail, reason := src.Available(ctx)
		if avail {
			return src, true, reason
		}
	}

	// Mode: auto -> 2. Try Caddy Admin API logging app
	if src, ok := m.sources["caddy_admin"]; ok {
		avail, reason := src.Available(ctx)
		if avail {
			return src, true, reason
		}
	}

	// Mode: auto -> 3. Try docker if registered
	if src, ok := m.sources["docker"]; ok {
		avail, reason := src.Available(ctx)
		if avail {
			return src, true, reason
		}
	}

	return nil, false, "logs_not_configured"
}
