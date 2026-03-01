// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	GoogleSub string
	CreatedAt time.Time
}

type UserStore struct {
	db *sql.DB
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{db: db}
}

// GetOrCreate upserts a user by google_sub, creating an org+membership on first insert.
// Returns the user and whether it was newly created.
func (s *UserStore) GetOrCreate(email, googleSub string) (*User, bool, error) {
	tx, err := s.db.Begin()
	if err != nil {
		return nil, false, fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	newID, err := uuid.NewV7()
	if err != nil {
		return nil, false, fmt.Errorf("generate user id: %w", err)
	}

	var user User
	var created bool

	// xmax = 0 means the row was inserted (not updated)
	err = tx.QueryRow(`
		INSERT INTO users (id, email, google_sub)
		VALUES ($1, $2, $3)
		ON CONFLICT (google_sub) DO UPDATE SET email = EXCLUDED.email
		RETURNING id, email, google_sub, created_at, (xmax = 0)
	`, newID, email, googleSub).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.CreatedAt, &created)
	if err != nil {
		return nil, false, fmt.Errorf("upsert user: %w", err)
	}

	if created {
		orgID, err := uuid.NewV7()
		if err != nil {
			return nil, false, fmt.Errorf("generate org id: %w", err)
		}
		if _, err = tx.Exec(`INSERT INTO orgs (id, name) VALUES ($1, $2)`, orgID, email); err != nil {
			return nil, false, fmt.Errorf("create org: %w", err)
		}
		if _, err = tx.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'owner')`, orgID, user.ID); err != nil {
			return nil, false, fmt.Errorf("create org member: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, false, fmt.Errorf("commit: %w", err)
	}

	return &user, created, nil
}

// CreateSession generates a random session token, stores its SHA-256 hash,
// and returns the raw token to the caller (only time it is ever in plaintext).
func (s *UserStore) CreateSession(userID uuid.UUID) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(buf)
	hash := sha256TokenHash(token)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err := s.db.Exec(
		`INSERT INTO user_sessions (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, hash, expiresAt,
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

func sha256TokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GetByID returns the user with the given ID.
func (s *UserStore) GetByID(id uuid.UUID) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT id, email, google_sub, created_at
		FROM users WHERE id = $1
	`, id).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	return &user, nil
}

// GetUserByToken hashes the provided token and looks up the matching non-expired session.
func (s *UserStore) GetUserByToken(token string) (*User, error) {
	var user User
	err := s.db.QueryRow(`
		SELECT u.id, u.email, u.google_sub, u.created_at
		FROM user_sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`, sha256TokenHash(token)).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get user by token: %w", err)
	}
	return &user, nil
}
