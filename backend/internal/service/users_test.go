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
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestUserService(t *testing.T) (context.Context, *UserService) {
	t.Helper()
	db := testutil.NewDB(t)
	return t.Context(), NewUserService(db, store.NewUserStore(), store.NewOrgStore())
}

func TestUserService_GetOrCreate(t *testing.T) {
	t.Run("creates new user and org", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		email := domain.Email("test@example.com")
		sub := domain.GoogleSub("google-123")

		user, created, err := svc.GetOrCreate(ctx, email, sub)
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		if !created {
			t.Error("expected created=true")
		}
		if user.Email != email {
			t.Errorf("got email %q, want %q", user.Email, email)
		}

		// Verify org exists
		var orgID domain.OrgID
		err = svc.db.QueryRowContext(ctx, "SELECT org_id FROM cove.org_members WHERE user_id = $1 LIMIT 1", user.ID).Scan(&orgID)
		if err != nil {
			t.Fatalf("verify org membership: %v", err)
		}
		if orgID.UUID == [16]byte{} {
			t.Error("expected non-nil orgID")
		}
	})

	t.Run("returns existing user", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		email := domain.Email("test@example.com")
		sub := domain.GoogleSub("google-123")

		_, _, _ = svc.GetOrCreate(ctx, email, sub)
		user, created, err := svc.GetOrCreate(ctx, email, sub)

		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		if created {
			t.Error("expected created=false for second call")
		}
		if user.Email != email {
			t.Errorf("got email %q, want %q", user.Email, email)
		}
	})
}

func TestUserService_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		created, _, _ := svc.GetOrCreate(ctx, "test@example.com", "sub")

		got, err := svc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Email != "test@example.com" {
			t.Errorf("got %q, want test@example.com", got.Email)
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		_, err := svc.Get(ctx, domain.NewUserID())
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestUserService_GetUserByToken(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		user, _, _ := svc.GetOrCreate(ctx, "test@example.com", "sub")
		token, _, _ := svc.CreateSession(ctx, user.ID, "1.2.3.4", "Chrome", "OS")

		gotUser, gotOrg, tokenID, err := svc.GetUserByToken(ctx, token, "1.2.3.4", "Chrome", "OS")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if gotUser.ID != user.ID {
			t.Errorf("got userID %v, want %v", gotUser.ID, user.ID)
		}
		if gotOrg == nil {
			t.Error("expected org")
		}
		if tokenID == uuid.Nil {
			t.Error("expected tokenID")
		}
	})

	t.Run("not found", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		_, _, _, err := svc.GetUserByToken(ctx, "invalid", "", "", "")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestUserService_Tokens(t *testing.T) {
	ctx, svc := newTestUserService(t)
	user, _, _ := svc.GetOrCreate(ctx, "test@example.com", "sub")
	var orgID domain.OrgID
	_ = svc.db.QueryRowContext(ctx, "SELECT org_id FROM cove.org_members WHERE user_id = $1 LIMIT 1", user.ID).Scan(&orgID)

	t.Run("session lifecycle", func(t *testing.T) {
		token, tokenID, err := svc.CreateSession(ctx, user.ID, "1.1.1.1", "B1", "O1")
		if err != nil {
			t.Fatal(err)
		}

		sessions, _ := svc.ListSessions(ctx, user.ID)
		if len(sessions) != 1 {
			t.Errorf("expected 1 session, got %d", len(sessions))
		}

		if err := svc.DeleteSession(ctx, user.ID, domain.SessionID(tokenID)); err != nil {
			t.Fatal(err)
		}

		_, _, _, err = svc.GetUserByToken(ctx, token, "", "", "")
		if !errors.Is(err, ErrNotFound) {
			t.Error("expected token to be deleted")
		}
	})

	t.Run("pat lifecycle", func(t *testing.T) {
		token, pat, err := svc.CreatePAT(ctx, user.ID, orgID, "My PAT", "1.1.1.1", "B1", "O1")
		if err != nil {
			t.Fatal(err)
		}

		pats, _ := svc.ListPATs(ctx, user.ID)
		if len(pats) != 1 {
			t.Errorf("expected 1 pat, got %d", len(pats))
		}

		if err := svc.DeletePAT(ctx, user.ID, domain.TokenID(pat.ID)); err != nil {
			t.Fatal(err)
		}

		_, _, _, err = svc.GetUserByToken(ctx, token, "", "", "")
		if !errors.Is(err, ErrNotFound) {
			t.Error("expected pat to be deleted")
		}
	})
}
