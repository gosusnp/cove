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
		app.SeedProgramExercise(p.ID, ps.ID, e.ID)
		app.SeedProgramExercise(p.ID, ps.ID, e.ID)

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
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)

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

	t.Run("weight without unit returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Squat", nil)

		body := fmt.Sprintf(`{"exercise_id": %d, "weight": 100}`, e.ID)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("weight with non-mass unit returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Squat", nil)

		body := fmt.Sprintf(`{"exercise_id": %d, "weight": 100, "weight_unit": "ml"}`, e.ID)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
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
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		full, _ := app.Programs.Get(ctx, p.ID)

		body := mustJSON(t, map[string]any{"exercise_id": int64(e.ID), "laterality": "updated notes", "updated_at": full.UpdatedAt})
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodPut, url, body)
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		got, _ := app.Programs.GetExercise(ctx, p.ID, ps.ID, pe.ID)
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
		body := mustJSON(t, map[string]any{"exercise_id": int64(e.ID), "updated_at": nil})
		r := httptest.NewRequest(http.MethodPut, url, body)
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("weight without unit returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Squat", nil)
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		full, _ := app.Programs.Get(ctx, p.ID)

		body := mustJSON(t, map[string]any{"exercise_id": int64(e.ID), "weight": 100, "updated_at": full.UpdatedAt})
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodPut, url, body)
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("weight with non-mass unit returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Squat", nil)
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		full, _ := app.Programs.Get(ctx, p.ID)

		body := mustJSON(t, map[string]any{"exercise_id": int64(e.ID), "weight": 100, "weight_unit": "ml", "updated_at": full.UpdatedAt})
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodPut, url, body)
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramExerciseHandler_Patch(t *testing.T) {
	t.Run("patches only provided fields", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)

		body := `{"reps": 12}`
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodPatch, url, strings.NewReader(body))
		w := app.DoRaw(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		got, _ := app.Programs.GetExercise(ctx, p.ID, ps.ID, pe.ID)
		if got.TargetReps == nil || *got.TargetReps != 12 {
			t.Errorf("got reps %v, want 12", got.TargetReps)
		}
		// exercise_id should be unchanged
		if got.ExerciseID != e.ID {
			t.Errorf("exercise_id changed unexpectedly")
		}
	})

	t.Run("weight without unit returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Squat", nil)
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)

		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodPatch, url, strings.NewReader(`{"weight": 80}`))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/999", p.ID, ps.ID)
		r := httptest.NewRequest(http.MethodPatch, url, strings.NewReader(`{"reps": 5}`))
		w := app.DoRaw(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		p := app.SeedProgram("Test")
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodPatch, url, strings.NewReader(`not json`))
		w := app.DoRaw(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cannot patch another org's exercise", func(t *testing.T) {
		app := NewTestApp(t)
		ownerID, ownerOrgID := app.SeedUserWithOrg("owner@test.com", "owner-sub")
		p := app.SeedProgramForUser(context.Background(), "Private Program", ownerID, ownerOrgID)
		ps := app.SeedProgramSet(p.ID, 1)
		e := app.SeedExercise("Pull-up", nil)
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)

		attackerID, _ := app.SeedUserWithOrg("attacker@test.com", "attacker-sub")
		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := app.AuthRequest(http.MethodPatch, url, strings.NewReader(`{"reps": 1}`), attackerID)
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
		pe := app.SeedProgramExercise(p.ID, ps.ID, e.ID)

		url := fmt.Sprintf("/programs/%d/sets/%d/exercises/%d", p.ID, ps.ID, pe.ID)
		r := httptest.NewRequest(http.MethodDelete, url, nil)
		w := app.DoRaw(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		ctx := domain.NewContext(context.Background(), app.programOwners[p.ID])
		_, err := app.Programs.GetExercise(ctx, p.ID, ps.ID, pe.ID)
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
