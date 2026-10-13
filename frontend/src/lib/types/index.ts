export interface User {
  id: number;
  username: string;
  created_at: string;
  last_login_at?: string;
}

export interface AuthState {
  authenticated: boolean;
  setup_required: boolean;
  user?: User;
}

export interface CaddyStatus {
  status: 'Connected' | 'Disconnected' | 'Unknown';
  version?: string;
}

export interface AppHealth {
  status: 'healthy' | 'unhealthy';
  database: 'connected' | 'disconnected';
  caddy_status: 'Connected' | 'Disconnected' | 'Unknown';
  setup_required: boolean;
}

export interface AuditLog {
  id: number;
  user_id?: number;
  username?: string;
  action: string;
  resource: string;
  resource_id?: string;
  metadata?: string;
  created_at: string;
}

export interface Upstream {
  address: string;
}

export interface Site {
  id: string;
  name: string;
  domain: string;
  upstreams: Upstream[];
  enabled: boolean;
  https: boolean;
  is_managed: boolean;
}

export interface SiteDiscoveryResult {
  sites: Site[];
  unmanaged_count: number;
}

export interface SiteSummaryMetrics {
  total_sites: number;
  enabled_sites: number;
  https_sites: number;
  managed_sites: number;
}

export interface SaveSiteRequest {
  id?: string;
  name: string;
  domain: string;
  upstreams: Upstream[];
  enabled: boolean;
  https?: boolean;
}

export interface ValidationResult {
  valid: boolean;
  message: string;
  details?: string[];
}

export interface DashboardData {
  caddy: CaddyStatus;
  health: AppHealth;
  recent_logs: AuditLog[];
  user_count: number;
  has_user: boolean;
  site_summary?: SiteSummaryMetrics;
}

export interface APIError {
  code: string;
  message: string;
}

export interface APIResponse<T> {
  data?: T;
  error?: APIError;
}

export interface ConfigResponse {
  config: string;
  version?: string;
  hash: string;
  retrieved_at: string;
}

export interface ValidateConfigRequest {
  base_hash?: string;
  config: string;
}

export interface ValidateConfigResponse {
  valid: boolean;
  message: string;
  details?: string[];
  current_hash: string;
  has_conflict: boolean;
}

export interface DiffConfigRequest {
  base_hash: string;
  config: string;
}

export interface DiffStats {
  additions: number;
  deletions: number;
  modified: number;
}

export interface ConfigDiffResponse {
  current_hash: string;
  base_hash: string;
  has_conflict: boolean;
  unified_diff: string;
  stats: DiffStats;
}

export interface ApplyConfigRequest {
  base_hash: string;
  config: string;
}

export interface ConfigHistoryItem {
  id: string;
  config_hash: string;
  created_by: string;
  created_at: string;
  source: string;
  size_bytes: number;
}

export interface ConfigSnapshot {
  id: string;
  config?: string;
  config_hash: string;
  created_by: string;
  created_at: string;
  source: string;
  size_bytes: number;
}

export interface RollbackRequest {
  base_hash: string;
}

export interface RuntimeStatusResponse {
  status: 'healthy' | 'degraded' | 'unreachable' | 'unknown';
  reachable: boolean;
  active_config_hash: string;
  caddy_version?: string;
  last_checked: string;
  server_count: number;
  site_count: number;
  logs_available: boolean;
  log_source?: string;
}

export interface RuntimeLogEntry {
  timestamp: string;
  level: string;
  logger?: string;
  message: string;
  fields?: Record<string, unknown>;
}

export interface RuntimeLogsResponse {
  available: boolean;
  reason?: string;
  source?: string;
  entries: RuntimeLogEntry[];
  has_more: boolean;
  next_cursor?: string;
}

export interface RuntimeLogFilters {
  level?: string;
  search?: string;
  limit?: number;
  cursor?: string;
}

