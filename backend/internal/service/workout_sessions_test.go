// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestWorkoutSessionService(t *testing.T) (*WorkoutSessionService, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)
	svc := NewWorkoutSessionService(db, store.NewWorkoutSessionStore(), crypto.NewTestEncryptor())

	uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
	user, _, _ := uSvc.GetOrCreate(context.Background(), "test@example.com", "sub123")

	var orgID domain.OrgID
	_ = db.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user.ID).Scan(&orgID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})

	return svc, ctx
}

func TestWorkoutSessionService_List(t *testing.T) {
	t.Run("returns sessions for current user only", func(t *testing.T) {
		db := testutil.NewDB(t)
		uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
		wsSvc := NewWorkoutSessionService(db, store.NewWorkoutSessionStore(), crypto.NewTestEncryptor())

		user1, _, _ := uSvc.GetOrCreate(context.Background(), "u1@example.com", "sub1")
		user2, _, _ := uSvc.GetOrCreate(context.Background(), "u2@example.com", "sub2")

		var orgID1, orgID2 domain.OrgID
		_ = db.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user1.ID).Scan(&orgID1)
		_ = db.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user2.ID).Scan(&orgID2)

		ctx1 := domain.NewContext(context.Background(), &domain.Identity{UserID: user1.ID, OrgID: orgID1})
		ctx2 := domain.NewContext(context.Background(), &domain.Identity{UserID: user2.ID, OrgID: orgID2})

		activity := "Run"
		_, _ = wsSvc.Create(ctx1, store.WorkoutSessionParams{Activity: &activity})

		list, err := wsSvc.List(ctx2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("expected 0 sessions for user2, got %d", len(list))
		}
	})

	t.Run("returns own sessions", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)
		activity1, activity2 := "Run", "Swim"
		_, _ = svc.Create(ctx, store.WorkoutSessionParams{Activity: &activity1})
		_, _ = svc.Create(ctx, store.WorkoutSessionParams{Activity: &activity2})

		list, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 sessions, got %d", len(list))
		}
	})
}

func TestWorkoutSessionService_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)
		activity := "Bike"
		created, _ := svc.Create(ctx, store.WorkoutSessionParams{Activity: &activity})

		got, err := svc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Activity == nil || *got.Activity != activity {
			t.Errorf("got activity %v, want %q", got.Activity, activity)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)

		_, err := svc.Get(ctx, domain.WorkoutSessionID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestWorkoutSessionService_Create_Unauthorized(t *testing.T) {
	svc, _ := newTestWorkoutSessionService(t)

	_, err := svc.Create(context.Background(), store.WorkoutSessionParams{})
	if !errors.Is(err, ErrUnauthorized) {
		t.Errorf("got %v, want ErrUnauthorized", err)
	}
}

func TestWorkoutSessionService_Create(t *testing.T) {
	t.Run("creates session with optional fields", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)
		activity := "Run"
		effort := 6
		duration := 1800

		ws, err := svc.Create(ctx, store.WorkoutSessionParams{
			Activity:  &activity,
			DurationS: &duration,
			SensitiveData: domain.SessionSensitiveData{
				PerceivedEffort: &effort,
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.Activity == nil || *ws.Activity != activity {
			t.Errorf("got activity %v, want %q", ws.Activity, activity)
		}
		if err := ws.UseSensitiveData(ctx, func(private domain.SessionSensitiveData) error {
			if private.PerceivedEffort == nil || *private.PerceivedEffort != effort {
				t.Errorf("got effort %v, want %d", private.PerceivedEffort, effort)
			}
			return nil
		}); err != nil {
			t.Fatalf("UseSensitiveData: %v", err)
		}
	})

	t.Run("creates empty session", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)

		ws, err := svc.Create(ctx, store.WorkoutSessionParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ID == domain.WorkoutSessionID(0) {
			t.Error("expected non-zero ID")
		}
	})
}

func TestWorkoutSessionService_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)
		activity := "Run"
		created, _ := svc.Create(ctx, store.WorkoutSessionParams{Activity: &activity})

		newActivity := "Swim"
		effort := 9
		updated, err := svc.Update(ctx, created.ID, &created.UpdatedAt, store.WorkoutSessionParams{
			Activity: &newActivity,
			SensitiveData: domain.SessionSensitiveData{
				PerceivedEffort: &effort,
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Activity == nil || *updated.Activity != newActivity {
			t.Errorf("got activity %v, want %q", updated.Activity, newActivity)
		}
		if err := updated.UseSensitiveData(ctx, func(private domain.SessionSensitiveData) error {
			if private.PerceivedEffort == nil || *private.PerceivedEffort != effort {
				t.Errorf("got effort %v, want %d", private.PerceivedEffort, effort)
			}
			return nil
		}); err != nil {
			t.Fatalf("UseSensitiveData: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)

		_, err := svc.Update(ctx, domain.WorkoutSessionID(999), nil, store.WorkoutSessionParams{})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestWorkoutSessionService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)
		created, _ := svc.Create(ctx, store.WorkoutSessionParams{})

		if err := svc.Delete(ctx, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.Get(ctx, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, ctx := newTestWorkoutSessionService(t)

		err := svc.Delete(ctx, domain.WorkoutSessionID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
