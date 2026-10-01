package audit

import (
	"database/sql"
	"fmt"
	"time"

	"caddy-manager/internal/database"
	"caddy-manager/internal/models"
)

type Logger struct {
	db *database.DB
}

func NewLogger(db *database.DB) *Logger {
	return &Logger{db: db}
}

func (l *Logger) Log(userID *int64, action, resource, resourceID, metadata string) error {
	now := time.Now().UTC()
	query := `
		INSERT INTO audit_logs (user_id, action, resource, resource_id, metadata, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := l.db.Exec(query, userID, action, resource, resourceID, metadata, now)
	if err != nil {
		return fmt.Errorf("failed to record audit log: %w", err)
	}
	return nil
}

func (l *Logger) GetRecentLogs(limit int) ([]models.AuditLog, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	query := `
		SELECT a.id, a.user_id, COALESCE(u.username, ''), a.action, a.resource, COALESCE(a.resource_id, ''), COALESCE(a.metadata, ''), a.created_at
		FROM audit_logs a
		LEFT JOIN users u ON a.user_id = u.id
		ORDER BY a.created_at DESC
		LIMIT ?
	`

	rows, err := l.db.Query(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to query audit logs: %w", err)
	}
	defer rows.Close()

	var logs []models.AuditLog
	for rows.Next() {
		var item models.AuditLog
		var userID sql.NullInt64

		if err := rows.Scan(&item.ID, &userID, &item.Username, &item.Action, &item.Resource, &item.ResourceID, &item.Metadata, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan audit log row: %w", err)
		}

		if userID.Valid {
			id := userID.Int64
			item.UserID = &id
		}

		logs = append(logs, item)
	}

	if logs == nil {
		logs = []models.AuditLog{}
	}

	return logs, nil
}
