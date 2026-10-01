package configstore

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
)

var (
	ErrSnapshotNotFound = errors.New("configuration snapshot not found")
)

type SnapshotStore struct {
	db        *database.DB
	retention int
}

func NewSnapshotStore(db *database.DB, retention int) *SnapshotStore {
	if retention <= 0 {
		retention = 50
	}
	return &SnapshotStore{
		db:        db,
		retention: retention,
	}
}

// CreateSnapshot saves a new configuration snapshot and prunes old entries beyond retention limit
func (s *SnapshotStore) CreateSnapshot(snapshot models.ConfigSnapshot) error {
	if snapshot.ID == "" {
		snapshot.ID = fmt.Sprintf("snap_%d", time.Now().UTC().UnixNano())
	}
	if snapshot.CreatedAt.IsZero() {
		snapshot.CreatedAt = time.Now().UTC()
	}
	if snapshot.SizeBytes == 0 {
		snapshot.SizeBytes = len([]byte(snapshot.Config))
	}

	query := `
		INSERT INTO config_snapshots (id, config, config_hash, created_by, created_at, source, size_bytes)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	_, err := s.db.Exec(query, snapshot.ID, snapshot.Config, snapshot.ConfigHash, snapshot.CreatedBy, snapshot.CreatedAt, snapshot.Source, snapshot.SizeBytes)
	if err != nil {
		return fmt.Errorf("failed to save config snapshot: %w", err)
	}

	// Prune old snapshots beyond retention count
	_ = s.PruneOldSnapshots()

	return nil
}

// GetSnapshotByID retrieves a single snapshot with full config body
func (s *SnapshotStore) GetSnapshotByID(id string) (models.ConfigSnapshot, error) {
	query := `
		SELECT id, config, config_hash, created_by, created_at, source, size_bytes
		FROM config_snapshots
		WHERE id = ?
	`

	var snap models.ConfigSnapshot
	err := s.db.QueryRow(query, id).Scan(&snap.ID, &snap.Config, &snap.ConfigHash, &snap.CreatedBy, &snap.CreatedAt, &snap.Source, &snap.SizeBytes)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return models.ConfigSnapshot{}, ErrSnapshotNotFound
		}
		return models.ConfigSnapshot{}, fmt.Errorf("failed to query snapshot by id: %w", err)
	}

	return snap, nil
}

// ListSnapshots retrieves metadata for all snapshots ordered by created_at DESC (without config body)
func (s *SnapshotStore) ListSnapshots(limit int) ([]models.ConfigHistoryItem, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	query := `
		SELECT id, config_hash, created_by, created_at, source, size_bytes
		FROM config_snapshots
		ORDER BY created_at DESC
		LIMIT ?
	`

	rows, err := s.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query config history: %w", err)
	}
	defer rows.Close()

	var history []models.ConfigHistoryItem
	for rows.Next() {
		var item models.ConfigHistoryItem
		if err := rows.Scan(&item.ID, &item.ConfigHash, &item.CreatedBy, &item.CreatedAt, &item.Source, &item.SizeBytes); err != nil {
			return nil, fmt.Errorf("failed scanning history row: %w", err)
		}
		history = append(history, item)
	}

	if history == nil {
		history = []models.ConfigHistoryItem{}
	}

	return history, nil
}

// PruneOldSnapshots retains only the most recent N snapshots (defined by retention setting)
func (s *SnapshotStore) PruneOldSnapshots() error {
	query := `
		DELETE FROM config_snapshots
		WHERE id NOT IN (
			SELECT id FROM config_snapshots
			ORDER BY created_at DESC
			LIMIT ?
		)
	`
	_, err := s.db.Exec(query, s.retention)
	return err
}
