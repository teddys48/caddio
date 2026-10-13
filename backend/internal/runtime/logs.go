package runtime

import (
	"context"
	"encoding/json"
	"strings"

	"caddy-manager/internal/caddy"
	"caddy-manager/internal/runtime/logsource"
)

const (
	DefaultLogLimit = 100
	MaxLogLimit     = 500
)

type LogManager struct {
	sourceMgr *logsource.LogSourceManager
	mockLogs  []RuntimeLogEntry
}

func NewLogManager(caddyClient caddy.CaddyClient) *LogManager {
	return &LogManager{
		sourceMgr: logsource.NewLogSourceManager(caddyClient),
	}
}

// SetMockLogs allows unit tests to inject mock log entries
func (m *LogManager) SetMockLogs(entries []RuntimeLogEntry) {
	m.mockLogs = entries
}

func (m *LogManager) GetLogs(ctx context.Context, filters RuntimeLogFilters) RuntimeLogsResponse {
	limit := filters.Limit
	if limit <= 0 {
		limit = DefaultLogLimit
	}
	if limit > MaxLogLimit {
		limit = MaxLogLimit
	}
	filters.Limit = limit

	if len(m.mockLogs) > 0 {
		filtered := filterAndSanitizeLogs(m.mockLogs, filters.Level, filters.Search, limit)
		return RuntimeLogsResponse{
			Available: true,
			Source:    "mock",
			Entries:   filtered,
			HasMore:   len(m.mockLogs) > len(filtered),
		}
	}

	activeSrc, available, reason := m.sourceMgr.GetActiveSource(ctx)
	if !available || activeSrc == nil {
		return RuntimeLogsResponse{
			Available: false,
			Reason:    reason,
			Source:    "",
			Entries:   []RuntimeLogEntry{},
			HasMore:   false,
		}
	}

	srcFilters := logsource.LogFilters{
		Level:  filters.Level,
		Search: filters.Search,
		Limit:  filters.Limit,
		Cursor: filters.Cursor,
	}

	result, err := activeSrc.Read(ctx, srcFilters)
	if err != nil {
		return RuntimeLogsResponse{
			Available: false,
			Reason:    "read_error",
			Source:    activeSrc.Name(),
			Entries:   []RuntimeLogEntry{},
			HasMore:   false,
		}
	}

	entries := make([]RuntimeLogEntry, len(result.Entries))
	for i, e := range result.Entries {
		entries[i] = RuntimeLogEntry{
			Timestamp: e.Timestamp,
			Level:     e.Level,
			Logger:    e.Logger,
			Message:   e.Message,
			Fields:    e.Fields,
		}
	}

	return RuntimeLogsResponse{
		Available:  true,
		Source:     activeSrc.Name(),
		Entries:    entries,
		HasMore:    result.HasMore,
		NextCursor: result.NextCursor,
	}
}

func filterAndSanitizeLogs(entries []RuntimeLogEntry, level string, search string, limit int) []RuntimeLogEntry {
	level = strings.ToUpper(strings.TrimSpace(level))
	search = strings.ToLower(strings.TrimSpace(search))

	result := make([]RuntimeLogEntry, 0, len(entries))

	for _, entry := range entries {
		if level != "" && level != "ALL" {
			if entry.Level != level {
				continue
			}
		}

		if search != "" {
			matches := strings.Contains(strings.ToLower(entry.Message), search) ||
				strings.Contains(strings.ToLower(entry.Logger), search)

			if !matches && entry.Fields != nil {
				for k, v := range entry.Fields {
					vStr := strings.ToLower(valToString(v))
					if strings.Contains(strings.ToLower(k), search) || strings.Contains(vStr, search) {
						matches = true
						break
					}
				}
			}

			if !matches {
				continue
			}
		}

		sanitizedFields := sanitizeFields(entry.Fields)

		sanitizedEntry := RuntimeLogEntry{
			Timestamp: entry.Timestamp,
			Level:     entry.Level,
			Logger:    entry.Logger,
			Message:   entry.Message,
			Fields:    sanitizedFields,
		}

		result = append(result, sanitizedEntry)

		if len(result) >= limit {
			break
		}
	}

	return result
}

func sanitizeFields(fields map[string]any) map[string]any {
	if fields == nil {
		return nil
	}
	sanitized := make(map[string]any, len(fields))
	for k, v := range fields {
		kLower := strings.ToLower(k)
		if strings.Contains(kLower, "authorization") || strings.Contains(kLower, "cookie") ||
			strings.Contains(kLower, "proxy-authorization") || strings.Contains(kLower, "token") ||
			strings.Contains(kLower, "secret") || strings.Contains(kLower, "password") ||
			strings.Contains(kLower, "api-key") {
			sanitized[k] = "[REDACTED]"
		} else {
			sanitized[k] = v
		}
	}
	return sanitized
}

func valToString(v interface{}) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	b, err := json.Marshal(v)
	if err != nil {
		return ""
	}
	return string(b)
}
