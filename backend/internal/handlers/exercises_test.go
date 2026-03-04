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

	"github.com/gosusnp/cove/backend/internal/domain"
)

func TestExerciseHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)

		r := httptest.NewRequest(http.MethodGet, "/exercises", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.ExerciseLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns exercises", func(t *testing.T) {
		app := NewTestApp(t)
		app.SeedExercise("Push-up", nil)
		app.SeedExercise("Air Squat", nil)

		r := httptest.NewRequest(http.MethodGet, "/exercises", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.ExerciseLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d exercises, want 2", len(got))
		}
	})
}

func TestExerciseHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		prog := "Add 1 rep"
		e := app.SeedExercise("Push-up", &prog)

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/exercises/%d", e.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got domain.Exercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Push-up" {
			t.Errorf("got name %q, want %q", got.Name, "Push-up")
		}
		if got.Progression == nil || *got.Progression != prog {
			t.Errorf("got progression %v, want %q", got.Progression, prog)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/exercises/999", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/exercises/abc", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestExerciseHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		body := `{"name":"New Exercise", "progression":"Test"}`
		r := httptest.NewRequest(http.MethodPost, "/exercises", strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.Exercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "New Exercise" {
			t.Errorf("got name %q, want %q", got.Name, "New Exercise")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/exercises", strings.NewReader(`{}`))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestExerciseHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		e := app.SeedExercise("Old Name", nil)

		body := `{"name":"New Name", "progression":"Updated"}`
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/exercises/%d", e.ID), strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		ex, _ := app.Exercises.Get(e.ID)
		if ex.Name != "New Name" {
			t.Errorf("got name %q, want %q", ex.Name, "New Name")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPut, "/exercises/999", strings.NewReader(`{"name":"test"}`))
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestExerciseHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		e := app.SeedExercise("To Delete", nil)

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/exercises/%d", e.ID), nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		_, err := app.Exercises.Get(e.ID)
		if err == nil {
			t.Error("expected error getting deleted exercise")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/exercises/999", nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
