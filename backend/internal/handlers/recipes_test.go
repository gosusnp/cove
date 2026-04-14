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
)

func defaultRecipeParams() domain.RecipeParams {
	return domain.RecipeParams{
		Name:     "Sourdough Bread",
		Servings: 4,
	}
}

func TestRecipeHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/recipes", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.RecipeLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("got nil, want empty slice")
		}
	})

	t.Run("returns seeded recipes", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		p2 := defaultRecipeParams()
		p2.Name = "Focaccia"
		app.SeedRecipeForUser(context.Background(), p2, u1, o1)

		r := app.AuthRequest(http.MethodGet, "/api/recipes", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.RecipeLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d recipes, want 2", len(got))
		}
	})

	t.Run("RLS: own and public visible, other org private not", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, o2 := app.SeedUserWithOrg("u2@test.com", "sub2")

		pub := defaultRecipeParams()
		pub.Name = "Public Bread"
		pub.IsPublic = true
		app.SeedRecipeForUser(context.Background(), pub, u1, o1)

		priv1 := defaultRecipeParams()
		priv1.Name = "U1 Private Recipe"
		app.SeedRecipeForUser(context.Background(), priv1, u1, o1)

		priv2 := defaultRecipeParams()
		priv2.Name = "U2 Private Recipe"
		app.SeedRecipeForUser(context.Background(), priv2, u2, o2)

		// U1 sees public + own private = 2
		r1 := app.AuthRequest(http.MethodGet, "/api/recipes", nil, u1)
		w1 := app.Do(r1)
		var got1 []domain.RecipeLite
		if err := json.NewDecoder(w1.Body).Decode(&got1); err != nil {
			t.Fatalf("decode U1: %v", err)
		}
		if len(got1) != 2 {
			t.Errorf("U1 should see 2, got %d", len(got1))
		}

		// U2 sees public + own private = 2
		r2 := app.AuthRequest(http.MethodGet, "/api/recipes", nil, u2)
		w2 := app.Do(r2)
		var got2 []domain.RecipeLite
		if err := json.NewDecoder(w2.Body).Decode(&got2); err != nil {
			t.Fatalf("decode U2: %v", err)
		}
		if len(got2) != 2 {
			t.Errorf("U2 should see 2, got %d", len(got2))
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/recipes", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestRecipeHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/recipes/%d", recipe.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var got domain.Recipe
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Sourdough Bread" {
			t.Errorf("got name %q, want %q", got.Name, "Sourdough Bread")
		}
		if got.Preparations == nil {
			t.Error("preparations should be empty slice, not nil")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/recipes/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/recipes/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 can get U1's public recipe", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		pub := defaultRecipeParams()
		pub.IsPublic = true
		recipe := app.SeedRecipeForUser(context.Background(), pub, u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/recipes/%d", recipe.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("cross-org: U2 cannot get U1's private recipe", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/recipes/%d", recipe.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/recipes/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestRecipeHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		body := `{"name":"Ciabatta","servings":2}`
		r := app.AuthRequest(http.MethodPost, "/api/recipes", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusCreated, w.Body)
		}
		var got domain.Recipe
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Ciabatta" {
			t.Errorf("got name %q, want %q", got.Name, "Ciabatta")
		}
	})

	t.Run("name normalization", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		body := `{"name":"  Baguette  ","servings":2}`
		r := app.AuthRequest(http.MethodPost, "/api/recipes", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.Recipe
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Baguette" {
			t.Errorf("got name %q, want trimmed %q", got.Name, "Baguette")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/recipes", strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/recipes", strings.NewReader(`{"servings":2}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("zero servings", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/recipes", strings.NewReader(`{"name":"Bread","servings":0}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/recipes", strings.NewReader(`{"name":"Bread","servings":2}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestRecipeHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		body := `{"name":"Updated Bread","servings":6}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d", recipe.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusOK, w.Body)
		}
		var got domain.Recipe
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Updated Bread" {
			t.Errorf("got name %q, want %q", got.Name, "Updated Bread")
		}
		if got.Servings != 6 {
			t.Errorf("got servings %d, want 6", got.Servings)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d", recipe.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPut, "/api/recipes/999", strings.NewReader(`{"name":"Ghost","servings":2}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPut, "/api/recipes/abc", strings.NewReader(`{"name":"X","servings":2}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot update U1's recipe", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d", recipe.ID), strings.NewReader(`{"name":"Hacked","servings":2}`), u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPut, "/api/recipes/1", strings.NewReader(`{"name":"X","servings":2}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestRecipeHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/recipes/%d", recipe.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		if _, err := app.Recipes.Get(ctx, recipe.ID); err == nil {
			t.Error("expected error getting deleted recipe")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/recipes/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/recipes/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot delete U1's recipe", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/recipes/%d", recipe.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/api/recipes/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestRecipeHandler_AddPreparation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		body := fmt.Sprintf(`{"preparation_id":%d,"position":1,"amount":500,"unit":"g"}`, prep.ID)
		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/recipes/%d/preparations", recipe.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusCreated, w.Body)
		}
		var got domain.RecipePreparation
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Position != 1 {
			t.Errorf("got position %d, want 1", got.Position)
		}
		if got.Amount != 500 {
			t.Errorf("got amount %v, want 500", got.Amount)
		}
		if got.Unit != "g" {
			t.Errorf("got unit %q, want %q", got.Unit, "g")
		}
	})

	t.Run("unknown unit returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		body := fmt.Sprintf(`{"preparation_id":%d,"position":1,"amount":1,"unit":"serving"}`, prep.ID)
		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/recipes/%d/preparations", recipe.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusBadRequest, w.Body)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/recipes/%d/preparations", recipe.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("recipe not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		body := fmt.Sprintf(`{"preparation_id":%d,"position":1,"amount":500,"unit":"g"}`, prep.ID)
		r := app.AuthRequest(http.MethodPost, "/api/recipes/999/preparations", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid recipe id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/recipes/abc/preparations", strings.NewReader(`{"preparation_id":1,"position":1,"amount":500,"unit":"g"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot add preparation to U1's recipe", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, o2 := app.SeedUserWithOrg("u2@test.com", "sub2")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u2, o2)

		body := fmt.Sprintf(`{"preparation_id":%d,"position":1,"amount":500,"unit":"g"}`, prep.ID)
		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/recipes/%d/preparations", recipe.ID), strings.NewReader(body), u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/recipes/1/preparations", strings.NewReader(`{"preparation_id":1,"position":1,"amount":500,"unit":"g"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestRecipeHandler_UpdatePreparation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		link, err := app.Recipes.AddPreparation(ctx, recipe.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("seed link: %v", err)
		}

		body := `{"position":2,"amount":250,"unit":"ml"}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d/preparations/%d", recipe.ID, link.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusOK, w.Body)
		}
		var got domain.RecipePreparation
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Position != 2 {
			t.Errorf("got position %d, want 2", got.Position)
		}
		if got.Unit != "ml" {
			t.Errorf("got unit %q, want %q", got.Unit, "ml")
		}
	})

	t.Run("unknown unit returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		link, err := app.Recipes.AddPreparation(ctx, recipe.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("seed link: %v", err)
		}

		body := `{"position":1,"amount":1,"unit":"serving"}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d/preparations/%d", recipe.ID, link.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusBadRequest, w.Body)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		link, err := app.Recipes.AddPreparation(ctx, recipe.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("seed link: %v", err)
		}

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d/preparations/%d", recipe.ID, link.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d/preparations/999", recipe.ID), strings.NewReader(`{"position":1,"amount":500,"unit":"g"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid link id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d/preparations/abc", recipe.ID), strings.NewReader(`{"position":1,"amount":500,"unit":"g"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot update U1's recipe preparation", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		link, err := app.Recipes.AddPreparation(ctx, recipe.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("seed link: %v", err)
		}

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/recipes/%d/preparations/%d", recipe.ID, link.ID), strings.NewReader(`{"position":1,"amount":250,"unit":"g"}`), u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPut, "/api/recipes/1/preparations/1", strings.NewReader(`{"position":1,"amount":500,"unit":"g"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestRecipeHandler_DeletePreparation(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		link, err := app.Recipes.AddPreparation(ctx, recipe.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("seed link: %v", err)
		}

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/recipes/%d/preparations/%d", recipe.ID, link.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusNoContent, w.Body)
		}

		// Verify link is gone
		got, err := app.Recipes.Get(ctx, recipe.ID)
		if err != nil {
			t.Fatalf("get recipe: %v", err)
		}
		if len(got.Preparations) != 0 {
			t.Errorf("expected 0 preparations, got %d", len(got.Preparations))
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/recipes/%d/preparations/999", recipe.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid link id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/recipes/%d/preparations/abc", recipe.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot delete U1's recipe preparation", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		recipe := app.SeedRecipeForUser(context.Background(), defaultRecipeParams(), u1, o1)
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		link, err := app.Recipes.AddPreparation(ctx, recipe.ID, domain.RecipePreparationParams{
			PreparationID: prep.ID,
			Position:      1,
			Amount:        500,
			Unit:          "g",
		})
		if err != nil {
			t.Fatalf("seed link: %v", err)
		}

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/recipes/%d/preparations/%d", recipe.ID, link.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/api/recipes/1/preparations/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
