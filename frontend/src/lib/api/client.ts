import type {
  APIResponse, AuthState, DashboardData, AuditLog, User, Site, SiteDiscoveryResult, SaveSiteRequest, ValidationResult,
  ConfigResponse, ValidateConfigRequest, ValidateConfigResponse, DiffConfigRequest, ConfigDiffResponse, ApplyConfigRequest,
  ConfigHistoryItem, ConfigSnapshot, RollbackRequest, RuntimeStatusResponse, RuntimeLogsResponse, RuntimeLogFilters
} from '../types';

async function request<T>(url: string, options: RequestInit = {}): Promise<T> {
  const defaultHeaders: Record<string, string> = {
    'Accept': 'application/json',
  };

  if (options.body && typeof options.body === 'string') {
    defaultHeaders['Content-Type'] = 'application/json';
  }

  const response = await fetch(url, {
    ...options,
    credentials: 'include',
    headers: {
      ...defaultHeaders,
      ...options.headers,
    },
  });

  const json = await response.json();

  if (!response.ok) {
    const errorMsg = json.error || json.error?.message || `HTTP error ${response.status}`;
    throw new Error(typeof errorMsg === 'string' ? errorMsg : JSON.stringify(errorMsg));
  }

  if (json.data !== undefined) {
    return json.data as T;
  }

  return json as T;
}

export const api = {
  // Auth
  async checkAuth(): Promise<AuthState> {
    return request<AuthState>('/api/auth/me');
  },

  async setup(payload: { username: string; password: string; confirm_password: string }): Promise<User> {
    return request<User>('/api/auth/setup', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async login(payload: { username: string; password: string }): Promise<User> {
    return request<User>('/api/auth/login', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async logout(): Promise<{ message: string }> {
    return request<{ message: string }>('/api/auth/logout', {
      method: 'POST',
    });
  },

  // System
  async getSystemCaddyVersion(): Promise<{ version: string }> {
    return request<{ version: string }>('/api/system/caddy/version');
  },

  // Operational Dashboard
  async getDashboard(): Promise<DashboardData> {
    return request<DashboardData>('/api/dashboard');
  },

  // Audit Logs
  async getAuditLogs(limit = 50): Promise<AuditLog[]> {
    return request<AuditLog[]>(`/api/audit-logs?limit=${limit}`);
  },

  // Sites Management
  async getSites(): Promise<SiteDiscoveryResult> {
    return request<SiteDiscoveryResult>('/api/sites');
  },

  async getSite(id: string): Promise<Site> {
    return request<Site>(`/api/sites/${encodeURIComponent(id)}`);
  },

  async validateSite(payload: SaveSiteRequest): Promise<ValidationResult> {
    return request<ValidationResult>('/api/sites/validate', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async createSite(payload: SaveSiteRequest): Promise<Site> {
    return request<Site>('/api/sites', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async updateSite(id: string, payload: SaveSiteRequest): Promise<Site> {
    return request<Site>(`/api/sites/${encodeURIComponent(id)}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    });
  },

  async deleteSite(id: string): Promise<{ message: string }> {
    return request<{ message: string }>(`/api/sites/${encodeURIComponent(id)}`, {
      method: 'DELETE',
    });
  },

  // Config Workspace API
  async getConfig(): Promise<ConfigResponse> {
    return request<ConfigResponse>('/api/config');
  },

  async validateConfig(payload: ValidateConfigRequest): Promise<ValidateConfigResponse> {
    return request<ValidateConfigResponse>('/api/config/validate', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async diffConfig(payload: DiffConfigRequest): Promise<ConfigDiffResponse> {
    return request<ConfigDiffResponse>('/api/config/diff', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async applyConfig(payload: ApplyConfigRequest): Promise<ConfigResponse> {
    return request<ConfigResponse>('/api/config/apply', {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  async getConfigHistory(): Promise<ConfigHistoryItem[]> {
    return request<ConfigHistoryItem[]>('/api/config/history');
  },

  async getConfigSnapshot(id: string): Promise<ConfigSnapshot> {
    return request<ConfigSnapshot>(`/api/config/history/${encodeURIComponent(id)}`);
  },

  async rollbackConfig(id: string, payload: RollbackRequest): Promise<ConfigResponse> {
    return request<ConfigResponse>(`/api/config/history/${encodeURIComponent(id)}/rollback`, {
      method: 'POST',
      body: JSON.stringify(payload),
    });
  },

  // Runtime Monitoring & Logs
  async getRuntimeStatus(): Promise<RuntimeStatusResponse> {
    return request<RuntimeStatusResponse>('/api/runtime/status');
  },

  async getRuntimeLogs(filters: RuntimeLogFilters = {}): Promise<RuntimeLogsResponse> {
    const params = new URLSearchParams();
    if (filters.level && filters.level !== 'all') {
      params.append('level', filters.level);
    }
    if (filters.search) {
      params.append('search', filters.search);
    }
    if (filters.limit) {
      params.append('limit', filters.limit.toString());
    }
    if (filters.cursor) {
      params.append('cursor', filters.cursor);
    }
    const queryString = params.toString();
    const url = `/api/runtime/logs${queryString ? `?${queryString}` : ''}`;
    return request<RuntimeLogsResponse>(url);
  },

  async enableRuntimeLogging(baseHash: string): Promise<{ changed: boolean; message: string; reason?: string; hash?: string }> {
    return request<{ changed: boolean; message: string; reason?: string; hash?: string }>('/api/runtime/enable-logging', {
      method: 'POST',
      body: JSON.stringify({ base_hash: baseHash }),
    });
  },
};

