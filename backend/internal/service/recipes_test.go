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

func newTestRecipeService(t *testing.T) (*RecipeService, *PreparationService, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)
	svc := NewRecipeService(db, store.NewRecipeStore())
	prepSvc := NewPreparationService(db, store.NewPreparationStore())

	uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
	user, _, _ := uSvc.GetOrCreate(context.Background(), "test@example.com", "sub123")

	var orgID domain.OrgID
	_ = db.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user.ID).Scan(&orgID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})

	return svc, prepSvc, ctx
}

func baseRecipe() domain.RecipeParams {
	return domain.RecipeParams{
		Name:     "Bread",
		Servings: 4,
	}
}

func seedPreparation(t *testing.T, prepSvc *PreparationService, ctx context.Context) *domain.Preparation {
	t.Helper()
	p, err := prepSvc.Create(ctx, domain.PreparationParams{
		Name:        "Basic Dough",
		YieldAmount: 500,
		YieldUnit:   "g",
		Steps:       []domain.PreparationStep{},
	})
	if err != nil {
		t.Fatalf("seed preparation: %v", err)
	}
	return p
}

func TestRecipeService_List(t *testing.T) {
	t.Run("empty returns slice not nil", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		list, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if list == nil {
			t.Error("got nil, want empty slice")
		}
	})

	t.Run("returns own recipes", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		_, _ = svc.Create(ctx, baseRecipe())
		r2 := baseRecipe()
		r2.Name = "Pizza"
		_, _ = svc.Create(ctx, r2)

		list, err := svc.List(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("got %d recipes, want 2", len(list))
		}
		// ordered by name: Bread, Pizza
		if list[0].Name != "Bread" {
			t.Errorf("got %q, want %q", list[0].Name, "Bread")
		}
	})

	t.Run("RLS: excludes other org recipes", func(t *testing.T) {
		svc1, _, ctx1 := newTestRecipeService(t)
		svc2, _, ctx2 := newTestRecipeService(t)

		_, _ = svc2.Create(ctx2, baseRecipe())

		list, err := svc1.List(ctx1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 0 {
			t.Errorf("got %d recipes, want 0 (cross-org)", len(list))
		}
	})
}

func TestRecipeService_Get(t *testing.T) {
	t.Run("found with preparations", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)
		_, _ = svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})

		got, err := svc.Get(ctx, r.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Bread" {
			t.Errorf("got name %q, want %q", got.Name, "Bread")
		}
		if len(got.Preparations) != 1 {
			t.Errorf("got %d preparations, want 1", len(got.Preparations))
		}
		if got.Preparations[0].Position != 1 {
			t.Errorf("got position %d, want 1", got.Preparations[0].Position)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		_, err := svc.Get(ctx, domain.RecipeID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("RLS: cannot get other org recipe", func(t *testing.T) {
		svc1, _, ctx1 := newTestRecipeService(t)
		svc2, _, ctx2 := newTestRecipeService(t)

		r, _ := svc2.Create(ctx2, baseRecipe())

		_, err := svc1.Get(ctx1, r.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestRecipeService_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		got, err := svc.Create(ctx, baseRecipe())
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Servings != 4 {
			t.Errorf("got servings %d, want 4", got.Servings)
		}
		if got.CreatedAt.IsZero() {
			t.Error("expected created_at to be set")
		}
		if len(got.Preparations) != 0 {
			t.Errorf("got %d preparations on new recipe, want 0", len(got.Preparations))
		}
	})

	t.Run("trims name", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		r := baseRecipe()
		r.Name = "  Sourdough  "

		got, err := svc.Create(ctx, r)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Sourdough" {
			t.Errorf("got %q, want %q", got.Name, "Sourdough")
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		r := baseRecipe()
		r.Name = ""

		_, err := svc.Create(ctx, r)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "name is required" {
			t.Errorf("got msg %q", ve.Msg)
		}
	})

	t.Run("unauthorized returns ErrUnauthorized", func(t *testing.T) {
		svc, _, _ := newTestRecipeService(t)
		_, err := svc.Create(context.Background(), baseRecipe())
		if !errors.Is(err, ErrUnauthorized) {
			t.Errorf("got %v, want ErrUnauthorized", err)
		}
	})

	t.Run("zero servings returns ValidationError", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		r := baseRecipe()
		r.Servings = 0

		_, err := svc.Create(ctx, r)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "servings must be greater than zero" {
			t.Errorf("got msg %q", ve.Msg)
		}
	})
}

func TestRecipeService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())

		updated, err := svc.Update(ctx, r.ID, domain.RecipeParams{
			Name:     "Updated Bread",
			Servings: 8,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Servings != 8 {
			t.Errorf("got servings %d, want 8", updated.Servings)
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())

		_, err := svc.Update(ctx, r.ID, domain.RecipeParams{Servings: 4})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		_, err := svc.Update(ctx, domain.RecipeID(999), baseRecipe())
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestRecipeService_Delete(t *testing.T) {
	t.Run("success cascades recipe_preparations", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)
		_, _ = svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})

		if err := svc.Delete(ctx, r.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.Get(ctx, r.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		err := svc.Delete(ctx, domain.RecipeID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestRecipeService_AddPreparation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)

		got, err := svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Position != 1 {
			t.Errorf("got position %d, want 1", got.Position)
		}
		if got.Amount != 500 {
			t.Errorf("got amount %v, want 500", got.Amount)
		}
	})

	t.Run("recipe not found returns ErrNotFound", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		prep := seedPreparation(t, prepSvc, ctx)

		_, err := svc.AddPreparation(ctx, domain.RecipeID(999), domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("cross-org recipe returns ErrNotFound", func(t *testing.T) {
		svc1, prepSvc1, ctx1 := newTestRecipeService(t)
		svc2, _, ctx2 := newTestRecipeService(t)
		r, _ := svc2.Create(ctx2, baseRecipe())
		prep := seedPreparation(t, prepSvc1, ctx1)

		_, err := svc1.AddPreparation(ctx1, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound (cross-org)", err)
		}
	})

	t.Run("zero position returns ValidationError", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)

		_, err := svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      0,
			Amount:        500,
			Unit:          "g",
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("zero amount returns ValidationError", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)

		_, err := svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Unit:          "g",
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("empty unit returns ValidationError", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)

		_, err := svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})
}

func TestRecipeService_UpdatePreparation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)
		added, _ := svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})

		updated, err := svc.UpdatePreparation(ctx, added.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      2,
			Amount:        750,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Position != 2 {
			t.Errorf("got position %d, want 2", updated.Position)
		}
		if updated.Amount != 750 {
			t.Errorf("got amount %v, want 750", updated.Amount)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		_, err := svc.UpdatePreparation(ctx, domain.RecipePreparationID(999), domain.RecipePreparationParams{
			Position: 1,
			Amount:   100,
			Unit:     "g",
		})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestRecipeService_DeletePreparation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, prepSvc, ctx := newTestRecipeService(t)
		r, _ := svc.Create(ctx, baseRecipe())
		prep := seedPreparation(t, prepSvc, ctx)
		added, _ := svc.AddPreparation(ctx, r.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})

		if err := svc.DeletePreparation(ctx, added.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		got, _ := svc.Get(ctx, r.ID)
		if len(got.Preparations) != 0 {
			t.Errorf("got %d preparations after delete, want 0", len(got.Preparations))
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, _, ctx := newTestRecipeService(t)
		err := svc.DeletePreparation(ctx, domain.RecipePreparationID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
