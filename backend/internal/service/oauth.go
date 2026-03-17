// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

// OAuthService handles OAuth 2.0 authorization server operations.
type OAuthService struct {
	db    *sql.DB
	oauth *store.OAuthStore
	users *store.UserStore
}

// NewOAuthService returns a new OAuthService.
func NewOAuthService(db *sql.DB, oauth *store.OAuthStore, users *store.UserStore) *OAuthService {
	return &OAuthService{db: db, oauth: oauth, users: users}
}

// RegisterClient validates and stores a new OAuth 2.0 client (RFC 7591).
// Returns the generated client ID.
func (s *OAuthService) RegisterClient(ctx context.Context, name string, redirectURIs []string) (string, error) {
	if len(redirectURIs) == 0 {
		return "", &ValidationError{Msg: "redirect_uris is required"}
	}
	for _, u := range redirectURIs {
		if err := validateRedirectURI(u); err != nil {
			return "", err
		}
	}
	clientID := uuid.New().String()
	if err := s.oauth.CreateClient(ctx, s.db, clientID, name, redirectURIs); err != nil {
		return "", fmt.Errorf("register client: %w", err)
	}
	return clientID, nil
}

// ValidateClientRedirectURI checks that clientID is registered and redirectURI is
// listed for that client. Returns ErrNotFound for unknown clients and
// *ValidationError for unregistered redirect URIs.
// Per RFC 6749 §4.1.2.1, these errors must be shown directly (not via redirect).
func (s *OAuthService) ValidateClientRedirectURI(ctx context.Context, clientID, redirectURI string) error {
	client, err := s.oauth.GetClient(ctx, s.db, clientID)
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("get client: %w", err)
	}
	if !containsURI(client.RedirectURIs, redirectURI) {
		return &ValidationError{Msg: "redirect_uri not registered for this client"}
	}
	return nil
}

// Authorize validates the client and redirect URI, then generates an authorization code.
func (s *OAuthService) Authorize(
	ctx context.Context,
	clientID, redirectURI, codeChallenge string,
	userID domain.UserID,
	orgID domain.OrgID,
) (string, error) {
	client, err := s.oauth.GetClient(ctx, s.db, clientID)
	if errors.Is(err, store.ErrNotFound) {
		return "", ErrNotFound
	}
	if err != nil {
		return "", fmt.Errorf("get client: %w", err)
	}
	if !containsURI(client.RedirectURIs, redirectURI) {
		return "", &ValidationError{Msg: "redirect_uri not registered for this client"}
	}

	code, err := oauthRandomHex(32)
	if err != nil {
		return "", fmt.Errorf("generate code: %w", err)
	}
	expiresAt := time.Now().Add(5 * time.Minute)
	if err := s.oauth.CreateCode(ctx, s.db, code, clientID, userID, orgID, redirectURI, codeChallenge, expiresAt); err != nil {
		return "", fmt.Errorf("create code: %w", err)
	}
	return code, nil
}

// Exchange validates an authorization code and issues an OAuth access token (RFC 6749 §4.1.3).
func (s *OAuthService) Exchange(ctx context.Context, code, redirectURI, clientID, codeVerifier string) (string, error) {
	authCode, err := s.oauth.UseCode(ctx, s.db, code)
	if errors.Is(err, store.ErrNotFound) {
		return "", &ValidationError{Msg: "invalid or expired authorization code"}
	}
	if err != nil {
		return "", fmt.Errorf("use code: %w", err)
	}
	if authCode.ClientID != clientID {
		return "", &ValidationError{Msg: "client_id mismatch"}
	}
	if authCode.RedirectURI != redirectURI {
		return "", &ValidationError{Msg: "redirect_uri mismatch"}
	}
	if !verifyPKCE(codeVerifier, authCode.CodeChallenge) {
		return "", &ValidationError{Msg: "invalid code_verifier"}
	}

	token, err := s.users.CreateOAuthToken(ctx, s.db, authCode.UserID, authCode.OrgID, "Claude (OAuth)")
	if err != nil {
		return "", fmt.Errorf("create oauth token: %w", err)
	}
	return token, nil
}

// Revoke deletes the token matching the given raw value (RFC 7009).
// It is a no-op if the token does not exist.
func (s *OAuthService) Revoke(ctx context.Context, token string) error {
	if err := s.users.RevokeToken(ctx, s.db, token); err != nil {
		return fmt.Errorf("revoke token: %w", err)
	}
	return nil
}

// validateRedirectURI returns an error if the URI is not safe for dynamic registration.
// Loopback addresses may use http://; all others must use https://.
func validateRedirectURI(rawURI string) error {
	u, err := url.Parse(rawURI)
	if err != nil {
		return &ValidationError{Msg: "invalid redirect_uri: " + rawURI}
	}
	if u.Fragment != "" {
		return &ValidationError{Msg: "redirect_uri must not contain a fragment"}
	}
	if isLoopback(u) {
		return nil
	}
	if u.Scheme != "https" {
		return &ValidationError{Msg: "redirect_uri must use https for non-loopback addresses"}
	}
	return nil
}

// isLoopback reports whether u points to a loopback address (RFC 8252 §8.3).
func isLoopback(u *url.URL) bool {
	host := u.Hostname()
	return (u.Scheme == "http" || u.Scheme == "https") &&
		(host == "localhost" || host == "127.0.0.1" || host == "::1")
}

// containsURI reports whether uris contains target using exact match.
func containsURI(uris []string, target string) bool {
	for _, u := range uris {
		if u == target {
			return true
		}
	}
	return false
}

// verifyPKCE checks that base64url(sha256(verifier)) == challenge (no padding).
func verifyPKCE(verifier, challenge string) bool {
	h := sha256.Sum256([]byte(verifier))
	got := base64.RawURLEncoding.EncodeToString(h[:])
	return got == challenge
}

// oauthRandomHex generates n random bytes and returns them hex-encoded.
func oauthRandomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("random bytes: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
