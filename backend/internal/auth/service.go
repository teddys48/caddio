package auth

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"caddy-manager/internal/database"
	"caddy-manager/internal/models"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrSetupAlreadyCompleted = errors.New("initial administrator setup has already been completed")
	ErrInvalidCredentials    = errors.New("invalid username or password")
	ErrSessionNotFound       = errors.New("session not found or expired")
)

const SessionDuration = 24 * 7 * time.Hour // 7 days

type Service struct {
	db *database.DB
}

func NewService(db *database.DB) *Service {
	return &Service{db: db}
}

func (s *Service) HasAnyUser() (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to count users: %w", err)
	}
	return count > 0, nil
}

func (s *Service) CreateInitialAdmin(username, password string) (*models.User, error) {
	hasUsers, err := s.HasAnyUser()
	if err != nil {
		return nil, err
	}
	if hasUsers {
		return nil, ErrSetupAlreadyCompleted
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	res, err := s.db.Exec(
		"INSERT INTO users (username, password_hash, created_at, updated_at) VALUES (?, ?, ?, ?)",
		username, string(hash), now, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to insert initial admin user: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	user := &models.User{
		ID:        id,
		Username:  username,
		CreatedAt: now,
		UpdatedAt: now,
	}

	return user, nil
}

func (s *Service) Authenticate(username, password string) (*models.User, error) {
	var user models.User
	var lastLogin sql.NullTime

	err := s.db.QueryRow(
		"SELECT id, username, password_hash, created_at, updated_at, last_login_at FROM users WHERE username = ?",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &user.UpdatedAt, &lastLogin)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("database query error: %w", err)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	now := time.Now().UTC()
	_, _ = s.db.Exec("UPDATE users SET last_login_at = ? WHERE id = ?", now, user.ID)
	user.LastLoginAt = &now

	return &user, nil
}

func (s *Service) CreateSession(userID int64) (*models.Session, error) {
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("failed to generate random session token: %w", err)
	}
	sessionID := hex.EncodeToString(tokenBytes)

	now := time.Now().UTC()
	expiresAt := now.Add(SessionDuration)

	_, err := s.db.Exec(
		"INSERT INTO sessions (id, user_id, expires_at, created_at) VALUES (?, ?, ?, ?)",
		sessionID, userID, expiresAt, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to save session to database: %w", err)
	}

	return &models.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: expiresAt,
		CreatedAt: now,
	}, nil
}

func (s *Service) GetSessionUser(sessionID string) (*models.User, error) {
	if sessionID == "" {
		return nil, ErrSessionNotFound
	}

	var user models.User
	var expiresAt time.Time
	var lastLogin sql.NullTime

	query := `
		SELECT u.id, u.username, u.created_at, u.updated_at, u.last_login_at, s.expires_at
		FROM sessions s
		JOIN users u ON s.user_id = u.id
		WHERE s.id = ?
	`

	err := s.db.QueryRow(query, sessionID).Scan(
		&user.ID, &user.Username, &user.CreatedAt, &user.UpdatedAt, &lastLogin, &expiresAt,
	)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrSessionNotFound
		}
		return nil, fmt.Errorf("failed to lookup session: %w", err)
	}

	if time.Now().UTC().After(expiresAt) {
		// Session expired, delete it
		_ = s.DeleteSession(sessionID)
		return nil, ErrSessionNotFound
	}

	if lastLogin.Valid {
		user.LastLoginAt = &lastLogin.Time
	}

	return &user, nil
}

func (s *Service) DeleteSession(sessionID string) error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE id = ?", sessionID)
	return err
}

func (s *Service) CleanupExpiredSessions() error {
	_, err := s.db.Exec("DELETE FROM sessions WHERE expires_at <= ?", time.Now().UTC())
	return err
}
