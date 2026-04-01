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

type UserStore struct{}

func NewUserStore() *UserStore {
	return &UserStore{}
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
		`INSERT INTO cove.users (id, email, google_sub, is_admin)
		 VALUES ($1, $2, $3, NOT EXISTS (SELECT 1 FROM cove.users WHERE NOT is_service_account))
		 ON CONFLICT (google_sub) DO UPDATE SET email = EXCLUDED.email
		 RETURNING id, email, google_sub, fitness_unit_system, cooking_unit_system, created_at, is_admin, display_name, first_name, last_name, (xmax = 0)`,
		id,
		email,
		googleSub,
	).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.FitnessUnitSystem, &user.CookingUnitSystem, &user.CreatedAt, &user.IsAdmin, &user.DisplayName, &user.FirstName, &user.LastName, &created)

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
		`SELECT id, email, google_sub, fitness_unit_system, cooking_unit_system, created_at, is_admin, display_name, first_name, last_name
		 FROM cove.users WHERE id = $1`,
		id,
	).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.FitnessUnitSystem, &user.CookingUnitSystem, &user.CreatedAt, &user.IsAdmin, &user.DisplayName, &user.FirstName, &user.LastName)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}

	if err != nil {
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

// UpdatePreferences updates the unit system preferences and name fields for the user.
func (s *UserStore) UpdatePreferences(
	ctx context.Context,
	q Querier,
	id domain.UserID,
	fitnessUnitSystem *domain.UnitSystem,
	cookingUnitSystem *domain.UnitSystem,
	displayName *string,
	firstName *string,
	lastName *string,
) (*domain.User, error) {
	var user domain.User
	err := q.QueryRowContext(
		ctx,
		`UPDATE cove.users
		 SET fitness_unit_system = $2, cooking_unit_system = $3, display_name = $4, first_name = $5, last_name = $6
		 WHERE id = $1
		 RETURNING id, email, google_sub, fitness_unit_system, cooking_unit_system, created_at, is_admin, display_name, first_name, last_name`,
		id, fitnessUnitSystem, cookingUnitSystem, displayName, firstName, lastName,
	).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.FitnessUnitSystem, &user.CookingUnitSystem, &user.CreatedAt, &user.IsAdmin, &user.DisplayName, &user.FirstName, &user.LastName)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update preferences: %w", err)
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
) (string, domain.SessionID, error) {
	var orgID domain.OrgID
	if err := q.QueryRowContext(
		ctx,
		`SELECT org_id FROM cove.org_members WHERE user_id = $1 LIMIT 1`,
		userID,
	).Scan(&orgID); err != nil {
		return "", domain.SessionID{}, fmt.Errorf("get org for session: %w", err)
	}

	sessionID := domain.NewSessionID()
	token, err := generateToken("sess_")
	if err != nil {
		return "", domain.SessionID{}, err
	}
	hash := sha256TokenHash(token)

	expiresAt := time.Now().Add(30 * 24 * time.Hour)
	_, err = q.ExecContext(
		ctx,
		`INSERT INTO cove.user_tokens (id, user_id, org_id, kind, token, expires_at, initial_ip_masked, initial_browser, initial_os)
		 VALUES ($1, $2, $3, 'session', $4, $5, $6, $7, $8)`,
		sessionID, userID, orgID, hash, expiresAt, ipMasked, browser, os,
	)
	if err != nil {
		return "", domain.SessionID{}, fmt.Errorf("create session: %w", err)
	}
	return token, sessionID, nil
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
		`INSERT INTO cove.user_tokens (id, user_id, org_id, kind, name, token, initial_ip_masked, initial_browser, initial_os)
		 VALUES ($1, $2, $3, 'pat', $4, $5, $6, $7, $8) RETURNING id, name, created_at`,
		id, userID, orgID, name, hash, ipMasked, browser, os,
	).Scan(&pat.ID, &pat.Name, &pat.CreatedAt)
	if err != nil {
		return "", nil, fmt.Errorf("create pat: %w", err)
	}
	return token, &pat, nil
}

// ListPATs returns all PATs for the given user, ordered by creation time.
func (s *UserStore) ListPATs(ctx context.Context, q Querier, userID domain.UserID) ([]domain.PAT, error) {
	rows, err := q.QueryContext(
		ctx,
		`SELECT id, name, created_at, last_used_at
		 FROM cove.user_tokens
		 WHERE user_id = $1 AND kind = 'pat'
		 ORDER BY created_at`,
		userID,
	)
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
		 FROM cove.user_tokens
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
func (s *UserStore) DeletePAT(ctx context.Context, q Querier, userID domain.UserID, tokenID domain.TokenID) error {
	res, err := q.ExecContext(
		ctx,
		`DELETE FROM cove.user_tokens WHERE id = $1 AND user_id = $2 AND kind = 'pat'`,
		tokenID,
		userID,
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
		`DELETE FROM cove.user_tokens WHERE id = $1 AND user_id = $2 AND kind = 'session'`,
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

// CreateOAuthToken generates an OAuth access token (prefixed with "pat_") that never expires.
// Used by the OAuth 2.0 authorization server after a successful code exchange.
// The raw token is returned once; only its SHA-256 hash is stored.
func (s *UserStore) CreateOAuthToken(
	ctx context.Context,
	q Querier,
	userID domain.UserID,
	orgID domain.OrgID,
	name string,
) (string, error) {
	id, err := uuid.NewV7()
	if err != nil {
		return "", fmt.Errorf("generate oauth token id: %w", err)
	}
	token, err := generateToken("pat_")
	if err != nil {
		return "", err
	}
	hash := sha256TokenHash(token)
	_, err = q.ExecContext(ctx,
		`INSERT INTO cove.user_tokens (id, user_id, org_id, kind, name, token)
		 VALUES ($1, $2, $3, 'oauth', $4, $5)`,
		id, userID, orgID, name, hash,
	)
	if err != nil {
		return "", fmt.Errorf("create oauth token: %w", err)
	}
	return token, nil
}

// RevokeToken deletes a pat or oauth token by its raw value.
// It is a no-op if the token does not exist, per RFC 7009 §2.2.
func (s *UserStore) RevokeToken(ctx context.Context, q Querier, token string) error {
	hash := sha256TokenHash(token)
	_, err := q.ExecContext(ctx,
		`DELETE FROM cove.user_tokens WHERE token = $1 AND kind IN ('pat', 'oauth')`,
		hash,
	)
	if err != nil {
		return fmt.Errorf("revoke token: %w", err)
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
// Returns the user, the org the token is scoped to (nil for service accounts), and the token ID.
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
	var orgUUID *uuid.UUID
	var tokenID uuid.UUID

	// Use UPDATE ... RETURNING to atomically update last used info and fetch user/org details.
	// The CASE handles the 1-minute throttle to minimize vacuum pressure.
	// COALESCE on email/google_sub handles service accounts which have neither.
	// org_id is nullable for service account tokens.
	err := q.QueryRowContext(
		ctx,
		`WITH t AS (
			UPDATE cove.user_tokens
			SET last_used_at = CASE
				WHEN last_used_at IS NULL
				  OR last_used_at < NOW() - INTERVAL '1 minute'
				  OR last_ip_masked IS DISTINCT FROM $2
				  OR last_browser   IS DISTINCT FROM $3
				  OR last_os        IS DISTINCT FROM $4
				THEN NOW()
				ELSE last_used_at
				END,
				last_ip_masked = $2,
				last_browser   = $3,
				last_os        = $4
			WHERE token = $1
			  AND (expires_at IS NULL OR expires_at > NOW())
			RETURNING user_id, org_id, id AS token_id
		)
		SELECT u.id, COALESCE(u.email, ''), COALESCE(u.google_sub, ''), u.fitness_unit_system, u.cooking_unit_system, u.created_at, u.is_service_account, u.is_admin, u.display_name, u.first_name, u.last_name, t.org_id, t.token_id
		FROM t
		JOIN cove.users u ON u.id = t.user_id`,
		sha256TokenHash(token), ipMasked, browser, os,
	).Scan(&user.ID, &user.Email, &user.GoogleSub, &user.FitnessUnitSystem, &user.CookingUnitSystem, &user.CreatedAt, &user.IsServiceAccount, &user.IsAdmin, &user.DisplayName, &user.FirstName, &user.LastName, &orgUUID, &tokenID)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, uuid.Nil, ErrNotFound
	}
	if err != nil {
		return nil, nil, uuid.Nil, fmt.Errorf("get user by token: %w", err)
	}

	var org *domain.Org
	if orgUUID != nil {
		org = &domain.Org{ID: domain.OrgID{UUID: *orgUUID}}
	}
	return &user, org, tokenID, nil
}
