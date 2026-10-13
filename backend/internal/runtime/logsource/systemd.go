package logsource

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type SystemdLogSource struct {
	unitName string
}

func NewSystemdLogSource() *SystemdLogSource {
	unit := strings.TrimSpace(os.Getenv("CADDIO_CADDY_SYSTEMD_UNIT"))
	if unit == "" {
		unit = "caddy.service"
	}
	return &SystemdLogSource{
		unitName: unit,
	}
}

func (s *SystemdLogSource) Name() string {
	return "systemd"
}

func findJournalctlPath() string {
	if path, err := exec.LookPath("journalctl"); err == nil {
		return path
	}
	for _, p := range []string{"/usr/bin/journalctl", "/bin/journalctl"} {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

func (s *SystemdLogSource) getUnitArgs() []string {
	unitBase := strings.TrimSuffix(s.unitName, ".service")
	return []string{"-u", unitBase + ".service", "-u", unitBase}
}

func (s *SystemdLogSource) Available(ctx context.Context) (bool, string) {
	journalPath := findJournalctlPath()
	if journalPath == "" {
		return false, "journalctl_not_found"
	}

	checkCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	args := append(s.getUnitArgs(), "-n", "1", "--output=json", "--no-pager")
	cmd := exec.CommandContext(checkCtx, journalPath, args...)
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok && exitErr.ExitCode() == 1 {
			return false, "no_logs"
		}
		return false, "permission_denied"
	}

	if len(bytes.TrimSpace(output)) == 0 {
		return false, "no_logs"
	}

	return true, "ok"
}

func (s *SystemdLogSource) Read(ctx context.Context, filters LogFilters) (LogResult, error) {
	journalPath := findJournalctlPath()
	if journalPath == "" {
		return LogResult{Entries: []LogEntry{}}, nil
	}

	limit := filters.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 500 {
		limit = 500
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	args := append(s.getUnitArgs(), "-n", strconv.Itoa(limit*2), "--output=json", "--no-pager")
	cmd := exec.CommandContext(reqCtx, journalPath, args...)
	out, err := cmd.Output()
	if err != nil {
		return LogResult{Entries: []LogEntry{}}, nil
	}

	rawEntries := make([]LogEntry, 0)
	scanner := bufio.NewScanner(bytes.NewReader(out))

	for scanner.Scan() {
		line := scanner.Bytes()
		if len(line) == 0 {
			continue
		}

		var record map[string]interface{}
		if err := json.Unmarshal(line, &record); err != nil {
			continue
		}

		entry := s.parseJournalRecord(record)
		if !entry.Timestamp.IsZero() {
			rawEntries = append(rawEntries, entry)
		}
	}

	// Reverse to get newest entries first
	for i, j := 0, len(rawEntries)-1; i < j; i, j = i+1, j-1 {
		rawEntries[i], rawEntries[j] = rawEntries[j], rawEntries[i]
	}

	filtered := filterAndSanitize(rawEntries, filters.Level, filters.Search, limit)

	return LogResult{
		Entries:    filtered,
		HasMore:    len(rawEntries) > len(filtered),
		NextCursor: "",
	}, nil
}

func (s *SystemdLogSource) parseJournalRecord(record map[string]interface{}) LogEntry {
	entry := LogEntry{
		Timestamp: time.Now().UTC(),
		Level:     "INFO",
		Logger:    "caddy",
		Message:   "",
		Fields:    make(map[string]any),
	}

	if realtime, ok := record["__REALTIME_TIMESTAMP"]; ok {
		var microsec int64
		switch v := realtime.(type) {
		case string:
			microsec, _ = strconv.ParseInt(v, 10, 64)
		case float64:
			microsec = int64(v)
		}
		if microsec > 0 {
			entry.Timestamp = time.Unix(0, microsec*1000).UTC()
		}
	}

	if ident, ok := record["SYSLOG_IDENTIFIER"].(string); ok && ident != "" {
		entry.Logger = ident
	}

	if prioVal, ok := record["PRIORITY"]; ok {
		var prio int
		switch v := prioVal.(type) {
		case string:
			prio, _ = strconv.Atoi(v)
		case float64:
			prio = int(v)
		}
		switch {
		case prio >= 0 && prio <= 3:
			entry.Level = "ERROR"
		case prio == 4:
			entry.Level = "WARN"
		case prio == 5 || prio == 6:
			entry.Level = "INFO"
		case prio >= 7:
			entry.Level = "DEBUG"
		}
	}

	if msgRaw, ok := record["MESSAGE"].(string); ok {
		msgRaw = strings.TrimSpace(msgRaw)
		if strings.HasPrefix(msgRaw, "{") && strings.HasSuffix(msgRaw, "}") {
			var caddyLog map[string]interface{}
			if err := json.Unmarshal([]byte(msgRaw), &caddyLog); err == nil {
				if levelStr, ok := caddyLog["level"].(string); ok && levelStr != "" {
					entry.Level = strings.ToUpper(levelStr)
				}
				if loggerStr, ok := caddyLog["logger"].(string); ok && loggerStr != "" {
					entry.Logger = loggerStr
				}
				if msgStr, ok := caddyLog["msg"].(string); ok {
					entry.Message = msgStr
				}
				for k, v := range caddyLog {
					if k != "level" && k != "logger" && k != "msg" && k != "ts" {
						entry.Fields[k] = v
					}
				}
				return entry
			}
		}
		entry.Message = msgRaw
	}

	return entry
}

func filterAndSanitize(entries []LogEntry, level string, search string, limit int) []LogEntry {
	level = strings.ToUpper(strings.TrimSpace(level))
	search = strings.ToLower(strings.TrimSpace(search))

	result := make([]LogEntry, 0, len(entries))

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

		sanitizedEntry := LogEntry{
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
