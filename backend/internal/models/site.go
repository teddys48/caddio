package models

// Upstream represents a backend target server address (host:port or URL).
type Upstream struct {
	Address string `json:"address"`
}

// Site represents a reverse proxy site managed or discovered in Caddy.
type Site struct {
	ID        string     `json:"id"`
	Name      string     `json:"name"`
	Domain    string     `json:"domain"`
	Upstreams []Upstream `json:"upstreams"`
	Enabled   bool       `json:"enabled"`
	HTTPS     bool       `json:"https"`
	IsManaged bool       `json:"is_managed"`
}

// SiteDiscoveryResult wraps discovered sites and unmanaged route stats.
type SiteDiscoveryResult struct {
	Sites          []Site `json:"sites"`
	UnmanagedCount int    `json:"unmanaged_count"`
}

// SiteSummaryMetrics wraps operational site counts for the dashboard.
type SiteSummaryMetrics struct {
	TotalSites   int `json:"total_sites"`
	EnabledSites int `json:"enabled_sites"`
	HTTPSSites   int `json:"https_sites"`
	ManagedSites int `json:"managed_sites"`
}

// SaveSiteRequest payload for create or update operations.
type SaveSiteRequest struct {
	ID        string     `json:"id,omitempty"`
	Name      string     `json:"name"`
	Domain    string     `json:"domain"`
	Upstreams []Upstream `json:"upstreams"`
	Enabled   bool       `json:"enabled"`
	HTTPS     bool       `json:"https"`
}

// ValidationResult represents dry-run validation status.
type ValidationResult struct {
	Valid   bool     `json:"valid"`
	Message string   `json:"message"`
	Details []string `json:"details,omitempty"`
}
