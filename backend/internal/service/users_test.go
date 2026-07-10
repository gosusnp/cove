// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"testing"
	"time"

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
		token, _, _, _ := svc.CreateSession(ctx, user.ID, "1.2.3.4", "Chrome", "OS")

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

func TestUserService_PatchPreferences(t *testing.T) {
	t.Run("sets fitness and cooking system", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		user, _, _ := svc.GetOrCreate(ctx, "pref@example.com", "sub-pref")

		fs := domain.UnitSystemImperial
		cs := domain.UnitSystemUSCustomary
		got, err := svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			FitnessUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &fs, Set: true},
			CookingUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &cs, Set: true},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FitnessUnitSystem == nil || *got.FitnessUnitSystem != domain.UnitSystemImperial {
			t.Errorf("got fitness_unit_system %v, want imperial", got.FitnessUnitSystem)
		}
		if got.CookingUnitSystem == nil || *got.CookingUnitSystem != domain.UnitSystemUSCustomary {
			t.Errorf("got cooking_unit_system %v, want us_customary", got.CookingUnitSystem)
		}
	})

	t.Run("absent fields retain current values", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		user, _, _ := svc.GetOrCreate(ctx, "pref-retain@example.com", "sub-pref-retain")

		fs := domain.UnitSystemImperial
		_, _ = svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			FitnessUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &fs, Set: true},
		})

		// Patch only cooking — fitness must remain imperial.
		cs := domain.UnitSystemUSCustomary
		got, err := svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			CookingUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &cs, Set: true},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FitnessUnitSystem == nil || *got.FitnessUnitSystem != domain.UnitSystemImperial {
			t.Errorf("got fitness_unit_system %v, want imperial", got.FitnessUnitSystem)
		}
	})

	t.Run("explicit null clears preference", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		user, _, _ := svc.GetOrCreate(ctx, "pref-clear@example.com", "sub-pref-clear")

		fs := domain.UnitSystemImperial
		_, _ = svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			FitnessUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &fs, Set: true},
		})

		got, err := svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			FitnessUnitSystem: domain.Optional[*domain.UnitSystem]{Value: nil, Set: true},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.FitnessUnitSystem != nil {
			t.Errorf("expected nil fitness_unit_system, got %v", *got.FitnessUnitSystem)
		}
	})

	t.Run("us_customary rejected for fitness", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		user, _, _ := svc.GetOrCreate(ctx, "pref-bad-fit@example.com", "sub-pref-bad-fit")

		fs := domain.UnitSystemUSCustomary
		_, err := svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			FitnessUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &fs, Set: true},
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("invalid fitness system returns ValidationError", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		user, _, _ := svc.GetOrCreate(ctx, "pref-bad-fit2@example.com", "sub-pref-bad-fit2")

		fs := domain.UnitSystem("furlong")
		_, err := svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			FitnessUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &fs, Set: true},
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("invalid cooking system returns ValidationError", func(t *testing.T) {
		ctx, svc := newTestUserService(t)
		user, _, _ := svc.GetOrCreate(ctx, "pref-bad-cook@example.com", "sub-pref-bad-cook")

		cs := domain.UnitSystem("stone")
		_, err := svc.PatchPreferences(ctx, user.ID, UserPreferencesPatch{
			CookingUnitSystem: domain.Optional[*domain.UnitSystem]{Value: &cs, Set: true},
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})
}

func TestUserService_Tokens(t *testing.T) {
	ctx, svc := newTestUserService(t)
	user, _, _ := svc.GetOrCreate(ctx, "test@example.com", "sub")
	var orgID domain.OrgID
	_ = svc.db.QueryRowContext(ctx, "SELECT org_id FROM cove.org_members WHERE user_id = $1 LIMIT 1", user.ID).Scan(&orgID)

	t.Run("session lifecycle", func(t *testing.T) {
		before := time.Now()
		token, tokenID, expiresAt, err := svc.CreateSession(ctx, user.ID, "1.1.1.1", "B1", "O1")
		if err != nil {
			t.Fatal(err)
		}
		wantExpiry := before.Add(SessionTTL)
		if expiresAt.Before(wantExpiry.Add(-time.Second)) || expiresAt.After(wantExpiry.Add(time.Second)) {
			t.Errorf("expiresAt %v not within 1s of expected %v", expiresAt, wantExpiry)
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
