package tests

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"caddy-manager/internal/api/routes"
	"caddy-manager/internal/audit"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/configstore"
	"caddy-manager/internal/database"
	"caddy-manager/internal/sites"

	"github.com/gofiber/fiber/v2"
)

type mockVersionProvider struct {
	version string
	err     error
}

func (m *mockVersionProvider) GetBinaryVersion(ctx context.Context) (string, error) {
	if m.err != nil {
		return "", m.err
	}
	return m.version, nil
}

func TestSystemVersionProvider_Unit(t *testing.T) {
	// 1. Success with whitespace & prefix normalization
	runnerSuccess := func(ctx context.Context, name string, arg ...string) ([]byte, error) {
		return []byte("  2.10.2 \n  build info  \n"), nil
	}
	providerSuccess := caddy.NewSystemVersionProviderWithRunner("/bin/true", runnerSuccess)
	ver, err := providerSuccess.GetBinaryVersion(context.Background())
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if ver != "v2.10.2" {
		t.Errorf("expected 'v2.10.2', got '%s'", ver)
	}

	// 2. Binary not found
	providerNotFound := caddy.NewSystemVersionProviderWithRunner("/nonexistent/path/caddy", runnerSuccess)
	_, errNotFound := providerNotFound.GetBinaryVersion(context.Background())
	if !errors.Is(errNotFound, caddy.ErrBinaryNotFound) {
		t.Errorf("expected ErrBinaryNotFound, got %v", errNotFound)
	}

	// 3. Command execution failure
	runnerFail := func(ctx context.Context, name string, arg ...string) ([]byte, error) {
		return nil, errors.New("exit status 1")
	}
	providerFail := caddy.NewSystemVersionProviderWithRunner("/bin/true", runnerFail)
	_, errFail := providerFail.GetBinaryVersion(context.Background())
	if !errors.Is(errFail, caddy.ErrExecutionFailed) {
		t.Errorf("expected ErrExecutionFailed, got %v", errFail)
	}
}

func TestSystemCaddyVersionEndpoint_Success200(t *testing.T) {
	db, _ := database.InitDB(":memory:")
	defer db.Close()

	authService := auth.NewService(db)
	mockCaddy := caddy.NewMockClient(true, "v2.8.4")
	auditLogger := audit.NewLogger(db)
	mockVersion := &mockVersionProvider{version: "v2.10.2"}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	configService := configstore.NewService(mockCaddy, mockVersion, configstore.NewSnapshotStore(db, 50), auditLogger)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	routes.SetupRoutes(app, db, authService, mockCaddy, auditLogger, mockVersion, sitesService, configService)

	req := httptest.NewRequest(http.MethodGet, "/api/system/caddy/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 OK, got %d", resp.StatusCode)
	}

	var res map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&res)
	if res["version"] != "v2.10.2" {
		t.Errorf("expected version 'v2.10.2', got '%s'", res["version"])
	}
}

func TestSystemCaddyVersionEndpoint_BinaryNotFound503(t *testing.T) {
	db, _ := database.InitDB(":memory:")
	defer db.Close()

	authService := auth.NewService(db)
	mockCaddy := caddy.NewMockClient(true, "v2.8.4")
	auditLogger := audit.NewLogger(db)
	mockVersion := &mockVersionProvider{err: caddy.ErrBinaryNotFound}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	configService := configstore.NewService(mockCaddy, mockVersion, configstore.NewSnapshotStore(db, 50), auditLogger)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	routes.SetupRoutes(app, db, authService, mockCaddy, auditLogger, mockVersion, sitesService, configService)

	req := httptest.NewRequest(http.MethodGet, "/api/system/caddy/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 Service Unavailable, got %d", resp.StatusCode)
	}

	var res map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&res)
	if res["error"] != "caddy binary not found" {
		t.Errorf("expected error 'caddy binary not found', got '%s'", res["error"])
	}
}

func TestSystemCaddyVersionEndpoint_CommandFailure503(t *testing.T) {
	db, _ := database.InitDB(":memory:")
	defer db.Close()

	authService := auth.NewService(db)
	mockCaddy := caddy.NewMockClient(true, "v2.8.4")
	auditLogger := audit.NewLogger(db)
	mockVersion := &mockVersionProvider{err: caddy.ErrExecutionFailed}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	configService := configstore.NewService(mockCaddy, mockVersion, configstore.NewSnapshotStore(db, 50), auditLogger)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	routes.SetupRoutes(app, db, authService, mockCaddy, auditLogger, mockVersion, sitesService, configService)

	req := httptest.NewRequest(http.MethodGet, "/api/system/caddy/version", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("API request failed: %v", err)
	}

	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503 Service Unavailable, got %d", resp.StatusCode)
	}

	var res map[string]string
	_ = json.NewDecoder(resp.Body).Decode(&res)
	if res["error"] != "failed to get Caddy version" {
		t.Errorf("expected error 'failed to get Caddy version', got '%s'", res["error"])
	}
}
