package configstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"caddy-manager/internal/audit"
	"caddy-manager/internal/caddy"
	"caddy-manager/internal/models"
)

var (
	ErrConfigConflict          = errors.New("Caddy configuration changed since this editor session was loaded")
	ErrConfigTooLarge          = errors.New("configuration payload exceeds maximum allowed size")
	ErrInvalidJSON             = errors.New("invalid JSON format")
	ErrInvalidCaddyStructure   = errors.New("invalid Caddy configuration structure (missing apps object)")
	ErrApplyVerificationFailed = errors.New("post-apply configuration verification failed")
)

const DefaultMaxConfigSize = 2 * 1024 * 1024 // 2MB

type Service struct {
	caddyClient     caddy.CaddyClient
	versionProvider caddy.SystemVersionProvider
	snapshotStore   *SnapshotStore
	auditLogger     *audit.Logger
	maxConfigSize   int
}

func NewService(caddyClient caddy.CaddyClient, versionProvider caddy.SystemVersionProvider, snapshotStore *SnapshotStore, auditLogger *audit.Logger) *Service {
	return &Service{
		caddyClient:     caddyClient,
		versionProvider: versionProvider,
		snapshotStore:   snapshotStore,
		auditLogger:     auditLogger,
		maxConfigSize:   DefaultMaxConfigSize,
	}
}

// NormalizeJSON formats JSON deterministically for stable hashing
func NormalizeJSON(rawJSON string) (string, string, error) {
	rawBytes := []byte(strings.TrimSpace(rawJSON))
	if len(rawBytes) == 0 {
		rawBytes = []byte("{}")
	}

	var root interface{}
	if err := json.Unmarshal(rawBytes, &root); err != nil {
		return "", "", fmt.Errorf("%w: %v", ErrInvalidJSON, err)
	}

	normalizedBytes, err := json.MarshalIndent(root, "", "  ")
	if err != nil {
		return "", "", fmt.Errorf("failed to format json: %w", err)
	}

	normalizedStr := string(normalizedBytes)
	hashBytes := sha256.Sum256(normalizedBytes)
	hashHex := hex.EncodeToString(hashBytes[:])

	return normalizedStr, hashHex, nil
}

// GetConfig fetches current active Caddy JSON, calculates SHA-256 hash, and retrieves host version
func (s *Service) GetConfig(ctx context.Context) (models.ConfigResponse, error) {
	rawBytes, err := s.caddyClient.GetConfig(ctx)
	if err != nil {
		return models.ConfigResponse{}, fmt.Errorf("failed to fetch active caddy config: %w", err)
	}

	normalizedJSON, hashHex, err := NormalizeJSON(string(rawBytes))
	if err != nil {
		normalizedJSON = string(rawBytes)
		hashBytes := sha256.Sum256(rawBytes)
		hashHex = hex.EncodeToString(hashBytes[:])
	}

	caddyVersion := ""
	if s.versionProvider != nil {
		if ver, err := s.versionProvider.GetBinaryVersion(ctx); err == nil {
			caddyVersion = ver
		}
	}

	return models.ConfigResponse{
		Config:      normalizedJSON,
		Version:     caddyVersion,
		Hash:        hashHex,
		RetrievedAt: time.Now().UTC(),
	}, nil
}

// ValidateConfig validates JSON syntax, Caddy structure, payload size, and checks base_hash against current_hash
func (s *Service) ValidateConfig(ctx context.Context, req models.ValidateConfigRequest) (models.ValidationResult, string, bool, error) {
	if len([]byte(req.Config)) > s.maxConfigSize {
		return models.ValidationResult{
			Valid:   false,
			Message: "Configuration payload exceeds maximum allowed size limit (2MB)",
		}, "", false, ErrConfigTooLarge
	}

	normalizedJSON, _, err := NormalizeJSON(req.Config)
	if err != nil {
		return models.ValidationResult{
			Valid:   false,
			Message: "Invalid JSON syntax",
			Details: []string{err.Error()},
		}, "", false, nil
	}

	var root map[string]interface{}
	if err := json.Unmarshal([]byte(normalizedJSON), &root); err != nil {
		return models.ValidationResult{
			Valid:   false,
			Message: "Invalid JSON structure",
		}, "", false, nil
	}

	if _, hasApps := root["apps"]; !hasApps {
		return models.ValidationResult{
			Valid:   false,
			Message: "Invalid Caddy configuration structure",
			Details: []string{"Root JSON object should contain an 'apps' module field (e.g. apps.http)"},
		}, "", false, nil
	}

	// Fetch current active config to evaluate current_hash and conflict status
	activeRes, err := s.GetConfig(ctx)
	currentHash := ""
	hasConflict := false
	if err == nil {
		currentHash = activeRes.Hash
		if req.BaseHash != "" && req.BaseHash != currentHash {
			hasConflict = true
		}
	}

	return models.ValidationResult{
		Valid:   true,
		Message: "Caddy configuration JSON is valid",
	}, currentHash, hasConflict, nil
}

// DiffConfig computes backend structured line diff comparing current active config vs proposed config
func (s *Service) DiffConfig(ctx context.Context, req models.DiffConfigRequest) (models.ConfigDiffResponse, error) {
	activeRes, err := s.GetConfig(ctx)
	if err != nil {
		return models.ConfigDiffResponse{}, fmt.Errorf("failed to fetch current caddy config for diff: %w", err)
	}

	proposedNormalized, _, err := NormalizeJSON(req.Config)
	if err != nil {
		return models.ConfigDiffResponse{}, err
	}

	hasConflict := req.BaseHash != "" && req.BaseHash != activeRes.Hash

	diffRes := GenerateDiff(activeRes.Config, proposedNormalized)
	diffRes.CurrentHash = activeRes.Hash
	diffRes.BaseHash = req.BaseHash
	diffRes.HasConflict = hasConflict

	return diffRes, nil
}

// ApplyConfig validates, checks optimistic concurrency, snapshots, applies to Caddy, and verifies post-apply hash
func (s *Service) ApplyConfig(ctx context.Context, userID int64, username string, req models.ApplyConfigRequest) (models.ConfigResponse, error) {
	if len([]byte(req.Config)) > s.maxConfigSize {
		return models.ConfigResponse{}, ErrConfigTooLarge
	}

	normalizedProposed, requestedHash, err := NormalizeJSON(req.Config)
	if err != nil {
		return models.ConfigResponse{}, err
	}

	// 1. Fetch current active configuration
	activeRes, err := s.GetConfig(ctx)
	if err != nil {
		return models.ConfigResponse{}, fmt.Errorf("failed to fetch current caddy config: %w", err)
	}

	// 2. Optimistic Concurrency Check
	if req.BaseHash != "" && req.BaseHash != activeRes.Hash {
		return models.ConfigResponse{}, ErrConfigConflict
	}

	// 3. Create Pre-Apply Backup Snapshot in SQLite
	snapshotID := fmt.Sprintf("snap_%d", time.Now().UTC().UnixNano())
	backupSnapshot := models.ConfigSnapshot{
		ID:         snapshotID,
		Config:     activeRes.Config,
		ConfigHash: activeRes.Hash,
		CreatedBy:  username,
		CreatedAt:  time.Now().UTC(),
		Source:     "manual_apply",
		SizeBytes:  len([]byte(activeRes.Config)),
	}
	if err := s.snapshotStore.CreateSnapshot(backupSnapshot); err != nil {
		return models.ConfigResponse{}, fmt.Errorf("failed to create backup snapshot: %w", err)
	}

	// 4. Post updated config to Caddy Admin API
	if err := s.caddyClient.ApplyConfig(ctx, []byte(normalizedProposed)); err != nil {
		return models.ConfigResponse{}, fmt.Errorf("%w: %v", caddy.ErrCaddyApplyFailed, err)
	}

	// 5. Post-Apply Verification Check
	postApplyRes, err := s.GetConfig(ctx)
	if err != nil {
		return models.ConfigResponse{}, fmt.Errorf("post-apply verification failed to read caddy config: %w", err)
	}

	if postApplyRes.Hash != requestedHash {
		return models.ConfigResponse{}, fmt.Errorf("%w: requested hash %s, got active hash %s", ErrApplyVerificationFailed, requestedHash, postApplyRes.Hash)
	}

	// 6. Record Audit Event (Sanitized without raw config text)
	metaMap := map[string]interface{}{
		"base_hash":    req.BaseHash,
		"new_hash":     requestedHash,
		"snapshot_id":  snapshotID,
		"size_bytes":   len([]byte(normalizedProposed)),
	}
	metaBytes, _ := json.Marshal(metaMap)

	var uID *int64
	if userID > 0 {
		uID = &userID
	}
	_ = s.auditLogger.Log(uID, "config.applied", "config", requestedHash, string(metaBytes))

	return postApplyRes, nil
}

// GetHistory returns snapshot metadata list
func (s *Service) GetHistory(ctx context.Context) ([]models.ConfigHistoryItem, error) {
	return s.snapshotStore.ListSnapshots(50)
}

// GetSnapshot returns single snapshot details by ID
func (s *Service) GetSnapshot(ctx context.Context, id string) (models.ConfigSnapshot, error) {
	return s.snapshotStore.GetSnapshotByID(id)
}

// RollbackConfig restores a historical snapshot with concurrency protection, backup snapshot creation, and post-apply verification
func (s *Service) RollbackConfig(ctx context.Context, userID int64, username string, snapshotID string, req models.RollbackRequest) (models.ConfigResponse, error) {
	snapshot, err := s.snapshotStore.GetSnapshotByID(snapshotID)
	if err != nil {
		return models.ConfigResponse{}, err
	}

	activeRes, err := s.GetConfig(ctx)
	if err != nil {
		return models.ConfigResponse{}, fmt.Errorf("failed to fetch active caddy config: %w", err)
	}

	// Optimistic Concurrency Check
	if req.BaseHash != "" && req.BaseHash != activeRes.Hash {
		return models.ConfigResponse{}, ErrConfigConflict
	}

	// Create backup snapshot of current active configuration before rollback
	backupSnapshotID := fmt.Sprintf("snap_backup_%d", time.Now().UTC().UnixNano())
	backupSnapshot := models.ConfigSnapshot{
		ID:         backupSnapshotID,
		Config:     activeRes.Config,
		ConfigHash: activeRes.Hash,
		CreatedBy:  username,
		CreatedAt:  time.Now().UTC(),
		Source:     "system_backup",
		SizeBytes:  len([]byte(activeRes.Config)),
	}
	_ = s.snapshotStore.CreateSnapshot(backupSnapshot)

	normalizedSnapshotJSON, targetHash, err := NormalizeJSON(snapshot.Config)
	if err != nil {
		return models.ConfigResponse{}, fmt.Errorf("invalid snapshot json: %w", err)
	}

	if err := s.caddyClient.ApplyConfig(ctx, []byte(normalizedSnapshotJSON)); err != nil {
		return models.ConfigResponse{}, fmt.Errorf("%w: %v", caddy.ErrCaddyApplyFailed, err)
	}

	postApplyRes, err := s.GetConfig(ctx)
	if err != nil {
		return models.ConfigResponse{}, fmt.Errorf("post-rollback verification failed: %w", err)
	}

	if postApplyRes.Hash != targetHash {
		return models.ConfigResponse{}, fmt.Errorf("%w: requested rollback hash %s, active %s", ErrApplyVerificationFailed, targetHash, postApplyRes.Hash)
	}

	metaMap := map[string]interface{}{
		"rollback_snapshot_id": snapshotID,
		"backup_snapshot_id":   backupSnapshotID,
		"restored_hash":        targetHash,
	}
	metaBytes, _ := json.Marshal(metaMap)

	var uID *int64
	if userID > 0 {
		uID = &userID
	}
	_ = s.auditLogger.Log(uID, "config.rollback", "config", targetHash, string(metaBytes))

	return postApplyRes, nil
}
