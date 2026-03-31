// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

// Tests for migration 023: service account RLS extensions across all tenant tables.
//
// Pattern per table:
//   - SELECT (Get)   → must succeed: USING extended for service accounts
//   - UPDATE         → must succeed: USING + WITH CHECK extended for service accounts
//   - INSERT         → must fail:    WITH CHECK not extended; org_id cannot be derived
//   - DELETE         → ErrNotFound:  USING not extended; row is invisible to service account
//
// Exception — programs:
//   programs UPDATE triggers program_history_trigger, which INSERTs into program_versions.
//   The program_versions INSERT policy is not extended for service accounts, so the trigger
//   INSERT fails, causing the whole UPDATE to fail. Service account UPDATE on programs is
//   therefore not supported and is not tested here.
//
// preparation_ingredients and recipe_preparations are covered implicitly:
//   - SELECT is exercised by PreparationStore.Get and RecipeStore.Get (which JOIN to those tables).
//   - UPDATE via UpdateIngredient is explicitly tested in TestPreparationStore_ServiceAccountRLS.

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

// svcRLSFixture holds shared state for service account RLS tests.
type svcRLSFixture struct {
	db      *sql.DB
	uID     domain.UserID
	oID     domain.OrgID
	svcUID  domain.UserID
	userCtx context.Context
	svcCtx  context.Context
}

func newSvcRLSFixture(t *testing.T) *svcRLSFixture {
	t.Helper()
	db := testutil.NewDB(t)

	uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")}
	oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")}
	svcUID := domain.NewUserID()

	_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'rls@test.com', 'sub-rls')`, uID)
	_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'rls-org')`, oID)
	_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)
	_, _ = db.Exec(`INSERT INTO cove.users (id, is_service_account) VALUES ($1, true)`, svcUID)

	userCtx := domain.NewContext(context.Background(), &domain.Identity{UserID: uID, OrgID: oID})
	svcCtx := domain.NewContext(context.Background(), &domain.Identity{UserID: svcUID, ServiceAccount: true})

	return &svcRLSFixture{
		db:      db,
		uID:     uID,
		oID:     oID,
		svcUID:  svcUID,
		userCtx: userCtx,
		svcCtx:  svcCtx,
	}
}

// commitUserTx begins a ScopedQuerier transaction, runs fn, then commits.
func (f *svcRLSFixture) commitUserTx(t *testing.T, fn func(Querier)) {
	t.Helper()
	tx, err := f.db.BeginTx(f.userCtx, nil)
	if err != nil {
		t.Fatalf("begin user tx: %v", err)
	}
	fn(NewScopedQuerier(tx, f.oID.String(), f.uID.String()))
	if err := tx.Commit(); err != nil {
		t.Fatalf("commit user tx: %v", err)
	}
}

// newSvcTx begins a service account transaction and registers Rollback on cleanup.
func (f *svcRLSFixture) newSvcTx(t *testing.T) Querier {
	t.Helper()
	tx, err := f.db.BeginTx(f.svcCtx, nil)
	if err != nil {
		t.Fatalf("begin svc tx: %v", err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })
	return NewServiceScopedQuerier(tx, f.svcUID.String())
}

// -----------------------------------------------------------------------------
// exercises
// -----------------------------------------------------------------------------

func TestExerciseStore_ServiceAccountRLS(t *testing.T) {
	f := newSvcRLSFixture(t)
	s := NewExerciseStore()

	var ex *domain.Exercise
	f.commitUserTx(t, func(q Querier) {
		var err error
		ex, err = s.Create(f.userCtx, q, "Squat", nil, nil, false)
		if err != nil {
			t.Fatalf("seed exercise: %v", err)
		}
	})

	t.Run("service account can SELECT exercises", func(t *testing.T) {
		q := f.newSvcTx(t)
		got, err := s.Get(f.svcCtx, q, f.oID, ex.ID)
		if err != nil {
			t.Fatalf("expected SELECT to succeed, got: %v", err)
		}
		if got.ID != ex.ID {
			t.Errorf("got ID %v, want %v", got.ID, ex.ID)
		}
	})

	t.Run("service account can UPDATE exercises", func(t *testing.T) {
		q := f.newSvcTx(t)
		_, err := s.Update(f.svcCtx, q, f.oID, ex.ID, "Squat Updated", nil, nil, false, nil)
		if err != nil {
			t.Fatalf("expected UPDATE to succeed, got: %v", err)
		}
	})

	t.Run("service account cannot INSERT exercises", func(t *testing.T) {
		q := f.newSvcTx(t)
		if _, err := s.Create(f.svcCtx, q, "Press", nil, nil, false); err == nil {
			t.Error("expected INSERT to be blocked for service account, got nil error")
		}
	})

	t.Run("service account cannot DELETE exercises", func(t *testing.T) {
		q := f.newSvcTx(t)
		// RLS-via-invisibility: USING clause not extended, row is invisible → ErrNotFound.
		if err := s.Delete(f.svcCtx, q, f.oID, ex.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected DELETE to be invisible via RLS (ErrNotFound), got: %v", err)
		}
	})
}

// -----------------------------------------------------------------------------
// programs
// -----------------------------------------------------------------------------

func TestProgramStore_ServiceAccountRLS(t *testing.T) {
	f := newSvcRLSFixture(t)
	s := NewProgramStore()

	var p *domain.ProgramLite
	f.commitUserTx(t, func(q Querier) {
		var err error
		p, err = s.Create(f.userCtx, q, f.oID, "Strength", nil, nil, false)
		if err != nil {
			t.Fatalf("seed program: %v", err)
		}
	})

	t.Run("service account can SELECT programs", func(t *testing.T) {
		q := f.newSvcTx(t)
		got, err := s.GetLite(f.svcCtx, q, f.oID, p.ID)
		if err != nil {
			t.Fatalf("expected SELECT to succeed, got: %v", err)
		}
		if got.ID != p.ID {
			t.Errorf("got ID %v, want %v", got.ID, p.ID)
		}
	})

	// UPDATE is not tested here: programs UPDATE fires program_history_trigger,
	// which INSERTs into program_versions. The program_versions INSERT policy is
	// not extended for service accounts, causing the trigger INSERT to fail and
	// rolling back the entire UPDATE. Service accounts cannot update programs.

	t.Run("service account cannot INSERT programs", func(t *testing.T) {
		q := f.newSvcTx(t)
		if _, err := s.Create(f.svcCtx, q, f.oID, "Cardio", nil, nil, false); err == nil {
			t.Error("expected INSERT to be blocked for service account, got nil error")
		}
	})

	t.Run("service account cannot DELETE programs", func(t *testing.T) {
		q := f.newSvcTx(t)
		// RLS-via-invisibility: USING clause not extended, row is invisible → ErrNotFound.
		if err := s.Delete(f.svcCtx, q, f.oID, p.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected DELETE to be invisible via RLS (ErrNotFound), got: %v", err)
		}
	})
}

// -----------------------------------------------------------------------------
// ingredients
// -----------------------------------------------------------------------------

func TestIngredientStore_ServiceAccountRLS(t *testing.T) {
	f := newSvcRLSFixture(t)
	s := NewIngredientStore()

	params := domain.IngredientParams{
		Name:            "Oats",
		CaloriesPer100g: 389,
		ProteinPer100g:  17,
		FatPer100g:      7,
		CarbsPer100g:    66,
	}

	var ing *domain.Ingredient
	f.commitUserTx(t, func(q Querier) {
		var err error
		ing, err = s.Create(f.userCtx, q, params)
		if err != nil {
			t.Fatalf("seed ingredient: %v", err)
		}
	})

	t.Run("service account can SELECT ingredients", func(t *testing.T) {
		q := f.newSvcTx(t)
		got, err := s.Get(f.svcCtx, q, f.oID, ing.ID)
		if err != nil {
			t.Fatalf("expected SELECT to succeed, got: %v", err)
		}
		if got.ID != ing.ID {
			t.Errorf("got ID %v, want %v", got.ID, ing.ID)
		}
	})

	t.Run("service account can UPDATE ingredients", func(t *testing.T) {
		q := f.newSvcTx(t)
		updated := params
		updated.Name = "Rolled Oats"
		if _, err := s.Update(f.svcCtx, q, f.oID, ing.ID, updated); err != nil {
			t.Fatalf("expected UPDATE to succeed, got: %v", err)
		}
	})

	t.Run("service account cannot INSERT ingredients", func(t *testing.T) {
		q := f.newSvcTx(t)
		if _, err := s.Create(f.svcCtx, q, domain.IngredientParams{
			Name: "Rice", CaloriesPer100g: 130, ProteinPer100g: 3, FatPer100g: 0.3, CarbsPer100g: 28,
		}); err == nil {
			t.Error("expected INSERT to be blocked for service account, got nil error")
		}
	})

	t.Run("service account cannot DELETE ingredients", func(t *testing.T) {
		q := f.newSvcTx(t)
		// RLS-via-invisibility: USING clause not extended, row is invisible → ErrNotFound.
		if err := s.Delete(f.svcCtx, q, f.oID, ing.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected DELETE to be invisible via RLS (ErrNotFound), got: %v", err)
		}
	})
}

// -----------------------------------------------------------------------------
// preparations (also covers preparation_ingredients SELECT via Get)
// -----------------------------------------------------------------------------

func TestPreparationStore_ServiceAccountRLS(t *testing.T) {
	f := newSvcRLSFixture(t)
	s := NewPreparationStore()
	is := NewIngredientStore()

	prepParams := domain.PreparationParams{
		Name:        "Oatmeal",
		YieldAmount: 1,
		YieldUnit:   "serving",
		Steps:       []domain.PreparationStep{{Description: "Mix and cook"}},
	}

	var prep *domain.Preparation
	var ing *domain.Ingredient
	var prepIngID domain.PreparationIngredientID
	f.commitUserTx(t, func(q Querier) {
		var err error
		ing, err = is.Create(f.userCtx, q, domain.IngredientParams{
			Name: "Oats", CaloriesPer100g: 389, ProteinPer100g: 17, FatPer100g: 7, CarbsPer100g: 66,
		})
		if err != nil {
			t.Fatalf("seed ingredient: %v", err)
		}
		prep, err = s.Create(f.userCtx, q, prepParams)
		if err != nil {
			t.Fatalf("seed preparation: %v", err)
		}
		pi, err := s.AddIngredient(f.userCtx, q, f.oID, prep.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID, Name: "Oats", Amount: 100, Unit: "g",
		})
		if err != nil {
			t.Fatalf("seed preparation ingredient: %v", err)
		}
		prepIngID = pi.ID
	})

	t.Run("service account can SELECT preparations (and preparation_ingredients via Get)", func(t *testing.T) {
		q := f.newSvcTx(t)
		got, err := s.Get(f.svcCtx, q, f.oID, prep.ID)
		if err != nil {
			t.Fatalf("expected SELECT to succeed, got: %v", err)
		}
		if got.ID != prep.ID {
			t.Errorf("got ID %v, want %v", got.ID, prep.ID)
		}
		if len(got.Ingredients) != 1 {
			t.Errorf("expected 1 ingredient via RLS-extended SELECT, got %d", len(got.Ingredients))
		}
	})

	t.Run("service account can UPDATE preparations", func(t *testing.T) {
		q := f.newSvcTx(t)
		updated := prepParams
		updated.Name = "Overnight Oats"
		if _, err := s.Update(f.svcCtx, q, f.oID, prep.ID, updated); err != nil {
			t.Fatalf("expected UPDATE to succeed, got: %v", err)
		}
	})

	t.Run("service account can UPDATE preparation_ingredients", func(t *testing.T) {
		q := f.newSvcTx(t)
		if _, err := s.UpdateIngredient(f.svcCtx, q, f.oID, prep.ID, prepIngID, domain.PreparationIngredientParams{
			IngredientID: ing.ID, Name: "Rolled Oats", Amount: 80, Unit: "g",
		}); err != nil {
			t.Fatalf("expected preparation_ingredient UPDATE to succeed, got: %v", err)
		}
	})

	t.Run("service account cannot INSERT preparations", func(t *testing.T) {
		q := f.newSvcTx(t)
		if _, err := s.Create(f.svcCtx, q, domain.PreparationParams{
			Name: "Smoothie", YieldAmount: 1, YieldUnit: "serving", Steps: []domain.PreparationStep{},
		}); err == nil {
			t.Error("expected INSERT to be blocked for service account, got nil error")
		}
	})

	t.Run("service account cannot DELETE preparations", func(t *testing.T) {
		q := f.newSvcTx(t)
		// RLS-via-invisibility: USING clause not extended, row is invisible → ErrNotFound.
		if err := s.Delete(f.svcCtx, q, f.oID, prep.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected DELETE to be invisible via RLS (ErrNotFound), got: %v", err)
		}
	})
}

// -----------------------------------------------------------------------------
// recipes (also covers recipe_preparations SELECT via Get)
// -----------------------------------------------------------------------------

func TestRecipeStore_ServiceAccountRLS(t *testing.T) {
	f := newSvcRLSFixture(t)
	s := NewRecipeStore()

	recipeParams := domain.RecipeParams{Name: "Oatmeal Bowl", Servings: 1}

	var recipe *domain.Recipe
	f.commitUserTx(t, func(q Querier) {
		var err error
		recipe, err = s.Create(f.userCtx, q, recipeParams)
		if err != nil {
			t.Fatalf("seed recipe: %v", err)
		}
	})

	t.Run("service account can SELECT recipes", func(t *testing.T) {
		q := f.newSvcTx(t)
		got, err := s.Get(f.svcCtx, q, f.oID, recipe.ID)
		if err != nil {
			t.Fatalf("expected SELECT to succeed, got: %v", err)
		}
		if got.ID != recipe.ID {
			t.Errorf("got ID %v, want %v", got.ID, recipe.ID)
		}
	})

	t.Run("service account can UPDATE recipes", func(t *testing.T) {
		q := f.newSvcTx(t)
		updated := recipeParams
		updated.Name = "Overnight Oats Bowl"
		if _, err := s.Update(f.svcCtx, q, f.oID, recipe.ID, updated); err != nil {
			t.Fatalf("expected UPDATE to succeed, got: %v", err)
		}
	})

	t.Run("service account cannot INSERT recipes", func(t *testing.T) {
		q := f.newSvcTx(t)
		if _, err := s.Create(f.svcCtx, q, domain.RecipeParams{Name: "Smoothie", Servings: 1}); err == nil {
			t.Error("expected INSERT to be blocked for service account, got nil error")
		}
	})

	t.Run("service account cannot DELETE recipes", func(t *testing.T) {
		q := f.newSvcTx(t)
		// RLS-via-invisibility: USING clause not extended, row is invisible → ErrNotFound.
		if err := s.Delete(f.svcCtx, q, f.oID, recipe.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected DELETE to be invisible via RLS (ErrNotFound), got: %v", err)
		}
	})
}

// -----------------------------------------------------------------------------
// program_versions (append-only; SELECT only extended)
// -----------------------------------------------------------------------------

func TestProgramVersionStore_ServiceAccountRLS(t *testing.T) {
	f := newSvcRLSFixture(t)
	s := NewProgramStore()

	// Seed a program and update it to trigger the history trigger (creating a version).
	var p *domain.ProgramLite
	f.commitUserTx(t, func(q Querier) {
		var err error
		p, err = s.Create(f.userCtx, q, f.oID, "Strength", nil, nil, false)
		if err != nil {
			t.Fatalf("seed program: %v", err)
		}
	})
	f.commitUserTx(t, func(q Querier) {
		if _, err := s.Update(f.userCtx, q, f.oID, p.ID, "Strength v2", nil, nil, false, nil); err != nil {
			t.Fatalf("update program to create version: %v", err)
		}
	})

	// Fetch the version ID via the regular user so we can look it up as service account.
	var versionID domain.ProgramVersionID
	f.commitUserTx(t, func(q Querier) {
		versions, err := s.ListVersions(f.userCtx, q, f.oID, p.ID)
		if err != nil || len(versions) == 0 {
			t.Fatalf("list versions: err=%v, count=%d", err, len(versions))
		}
		versionID = versions[0].ID
	})

	t.Run("service account can SELECT program_versions", func(t *testing.T) {
		q := f.newSvcTx(t)
		got, err := s.GetVersion(f.svcCtx, q, f.oID, versionID)
		if err != nil {
			t.Fatalf("expected SELECT to succeed, got: %v", err)
		}
		if got.ID != versionID {
			t.Errorf("got ID %v, want %v", got.ID, versionID)
		}
	})

	// No UPDATE policy exists on program_versions (append-only table).
	// No INSERT test: program_versions has no direct INSERT path; rows are written
	// by the program_history_trigger during programs UPDATE, which is not a service
	// account use case.
	// No DELETE test: program_versions has no DELETE policy.
}
