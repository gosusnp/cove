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
	"github.com/gosusnp/cove/backend/internal/store"
)

func TestProgramSetHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets", p.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns sets for program", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		app.SeedProgramSet(p.ID, 1)
		app.SeedProgramSet(p.ID, 1)

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets", p.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d sets, want 2", len(got))
		}
	})
}

func TestProgramSetHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		created := app.SeedProgramSet(p.ID, 3)

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets/%d", p.ID, created.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Rounds != 3 {
			t.Errorf("got rounds %d, want 3", got.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets/999", p.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets/abc", p.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramSetHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")

		body := `{"rounds":3}`
		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/programs/%d/sets", p.ID), strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/programs/%d/sets", p.ID), strings.NewReader(`not json`))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramSetHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		ps := app.SeedProgramSet(p.ID, 1)
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		full, _ := app.Programs.Get(ctx, p.ID)

		body := mustJSON(t, map[string]any{"rounds": 5, "updated_at": full.UpdatedAt})
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/programs/%d/sets/%d", p.ID, ps.ID), body)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		got, _ := app.Programs.GetSet(ctx, p.ID, ps.ID)
		if got.Rounds != 5 {
			t.Errorf("got rounds %d, want 5", got.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		body := mustJSON(t, map[string]any{"rounds": 1, "updated_at": nil})
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/programs/%d/sets/999", p.ID), body)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramSetHandler_Patch(t *testing.T) {
	t.Run("patches only provided fields", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		ps := app.SeedProgramSet(p.ID, 3)

		body := `{"rounds": 5}`
		r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/programs/%d/sets/%d", p.ID, ps.ID), strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		got, _ := app.Programs.GetSet(ctx, p.ID, ps.ID)
		if got.Rounds != 5 {
			t.Errorf("got rounds %d, want 5", got.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/programs/%d/sets/999", p.ID), strings.NewReader(`{"rounds": 1}`))
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		ps := app.SeedProgramSet(p.ID, 1)
		r := httptest.NewRequest(http.MethodPatch, fmt.Sprintf("/programs/%d/sets/%d", p.ID, ps.ID), strings.NewReader(`not json`))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cannot patch another org's set", func(t *testing.T) {
		app := NewTestApp(t)
		ownerID, ownerOrgID := app.SeedUserWithOrg("owner@test.com", "owner-sub")
		p := app.SeedProgramForUser(context.Background(), "Private Program", ownerID, ownerOrgID)
		ps := app.SeedProgramSet(p.ID, 3)

		attackerID, _ := app.SeedUserWithOrg("attacker@test.com", "attacker-sub")
		r := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/programs/%d/sets/%d", p.ID, ps.ID), strings.NewReader(`{"rounds": 1}`), attackerID)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramSetHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		ps := app.SeedProgramSet(p.ID, 1)

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/programs/%d/sets/%d", p.ID, ps.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		_, err := app.Programs.GetSet(ctx, p.ID, ps.ID)
		if err == nil {
			t.Error("expected error getting deleted set")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test Program")
		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/programs/%d/sets/999", p.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
