// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/store"
)

func newTestUserService(t *testing.T) (*UserService, *store.UserStore) {
	t.Helper()
	us := store.NewUserStore(newTestDB(t))
	return NewUserService(us), us
}

func TestUserService_Get(t *testing.T) {
	t.Run("returns existing user", func(t *testing.T) {
		svc, us := newTestUserService(t)

		created, _, err := us.GetOrCreate("get@example.com", "sub-get")
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
		svc, us := newTestUserService(t)

		// Create a user to generate a valid UUID format, then discard it.
		user, _, err := us.GetOrCreate("other@example.com", "sub-other")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}

		// Flip one byte so the UUID is valid but unknown.
		id := user.ID
		id[0] ^= 0xff

		_, err = svc.Get(id)
		if !errors.Is(err, store.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func setupUserWithOrg(t *testing.T, us *store.UserStore) (*store.User, uuid.UUID) {
	t.Helper()
	user, _, err := us.GetOrCreate("pat@example.com", "sub-pat-svc")
	if err != nil {
		t.Fatalf("GetOrCreate: %v", err)
	}
	// Use CreateSession + GetUserByToken to discover the org without DB access.
	token, err := us.CreateSession(user.ID, "", "", "")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_, org, err := us.GetUserByToken(token, "", "", "")
	if err != nil {
		t.Fatalf("GetUserByToken: %v", err)
	}
	return user, org.ID
}

func TestUserService_CreatePAT(t *testing.T) {
	t.Run("returns token with pat_ prefix", func(t *testing.T) {
		svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, us)

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
		svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, us)

		_, _, err := svc.CreatePAT(user.ID, orgID, "  ", "", "", "")
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Errorf("got %v, want ValidationError", err)
		}
	})
}

func TestUserService_ListPATs(t *testing.T) {
	t.Run("returns empty list when none exist", func(t *testing.T) {
		svc, us := newTestUserService(t)
		user, _ := setupUserWithOrg(t, us)

		pats, err := svc.ListPATs(user.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(pats) != 0 {
			t.Errorf("expected 0 pats, got %d", len(pats))
		}
	})

	t.Run("returns created PATs", func(t *testing.T) {
		svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, us)

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
		svc, us := newTestUserService(t)
		user, orgID := setupUserWithOrg(t, us)

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
		svc, us := newTestUserService(t)
		user, _ := setupUserWithOrg(t, us)

		err := svc.DeletePAT(user.ID, uuid.Max)
		if !errors.Is(err, store.ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
