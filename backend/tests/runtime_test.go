package tests

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"caddy-manager/internal/api/routes"
	"caddy-manager/internal/audit"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/configstore"
	"caddy-manager/internal/runtime"
	"caddy-manager/internal/sites"

	"github.com/gofiber/fiber/v2"
)

func TestRuntimeStatusHealthy(t *testing.T) {
	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	mockCaddy.ConfigData = []byte(`{"apps":{"http":{"servers":{"srv0":{"listen":[":443"]}}},"logging":{}}}`)
	sitesSvc := sites.NewService(mockCaddy, nil)
	mockVer := &mockVersionProvider{version: "v2.10.2"}

	runtimeSvc := runtime.NewService(mockCaddy, mockVer, sitesSvc, nil)
	runtimeSvc.SetMockLogs([]runtime.RuntimeLogEntry{
		{Timestamp: time.Now(), Level: "INFO", Message: "caddy started"},
	})

	ctx := context.Background()
	status := runtimeSvc.GetStatus(ctx)

	if !status.Reachable {
		t.Errorf("expected reachable to be true")
	}
	if status.Status != "healthy" {
		t.Errorf("expected status healthy, got %s", status.Status)
	}
	if status.ActiveConfigHash == "" {
		t.Errorf("expected non-empty active config hash")
	}
	if status.ServerCount != 1 {
		t.Errorf("expected server count 1, got %d", status.ServerCount)
	}
}

func TestRuntimeStatusUnreachable(t *testing.T) {
	mockCaddy := caddy.NewMockClient(false, "")
	runtimeSvc := runtime.NewService(mockCaddy, nil, nil, nil)

	ctx := context.Background()
	status := runtimeSvc.GetStatus(ctx)

	if status.Reachable {
		t.Errorf("expected reachable to be false")
	}
	if status.Status != "unreachable" {
		t.Errorf("expected status unreachable, got %s", status.Status)
	}
	if status.LogsAvailable {
		t.Errorf("expected logs available to be false when caddy is unreachable")
	}
}

func TestRuntimeLogsFiltering(t *testing.T) {
	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	runtimeSvc := runtime.NewService(mockCaddy, nil, nil, nil)

	mockEntries := []runtime.RuntimeLogEntry{
		{Timestamp: time.Now(), Level: "INFO", Logger: "http.log.access", Message: "handled request GET /api"},
		{Timestamp: time.Now(), Level: "WARN", Logger: "http.handlers.proxy", Message: "upstream latency high"},
		{Timestamp: time.Now(), Level: "ERROR", Logger: "tls", Message: "certificate renewal failed"},
	}
	runtimeSvc.SetMockLogs(mockEntries)

	ctx := context.Background()

	// 1. Filter by Level ERROR
	errLogs := runtimeSvc.GetLogs(ctx, runtime.RuntimeLogFilters{Level: "ERROR", Limit: 10})
	if !errLogs.Available || len(errLogs.Entries) != 1 || errLogs.Entries[0].Level != "ERROR" {
		t.Errorf("expected 1 ERROR log entry, got %v", errLogs)
	}

	// 2. Filter by Search Query "proxy"
	searchLogs := runtimeSvc.GetLogs(ctx, runtime.RuntimeLogFilters{Search: "proxy", Limit: 10})
	if !searchLogs.Available || len(searchLogs.Entries) != 1 || searchLogs.Entries[0].Logger != "http.handlers.proxy" {
		t.Errorf("expected 1 log matching 'proxy', got %v", searchLogs)
	}
}

func TestRuntimeLogsLimit(t *testing.T) {
	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	runtimeSvc := runtime.NewService(mockCaddy, nil, nil, nil)

	entries := make([]runtime.RuntimeLogEntry, 600)
	for i := 0; i < 600; i++ {
		entries[i] = runtime.RuntimeLogEntry{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "log entry",
		}
	}
	runtimeSvc.SetMockLogs(entries)

	ctx := context.Background()

	// Request limit 1000 -> should cap at 500
	logsRes := runtimeSvc.GetLogs(ctx, runtime.RuntimeLogFilters{Limit: 1000})
	if len(logsRes.Entries) > 500 {
		t.Errorf("expected max limit 500, got %d", len(logsRes.Entries))
	}
}

func TestEnableLoggingCreatesSafeConfig(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	// Active config with unmanaged site & Caddio site
	mockCaddy.ConfigData = []byte(`{
		"apps": {
			"http": {
				"servers": {
					"srv0": {
						"listen": [":443"],
						"routes": [
							{
								"@caddio": {"site_id": "site_1", "managed": true},
								"match": [{"host": ["example.com"]}]
							},
							{
								"match": [{"host": ["unmanaged.com"]}]
							}
						]
					}
				}
			}
		}
	}`)

	authService := auth.NewService(db)
	adminUser, _ := authService.CreateInitialAdmin("admin", "Password123!")

	auditLogger := audit.NewLogger(db)
	mockVersion := &mockVersionProvider{version: "v2.10.2"}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	snapshotStore := configstore.NewSnapshotStore(db, 50)
	configService := configstore.NewService(mockCaddy, mockVersion, snapshotStore, auditLogger)
	runtimeService := runtime.NewService(mockCaddy, mockVersion, sitesService, configService)

	ctx := context.Background()
	activeConfig, err := configService.GetConfig(ctx)
	if err != nil {
		t.Fatalf("failed to fetch active config: %v", err)
	}

	// 1. Enable Logging via Phase 3 Safe Pipeline
	enableRes, err := runtimeService.EnableLogging(ctx, adminUser.ID, "admin", runtime.EnableLoggingRequest{
		BaseHash: activeConfig.Hash,
	})
	if err != nil {
		t.Fatalf("EnableLogging failed: %v", err)
	}
	if !enableRes.Changed {
		t.Errorf("expected Changed to be true")
	}

	// 2. Verify snapshot was created in database
	snapshots, err := snapshotStore.ListSnapshots(10)
	if err != nil || len(snapshots) == 0 {
		t.Errorf("expected snapshot to be created in DB, got snapshots %v", snapshots)
	}

	// 3. Verify audit log entry was created
	auditLogs, err := auditLogger.GetRecentLogs(10)
	if err != nil || len(auditLogs) == 0 {
		t.Errorf("expected audit log record for enable logging, got %v", auditLogs)
	}
	foundAudit := false
	for _, l := range auditLogs {
		if l.Action == "config.applied" {
			foundAudit = true
			break
		}
	}
	if !foundAudit {
		t.Errorf("expected config.applied audit record")
	}

	// 4. Verify root-level logging was added while HTTP apps remain 100% intact
	updatedConfig, _ := configService.GetConfig(ctx)
	var root map[string]interface{}
	_ = json.Unmarshal([]byte(updatedConfig.Config), &root)

	if _, hasLogging := root["logging"]; !hasLogging {
		t.Errorf("expected root logging object to be added")
	}
	apps, hasApps := root["apps"].(map[string]interface{})
	if !hasApps {
		t.Errorf("expected root apps object to exist")
	} else if _, hasHTTP := apps["http"]; !hasHTTP {
		t.Errorf("expected apps.http to remain intact")
	}
}

func TestEnableLoggingOptimisticConcurrencyConflict(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	mockCaddy.ConfigData = []byte(`{"apps":{"http":{}}}`)

	auditLogger := audit.NewLogger(db)
	mockVersion := &mockVersionProvider{version: "v2.10.2"}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	snapshotStore := configstore.NewSnapshotStore(db, 50)
	configService := configstore.NewService(mockCaddy, mockVersion, snapshotStore, auditLogger)
	runtimeService := runtime.NewService(mockCaddy, mockVersion, sitesService, configService)

	ctx := context.Background()

	// Pass stale base_hash
	_, err := runtimeService.EnableLogging(ctx, 1, "admin", runtime.EnableLoggingRequest{
		BaseHash: "stale_hash_12345",
	})
	if err == nil {
		t.Fatalf("expected optimistic concurrency error for stale hash")
	}
}

func TestRuntimeRequiresAuthentication(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	authService := auth.NewService(db)
	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	auditLogger := audit.NewLogger(db)
	mockVersion := &mockVersionProvider{version: "v2.10.2"}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	configService := configstore.NewService(mockCaddy, mockVersion, configstore.NewSnapshotStore(db, 50), auditLogger)
	runtimeService := runtime.NewService(mockCaddy, mockVersion, sitesService, configService)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	routes.SetupRoutes(app, db, authService, mockCaddy, auditLogger, mockVersion, sitesService, configService, runtimeService)

	// Unauthenticated request to /api/runtime/status
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/runtime/status", nil)
	respStatus, err := app.Test(reqStatus)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respStatus.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", respStatus.StatusCode)
	}

	// Unauthenticated request to /api/runtime/logs
	reqLogs := httptest.NewRequest(http.MethodGet, "/api/runtime/logs", nil)
	respLogs, err := app.Test(reqLogs)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if respLogs.StatusCode != http.StatusUnauthorized {
		t.Errorf("expected status 401 Unauthorized, got %d", respLogs.StatusCode)
	}
}

func TestRuntimeMonitoringNoAuditLogCreated(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	authService := auth.NewService(db)
	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	auditLogger := audit.NewLogger(db)
	mockVersion := &mockVersionProvider{version: "v2.10.2"}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	configService := configstore.NewService(mockCaddy, mockVersion, configstore.NewSnapshotStore(db, 50), auditLogger)
	runtimeService := runtime.NewService(mockCaddy, mockVersion, sitesService, configService)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	routes.SetupRoutes(app, db, authService, mockCaddy, auditLogger, mockVersion, sitesService, configService, runtimeService)

	adminUser, _ := authService.CreateInitialAdmin("admin", "Password123!")
	session, _ := authService.CreateSession(adminUser.ID)
	sessionCookie := &http.Cookie{Name: "caddio_session", Value: session.ID}

	// Make monitoring GET requests
	reqStatus := httptest.NewRequest(http.MethodGet, "/api/runtime/status", nil)
	reqStatus.AddCookie(sessionCookie)
	_, _ = app.Test(reqStatus)

	reqLogs := httptest.NewRequest(http.MethodGet, "/api/runtime/logs", nil)
	reqLogs.AddCookie(sessionCookie)
	_, _ = app.Test(reqLogs)

	logs, err := auditLogger.GetRecentLogs(50)
	if err != nil {
		t.Fatalf("failed to query audit logs: %v", err)
	}

	for _, l := range logs {
		if l.Action == "runtime.status" || l.Action == "runtime.logs" {
			t.Errorf("unexpected audit log created for read-only monitoring: %v", l)
		}
	}
}
