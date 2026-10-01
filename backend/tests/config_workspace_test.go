package tests

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"caddy-manager/internal/audit"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/configstore"
	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
	"caddy-manager/internal/sites"
)

func setupTestDB(t *testing.T) (*database.DB, func()) {
	t.Helper()
	tmpDir, err := os.MkdirTemp("", "caddio-config-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, "test.db")
	db, err := database.InitDB(dbPath)
	if err != nil {
		os.RemoveAll(tmpDir)
		t.Fatalf("failed to init db: %v", err)
	}

	cleanup := func() {
		db.Close()
		os.RemoveAll(tmpDir)
	}

	return db, cleanup
}

func TestNormalizeJSON(t *testing.T) {
	input := `{"apps":{"http":{"servers":{"srv0":{"listen":[":443"]}}}}}`
	norm, hash, err := configstore.NormalizeJSON(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if norm == "" || hash == "" {
		t.Fatalf("expected non-empty normalized string and hash")
	}

	// Test invalid JSON
	_, _, err = configstore.NormalizeJSON("{invalid json")
	if err == nil || !errors.Is(err, configstore.ErrInvalidJSON) {
		t.Fatalf("expected ErrInvalidJSON, got %v", err)
	}
}

func TestGenerateDiff(t *testing.T) {
	oldJSON := "{\n  \"apps\": {}\n}"
	newJSON := "{\n  \"apps\": {\n    \"http\": {}\n  }\n}"

	diff := configstore.GenerateDiff(oldJSON, newJSON)
	if diff.Stats.Additions == 0 {
		t.Errorf("expected additions in diff stats, got 0")
	}
	if diff.UnifiedDiff == "" {
		t.Errorf("expected non-empty unified diff string")
	}
}

func TestConfigService_ApplyAndHistory(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	mockCaddy.ConfigData = []byte(`{"apps":{}}`)
	auditLog := audit.NewLogger(db)
	snapStore := configstore.NewSnapshotStore(db, 50)
	svc := configstore.NewService(mockCaddy, nil, snapStore, auditLog)

	ctx := context.Background()

	// 1. Get initial config
	cfg, err := svc.GetConfig(ctx)
	if err != nil {
		t.Fatalf("failed to get config: %v", err)
	}
	initialHash := cfg.Hash

	// 2. Apply updated config
	proposedJSON := `{"apps":{"http":{"servers":{"srv0":{"listen":[":80"]}}}}}`
	req := models.ApplyConfigRequest{
		Config:   proposedJSON,
		BaseHash: initialHash,
	}

	res, err := svc.ApplyConfig(ctx, 1, "admin", req)
	if err != nil {
		t.Fatalf("failed to apply config: %v", err)
	}

	if res.Hash == initialHash {
		t.Errorf("expected hash to change after apply")
	}

	// 3. Verify history item recorded
	history, err := svc.GetHistory(ctx)
	if err != nil {
		t.Fatalf("failed to fetch history: %v", err)
	}
	if len(history) != 1 {
		t.Fatalf("expected 1 history item, got %d", len(history))
	}
	if history[0].Source != "manual_apply" {
		t.Errorf("expected source manual_apply, got %s", history[0].Source)
	}
}

func TestConfigService_OptimisticConcurrencyConflict(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	mockCaddy.ConfigData = []byte(`{"apps":{}}`)
	auditLog := audit.NewLogger(db)
	snapStore := configstore.NewSnapshotStore(db, 50)
	svc := configstore.NewService(mockCaddy, nil, snapStore, auditLog)

	ctx := context.Background()

	// Attempt apply with outdated base_hash
	wrongBaseHash := "11223344556677889900aabbccddeeff11223344556677889900aabbccddeeff"
	req := models.ApplyConfigRequest{
		Config:   `{"apps":{"http":{}}}`,
		BaseHash: wrongBaseHash,
	}

	_, err := svc.ApplyConfig(ctx, 1, "admin", req)
	if err == nil || !errors.Is(err, configstore.ErrConfigConflict) {
		t.Fatalf("expected ErrConfigConflict, got %v", err)
	}
}

func TestApplyFailureHandling(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	mockCaddy.ConfigData = []byte(`{"apps":{}}`)
	mockCaddy.ApplyError = errors.New("caddy connection refused")

	auditLog := audit.NewLogger(db)
	snapStore := configstore.NewSnapshotStore(db, 50)
	svc := configstore.NewService(mockCaddy, nil, snapStore, auditLog)

	ctx := context.Background()
	initialConfig, _ := svc.GetConfig(ctx)

	req := models.ApplyConfigRequest{
		Config:   `{"apps":{"http":{}}}`,
		BaseHash: initialConfig.Hash,
	}

	_, err := svc.ApplyConfig(ctx, 1, "admin", req)
	if err == nil || !errors.Is(err, caddy.ErrCaddyApplyFailed) {
		t.Fatalf("expected ErrCaddyApplyFailed wrapper error, got %v", err)
	}
}

func TestPostApplyVerificationFailure(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	// Create custom client mock where ApplyConfig accepts payload, but GetConfig returns unchanged payload
	mockCaddy := &StubVerificationFailClient{
		currentConfig: []byte(`{"apps":{}}`),
	}

	auditLog := audit.NewLogger(db)
	snapStore := configstore.NewSnapshotStore(db, 50)
	svc := configstore.NewService(mockCaddy, nil, snapStore, auditLog)

	ctx := context.Background()
	initialConfig, _ := svc.GetConfig(ctx)

	req := models.ApplyConfigRequest{
		Config:   `{"apps":{"http":{"servers":{}}}}`,
		BaseHash: initialConfig.Hash,
	}

	_, err := svc.ApplyConfig(ctx, 1, "admin", req)
	if err == nil || !errors.Is(err, configstore.ErrApplyVerificationFailed) {
		t.Fatalf("expected ErrApplyVerificationFailed, got %v", err)
	}
}

type StubVerificationFailClient struct {
	currentConfig []byte
}

func (s *StubVerificationFailClient) Health(ctx context.Context) error                          { return nil }
func (s *StubVerificationFailClient) GetVersion(ctx context.Context) (string, error)             { return "v2.10.2", nil }
func (s *StubVerificationFailClient) GetConfig(ctx context.Context) ([]byte, error)             { return s.currentConfig, nil }
func (s *StubVerificationFailClient) ApplyConfig(ctx context.Context, configBytes []byte) error { return nil }
func (s *StubVerificationFailClient) GetSites(ctx context.Context) (models.SiteDiscoveryResult, error) {
	return models.SiteDiscoveryResult{}, nil
}
func (s *StubVerificationFailClient) ApplySite(ctx context.Context, site models.Site) error {
	return nil
}
func (s *StubVerificationFailClient) DeleteSite(ctx context.Context, siteID string) error {
	return nil
}

func TestConfigService_Rollback(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	mockCaddy.ConfigData = []byte(`{"apps":{"http":{"servers":{"srv0":{"listen":[":80"]}}}}}`)
	auditLog := audit.NewLogger(db)
	snapStore := configstore.NewSnapshotStore(db, 50)
	svc := configstore.NewService(mockCaddy, nil, snapStore, auditLog)

	ctx := context.Background()

	// Save a historical snapshot to rollback to
	oldConfig := `{"apps":{"http":{"servers":{"srv0":{"listen":[":443"]}}}}}`
	oldNorm, oldHash, _ := configstore.NormalizeJSON(oldConfig)
	snapID := "snap_test_1"
	_ = snapStore.CreateSnapshot(models.ConfigSnapshot{
		ID:         snapID,
		Config:     oldNorm,
		ConfigHash: oldHash,
		CreatedBy:  "admin",
		Source:     "manual_apply",
	})

	activeCfg, _ := svc.GetConfig(ctx)

	// Perform Rollback
	rollbackRes, err := svc.RollbackConfig(ctx, 1, "admin", snapID, models.RollbackRequest{
		BaseHash: activeCfg.Hash,
	})
	if err != nil {
		t.Fatalf("rollback failed: %v", err)
	}

	if rollbackRes.Hash != oldHash {
		t.Errorf("expected active hash after rollback %s, got %s", oldHash, rollbackRes.Hash)
	}
}

func TestPhase2SitePreservationAfterRawApply(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	caddyJSON := `{
		"apps": {
			"http": {
				"servers": {
					"srv0": {
						"listen": [":443"],
						"routes": [
							{
								"@caddio": { "site_id": "site_test_1", "managed": true, "updated_at": "2026-10-01T00:00:00Z" },
								"match": [{ "host": ["example.com"] }],
								"handle": [{ "handler": "subroute", "routes": [{ "handle": [{ "handler": "reverse_proxy", "upstreams": [{ "dial": "127.0.0.1:8080" }] }] }] }],
								"terminal": true
							}
						]
					}
				}
			}
		}
	}`

	mockCaddy := caddy.NewMockClient(true, "v2.10.2")
	mockCaddy.ConfigData = []byte(caddyJSON)
	auditLog := audit.NewLogger(db)
	snapStore := configstore.NewSnapshotStore(db, 50)
	svc := configstore.NewService(mockCaddy, nil, snapStore, auditLog)
	sitesSvc := sites.NewService(mockCaddy, auditLog)

	ctx := context.Background()

	// Discover sites before workspace edit
	sitesBefore, err := sitesSvc.GetSites(ctx)
	if err != nil {
		t.Fatalf("failed to get sites before apply: %v", err)
	}
	if len(sitesBefore.Sites) != 1 || sitesBefore.Sites[0].Domain != "example.com" {
		t.Fatalf("expected 1 managed site example.com, got %v", sitesBefore.Sites)
	}

	// Apply raw config edit via workspace service
	activeCfg, _ := svc.GetConfig(ctx)
	newCaddyJSON := `{
		"apps": {
			"http": {
				"servers": {
					"srv0": {
						"listen": [":443"],
						"routes": [
							{
								"@caddio": { "site_id": "site_test_1", "managed": true, "updated_at": "2026-10-01T00:00:00Z" },
								"match": [{ "host": ["example.com"] }],
								"handle": [{ "handler": "subroute", "routes": [{ "handle": [{ "handler": "reverse_proxy", "upstreams": [{ "dial": "127.0.0.1:9090" }] }] }] }],
								"terminal": true
							}
						]
					}
				}
			}
		}
	}`

	_, err = svc.ApplyConfig(ctx, 1, "admin", models.ApplyConfigRequest{
		Config:   newCaddyJSON,
		BaseHash: activeCfg.Hash,
	})
	if err != nil {
		t.Fatalf("failed to apply config: %v", err)
	}

	// Discover sites after workspace edit to verify Phase 2 compatibility remains intact
	sitesAfter, err := sitesSvc.GetSites(ctx)
	if err != nil {
		t.Fatalf("failed to get sites after apply: %v", err)
	}
	if len(sitesAfter.Sites) != 1 || sitesAfter.Sites[0].Upstreams[0].Address != "127.0.0.1:9090" {
		t.Fatalf("expected managed site to reflect upstream 127.0.0.1:9090, got %v", sitesAfter.Sites)
	}
}
