// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

func newTestUserService(t *testing.T) (context.Context, *UserService, *store.UserStore) {
	t.Helper()
	db := newTestDB(t)
	us := store.NewUserStore(db)
	orgs := store.NewOrgStore()
	return t.Context(), NewUserService(db, us, orgs), us
}

func TestUserService_GetOrCreate(t *testing.T) {
	t.Run("creates new user", func(t *testing.T) {
		ctx, svc, _ := newTestUserService(t)

		user, created, err := svc.GetOrCreate(ctx, "alice@example.com", "sub-alice")
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
		if user.ID.UUID == uuid.Nil {
			t.Error("expected non-zero UUID")
		}
	})

	t.Run("creates org and membership for new user", func(t *testing.T) {
		ctx, svc, us := newTestUserService(t)

		user, _, err := svc.GetOrCreate(ctx, "bob@example.com", "sub-bob")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var count int
		err = us.DB().QueryRow(
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
		ctx, svc, us := newTestUserService(t)

		if _, _, err := svc.GetOrCreate(ctx, "carol@example.com", "sub-carol"); err != nil {
			t.Fatalf("first call: %v", err)
		}
		user, _, err := svc.GetOrCreate(ctx, "carol@example.com", "sub-carol")
		if err != nil {
			t.Fatalf("second call: %v", err)
		}

		var count int
		if err := us.DB().QueryRowContext(ctx, `SELECT count(*) FROM org_members WHERE user_id = $1`, user.ID).Scan(&count); err != nil {
			t.Fatalf("query org_members: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 org membership, got %d", count)
		}
	})

	t.Run("returns existing user on second call", func(t *testing.T) {
		ctx, svc, _ := newTestUserService(t)

		first, _, err := svc.GetOrCreate(ctx, "carol2@example.com", "sub-carol2")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		second, created, err := svc.GetOrCreate(ctx, "carol2@example.com", "sub-carol2")
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
		ctx, svc, _ := newTestUserService(t)

		if _, _, err := svc.GetOrCreate(ctx, "old@example.com", "sub-dave"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		updated, created, err := svc.GetOrCreate(ctx, "new@example.com", "sub-dave")
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

func TestUserService_Get(t *testing.T) {
	t.Run("returns existing user", func(t *testing.T) {
		ctx, svc, _ := newTestUserService(t)

		created, _, err := svc.GetOrCreate(ctx, "get@example.com", "sub-get")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}

		got, err := svc.Get(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Email != "get@example.com" {
			t.Errorf("got email %q, want %q", got.Email, "get@example.com")
		}
		if got.ID != created.ID {
			t.Errorf("got id %v, want %v", got.ID, created.ID)
		}
	})

	t.Run("unknown id returns ErrNotFound", func(t *testing.T) {
		ctx, svc, _ := newTestUserService(t)

		// Create a user to generate a valid UUID format, then discard it.
		user, _, err := svc.GetOrCreate(ctx, "other@example.com", "sub-other")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}

		// Flip one byte so the UUID is valid but unknown.
		id := user.ID
		id.UUID[0] ^= 0xff

		_, err = svc.Get(id)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func setupUserWithOrg(t *testing.T, svc *UserService, us *store.UserStore) (*domain.User, domain.OrgID) {
	t.Helper()
	user, _, err := svc.GetOrCreate(t.Context(), "pat@example.com", "sub-pat-svc")
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}
	// Use CreateSession + GetUserByToken to discover the org without DB access.
	token, err := us.CreateSession(user.ID, "", "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_, org, _, err := us.GetUserByToken(token, "", "", "")
	if err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}
	return user, org.ID
}

func TestUserService_CreatePAT(t *testing.T) {
	t.Run("returns token with pat_ prefix", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, svc, us)

		token, pat, err := svc.CreatePAT(user.ID, orgID, "my-key", "", "", "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(token) < 4 || token[:4] != "pat_" {
			t.Errorf("expected pat_ prefix, got %q", token[:min(len(token), 10)])
		}
		if pat.Name != "my-key" {
			t.Errorf("got name %q, want %q", pat.Name, "my-key")
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, svc, us)

		_, _, err := svc.CreatePAT(user.ID, orgID, "  ", "", "", "")
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("got %v, want ValidationError", err)
		}
	})
}

func TestUserService_ListPATs(t *testing.T) {
	t.Run("returns empty list when none exist", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, _ := setupUserWithOrg(t, svc, us)

		pats, err := svc.ListPATs(user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pats) != 0 {
			t.Errorf("expected 0 pats, got %d", len(pats))
		}
	})

	t.Run("returns created PATs", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, svc, us)

		if _, _, err := svc.CreatePAT(user.ID, orgID, "key-a", "", "", ""); err != nil {
			t.Fatalf("CreatePAT: %v", err)
		}
		if _, _, err := svc.CreatePAT(user.ID, orgID, "key-b", "", "", ""); err != nil {
			t.Fatalf("CreatePAT: %v", err)
		}

		pats, err := svc.ListPATs(user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pats) != 2 {
			t.Fatalf("expected 2 pats, got %d", len(pats))
		}
		if pats[0].Name != "key-a" {
			t.Errorf("expected first pat name %q, got %q", "key-a", pats[0].Name)
		}
	})
}

func TestUserService_DeletePAT(t *testing.T) {
	t.Run("deletes existing PAT", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, svc, us)

		_, pat, err := svc.CreatePAT(user.ID, orgID, "to-delete", "", "", "")
		if err != nil {
			t.Fatalf("CreatePAT: %v", err)
		}

		if err := svc.DeletePAT(user.ID, pat.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		pats, err := svc.ListPATs(user.ID)
		if err != nil {
			t.Fatalf("ListPATs: %v", err)
		}
		if len(pats) != 0 {
			t.Errorf("expected 0 pats after delete, got %d", len(pats))
		}
	})

	t.Run("unknown id returns ErrNotFound", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, _ := setupUserWithOrg(t, svc, us)

		err := svc.DeletePAT(user.ID, domain.NewTokenID(uuid.Max))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestUserService_Sessions(t *testing.T) {
	t.Run("lists sessions", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, _ := setupUserWithOrg(t, svc, us)

		sessions, err := svc.ListSessions(user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		// setupUserWithOrg creates 1 session via CreateSession.
		if len(sessions) != 1 {
			t.Errorf("expected 1 session, got %d", len(sessions))
		}
	})

	t.Run("deletes session", func(t *testing.T) {
		_, svc, us := newTestUserService(t)
		user, _ := setupUserWithOrg(t, svc, us)

		sessions, _ := svc.ListSessions(user.ID)
		if err := svc.DeleteSession(user.ID, sessions[0].ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		sessions, _ = svc.ListSessions(user.ID)
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions after delete, got %d", len(sessions))
		}
	})
}
