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

func defaultPrepParams() domain.PreparationParams {
	return domain.PreparationParams{
		Name:        "Basic Vinaigrette",
		YieldAmount: 200,
		YieldUnit:   "ml",
		Steps:       []domain.PreparationStep{},
	}
}

func TestPreparationHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/preparations", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.PreparationLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("got nil, want empty slice")
		}
	})

	t.Run("returns seeded preparations", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		p2 := defaultPrepParams()
		p2.Name = "Beurre Blanc"
		app.SeedPreparationForUser(context.Background(), p2, u1, o1)

		r := app.AuthRequest(http.MethodGet, "/api/preparations", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.PreparationLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d preparations, want 2", len(got))
		}
	})

	t.Run("RLS: own and public visible, other org private not", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, o2 := app.SeedUserWithOrg("u2@test.com", "sub2")

		pub := defaultPrepParams()
		pub.Name = "Public Sauce"
		pub.IsPublic = true
		id1 := &domain.Identity{UserID: u1, OrgID: o1}
		ctx1 := domain.NewContext(context.Background(), id1)
		if _, err := app.Preparations.Create(ctx1, pub); err != nil {
			t.Fatal(err)
		}

		priv1 := defaultPrepParams()
		priv1.Name = "U1 Private Prep"
		if _, err := app.Preparations.Create(ctx1, priv1); err != nil {
			t.Fatal(err)
		}

		priv2 := defaultPrepParams()
		priv2.Name = "U2 Private Prep"
		id2 := &domain.Identity{UserID: u2, OrgID: o2}
		ctx2 := domain.NewContext(context.Background(), id2)
		if _, err := app.Preparations.Create(ctx2, priv2); err != nil {
			t.Fatal(err)
		}

		// U1 sees public + own private = 2
		r1 := app.AuthRequest(http.MethodGet, "/api/preparations", nil, u1)
		w1 := app.Do(r1)
		var got1 []domain.PreparationLite
		if err := json.NewDecoder(w1.Body).Decode(&got1); err != nil {
			t.Fatalf("decode U1: %v", err)
		}
		if len(got1) != 2 {
			t.Errorf("U1 should see 2, got %d", len(got1))
		}

		// U2 sees public + own private = 2
		r2 := app.AuthRequest(http.MethodGet, "/api/preparations", nil, u2)
		w2 := app.Do(r2)
		var got2 []domain.PreparationLite
		if err := json.NewDecoder(w2.Body).Decode(&got2); err != nil {
			t.Fatalf("decode U2: %v", err)
		}
		if len(got2) != 2 {
			t.Errorf("U2 should see 2, got %d", len(got2))
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/preparations", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreparationHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/preparations/%d", prep.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d", w.Code, http.StatusOK)
		}
		var got domain.Preparation
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Basic Vinaigrette" {
			t.Errorf("got name %q, want %q", got.Name, "Basic Vinaigrette")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/preparations/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/preparations/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot get U1's private preparation", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/preparations/%d", prep.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/preparations/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreparationHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		body := `{"name":"Vinaigrette","yield_amount":200,"yield_unit":"ml","steps":[]}`
		r := app.AuthRequest(http.MethodPost, "/api/preparations", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusCreated, w.Body)
		}
		var got domain.Preparation
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Vinaigrette" {
			t.Errorf("got name %q, want %q", got.Name, "Vinaigrette")
		}
	})

	t.Run("name normalization", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		body := `{"name":"  Hollandaise  ","yield_amount":100,"yield_unit":"ml","steps":[]}`
		r := app.AuthRequest(http.MethodPost, "/api/preparations", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.Preparation
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Hollandaise" {
			t.Errorf("got name %q, want %q after normalization", got.Name, "Hollandaise")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/preparations", strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/preparations", strings.NewReader(`{"yield_amount":100,"yield_unit":"ml"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("yield_amount zero", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/preparations", strings.NewReader(`{"name":"Sauce","yield_amount":0,"yield_unit":"ml"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing yield_unit", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/preparations", strings.NewReader(`{"name":"Sauce","yield_amount":100}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/preparations", strings.NewReader(`{"name":"Sauce","yield_amount":100,"yield_unit":"ml"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreparationHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		body := `{"name":"Updated Vinaigrette","yield_amount":300,"yield_unit":"ml","steps":[]}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/preparations/%d", prep.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusOK, w.Body)
		}
		var got domain.Preparation
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Updated Vinaigrette" {
			t.Errorf("got name %q, want %q", got.Name, "Updated Vinaigrette")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/preparations/%d", prep.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPut, "/api/preparations/999", strings.NewReader(`{"name":"Ghost","yield_amount":1,"yield_unit":"ml","steps":[]}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPut, "/api/preparations/abc", strings.NewReader(`{"name":"X","yield_amount":1,"yield_unit":"ml","steps":[]}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot update U1's preparation", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		body := `{"name":"Hacked","yield_amount":1,"yield_unit":"ml","steps":[]}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/preparations/%d", prep.ID), strings.NewReader(body), u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPut, "/api/preparations/1", strings.NewReader(`{"name":"X","yield_amount":1,"yield_unit":"ml","steps":[]}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreparationHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/preparations/%d", prep.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		if _, err := app.Preparations.Get(ctx, prep.ID); err == nil {
			t.Error("expected error getting deleted preparation")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/preparations/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/preparations/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cross-org: U2 cannot delete U1's preparation", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/preparations/%d", prep.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/api/preparations/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreparationHandler_AddIngredient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		body := fmt.Sprintf(`{"ingredient_id":%d,"name":"Flour","amount":100,"unit":"g"}`, ing.ID)
		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/preparations/%d/ingredients", prep.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusCreated, w.Body)
		}
		var got domain.PreparationIngredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Amount != 100 {
			t.Errorf("got amount %v, want 100", got.Amount)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)

		r := app.AuthRequest(http.MethodPost, fmt.Sprintf("/api/preparations/%d/ingredients", prep.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("preparation not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)

		body := fmt.Sprintf(`{"ingredient_id":%d,"name":"Flour","amount":100,"unit":"g"}`, ing.ID)
		r := app.AuthRequest(http.MethodPost, "/api/preparations/999/ingredients", strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid prep id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/preparations/abc/ingredients", strings.NewReader(`{"ingredient_id":1,"name":"Flour","amount":100,"unit":"g"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/preparations/1/ingredients", strings.NewReader(`{"ingredient_id":1,"name":"Flour","amount":100,"unit":"g"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreparationHandler_UpdateIngredient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)
		prepIng := app.SeedPreparationIngredientForUser(context.Background(), prep.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID,
			Name:         "Flour",
			Amount:       100,
			Unit:         "g",
		}, u1, o1)

		body := fmt.Sprintf(`{"ingredient_id":%d,"name":"Bread Flour","amount":200,"unit":"g"}`, ing.ID)
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/preparations/%d/ingredients/%d", prep.ID, prepIng.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusOK, w.Body)
		}
		var got domain.PreparationIngredient
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Bread Flour" {
			t.Errorf("got name %q, want %q", got.Name, "Bread Flour")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)
		prepIng := app.SeedPreparationIngredientForUser(context.Background(), prep.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID, Name: "Flour", Amount: 100, Unit: "g",
		}, u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/preparations/%d/ingredients/%d", prep.ID, prepIng.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("wrong preparation returns 404", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		p1 := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		p2params := defaultPrepParams()
		p2params.Name = "Other Prep"
		p2 := app.SeedPreparationForUser(context.Background(), p2params, u1, o1)
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)
		// Ingredient belongs to p2
		prepIng := app.SeedPreparationIngredientForUser(context.Background(), p2.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID, Name: "Flour", Amount: 100, Unit: "g",
		}, u1, o1)

		// Try to update it via p1
		body := fmt.Sprintf(`{"ingredient_id":%d,"name":"Flour","amount":200,"unit":"g"}`, ing.ID)
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/preparations/%d/ingredients/%d", p1.ID, prepIng.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid prep id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPut, "/api/preparations/abc/ingredients/1", strings.NewReader(`{"ingredient_id":1,"name":"Flour","amount":100,"unit":"g"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid ingredient id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/preparations/%d/ingredients/abc", prep.ID), strings.NewReader(`{"ingredient_id":1,"name":"Flour","amount":100,"unit":"g"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPut, "/api/preparations/1/ingredients/1", strings.NewReader(`{"ingredient_id":1,"name":"Flour","amount":100,"unit":"g"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestPreparationHandler_DeleteIngredient(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)
		prepIng := app.SeedPreparationIngredientForUser(context.Background(), prep.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID, Name: "Flour", Amount: 100, Unit: "g",
		}, u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/preparations/%d/ingredients/%d", prep.ID, prepIng.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got %d, want %d: %s", w.Code, http.StatusNoContent, w.Body)
		}

		// Verify ingredient is gone from preparation
		id := &domain.Identity{UserID: u1, OrgID: o1}
		ctx := domain.NewContext(context.Background(), id)
		got, err := app.Preparations.Get(ctx, prep.ID)
		if err != nil {
			t.Fatalf("get prep: %v", err)
		}
		if len(got.Ingredients) != 0 {
			t.Errorf("expected 0 ingredients, got %d", len(got.Ingredients))
		}
	})

	t.Run("wrong preparation returns 404", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		p1 := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		p2params := defaultPrepParams()
		p2params.Name = "Other Prep"
		p2 := app.SeedPreparationForUser(context.Background(), p2params, u1, o1)
		ing := app.SeedIngredientForUser(context.Background(), defaultIngredientParams(), u1, o1)
		prepIng := app.SeedPreparationIngredientForUser(context.Background(), p2.ID, domain.PreparationIngredientParams{
			IngredientID: ing.ID, Name: "Flour", Amount: 100, Unit: "g",
		}, u1, o1)

		// Try to delete via p1
		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/preparations/%d/ingredients/%d", p1.ID, prepIng.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid prep id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/preparations/abc/ingredients/1", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid ingredient id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prep := app.SeedPreparationForUser(context.Background(), defaultPrepParams(), u1, o1)
		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/preparations/%d/ingredients/abc", prep.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/api/preparations/1/ingredients/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
