package tests

import (
	"bytes"
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
	"caddy-manager/internal/models"
	"caddy-manager/internal/runtime"
	"caddy-manager/internal/sites"

	"github.com/gofiber/fiber/v2"
)

type CapturingMockCaddy struct {
	ConfigData   []byte
	AppliedBytes []byte
}

func (m *CapturingMockCaddy) GetVersion(ctx context.Context) (string, error) {
	return "v2.8.4", nil
}

func (m *CapturingMockCaddy) Health(ctx context.Context) error {
	return nil
}

func (m *CapturingMockCaddy) GetConfig(ctx context.Context) ([]byte, error) {
	return m.ConfigData, nil
}

func (m *CapturingMockCaddy) ApplyConfig(ctx context.Context, configBytes []byte) error {
	m.ConfigData = configBytes
	m.AppliedBytes = configBytes
	return nil
}

func (m *CapturingMockCaddy) GetSites(ctx context.Context) (models.SiteDiscoveryResult, error) {
	return caddy.ParseSitesFromConfig(m.ConfigData)
}

func (m *CapturingMockCaddy) ApplySite(ctx context.Context, site models.Site) error {
	client := caddy.NewClient("http://127.0.0.1:2019")
	_ = client
	// Use caddy site adapter logic directly on config bytes
	adapterClient := &mockAdapterClient{ConfigData: m.ConfigData}
	err := adapterClient.ApplySite(ctx, site)
	if err == nil {
		m.ConfigData = adapterClient.AppliedBytes
		m.AppliedBytes = adapterClient.AppliedBytes
	}
	return err
}

func (m *CapturingMockCaddy) DeleteSite(ctx context.Context, siteID string) error {
	adapterClient := &mockAdapterClient{ConfigData: m.ConfigData}
	err := adapterClient.DeleteSite(ctx, siteID)
	if err == nil {
		m.ConfigData = adapterClient.AppliedBytes
		m.AppliedBytes = adapterClient.AppliedBytes
	}
	return err
}

type mockAdapterClient struct {
	ConfigData   []byte
	AppliedBytes []byte
}

func (m *mockAdapterClient) GetConfig(ctx context.Context) ([]byte, error) {
	return m.ConfigData, nil
}

func (m *mockAdapterClient) LoadConfig(ctx context.Context, configBytes []byte) error {
	m.AppliedBytes = configBytes
	return nil
}

func (m *mockAdapterClient) ApplySite(ctx context.Context, site models.Site) error {
	c := &caddy.Client{}
	_ = c
	return caddy.ApplySiteOnMock(m, ctx, site)
}

func (m *mockAdapterClient) DeleteSite(ctx context.Context, siteID string) error {
	return caddy.DeleteSiteOnMock(m, ctx, siteID)
}

func TestUnrelatedCaddyConfigIntact(t *testing.T) {
	// Complex initial Caddy config with file_server, custom API routes, TLS options, and extra server blocks
	initialConfig := []byte(`{
		"apps": {
			"http": {
				"servers": {
					"custom_srv": {
						"listen": [":8080"],
						"routes": [
							{
								"@id": "ext_static_file_server",
								"match": [{"host": ["static.example.com"]}],
								"handle": [{"handler": "file_server", "root": "/var/www/static"}]
							},
							{
								"@id": "caddio_site_app_example_com_12345678",
								"match": [{"host": ["app.example.com"]}],
								"handle": [{"handler": "reverse_proxy", "upstreams": [{"dial": "127.0.0.1:3000"}]}],
								"comment": "managed_by: caddio"
							}
						]
					}
				}
			},
			"tls": {
				"automation": {
					"policies": [{"subjects": ["example.com"]}]
				}
			}
		}
	}`)

	mockCaddy := &CapturingMockCaddy{ConfigData: initialConfig}
	db, _ := database.InitDB(":memory:")
	defer db.Close()
	auditLogger := audit.NewLogger(db)
	service := sites.NewService(mockCaddy, auditLogger)

	// 1. Add a new site (api.example.com)
	newSiteReq := models.SaveSiteRequest{
		Name:      "API Service",
		Domain:    "api.example.com",
		Upstreams: []models.Upstream{{Address: "127.0.0.1:8080"}},
		Enabled:   true,
	}

	createdSite, err := service.SaveAndApplySite(context.Background(), 1, newSiteReq)
	if err != nil {
		t.Fatalf("failed to create new site: %v", err)
	}

	// Verify resulting config maintains custom_srv, tls config, static file server, and old site intact
	var resMap map[string]interface{}
	if err := json.Unmarshal(mockCaddy.ConfigData, &resMap); err != nil {
		t.Fatalf("invalid json result: %v", err)
	}

	apps := resMap["apps"].(map[string]interface{})
	if _, hasTLS := apps["tls"]; !hasTLS {
		t.Errorf("CRITICAL: tls app configuration was lost during site apply!")
	}

	servers := apps["http"].(map[string]interface{})["servers"].(map[string]interface{})
	customSrv := servers["custom_srv"].(map[string]interface{})
	routesList := customSrv["routes"].([]interface{})

	if len(routesList) != 3 {
		t.Fatalf("expected 3 routes in custom_srv, got %d", len(routesList))
	}

	// Verify static file server route is 100% intact
	route0 := routesList[0].(map[string]interface{})
	if route0["@id"] != "ext_static_file_server" {
		t.Errorf("expected route 0 to remain ext_static_file_server, got %v", route0["@id"])
	}

	// 2. Delete the created site
	if err := service.DeleteSite(context.Background(), 1, createdSite.ID); err != nil {
		t.Fatalf("failed to delete site: %v", err)
	}

	// Verify static file server route and original caddio site are STILL intact
	var afterDelete map[string]interface{}
	_ = json.Unmarshal(mockCaddy.ConfigData, &afterDelete)
	routesAfter := afterDelete["apps"].(map[string]interface{})["http"].(map[string]interface{})["servers"].(map[string]interface{})["custom_srv"].(map[string]interface{})["routes"].([]interface{})

	if len(routesAfter) != 2 {
		t.Fatalf("expected 2 routes remaining after delete, got %d", len(routesAfter))
	}
	r0 := routesAfter[0].(map[string]interface{})
	if r0["@id"] != "ext_static_file_server" {
		t.Errorf("static file server route altered after delete: %v", r0["@id"])
	}
}

func TestCaddioOwnershipScoping(t *testing.T) {
	initialConfig := []byte(`{
		"apps": {
			"http": {
				"servers": {
					"srv0": {
						"listen": [":443"],
						"routes": [
							{
								"@id": "unmanaged_external_route",
								"match": [{"host": ["external.com"]}],
								"handle": [{"handler": "reverse_proxy", "upstreams": [{"dial": "10.0.0.1:80"}]}]
							}
						]
					}
				}
			}
		}
	}`)

	mockCaddy := &CapturingMockCaddy{ConfigData: initialConfig}
	db, _ := database.InitDB(":memory:")
	defer db.Close()
	auditLogger := audit.NewLogger(db)
	service := sites.NewService(mockCaddy, auditLogger)

	// 1. Attempt to update unmanaged route -> MUST fail with ErrUnmanagedRoute
	updateReq := models.SaveSiteRequest{
		ID:        "unmanaged_external_route",
		Name:      "Hacked External Route",
		Domain:    "external.com",
		Upstreams: []models.Upstream{{Address: "1.1.1.1:80"}},
	}
	_, errUpdate := service.SaveAndApplySite(context.Background(), 1, updateReq)
	if !errors.Is(errUpdate, caddy.ErrUnmanagedRoute) {
		t.Errorf("expected ErrUnmanagedRoute on updating external route, got %v", errUpdate)
	}

	// 2. Attempt to delete unmanaged route -> MUST fail with ErrUnmanagedRoute
	errDelete := service.DeleteSite(context.Background(), 1, "unmanaged_external_route")
	if !errors.Is(errDelete, caddy.ErrUnmanagedRoute) {
		t.Errorf("expected ErrUnmanagedRoute on deleting external route, got %v", errDelete)
	}
}

func TestDynamicServerBlockHandling(t *testing.T) {
	// Server block is named "my_custom_proxy" (NOT srv0)
	initialConfig := []byte(`{
		"apps": {
			"http": {
				"servers": {
					"my_custom_proxy": {
						"listen": [":80", ":443"],
						"routes": []
					}
				}
			}
		}
	}`)

	mockCaddy := &CapturingMockCaddy{ConfigData: initialConfig}
	db, _ := database.InitDB(":memory:")
	defer db.Close()
	auditLogger := audit.NewLogger(db)
	service := sites.NewService(mockCaddy, auditLogger)

	// Create site
	req := models.SaveSiteRequest{
		Name:      "Dynamic Server Test",
		Domain:    "dynamic.internal",
		Upstreams: []models.Upstream{{Address: "127.0.0.1:9000"}},
		Enabled:   true,
	}

	_, err := service.SaveAndApplySite(context.Background(), 1, req)
	if err != nil {
		t.Fatalf("failed to apply site to dynamic server block: %v", err)
	}

	var resMap map[string]interface{}
	_ = json.Unmarshal(mockCaddy.ConfigData, &resMap)
	servers := resMap["apps"].(map[string]interface{})["http"].(map[string]interface{})["servers"].(map[string]interface{})

	if _, hasMyCustomProxy := servers["my_custom_proxy"]; !hasMyCustomProxy {
		t.Fatalf("my_custom_proxy server block was destroyed or renamed!")
	}

	routesList := servers["my_custom_proxy"].(map[string]interface{})["routes"].([]interface{})
	if len(routesList) != 1 {
		t.Fatalf("expected 1 route in my_custom_proxy, got %d", len(routesList))
	}
}

func TestSitesAPI_HTTPHandlers(t *testing.T) {
	db, _ := database.InitDB(":memory:")
	defer db.Close()
	authService := auth.NewService(db)
	auditLogger := audit.NewLogger(db)
	mockCaddy := &CapturingMockCaddy{
		ConfigData: []byte(`{"apps":{"http":{"servers":{"srv0":{"listen":[":443"],"routes":[]}}}}}`),
	}
	sitesService := sites.NewService(mockCaddy, auditLogger)
	mockVersion := &mockVersionProvider{version: "v2.10.2"}
	configService := configstore.NewService(mockCaddy, mockVersion, configstore.NewSnapshotStore(db, 50), auditLogger)
	runtimeService := runtime.NewService(mockCaddy, mockVersion, sitesService, configService)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	routes.SetupRoutes(app, db, authService, mockCaddy, auditLogger, mockVersion, sitesService, configService, runtimeService)

	// Authenticate admin user
	adminUser, _ := authService.CreateInitialAdmin("admin", "Password123!")
	session, _ := authService.CreateSession(adminUser.ID)
	sessionCookie := &http.Cookie{Name: "caddio_session", Value: session.ID}

	// 1. GET /api/sites -> empty list initially
	getReq := httptest.NewRequest(http.MethodGet, "/api/sites", nil)
	getReq.AddCookie(sessionCookie)
	getResp, _ := app.Test(getReq)
	if getResp.StatusCode != http.StatusOK {
		t.Fatalf("GET /api/sites returned status %d", getResp.StatusCode)
	}

	// 2. POST /api/sites/validate -> dry-run validation
	valPayload := models.SaveSiteRequest{
		Domain:    "valid.com",
		Upstreams: []models.Upstream{{Address: "127.0.0.1:8080"}},
	}
	valBody, _ := json.Marshal(valPayload)
	valReq := httptest.NewRequest(http.MethodPost, "/api/sites/validate", bytes.NewReader(valBody))
	valReq.Header.Set("Content-Type", "application/json")
	valReq.AddCookie(sessionCookie)
	valResp, _ := app.Test(valReq)

	if valResp.StatusCode != http.StatusOK {
		t.Fatalf("POST /api/sites/validate returned status %d", valResp.StatusCode)
	}
	var valAPIResp models.APIResponse
	_ = json.NewDecoder(valResp.Body).Decode(&valAPIResp)

	// 3. POST /api/sites -> create site
	createPayload := models.SaveSiteRequest{
		Name:      "Valid Site",
		Domain:    "valid.com",
		Upstreams: []models.Upstream{{Address: "127.0.0.1:8080"}},
		Enabled:   true,
	}
	createBody, _ := json.Marshal(createPayload)
	createReq := httptest.NewRequest(http.MethodPost, "/api/sites", bytes.NewReader(createBody))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(sessionCookie)
	createResp, _ := app.Test(createReq)

	if createResp.StatusCode != http.StatusCreated {
		t.Fatalf("POST /api/sites returned status %d", createResp.StatusCode)
	}

	// 4. GET /api/dashboard -> verify site_summary metrics present
	dashReq := httptest.NewRequest(http.MethodGet, "/api/dashboard", nil)
	dashReq.AddCookie(sessionCookie)
	dashResp, _ := app.Test(dashReq)

	var dashAPIResp models.APIResponse
	_ = json.NewDecoder(dashResp.Body).Decode(&dashAPIResp)
	dashData, ok := dashAPIResp.Data.(map[string]interface{})
	if !ok {
		t.Fatalf("invalid dashboard data response")
	}

	siteSummary, ok := dashData["site_summary"].(map[string]interface{})
	if !ok {
		t.Fatalf("site_summary field missing from dashboard response")
	}

	if total, ok := siteSummary["total_sites"].(float64); !ok || total != 1 {
		t.Errorf("expected total_sites = 1, got %v", siteSummary["total_sites"])
	}
}

func TestSubrouteParsingUserConfig(t *testing.T) {
	userConfig := []byte(`{
		"apps": {
			"http": {
				"servers": {
					"srv0": {
						"listen": [":443"],
						"routes": [
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:9005"}]}]}]}],"match":[{"host":["gallery-example.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:9001"}]}]}]}],"match":[{"host":["api-url-short.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:9004"}]}]}]}],"match":[{"host":["youtube-dl.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:9003"}]}]}]}],"match":[{"host":["api-yt-dlp.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:9002"}]}]}]}],"match":[{"host":["portofolio.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:5000"}]}]}]}],"match":[{"host":["registry.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:51820"}]}]}]}],"match":[{"host":["wg-api.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:9000"}]}]}]}],"match":[{"host":["kort.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:3000"}]}]}]}],"match":[{"host":["dns.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:51821"}]}]}]}],"match":[{"host":["vpn.devteddy.my.id"]}],"terminal":true},
							{"handle":[{"handler":"subroute","routes":[{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:9999"}]}]}]}],"match":[{"host":["pm.devteddy.my.id"]}],"terminal":true}
						]
					},
					"srv1": {
						"listen": [":80"],
						"routes": [
							{"handle":[{"handler":"reverse_proxy","upstreams":[{"dial":"localhost:8080"}]}]}
						]
					}
				}
			}
		}
	}`)

	res, err := caddy.ParseSitesFromConfig(userConfig)
	if err != nil {
		t.Fatalf("ParseSitesFromConfig returned error: %v", err)
	}

	if len(res.Sites) != 11 {
		t.Fatalf("expected 11 discovered sites, got %d", len(res.Sites))
	}

	// Verify domain and upstream mapping for first site
	site0 := res.Sites[0]
	if site0.Domain != "gallery-example.devteddy.my.id" {
		t.Errorf("expected domain gallery-example.devteddy.my.id, got %s", site0.Domain)
	}
	if len(site0.Upstreams) != 1 || site0.Upstreams[0].Address != "localhost:9005" {
		t.Errorf("expected upstream localhost:9005, got %v", site0.Upstreams)
	}
}
