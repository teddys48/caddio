package models

import "time"

// ConfigSnapshot represents a historical Caddy configuration snapshot stored in SQLite.
type ConfigSnapshot struct {
	ID         string    `json:"id"`
	Config     string    `json:"config,omitempty"`
	ConfigHash string    `json:"config_hash"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	Source     string    `json:"source"` // 'manual_apply', 'rollback', 'system_backup'
	SizeBytes  int       `json:"size_bytes"`
}

// ConfigResponse returned by GET /api/config
type ConfigResponse struct {
	Config      string    `json:"config"`
	Version     string    `json:"version,omitempty"`
	Hash        string    `json:"hash"`
	RetrievedAt time.Time `json:"retrieved_at"`
}

// ValidateConfigRequest payload for POST /api/config/validate
type ValidateConfigRequest struct {
	BaseHash string `json:"base_hash,omitempty"`
	Config   string `json:"config"`
}

// DiffConfigRequest payload for POST /api/config/diff
type DiffConfigRequest struct {
	BaseHash string `json:"base_hash"`
	Config   string `json:"config"`
}

// ApplyConfigRequest payload for POST /api/config/apply
type ApplyConfigRequest struct {
	BaseHash string `json:"base_hash"`
	Config   string `json:"config"`
}

// RollbackRequest payload for POST /api/config/history/:id/rollback
type RollbackRequest struct {
	BaseHash string `json:"base_hash"`
}

// DiffStats summary of additions, deletions, and modified lines
type DiffStats struct {
	Additions int `json:"additions"`
	Deletions int `json:"deletions"`
	Modified  int `json:"modified"`
}

// ConfigDiffResponse returned by POST /api/config/diff
type ConfigDiffResponse struct {
	CurrentHash string    `json:"current_hash"`
	BaseHash    string    `json:"base_hash"`
	HasConflict bool      `json:"has_conflict"`
	UnifiedDiff string    `json:"unified_diff"`
	Stats       DiffStats `json:"stats"`
}

// ConfigHistoryItem light summary for history list (without full config body)
type ConfigHistoryItem struct {
	ID         string    `json:"id"`
	ConfigHash string    `json:"config_hash"`
	CreatedBy  string    `json:"created_by"`
	CreatedAt  time.Time `json:"created_at"`
	Source     string    `json:"source"`
	SizeBytes  int       `json:"size_bytes"`
}
