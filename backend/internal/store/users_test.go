// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestUserStore(t *testing.T) (context.Context, *sql.DB, *UserStore, *OrgStore) {
	t.Helper()
	db := testutil.NewDB(t)
	return t.Context(), db, NewUserStore(), NewOrgStore()
}

// createTestUser seeds a user with an org and owner membership, mirroring the
// full setup that UserService.GetOrCreate performs in production.
func createTestUser(t *testing.T, db *sql.DB, s *UserStore, os *OrgStore, email string, googleSub string) (*domain.User, domain.OrgID) {
	t.Helper()

	ctx := t.Context()

	userID := domain.NewUserID()
	user, created, err := s.UpsertUser(ctx, db, userID, domain.Email(email), domain.GoogleSub(googleSub))
	if err != nil {
		t.Fatalf("UpsertUser: %v", err)
	}

	var orgID domain.OrgID
	if created {
		orgID = domain.NewOrgID()
		if err := os.CreateOrg(ctx, db, orgID, email); err != nil {
			t.Fatalf("CreateOrg: %v", err)
		}
		if err := os.CreateOrgMember(ctx, db, orgID, user.ID, "owner"); err != nil {
			t.Fatalf("CreateOrgMember: %v", err)
		}
	} else {
		err = db.QueryRowContext(ctx, "SELECT org_id FROM org_members WHERE user_id = $1 LIMIT 1", user.ID).Scan(&orgID)
		if err != nil {
			t.Fatalf("get existing org: %v", err)
		}
	}

	return user, orgID
}

func TestUserStore_UpsertUser(t *testing.T) {
	t.Run("create new user", func(t *testing.T) {
		ctx, db, s, _ := newTestUserStore(t)
		id := domain.NewUserID()

		user, created, err := s.UpsertUser(ctx, db, id, "test@example.com", "google-123")
		if err != nil {
			t.Fatalf("UpsertUser: %v", err)
		}
		if !created {
			t.Error("expected created=true")
		}
		if user.Email != "test@example.com" {
			t.Errorf("got email %q, want test@example.com", user.Email)
		}
		if user.ID != id {
			t.Errorf("got id %v, want %v", user.ID, id)
		}
	})

	t.Run("update existing user (same sub, same email)", func(t *testing.T) {
		ctx, db, s, _ := newTestUserStore(t)
		id := domain.NewUserID()

		_, _, _ = s.UpsertUser(ctx, db, id, "test@example.com", "google-123")
		user, created, err := s.UpsertUser(ctx, db, id, "test@example.com", "google-123")

		if err != nil {
			t.Fatalf("UpsertUser: %v", err)
		}
		if created {
			t.Error("expected created=false")
		}
		if user.ID != id {
			t.Errorf("got id %v, want %v", user.ID, id)
		}
	})
}

func TestUserStore_GetByID(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		ctx, db, s, _ := newTestUserStore(t)
		id := domain.NewUserID()
		_, _, _ = s.UpsertUser(ctx, db, id, "test@example.com", "sub")

		user, err := s.GetByID(ctx, db, id)
		if err != nil {
			t.Fatalf("GetByID: %v", err)
		}
		if user.Email != "test@example.com" {
			t.Errorf("got email %q, want test@example.com", user.Email)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctx, db, s, _ := newTestUserStore(t)
		_, err := s.GetByID(ctx, db, domain.NewUserID())
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestUserStore_CreateSession(t *testing.T) {
	ctx, db, s, os := newTestUserStore(t)
	user, _ := createTestUser(t, db, s, os, "test@example.com", "sub")

	t.Run("success", func(t *testing.T) {
		token, _, err := s.CreateSession(ctx, db, user.ID, "127.0.0.1", "Chrome", "macOS")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		// Verify it exists (hashed)
		hash := sha256TokenHash(token)
		var count int
		err = db.QueryRow("SELECT count(*) FROM user_tokens WHERE token = $1 AND kind = 'session'", hash).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Error("session token not created")
		}
	})
}

func TestUserStore_CreatePAT(t *testing.T) {
	ctx, db, s, os := newTestUserStore(t)
	user, orgID := createTestUser(t, db, s, os, "test@example.com", "sub")

	t.Run("success", func(t *testing.T) {
		token, pat, err := s.CreatePAT(ctx, db, user.ID, orgID, "My PAT", "127.0.0.1", "Chrome", "macOS")
		if err != nil {
			t.Fatalf("CreatePAT: %v", err)
		}

		if pat.Name != "My PAT" {
			t.Errorf("got name %q, want 'My PAT'", pat.Name)
		}

		// Verify it exists (hashed)
		hash := sha256TokenHash(token)
		var count int
		err = db.QueryRow("SELECT count(*) FROM user_tokens WHERE token = $1 AND kind = 'pat'", hash).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 1 {
			t.Error("pat token not created")
		}
	})
}

func TestUserStore_Lists(t *testing.T) {
	ctx, db, s, os := newTestUserStore(t)
	user, orgID := createTestUser(t, db, s, os, "test@example.com", "sub")

	t.Run("sessions", func(t *testing.T) {
		_, _, _ = s.CreateSession(ctx, db, user.ID, "1.1.1.1", "B1", "O1")
		_, _, _ = s.CreateSession(ctx, db, user.ID, "2.2.2.2", "B2", "O2")

		list, err := s.ListSessions(ctx, db, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 sessions, got %d", len(list))
		}
	})

	t.Run("pats", func(t *testing.T) {
		_, _, _ = s.CreatePAT(ctx, db, user.ID, orgID, "PAT 1", "1.1.1.1", "B1", "O1")
		_, _, _ = s.CreatePAT(ctx, db, user.ID, orgID, "PAT 2", "2.2.2.2", "B2", "O2")

		list, err := s.ListPATs(ctx, db, user.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 pats, got %d", len(list))
		}
	})
}

func TestUserStore_GetUserByToken(t *testing.T) {
	ctx, db, s, os := newTestUserStore(t)
	user, _ := createTestUser(t, db, s, os, "test@example.com", "sub")

	t.Run("found active session", func(t *testing.T) {
		token, _, _ := s.CreateSession(ctx, db, user.ID, "1.1.1.1", "B1", "O1")

		gotUser, gotOrg, tokenID, err := s.GetUserByToken(ctx, db, token, "2.2.2.2", "B2", "O2")
		if err != nil {
			t.Fatalf("GetUserByToken: %v", err)
		}
		if gotUser.ID != user.ID {
			t.Errorf("got userID %v, want %v", gotUser.ID, user.ID)
		}
		if gotOrg.ID == (domain.OrgID{}) {
			t.Error("expected non-empty org ID")
		}
		if tokenID == (uuid.Nil) {
			t.Error("expected non-empty token ID")
		}
	})

	t.Run("not found", func(t *testing.T) {
		_, _, _, err := s.GetUserByToken(ctx, db, "nonexistent", "", "", "")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestUserStore_Deletes(t *testing.T) {
	ctx, db, s, os := newTestUserStore(t)
	user, orgID := createTestUser(t, db, s, os, "test@example.com", "sub")

	t.Run("delete pat", func(t *testing.T) {
		_, pat, _ := s.CreatePAT(ctx, db, user.ID, orgID, "To Delete", "", "", "")

		err := s.DeletePAT(ctx, db, user.ID, domain.TokenID(pat.ID))
		if err != nil {
			t.Fatalf("DeletePAT: %v", err)
		}

		// Verify gone
		var count int
		err = db.QueryRow("SELECT count(*) FROM user_tokens WHERE id = $1", pat.ID).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Error("token still exists")
		}
	})

	t.Run("delete session", func(t *testing.T) {
		_, sessionID, _ := s.CreateSession(ctx, db, user.ID, "", "", "")

		err := s.DeleteSession(ctx, db, user.ID, sessionID)
		if err != nil {
			t.Fatalf("DeleteSession: %v", err)
		}

		// Verify gone
		var count int
		err = db.QueryRow("SELECT count(*) FROM user_tokens WHERE id = $1", sessionID).Scan(&count)
		if err != nil {
			t.Fatal(err)
		}
		if count != 0 {
			t.Error("token still exists")
		}
	})
}
