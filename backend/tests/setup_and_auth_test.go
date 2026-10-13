package tests

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"caddy-manager/internal/api/routes"
	"caddy-manager/internal/audit"
	"caddy-manager/internal/auth"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/configstore"
	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
	"caddy-manager/internal/runtime"
	"caddy-manager/internal/sites"

	"github.com/gofiber/fiber/v2"
)

func setupTestApp(t *testing.T) (*fiber.App, *database.DB, *auth.Service, *caddy.MockClient) {
	t.Helper()
	db, err := database.InitDB(":memory:")
	if err != nil {
		t.Fatalf("failed to init in-memory database: %v", err)
	}

	authService := auth.NewService(db)
	mockCaddy := caddy.NewMockClient(true, "v2.8.4")
	auditLogger := audit.NewLogger(db)
	sitesService := sites.NewService(mockCaddy, auditLogger)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: true,
	})

	mockVersion := &mockVersionProvider{version: "v2.10.2"}
	configService := configstore.NewService(mockCaddy, mockVersion, configstore.NewSnapshotStore(db, 50), auditLogger)
	runtimeService := runtime.NewService(mockCaddy, mockVersion, sitesService, configService)
	routes.SetupRoutes(app, db, authService, mockCaddy, auditLogger, mockVersion, sitesService, configService, runtimeService)

	return app, db, authService, mockCaddy
}

func TestFirstRunSetup_SuccessAndDuplicatePrevention(t *testing.T) {
	app, db, _, _ := setupTestApp(t)
	defer db.Close()

	// 1. First-run setup
	setupPayload := models.SetupRequest{
		Username:        "admin",
		Password:        "SecurePassword123!",
		ConfirmPassword: "SecurePassword123!",
	}
	body, _ := json.Marshal(setupPayload)

	req := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("setup request failed: %v", err)
	}

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("expected status 201 Created, got %d", resp.StatusCode)
	}

	var apiResp models.APIResponse
	_ = json.NewDecoder(resp.Body).Decode(&apiResp)
	if apiResp.Data == nil {
		t.Fatalf("expected user data in setup response")
	}

	// 2. Duplicate setup prevention
	req2 := httptest.NewRequest(http.MethodPost, "/api/auth/setup", bytes.NewReader(body))
	req2.Header.Set("Content-Type", "application/json")

	resp2, err := app.Test(req2)
	if err != nil {
		t.Fatalf("duplicate setup request failed: %v", err)
	}

	if resp2.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected status 400 Bad Request for duplicate setup, got %d", resp2.StatusCode)
	}
}

func TestAuth_LoginSuccessAndInvalidCredentials(t *testing.T) {
	app, db, authService, _ := setupTestApp(t)
	defer db.Close()

	// Create admin user
	_, err := authService.CreateInitialAdmin("sysadmin", "MyPassword123!")
	if err != nil {
		t.Fatalf("failed creating test admin: %v", err)
	}

	// 1. Login with invalid password
	invalidPayload := models.LoginRequest{Username: "sysadmin", Password: "WrongPassword"}
	invBody, _ := json.Marshal(invalidPayload)
	invReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(invBody))
	invReq.Header.Set("Content-Type", "application/json")

	invResp, _ := app.Test(invReq)
	if invResp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401 Unauthorized for bad password, got %d", invResp.StatusCode)
	}

	// 2. Login with valid password
	validPayload := models.LoginRequest{Username: "sysadmin", Password: "MyPassword123!"}
	valBody, _ := json.Marshal(validPayload)
	valReq := httptest.NewRequest(http.MethodPost, "/api/auth/login", bytes.NewReader(valBody))
	valReq.Header.Set("Content-Type", "application/json")

	valResp, _ := app.Test(valReq)
	if valResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for valid login, got %d", valResp.StatusCode)
	}

	// Check HTTP-only cookie set
	cookies := valResp.Cookies()
	var sessionCookie *http.Cookie
	for _, c := range cookies {
		if c.Name == "caddio_session" {
			sessionCookie = c
			break
		}
	}

	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatalf("expected caddio_session cookie to be set")
	}

	if !sessionCookie.HttpOnly {
		t.Errorf("expected session cookie to be HttpOnly")
	}

	// 3. Test /api/auth/me with session cookie
	meReq := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReq.AddCookie(sessionCookie)
	meResp, _ := app.Test(meReq)
	if meResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for /api/auth/me with cookie, got %d", meResp.StatusCode)
	}

	// 4. Test Logout
	logoutReq := httptest.NewRequest(http.MethodPost, "/api/auth/logout", nil)
	logoutReq.AddCookie(sessionCookie)
	logoutResp, _ := app.Test(logoutReq)
	if logoutResp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 OK for logout, got %d", logoutResp.StatusCode)
	}

	// 5. Verify /api/auth/me rejects deleted session
	meReqAfter := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	meReqAfter.AddCookie(sessionCookie)
	meRespAfter, _ := app.Test(meReqAfter)

	var meData models.APIResponse
	_ = json.NewDecoder(meRespAfter.Body).Decode(&meData)
	dataMap, _ := meData.Data.(map[string]interface{})
	if dataMap["authenticated"] == true {
		t.Errorf("expected authenticated=false after logout")
	}
}

func TestHealthAndReadinessEndpoints(t *testing.T) {
	app, db, _, mockCaddy := setupTestApp(t)
	defer db.Close()

	// 1. Test /health
	reqHealth := httptest.NewRequest(http.MethodGet, "/health", nil)
	respHealth, _ := app.Test(reqHealth)
	if respHealth.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /health, got %d", respHealth.StatusCode)
	}

	// 2. Test /ready when Caddy reachable
	reqReady := httptest.NewRequest(http.MethodGet, "/ready", nil)
	respReady, _ := app.Test(reqReady)
	if respReady.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /ready, got %d", respReady.StatusCode)
	}

	// 3. Test /ready when Caddy unreachable (app process still ready)
	mockCaddy.IsReachable = false
	respReady2, _ := app.Test(reqReady)
	if respReady2.StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for /ready even when caddy unreachable, got %d", respReady2.StatusCode)
	}
}
