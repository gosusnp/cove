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
	_ = db.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user.ID).Scan(&orgID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})

	return svc, ctx
}

func TestExerciseService_List(t *testing.T) {
	t.Run("returns all exercises", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)
		prog := "Add 1 rep"
		_, _ = svc.Create(ctx, "Push-up", &prog, nil, true)
		_, _ = svc.Create(ctx, "Pull-up", nil, nil, true)

		list, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 exercises, got %d", len(list))
		}

		// Pull-up, Push-up (ordered by name)
		if list[1].Name != "Push-up" {
			t.Errorf("expected second exercise to be Push-up, got %q", list[1].Name)
		}
		if list[1].Progression == nil || *list[1].Progression != prog {
			t.Errorf("expected progression %q, got %v", prog, list[1].Progression)
		}
	})
}

func TestExerciseService_GetByIDs(t *testing.T) {
	t.Run("returns found and computes missing", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)
		e1, _ := svc.Create(ctx, "A", nil, nil, true)
		e2, _ := svc.Create(ctx, "B", nil, nil, true)
		_, _ = svc.Create(ctx, "C", nil, nil, true)

		result, err := svc.GetByIDs(ctx, []domain.ExerciseID{e1.ID, e2.ID, domain.ExerciseID(99999)})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Found) != 2 {
			t.Errorf("expected 2 found, got %d", len(result.Found))
		}
		if result.Found[0].Name != "A" || result.Found[1].Name != "B" {
			t.Errorf("got %v, want A and B", result.Found)
		}
		if len(result.Missing) != 1 || result.Missing[0] != domain.ExerciseID(99999) {
			t.Errorf("got missing %v, want [99999]", result.Missing)
		}
	})

	t.Run("RLS: excludes exercises from other orgs", func(t *testing.T) {
		svc1, ctx1 := newTestExerciseService(t)
		svc2, ctx2 := newTestExerciseService(t)

		// Create a private exercise belonging to org2
		e, _ := svc2.Create(ctx2, "Org2 Secret", nil, nil, false)

		// Org1 requests that ID — should appear in Missing
		result, err := svc1.GetByIDs(ctx1, []domain.ExerciseID{e.ID})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result.Found) != 0 {
			t.Errorf("expected 0 found (cross-org private), got %d", len(result.Found))
		}
		if len(result.Missing) != 1 {
			t.Errorf("expected 1 missing, got %d", len(result.Missing))
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

		_, err := svc.Update(ctx, e.ID, nil, "", nil, nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, ctx := newTestExerciseService(t)

		_, err := svc.Update(ctx, domain.ExerciseID(999), nil, "Valid Name", nil, nil, true)
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
