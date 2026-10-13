package runtime

import "time"

// RuntimeStatusResponse represents the operational state and capabilities of Caddy.
type RuntimeStatusResponse struct {
	Status           string    `json:"status"` // 'healthy', 'degraded', 'unreachable', 'unknown'
	Reachable        bool      `json:"reachable"`
	ActiveConfigHash string    `json:"active_config_hash"`
	CaddyVersion     *string   `json:"caddy_version,omitempty"`
	LastChecked      time.Time `json:"last_checked"`
	ServerCount      int       `json:"server_count"`
	SiteCount        int       `json:"site_count"`
	LogsAvailable    bool      `json:"logs_available"`
	LogSource        string    `json:"log_source,omitempty"`
}

// RuntimeLogEntry represents a single normalized log entry from Caddy.
type RuntimeLogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"` // 'DEBUG', 'INFO', 'WARN', 'ERROR'
	Logger    string         `json:"logger,omitempty"`
	Message   string         `json:"message"`
	Fields    map[string]any `json:"fields,omitempty"`
}

// LogResult represents the result returned by a LogSource implementation.
type LogResult struct {
	Entries    []RuntimeLogEntry `json:"entries"`
	HasMore    bool              `json:"has_more"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

// RuntimeLogsResponse wraps log entries or explains unavailability.
type RuntimeLogsResponse struct {
	Available  bool              `json:"available"`
	Reason     string            `json:"reason,omitempty"` // 'ok', 'logs_not_configured', 'logs_unavailable', 'caddy_unreachable'
	Source     string            `json:"source,omitempty"`
	Entries    []RuntimeLogEntry `json:"entries"`
	HasMore    bool              `json:"has_more"`
	NextCursor string            `json:"next_cursor,omitempty"`
}

// RuntimeLogFilters parameters for querying runtime logs.
type RuntimeLogFilters struct {
	Level  string `json:"level"`  // 'all', 'debug', 'info', 'warn', 'error'
	Search string `json:"search"` // free text search
	Limit  int    `json:"limit"`  // default 100, max 500
	Cursor string `json:"cursor"`
}

// EnableLoggingRequest payload for POST /api/runtime/enable-logging
type EnableLoggingRequest struct {
	BaseHash string `json:"base_hash"`
}

// EnableLoggingResponse payload returned after logging provisioning
type EnableLoggingResponse struct {
	Changed bool   `json:"changed"`
	Message string `json:"message"`
	Reason  string `json:"reason,omitempty"`
	Hash    string `json:"hash,omitempty"`
}
