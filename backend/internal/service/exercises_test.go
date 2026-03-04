// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestExerciseService(t *testing.T) (*ExerciseService, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)
	svc := NewExerciseService(db, store.NewExerciseStore())

	// Create a test user and org
	uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
	user, _, _ := uSvc.GetOrCreate(context.Background(), "test@example.com", "sub123")

	// Get org ID from org_members
	var orgID domain.OrgID
	_ = db.QueryRow(`SELECT org_id FROM org_members WHERE user_id = $1`, user.ID).Scan(&orgID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})

	return svc, ctx
}

func TestExerciseService_List(t *testing.T) {
	t.Run("returns all exercises", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)
		_, _ = svc.Create(ctx, "Push-up", nil, nil, true)
		_, _ = svc.Create(ctx, "Pull-up", nil, nil, true)

		list, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 exercises, got %d", len(list))
		}
	})
}

func TestExerciseService_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)
		created, _ := svc.Create(ctx, "Push-up", nil, nil, true)

		got, err := svc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Push-up" {
			t.Errorf("got %q, want Push-up", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)
		_, err := svc.Get(ctx, domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseService_Create(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)

		_, err := svc.Create(ctx, "", nil, nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "name is required" {
			t.Errorf("got msg %q, want %q", ve.Msg, "name is required")
		}
	})

	t.Run("trims name", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)

		e, err := svc.Create(ctx, "  Push-up  ", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		if e.Name != "Push-up" {
			t.Errorf("got %q, want %q", e.Name, "Push-up")
		}
	})
}

func TestExerciseService_Update(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)
		e, _ := svc.Create(ctx, "Push-up", nil, nil, true)

		_, err := svc.Update(ctx, e.ID, "", nil, nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)

		_, err := svc.Update(ctx, domain.ExerciseID(999), "Valid Name", nil, nil, true)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)
		created, _ := svc.Create(ctx, "Push-up", nil, nil, true)

		if err := svc.Delete(ctx, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.Get(ctx, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)

		err := svc.Delete(ctx, domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
