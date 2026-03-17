// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// OAuthClient represents a dynamically registered OAuth 2.0 client.
type OAuthClient struct {
	ID           string
	Name         string
	RedirectURIs []string
}

// OAuthCode represents a short-lived authorization code.
type OAuthCode struct {
	Code          string
	ClientID      string
	UserID        domain.UserID
	OrgID         domain.OrgID
	RedirectURI   string
	CodeChallenge string
	ExpiresAt     time.Time
}

// OAuthStore handles persistence for OAuth 2.0 clients and authorization codes.
type OAuthStore struct{}

// NewOAuthStore returns a new OAuthStore.
func NewOAuthStore() *OAuthStore {
	return &OAuthStore{}
}

// CreateClient inserts a new OAuth 2.0 client registration.
func (s *OAuthStore) CreateClient(ctx context.Context, q Querier, id, name string, redirectURIs []string) error {
	data, err := json.Marshal(redirectURIs)
	if err != nil {
		return fmt.Errorf("marshal redirect_uris: %w", err)
	}
	_, err = q.ExecContext(ctx,
		`INSERT INTO cove.oauth_clients (id, name, redirect_uris) VALUES ($1, $2, $3)`,
		id, name, string(data),
	)
	if err != nil {
		return fmt.Errorf("create oauth client: %w", err)
	}
	return nil
}

// GetClient returns the registered client with the given ID.
func (s *OAuthStore) GetClient(ctx context.Context, q Querier, id string) (*OAuthClient, error) {
	var c OAuthClient
	var raw string
	err := q.QueryRowContext(ctx,
		`SELECT id, name, redirect_uris FROM cove.oauth_clients WHERE id = $1`,
		id,
	).Scan(&c.ID, &c.Name, &raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get oauth client: %w", err)
	}
	if err := json.Unmarshal([]byte(raw), &c.RedirectURIs); err != nil {
		return nil, fmt.Errorf("unmarshal redirect_uris: %w", err)
	}
	return &c, nil
}

// CreateCode inserts a new authorization code.
func (s *OAuthStore) CreateCode(
	ctx context.Context,
	q Querier,
	code, clientID string,
	userID domain.UserID,
	orgID domain.OrgID,
	redirectURI, codeChallenge string,
	expiresAt time.Time,
) error {
	_, err := q.ExecContext(ctx,
		`INSERT INTO cove.oauth_codes
		 (code, client_id, user_id, org_id, redirect_uri, code_challenge, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		code, clientID, userID, orgID, redirectURI, codeChallenge, expiresAt,
	)
	if err != nil {
		return fmt.Errorf("create oauth code: %w", err)
	}
	return nil
}

// UseCode atomically marks a code as used and returns its data.
// Returns ErrNotFound if the code does not exist, is already used, or has expired.
func (s *OAuthStore) UseCode(ctx context.Context, q Querier, code string) (*OAuthCode, error) {
	var c OAuthCode
	err := q.QueryRowContext(ctx,
		`UPDATE cove.oauth_codes
		 SET used_at = NOW()
		 WHERE code = $1 AND used_at IS NULL AND expires_at > NOW()
		 RETURNING code, client_id, user_id, org_id, redirect_uri, code_challenge, expires_at`,
		code,
	).Scan(&c.Code, &c.ClientID, &c.UserID, &c.OrgID, &c.RedirectURI, &c.CodeChallenge, &c.ExpiresAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("use oauth code: %w", err)
	}
	return &c, nil
}
