package models

import "time"

// User represents an administrator in the system.
type User struct {
	ID           int64      `json:"id"`
	Username     string     `json:"username"`
	PasswordHash string     `json:"-"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	LastLoginAt  *time.Time `json:"last_login_at,omitempty"`
}

// Session represents an active opaque user session stored in SQLite.
type Session struct {
	ID        string    `json:"id"`
	UserID    int64     `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
	CreatedAt time.Time `json:"created_at"`
}

// Setting represents application key-value configuration.
type Setting struct {
	Key       string    `json:"key"`
	Value     string    `json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuditLog represents an audited action in the system.
type AuditLog struct {
	ID         int64     `json:"id"`
	UserID     *int64    `json:"user_id,omitempty"`
	Username   string    `json:"username,omitempty"`
	Action     string    `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id,omitempty"`
	Metadata   string    `json:"metadata,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}

// Standard API response structures
type APIResponse struct {
	Data  interface{} `json:"data,omitempty"`
	Error *APIError   `json:"error,omitempty"`
}

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// DTOs
type SetupRequest struct {
	Username        string `json:"username"`
	Password        string `json:"password"`
	ConfirmPassword string `json:"confirm_password"`
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID          int64      `json:"id"`
	Username    string     `json:"username"`
	CreatedAt   time.Time  `json:"created_at"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
}

type CaddyStatusResponse struct {
	Status  string `json:"status"` // "Connected", "Disconnected", "Unknown"
	Version string `json:"version,omitempty"`
}

type AppHealthResponse struct {
	Status        string `json:"status"`   // "healthy", "unhealthy"
	Database      string `json:"database"` // "connected", "disconnected"
	CaddyStatus   string `json:"caddy_status"`
	SetupRequired bool   `json:"setup_required"`
}

type DashboardResponse struct {
	Caddy       CaddyStatusResponse `json:"caddy"`
	Health      AppHealthResponse   `json:"health"`
	RecentLogs  []AuditLog          `json:"recent_logs"`
	UserCount   int                 `json:"user_count"`
	HasUser     bool                `json:"has_user"`
	SiteSummary SiteSummaryMetrics  `json:"site_summary"`
}
