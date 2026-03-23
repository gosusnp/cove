// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/fdc"
)

// defaultIngredientParams returns a valid IngredientParams for use in tests.
func defaultIngredientParams() domain.IngredientParams {
	return domain.IngredientParams{
		Name:            "Chicken Breast",
		CaloriesPer100g: 165,
		ProteinPer100g:  31,
		FatPer100g:      3.6,
		CarbsPer100g:    0,
		IsPublic:        false,
	}
}

func TestIngredientHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/ingredients", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.Ingredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("got nil, want empty slice")
		}
	})

	t.Run("returns seeded ingredients", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		p1 := defaultIngredientParams()
		p1.Name = "Chicken Breast"
		p2 := defaultIngredientParams()
		p2.Name = "Oats"
		app.SeedIngredientForUser(context.Background(), p1, u1, o1)
		app.SeedIngredientForUser(context.Background(), p2, u1, o1)

		r := app.AuthRequest(http.MethodGet, "/api/ingredients", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.Ingredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d ingredients, want 2", len(got))
		}
	})

	t.Run("RLS: list only returns own or public", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, o2 := app.SeedUserWithOrg("u2@test.com", "sub2")

		// Public ingredient owned by U1
		id1 := &domain.Identity{UserID: u1, OrgID: o1}
		ctx1 := domain.NewContext(context.Background(), id1)
		pubParams := defaultIngredientParams()
		pubParams.Name = "Public Rice"
		pubParams.IsPublic = true
		if _, err := app.Ingredients.Create(ctx1, pubParams); err != nil {
			t.Fatal(err)
		}

		// Private ingredient for U1
		privParams1 := defaultIngredientParams()
		privParams1.Name = "U1 Private Oats"
		if _, err := app.Ingredients.Create(ctx1, privParams1); err != nil {
			t.Fatal(err)
		}

		// Private ingredient for U2
		id2 := &domain.Identity{UserID: u2, OrgID: o2}
		ctx2 := domain.NewContext(context.Background(), id2)
		privParams2 := defaultIngredientParams()
		privParams2.Name = "U2 Private Beef"
		if _, err := app.Ingredients.Create(ctx2, privParams2); err != nil {
			t.Fatal(err)
		}

		// Request as U1 — should see public + own private = 2
		r1 := app.AuthRequest(http.MethodGet, "/api/ingredients", nil, u1)
		w1 := app.Do(r1)

		if w1.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w1.Code, http.StatusOK)
		}
		var got1 []domain.Ingredient
		if err := json.NewDecoder(w1.Body).Decode(&got1); err != nil {
			t.Fatalf("decode U1: %v", err)
		}
		if len(got1) != 2 {
			t.Errorf("U1 should see 2 ingredients (public + own), got %d: %+v", len(got1), got1)
		}

		// Request as U2 — should see public + own private = 2
		r2 := app.AuthRequest(http.MethodGet, "/api/ingredients", nil, u2)
		w2 := app.Do(r2)
		if w2.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w2.Code, http.StatusOK)
		}
		var got2 []domain.Ingredient
		if err := json.NewDecoder(w2.Body).Decode(&got2); err != nil {
			t.Fatalf("decode U2: %v", err)
		}
		if len(got2) != 2 {
			t.Errorf("U2 should see 2 ingredients (public + own), got %d: %+v", len(got2), got2)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/ingredients", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestIngredientHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/ingredients/%d", ing.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got domain.Ingredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Chicken Breast" {
			t.Errorf("got name %q, want %q", got.Name, "Chicken Breast")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/ingredients/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/ingredients/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot get U1's private ingredient", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/ingredients/%d", ing.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/ingredients/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestIngredientHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		body := `{"name":"Chicken Breast","calories_per_100g":165,"protein_per_100g":31,"fat_per_100g":3.6,"carbs_per_100g":0}`
		r := app.AuthRequest(http.MethodPost, "/api/ingredients", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.Ingredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Chicken Breast" {
			t.Errorf("got name %q, want %q", got.Name, "Chicken Breast")
		}

		// Verify bookkeeping via trigger
		if got.CreatedBy.UUID != u1.UUID {
			t.Errorf("got created_by %v, want %v", got.CreatedBy, u1)
		}
		if got.CreatedAt.IsZero() {
			t.Error("expected created_at to be populated")
		}
	})

	t.Run("name normalization", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		body := `{"name":"  Oats  ","calories_per_100g":389,"protein_per_100g":17,"fat_per_100g":7,"carbs_per_100g":66}`
		r := app.AuthRequest(http.MethodPost, "/api/ingredients", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.Ingredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Oats" {
			t.Errorf("got name %q after normalization, want %q", got.Name, "Oats")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/ingredients", strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/ingredients", strings.NewReader(`{"calories_per_100g":100}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/ingredients", strings.NewReader(`{"name":"Oats"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestIngredientHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		body := `{"name":"Updated Chicken","calories_per_100g":165,"protein_per_100g":31,"fat_per_100g":3.6,"carbs_per_100g":0}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/ingredients/%d", ing.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change via service
		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		updated, err := app.Ingredients.Get(ctx, ing.ID)
		if err != nil {
			t.Fatalf("get after update: %v", err)
		}
		if updated.Name != "Updated Chicken" {
			t.Errorf("got name %q, want %q", updated.Name, "Updated Chicken")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/ingredients/%d", ing.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := `{"name":"Ghost","calories_per_100g":0,"protein_per_100g":0,"fat_per_100g":0,"carbs_per_100g":0}`
		r := app.AuthRequest(http.MethodPut, "/api/ingredients/999", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		body := `{"name":"Ghost","calories_per_100g":0,"protein_per_100g":0,"fat_per_100g":0,"carbs_per_100g":0}`
		r := httptest.NewRequest(http.MethodPut, "/api/ingredients/1", strings.NewReader(body))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("cross-org: U2 cannot update U1's ingredient", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		body := `{"name":"Hacked","calories_per_100g":0,"protein_per_100g":0,"fat_per_100g":0,"carbs_per_100g":0}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/ingredients/%d", ing.ID), strings.NewReader(body), u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestIngredientHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/ingredients/%d", ing.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		_, err := app.Ingredients.Get(ctx, ing.ID)
		if err == nil {
			t.Error("expected error getting deleted ingredient")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/ingredients/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/api/ingredients/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("cross-org: U2 cannot delete U1's ingredient", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/ingredients/%d", ing.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestIngredientHandler_FDCSync(t *testing.T) {
	t.Run("success: updates nutrition from FDC", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"fdcId":       42,
				"description": "Chicken Breast",
				"foodNutrients": []map[string]any{
					{"nutrient": map[string]any{"id": 1008}, "amount": 165},
					{"nutrient": map[string]any{"id": 1003}, "amount": 31},
					{"nutrient": map[string]any{"id": 1004}, "amount": 3.6},
					{"nutrient": map[string]any{"id": 1005}, "amount": 0},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		fdcID := 42
		p := defaultIngredientParams()
		p.FdcID = &fdcID
		p.CaloriesPer100g = 0 // will be overwritten by sync

		app := NewTestAppWithFDC(t, fdc.NewClientWithBaseURL("testkey", srv.URL))
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ing := app.SeedIngredientForUser(context.Background(), p, u1, o1)

		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/ingredients/%d/fdc-sync", ing.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
		}
		var got domain.Ingredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.CaloriesPer100g != 165 {
			t.Errorf("got calories %v, want 165", got.CaloriesPer100g)
		}
	})

	t.Run("no fdc_id returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/ingredients/%d/fdc-sync", ing.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found returns 404", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		r := app.AuthRequest(http.MethodPost, "/api/ingredients/999/fdc-sync", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized returns 401", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/ingredients/1/fdc-sync", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("cross-org: U2 cannot sync U1's ingredient", func(t *testing.T) {
		fdcID := 1
		p := defaultIngredientParams()
		p.FdcID = &fdcID
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ing := app.SeedIngredientForUser(context.Background(), p, u1, o1)

		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/ingredients/%d/fdc-sync", ing.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
