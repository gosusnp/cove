// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/store"
)

func TestProgramHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)

		r := httptest.NewRequest(http.MethodGet, "/programs", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.Program
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns programs", func(t *testing.T) {
		app := NewTestApp(t)
		app.SeedProgram("Strength")
		app.SeedProgram("Hypertrophy")

		r := httptest.NewRequest(http.MethodGet, "/programs", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.Program
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d programs, want 2", len(got))
		}
	})
}

func TestProgramHandler_Get(t *testing.T) {
	t.Run("found returns full hierarchy", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Strength")
		ps := app.SeedProgramSet(p.ID, 3)
		e := app.SeedExercise("Pull-up", nil)
		app.SeedProgramExercise(ps.ID, e.ID)

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d", p.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.ProgramDetail
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got name %q, want %q", got.Name, "Strength")
		}
		if len(got.Sets) != 1 {
			t.Errorf("got %d sets, want 1", len(got.Sets))
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/programs/999", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/programs/abc", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		body := `{"name":"New Program"}`
		r := httptest.NewRequest(http.MethodPost, "/programs", strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got store.Program
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "New Program" {
			t.Errorf("got name %q, want %q", got.Name, "New Program")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/programs", strings.NewReader(`{}`))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Old Name")

		body := `{"name":"New Name"}`
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/programs/%d", p.ID), strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		ex, _ := app.Programs.GetDetail(p.ID)
		if ex.Name != "New Name" {
			t.Errorf("got name %q, want %q", ex.Name, "New Name")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPut, "/programs/999", strings.NewReader(`{"name":"test"}`))
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("To Delete")

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/programs/%d", p.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		_, err := app.Programs.GetDetail(p.ID)
		if err == nil {
			t.Error("expected error getting deleted program")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/programs/999", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
