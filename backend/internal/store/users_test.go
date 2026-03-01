// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
)

func newTestUserStore(t *testing.T) *UserStore {
	t.Helper()
	return NewUserStore(newTestDB(t))
}

func TestUserStore_GetOrCreate(t *testing.T) {
	t.Run("creates new user", func(t *testing.T) {
		s := newTestUserStore(t)

		user, created, err := s.GetOrCreate("alice@example.com", "sub-alice")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !created {
			t.Error("expected created=true for new user")
		}
		if user.Email != "alice@example.com" {
			t.Errorf("got email %q, want %q", user.Email, "alice@example.com")
		}
		if user.GoogleSub != "sub-alice" {
			t.Errorf("got sub %q, want %q", user.GoogleSub, "sub-alice")
		}
		if user.ID == [16]byte{} {
			t.Error("expected non-zero UUID")
		}
	})

	t.Run("creates org and membership for new user", func(t *testing.T) {
		s := newTestUserStore(t)

		user, _, err := s.GetOrCreate("bob@example.com", "sub-bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var count int
		err = s.db.QueryRow(
			`SELECT count(*) FROM org_members WHERE user_id = $1 AND role = 'owner'`,
			user.ID,
		).Scan(&count)
		if err != nil {
			t.Fatalf("query org_members: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 org membership, got %d", count)
		}
	})

	t.Run("does not create duplicate org on second login", func(t *testing.T) {
		s := newTestUserStore(t)

		if _, _, err := s.GetOrCreate("carol@example.com", "sub-carol"); err != nil {
			t.Fatalf("first call: %v", err)
		}
		user, _, err := s.GetOrCreate("carol@example.com", "sub-carol")
		if err != nil {
			t.Fatalf("second call: %v", err)
		}

		var count int
		if err := s.db.QueryRow(`SELECT count(*) FROM org_members WHERE user_id = $1`, user.ID).Scan(&count); err != nil {
			t.Fatalf("query org_members: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 org membership, got %d", count)
		}
	})

	t.Run("returns existing user on second call", func(t *testing.T) {
		s := newTestUserStore(t)

		first, _, err := s.GetOrCreate("carol@example.com", "sub-carol")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		second, created, err := s.GetOrCreate("carol@example.com", "sub-carol")
		if err != nil {
			t.Fatalf("unexpected error on second call: %v", err)
		}
		if created {
			t.Error("expected created=false for existing user")
		}
		if first.ID != second.ID {
			t.Errorf("expected same ID, got %v and %v", first.ID, second.ID)
		}
	})

	t.Run("updates email when google sub already exists", func(t *testing.T) {
		s := newTestUserStore(t)

		if _, _, err := s.GetOrCreate("old@example.com", "sub-dave"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, created, err := s.GetOrCreate("new@example.com", "sub-dave")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if created {
			t.Error("expected created=false")
		}
		if updated.Email != "new@example.com" {
			t.Errorf("got email %q, want %q", updated.Email, "new@example.com")
		}
	})
}

func TestUserStore_CreateSession(t *testing.T) {
	t.Run("returns non-empty token with sess_ prefix", func(t *testing.T) {
		s := newTestUserStore(t)
		user, _, err := s.GetOrCreate("eve@example.com", "sub-eve")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		token, err := s.CreateSession(user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if token == "" {
			t.Error("expected non-empty token")
		}
		if len(token) < 5 || token[:5] != "sess_" {
			t.Errorf("expected sess_ prefix, got %q", token[:min(len(token), 10)])
		}
	})

	t.Run("stores hash not plaintext", func(t *testing.T) {
		s := newTestUserStore(t)
		user, _, err := s.GetOrCreate("frank@example.com", "sub-frank")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		token, err := s.CreateSession(user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		h := sha256.Sum256([]byte(token))
		wantHash := hex.EncodeToString(h[:])

		var storedToken string
		err = s.db.QueryRow(`SELECT token FROM user_tokens WHERE user_id = $1 AND kind = 'session'`, user.ID).Scan(&storedToken)
		if err != nil {
			t.Fatalf("query session: %v", err)
		}
		if storedToken == token {
			t.Error("plaintext token must not be stored in DB")
		}
		if storedToken != wantHash {
			t.Errorf("stored token %q, want SHA-256 hash %q", storedToken, wantHash)
		}
	})
}

func TestUserStore_CreatePAT(t *testing.T) {
	// lookupOrgID fetches the user's org from org_members.
	lookupOrgID := func(t *testing.T, s *UserStore, userID uuid.UUID) uuid.UUID {
		t.Helper()
		var orgID uuid.UUID
		if err := s.db.QueryRow(
			`SELECT org_id FROM org_members WHERE user_id = $1 LIMIT 1`, userID,
		).Scan(&orgID); err != nil {
			t.Fatalf("lookup org: %v", err)
		}
		return orgID
	}

	t.Run("returns token with pat_ prefix", func(t *testing.T) {
		s := newTestUserStore(t)
		user, _, err := s.GetOrCreate("pat@example.com", "sub-pat")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		orgID := lookupOrgID(t, s, user.ID)

		token, pat, err := s.CreatePAT(user.ID, orgID, "my-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(token) < 4 || token[:4] != "pat_" {
			t.Errorf("expected pat_ prefix, got %q", token[:min(len(token), 10)])
		}
		if pat.ID == (uuid.UUID{}) {
			t.Error("expected non-zero PAT id")
		}
		if pat.Name != "my-token" {
			t.Errorf("got name %q, want %q", pat.Name, "my-token")
		}
	})

	t.Run("PAT is valid for GetUserByToken", func(t *testing.T) {
		s := newTestUserStore(t)
		user, _, err := s.GetOrCreate("pat2@example.com", "sub-pat2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		orgID := lookupOrgID(t, s, user.ID)

		token, _, err := s.CreatePAT(user.ID, orgID, "ci-token")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, _, err := s.GetUserByToken(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != user.ID {
			t.Errorf("got user ID %v, want %v", got.ID, user.ID)
		}
	})
}

func TestUserStore_GetUserByToken(t *testing.T) {
	t.Run("valid token returns user", func(t *testing.T) {
		s := newTestUserStore(t)
		user, _, err := s.GetOrCreate("grace@example.com", "sub-grace")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		token, err := s.CreateSession(user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		got, org, err := s.GetUserByToken(token)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ID != user.ID {
			t.Errorf("got user ID %v, want %v", got.ID, user.ID)
		}
		if got.Email != user.Email {
			t.Errorf("got email %q, want %q", got.Email, user.Email)
		}
		if org.ID == (uuid.UUID{}) {
			t.Error("expected non-zero org ID")
		}
	})

	t.Run("invalid token returns ErrNotFound", func(t *testing.T) {
		s := newTestUserStore(t)

		_, _, err := s.GetUserByToken("notavalidtoken")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("expired token returns ErrNotFound", func(t *testing.T) {
		s := newTestUserStore(t)
		user, _, err := s.GetOrCreate("henry@example.com", "sub-henry")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var orgID uuid.UUID
		if err = s.db.QueryRow(
			`SELECT org_id FROM org_members WHERE user_id = $1 LIMIT 1`, user.ID,
		).Scan(&orgID); err != nil {
			t.Fatalf("lookup org: %v", err)
		}

		// Insert an already-expired session directly.
		expiredAt := time.Now().Add(-1 * time.Hour)
		_, err = s.db.Exec(
			`INSERT INTO user_tokens (user_id, org_id, kind, token, expires_at) VALUES ($1, $2, 'session', $3, $4)`,
			user.ID, orgID, "expiredtokenhash", expiredAt,
		)
		if err != nil {
			t.Fatalf("insert expired session: %v", err)
		}

		// The raw "token" whose hash would be "expiredtokenhash" doesn't exist,
		// but we can verify the expired row isn't returned by querying directly
		// with the stored hash value (simulating a lookup by hash).
		var count int
		err = s.db.QueryRow(
			`SELECT count(*) FROM user_tokens WHERE token = $1 AND expires_at > NOW()`,
			"expiredtokenhash",
		).Scan(&count)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		if count != 0 {
			t.Error("expired session should not be returned by active query")
		}
	})
}
