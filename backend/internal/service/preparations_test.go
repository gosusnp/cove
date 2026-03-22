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

func newTestPreparationService(t *testing.T) (*PreparationService, *IngredientService, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)
	svc := NewPreparationService(db, store.NewPreparationStore())
	ingSvc := NewIngredientService(db, store.NewIngredientStore())

	uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
	user, _, _ := uSvc.GetOrCreate(context.Background(), "test@example.com", "sub123")

	var orgID domain.OrgID
	_ = db.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user.ID).Scan(&orgID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})

	return svc, ingSvc, ctx
}

func seedIngredient(t *testing.T, ingSvc *IngredientService, ctx context.Context) *domain.Ingredient {
	t.Helper()
	ing, err := ingSvc.Create(ctx, domain.IngredientParams{
		Name:            "flour",
		CaloriesPer100g: 364,
		ProteinPer100g:  10,
		FatPer100g:      1,
		CarbsPer100g:    76,
	})
	if err != nil {
		t.Fatalf("seed ingredient: %v", err)
	}
	return ing
}

func basePrep() domain.PreparationParams {
	return domain.PreparationParams{
		Name:        "Basic Vinaigrette",
		YieldAmount: 200,
		YieldUnit:   "ml",
		Steps:       []domain.PreparationStep{},
	}
}

func TestPreparationService_List(t *testing.T) {
	t.Run("empty returns slice not nil", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		list, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if list == nil {
			t.Error("got nil, want empty slice")
		}
	})

	t.Run("returns own preparations", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		_, _ = svc.Create(ctx, basePrep())
		p2 := basePrep()
		p2.Name = "Beurre Blanc"
		_, _ = svc.Create(ctx, p2)

		list, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("got %d preparations, want 2", len(list))
		}
		// ordered by name: Basic Vinaigrette, Beurre Blanc
		if list[0].Name != "Basic Vinaigrette" {
			t.Errorf("got %q, want %q", list[0].Name, "Basic Vinaigrette")
		}
	})

	t.Run("RLS: excludes other org preparations", func(t *testing.T) {
		svc1, _, ctx1 := newTestPreparationService(t)
		svc2, _, ctx2 := newTestPreparationService(t)

		_, _ = svc2.Create(ctx2, basePrep())

		list, err := svc1.List(ctx1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d preparations, want 0 (cross-org)", len(list))
		}
	})
}

func TestPreparationService_Get(t *testing.T) {
	t.Run("found with ingredients", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		ing := seedIngredient(t, ingSvc, ctx)
		_, _ = svc.AddIngredient(ctx, p.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})

		got, err := svc.Get(ctx, p.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Basic Vinaigrette" {
			t.Errorf("got name %q, want %q", got.Name, "Basic Vinaigrette")
		}
		if len(got.Ingredients) != 1 {
			t.Errorf("got %d ingredients, want 1", len(got.Ingredients))
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		_, err := svc.Get(ctx, domain.PreparationID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("RLS: cannot get other org preparation", func(t *testing.T) {
		svc1, _, ctx1 := newTestPreparationService(t)
		svc2, _, ctx2 := newTestPreparationService(t)

		p, _ := svc2.Create(ctx2, basePrep())

		_, err := svc1.Get(ctx1, p.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestPreparationService_Create(t *testing.T) {
	t.Run("success with steps", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p := basePrep()
		p.Steps = []domain.PreparationStep{
			{Description: "Whisk oil and vinegar"},
		}

		got, err := svc.Create(ctx, p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.YieldAmount != 200 {
			t.Errorf("got yield_amount %v, want 200", got.YieldAmount)
		}
		if len(got.Steps) != 1 {
			t.Errorf("got %d steps, want 1", len(got.Steps))
		}
		if got.CreatedAt.IsZero() {
			t.Error("expected created_at to be set")
		}
	})

	t.Run("trims name", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p := basePrep()
		p.Name = "  Beurre Blanc  "

		got, err := svc.Create(ctx, p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Beurre Blanc" {
			t.Errorf("got %q, want %q", got.Name, "Beurre Blanc")
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p := basePrep()
		p.Name = ""

		_, err := svc.Create(ctx, p)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "name is required" {
			t.Errorf("got msg %q", ve.Msg)
		}
	})

	t.Run("zero yield_amount returns ValidationError", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p := basePrep()
		p.YieldAmount = 0

		_, err := svc.Create(ctx, p)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("empty yield_unit returns ValidationError", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p := basePrep()
		p.YieldUnit = ""

		_, err := svc.Create(ctx, p)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("unauthorized returns ErrUnauthorized", func(t *testing.T) {
		svc, _, _ := newTestPreparationService(t)
		_, err := svc.Create(context.Background(), basePrep())
		if !errors.Is(err, ErrUnauthorized) {
			t.Errorf("got %v, want ErrUnauthorized", err)
		}
	})

	t.Run("nil steps defaults to empty slice", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p := basePrep()
		p.Steps = nil

		got, err := svc.Create(ctx, p)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Steps == nil {
			t.Error("got nil steps, want empty slice")
		}
	})
}

func TestPreparationService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())

		updated, err := svc.Update(ctx, p.ID, domain.PreparationParams{
			Name:        "Updated Vinaigrette",
			YieldAmount: 300,
			YieldUnit:   "ml",
			Steps:       []domain.PreparationStep{},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.YieldAmount != 300 {
			t.Errorf("got yield_amount %v, want 300", updated.YieldAmount)
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())

		_, err := svc.Update(ctx, p.ID, domain.PreparationParams{YieldAmount: 100, YieldUnit: "ml"})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		_, err := svc.Update(ctx, domain.PreparationID(999), basePrep())
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestPreparationService_Delete(t *testing.T) {
	t.Run("success cascades ingredients", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		ing := seedIngredient(t, ingSvc, ctx)
		_, _ = svc.AddIngredient(ctx, p.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})

		if err := svc.Delete(ctx, p.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.Get(ctx, p.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		err := svc.Delete(ctx, domain.PreparationID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestPreparationService_AddIngredient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		ing := seedIngredient(t, ingSvc, ctx)

		got, err := svc.AddIngredient(ctx, p.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       150,
			Unit:         "g",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Amount != 150 {
			t.Errorf("got amount %v, want 150", got.Amount)
		}
	})

	t.Run("preparation not found returns ErrNotFound", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		ing := seedIngredient(t, ingSvc, ctx)

		_, err := svc.AddIngredient(ctx, domain.PreparationID(999), domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("cross-org preparation returns ErrNotFound", func(t *testing.T) {
		svc1, ingSvc1, ctx1 := newTestPreparationService(t)
		svc2, _, ctx2 := newTestPreparationService(t)
		p, _ := svc2.Create(ctx2, basePrep())
		ing := seedIngredient(t, ingSvc1, ctx1)

		_, err := svc1.AddIngredient(ctx1, p.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound (cross-org)", err)
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		ing := seedIngredient(t, ingSvc, ctx)

		_, err := svc.AddIngredient(ctx, p.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Amount:       100,
			Unit:         "g",
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})
}

func TestPreparationService_UpdateIngredient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		ing := seedIngredient(t, ingSvc, ctx)
		added, _ := svc.AddIngredient(ctx, p.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})

		updated, err := svc.UpdateIngredient(ctx, p.ID, added.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "bread flour",
			Amount:       200,
			Unit:         "g",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Amount != 200 {
			t.Errorf("got amount %v, want 200", updated.Amount)
		}
		if updated.Name != "bread flour" {
			t.Errorf("got name %q, want %q", updated.Name, "bread flour")
		}
	})

	t.Run("ingredient not in preparation returns ErrNotFound", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p1, _ := svc.Create(ctx, basePrep())
		p2 := basePrep()
		p2.Name = "Other Prep"
		p2r, _ := svc.Create(ctx, p2)
		ing := seedIngredient(t, ingSvc, ctx)
		// Add ingredient to p2 only
		added, _ := svc.AddIngredient(ctx, p2r.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})

		// Try to update it via p1 — should fail
		_, err := svc.UpdateIngredient(ctx, p1.ID, added.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       200,
			Unit:         "g",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound (wrong preparation)", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		ing := seedIngredient(t, ingSvc, ctx)

		_, err := svc.UpdateIngredient(ctx, p.ID, domain.PreparationIngredientID(999), domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestPreparationService_DeleteIngredient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		ing := seedIngredient(t, ingSvc, ctx)
		added, _ := svc.AddIngredient(ctx, p.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})

		if err := svc.DeleteIngredient(ctx, p.ID, added.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, _ := svc.Get(ctx, p.ID)
		if len(got.Ingredients) != 0 {
			t.Errorf("got %d ingredients after delete, want 0", len(got.Ingredients))
		}
	})

	t.Run("ingredient not in preparation returns ErrNotFound", func(t *testing.T) {
		svc, ingSvc, ctx := newTestPreparationService(t)
		p1, _ := svc.Create(ctx, basePrep())
		p2 := basePrep()
		p2.Name = "Other Prep"
		p2r, _ := svc.Create(ctx, p2)
		ing := seedIngredient(t, ingSvc, ctx)
		// Add ingredient to p2 only
		added, _ := svc.AddIngredient(ctx, p2r.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "flour",
			Amount:       100,
			Unit:         "g",
		})

		// Try to delete it via p1 — should fail, ingredient still exists in p2
		err := svc.DeleteIngredient(ctx, p1.ID, added.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound (wrong preparation)", err)
		}
		// Verify the ingredient is still in p2
		got, _ := svc.Get(ctx, p2r.ID)
		if len(got.Ingredients) != 1 {
			t.Errorf("got %d ingredients in p2, want 1 (should not have been deleted)", len(got.Ingredients))
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, _, ctx := newTestPreparationService(t)
		p, _ := svc.Create(ctx, basePrep())
		err := svc.DeleteIngredient(ctx, p.ID, domain.PreparationIngredientID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
