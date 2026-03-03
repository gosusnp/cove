// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/domain"
)

type UserStore struct {
	baseStore
}

func NewUserStore(db *sql.DB) *UserStore {
	return &UserStore{baseStore{db: db}}
}

// DB returns the underlying query executor, used for direct database access in tests.
func (s *UserStore) DB() Querier {
	return s.db
}

// WithTx returns a UserStore that executes queries within tx.
func (s *UserStore) WithTx(tx *sql.Tx) *UserStore {
	return &UserStore{s.withTx(tx)}
}

// UpsertUser inserts or updates a user by google_sub.
// Returns the user and whether it was newly created.
func (s *UserStore) UpsertUser(
	ctx context.Context,
	q Querier,
	id domain.UserID,
	email domain.Email,
	googleSub domain.GoogleSub,
) (*domain.User, bool, error) {
	var user domain.User
	var created bool
	err := q.QueryRowContext(
		ctx,
		`INSERT INTO users (id, email, google_sub)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (google_sub) DO UPDATE SET email = EXCLUDED.email
		 RETURNING id, email, google_sub, created_at, (xmax = 0)`,
		id,
		email,
		googleSub,
	).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.CreatedAt, &created)

	if err != nil {
		return nil, false, fmt.Errorf("upsert user: %w", err)
	}
	return &user, created, nil
}

// GetByID returns the user with the given ID.
func (s *UserStore) GetByID(
	ctx context.Context,
	q Querier,
	id domain.UserID,
) (*domain.User, error) {
	var user domain.User
	err := q.QueryRowContext(
		ctx,
		`SELECT id, email, google_sub, created_at
		 FROM users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

// CreateSession generates a random session token (prefixed with "sess_"), stores its
// SHA-256 hash in user_tokens, and returns the raw token (only time it is ever plaintext).
func (s *UserStore) CreateSession(
	ctx context.Context,
	q Querier,
	userID domain.UserID,
	ipMasked domain.MaskedIP,
	browser string,
	os string,
) (string, error) {
	var orgID domain.OrgID
	if err := q.QueryRowContext(
		ctx,
		`SELECT org_id FROM org_members WHERE user_id = $1 LIMIT 1`,
		userID,
	).Scan(&orgID); err != nil {
		return "", fmt.Errorf("get org for session: %w", err)
	}

	sessionID := domain.NewSessionID()
	token, err := generateToken("sess_")
	if err != nil {
		return "", err
	}
	hash := sha256TokenHash(token)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = s.db.Exec(
		`INSERT INTO user_tokens (id, user_id, org_id, kind, token, expires_at, initial_ip_masked, initial_browser, initial_os)
		 VALUES ($1, $2, $3, 'session', $4, $5, $6, $7, $8)`,
		sessionID, userID, orgID, hash, expiresAt, ipMasked, browser, os,
	)
	if err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

// CreatePAT generates a named personal access token (prefixed with "pat_") that never expires.
// The raw token is returned once; only its SHA-256 hash is stored.
func (s *UserStore) CreatePAT(
	ctx context.Context,
	q Querier,
	userID domain.UserID,
	orgID domain.OrgID,
	name string,
	ipMasked domain.MaskedIP,
	browser string, os string,
) (string, *domain.PAT, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", nil, fmt.Errorf("generate pat id: %w", err)
	}
	token, err := generateToken("pat_")
	if err != nil {
		return "", nil, err
	}
	hash := sha256TokenHash(token)

	var pat domain.PAT
	err = q.QueryRowContext(
		ctx,
		`INSERT INTO user_tokens (id, user_id, org_id, kind, name, token, initial_ip_masked, initial_browser, initial_os)
		 VALUES ($1, $2, $3, 'pat', $4, $5, $6, $7, $8) RETURNING id, name, created_at`,
		id, userID, orgID, name, hash, ipMasked, browser, os,
	).Scan(&pat.ID, &pat.Name, &pat.CreatedAt)
	if err != nil {
		return "", nil, fmt.Errorf("create pat: %w", err)
	}
	return token, &pat, nil
}

// ListPATs returns all PATs for the given user, ordered by creation time.
func (s *UserStore) ListPATs(userID domain.UserID) ([]domain.PAT, error) {
	rows, err := s.db.Query(`
		SELECT id, name, created_at, last_used_at
		FROM user_tokens
		WHERE user_id = $1 AND kind = 'pat'
		ORDER BY created_at
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list pats: %w", err)
	}
	defer rows.Close()

	pats := []domain.PAT{}
	for rows.Next() {
		var p domain.PAT
		if err := rows.Scan(&p.ID, &p.Name, &p.CreatedAt, &p.LastUsedAt); err != nil {
			return nil, fmt.Errorf("scan pat: %w", err)
		}
		pats = append(pats, p)
	}
	return pats, rows.Err()
}

// ListSessions returns all active sessions for the given user, ordered by last used time.
func (s *UserStore) ListSessions(ctx context.Context, q Querier, userID domain.UserID) ([]domain.Session, error) {
	rows, err := q.QueryContext(
		ctx,
		`SELECT id, created_at, last_used_at, initial_ip_masked, initial_browser, initial_os, last_ip_masked, last_browser, last_os
		 FROM user_tokens
		 WHERE user_id = $1 AND kind = 'session' AND (expires_at IS NULL OR expires_at > NOW())
		 ORDER BY COALESCE(last_used_at, created_at) DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	sessions := []domain.Session{}
	for rows.Next() {
		var sess domain.Session
		if err := rows.Scan(&sess.ID, &sess.CreatedAt, &sess.LastUsedAt, &sess.InitialIPMasked, &sess.InitialBrowser, &sess.InitialOS, &sess.LastIPMasked, &sess.LastBrowser, &sess.LastOS); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, sess)
	}
	return sessions, rows.Err()
}

// DeletePAT deletes the PAT with the given id scoped to the user. Returns ErrNotFound if no row was deleted.
func (s *UserStore) DeletePAT(userID domain.UserID, tokenID domain.TokenID) error {
	res, err := s.db.Exec(
		`DELETE FROM user_tokens WHERE id = $1 AND user_id = $2 AND kind = 'pat'`,
		tokenID, userID,
	)
	if err != nil {
		return fmt.Errorf("delete pat: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// DeleteSession deletes the session with the given id scoped to the user. Returns ErrNotFound if no row was deleted.
func (s *UserStore) DeleteSession(ctx context.Context, q Querier, userID domain.UserID, sessionID domain.SessionID) error {
	res, err := q.ExecContext(
		ctx,
		`DELETE FROM user_tokens WHERE id = $1 AND user_id = $2 AND kind = 'session'`,
		sessionID,
		userID,
	)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func generateToken(prefix string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate token: %w", err)
	}
	return prefix + hex.EncodeToString(buf), nil
}

func sha256TokenHash(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}

// GetUserByToken hashes the provided token and looks up the matching non-expired token.
// Returns both the user and the org the token is scoped to.
// As a side effect, updates last_used_at and last session info when either the throttle
// window (1 minute) has elapsed or the client info has changed, to minimise dead tuples.
func (s *UserStore) GetUserByToken(
	ctx context.Context,
	q Querier,
	token string,
	ipMasked domain.MaskedIP,
	browser string,
	os string,
) (*domain.User, *domain.Org, uuid.UUID, error) {
	var user domain.User
	var org domain.Org
	var tokenID uuid.UUID
	err := q.QueryRowContext(
		ctx,
		`WITH updated AS (
			UPDATE user_tokens
			SET last_used_at   = NOW(),
			    last_ip_masked = $2,
			    last_browser   = $3,
			    last_os        = $4
			WHERE token = $1
			  AND (expires_at IS NULL OR expires_at > NOW())
			  AND (
			      last_used_at IS NULL
			   OR last_used_at   < NOW() - INTERVAL '1 minute'
			   OR last_ip_masked IS DISTINCT FROM $2
			   OR last_browser   IS DISTINCT FROM $3
			   OR last_os        IS DISTINCT FROM $4
			  )
		 )
		 SELECT u.id, u.email, u.google_sub, u.created_at, t.org_id, t.id
		 FROM user_tokens t
		 JOIN users u ON u.id = t.user_id
		 WHERE t.token = $1 AND (t.expires_at IS NULL OR t.expires_at > NOW())`,
		sha256TokenHash(token),
		ipMasked,
		browser,
		os,
	).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.CreatedAt, &org.ID, &tokenID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, uuid.Nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("get user by token: %w", err)
	}
	return &user, &org, tokenID, nil
}
