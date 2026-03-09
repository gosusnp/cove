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

func TestProgramExerciseHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)

		url := fmt.Sprintf("/programs/%d/sets/%d/exercises", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns exercises for set", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		app.SeedProgramExercise(ps.ID, e.ID)
		app.SeedProgramExercise(ps.ID, e.ID)

		url := fmt.Sprintf("/programs/%d/sets/%d/exercises", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d exercises, want 2", len(got))
		}
	})
}

func TestProgramExerciseHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		pe := app.SeedProgramExercise(ps.ID, e.ID)

		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.ProgramExercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ExerciseID != e.ID {
			t.Errorf("got exercise ID %d, want %d", got.ExerciseID, e.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/999", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodGet, url, nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramExerciseHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)

		body := fmt.Sprintf(`{"exercise_id": %d}`, e.ID)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(`not json`))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramExerciseHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		pe := app.SeedProgramExercise(ps.ID, e.ID)

		body := fmt.Sprintf(`{"exercise_id": %d, "laterality": "updated notes"}`, e.ID)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		got, _ := app.ProgramExercises.Get(ctx, ps.ID, pe.ID)
		if got.Laterality == nil || *got.Laterality != "updated notes" {
			t.Errorf("got laterality %v, want 'updated notes'", got.Laterality)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/999", p.ID, ps.ID)
		body := fmt.Sprintf(`{"exercise_id": %d}`, e.ID)
		r := httptest.NewRequest(http.MethodPut, url, strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramExerciseHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		pe := app.SeedProgramExercise(ps.ID, e.ID)

		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		_, err := app.ProgramExercises.Get(ctx, ps.ID, pe.ID)
		if err == nil {
			t.Error("expected error getting deleted program exercise")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/999", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
