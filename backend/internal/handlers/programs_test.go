// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
)

func TestProgramHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")

		r := app.AuthRequest(http.MethodGet, "/api/programs", nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.ProgramLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns programs", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		app.SeedProgramForUser(t.Context(), "Strength", uID, oID)
		app.SeedProgramForUser(t.Context(), "Hypertrophy", uID, oID)

		r := app.AuthRequest(http.MethodGet, "/api/programs", nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.ProgramLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d programs, want 2", len(got))
		}
	})

	t.Run("RLS: list only returns own or public", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, o2 := app.SeedUserWithOrg("u2@test.com", "sub2")

		// Public program owned by U1
		ctx1 := domain.NewContext(t.Context(), &domain.Identity{UserID: u1, OrgID: o1})
		if _, err := app.Programs.Create(ctx1, "Public Strength", nil, nil, true); err != nil {
			t.Fatal(err)
		}

		// Private program for U1
		if _, err := app.Programs.Create(ctx1, "U1 Secret Bodyweight", nil, nil, false); err != nil {
			t.Fatal(err)
		}

		// Private program for U2
		ctx2 := domain.NewContext(t.Context(), &domain.Identity{UserID: u2, OrgID: o2})
		if _, err := app.Programs.Create(ctx2, "U2 Secret Yoga", nil, nil, false); err != nil {
			t.Fatal(err)
		}

		// Request as U1
		r1 := app.AuthRequest(http.MethodGet, "/api/programs", nil, u1)
		w1 := app.Do(r1)

		if w1.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w1.Code, http.StatusOK)
		}
		var got1 []domain.ProgramLite
		if err := json.NewDecoder(w1.Body).Decode(&got1); err != nil {
			t.Fatalf("decode U1: %v", err)
		}

		if len(got1) != 2 {
			t.Errorf("U1 should see 2 programs (public + own), got %d: %+v", len(got1), got1)
		}

		// Request as U2
		r2 := app.AuthRequest(http.MethodGet, "/api/programs", nil, u2)
		w2 := app.Do(r2)
		if w2.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w2.Code, http.StatusOK)
		}
		var got2 []domain.ProgramLite
		if err := json.NewDecoder(w2.Body).Decode(&got2); err != nil {
			t.Fatalf("decode U2: %v", err)
		}

		if len(got2) != 2 {
			t.Errorf("U2 should see 2 programs (public + own), got %d: %+v", len(got2), got2)
		}
	})
}

func TestProgramHandler_Get(t *testing.T) {
	t.Run("found returns full hierarchy", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Strength", uID, oID)
		ps := app.SeedProgramSet(p.ID, 3)
		e := app.SeedExercise("Pull-up", nil)
		app.SeedProgramExercise(p.ID, ps.ID, e.ID)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/programs/%d", p.ID), nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got domain.Program
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

	t.Run("response includes structure field", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Strength", uID, oID)
		ps := app.SeedProgramSet(p.ID, 3)
		e := app.SeedExercise("Squat", nil)
		app.SeedProgramExercise(p.ID, ps.ID, e.ID)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/programs/%d", p.ID), nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		structure, ok := got["structure"].(string)
		if !ok || structure == "" {
			t.Errorf("expected non-empty structure field, got: %v", got["structure"])
		}
		if strings.Contains(structure, "Strength") {
			t.Errorf("structure should not contain program name, got: %q", structure)
		}
		if !strings.Contains(structure, "Squat") {
			t.Errorf("structure should contain exercise name, got: %q", structure)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")
		r := app.AuthRequest(http.MethodGet, "/api/programs/999", nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")
		r := app.AuthRequest(http.MethodGet, "/api/programs/abc", nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("RLS: cannot get another org's private program", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")

		p1 := app.SeedProgramForUser(t.Context(), "U1 Secret Strength", u1, o1)

		// Request as U2
		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/programs/%d", p1.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("RLS: can get public program from another org", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")

		// Create public program for U1
		ctx1 := domain.NewContext(t.Context(), &domain.Identity{UserID: u1, OrgID: o1})
		p1, err := app.Programs.Create(ctx1, "Public Strength", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		// Request as U2
		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/programs/%d", p1.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
	})
}

func TestProgramHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")
		body := `{"name":"New Program"}`
		r := app.AuthRequest(http.MethodPost, "/api/programs", strings.NewReader(body), uID)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.ProgramLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "New Program" {
			t.Errorf("got name %q, want %q", got.Name, "New Program")
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")
		r := app.AuthRequest(http.MethodPost, "/api/programs", strings.NewReader(`{}`), uID)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Old Name", uID, oID)
		ctx := domain.NewContext(t.Context(), &domain.Identity{UserID: uID, OrgID: oID})
		full, _ := app.Programs.Get(ctx, p.ID)

		body := mustJSON(t, map[string]any{"name": "New Name", "updated_at": full.UpdatedAt})
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d", p.ID), body, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		ex, _ := app.Programs.Get(ctx, p.ID)
		if ex.Name != "New Name" {
			t.Errorf("got name %q, want %q", ex.Name, "New Name")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")
		body := mustJSON(t, map[string]any{"name": "test", "updated_at": nil})
		r := app.AuthRequest(http.MethodPut, "/api/programs/999", body, uID)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("conflict when updated_at is stale", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Old Name", uID, oID)
		ctx := domain.NewContext(t.Context(), &domain.Identity{UserID: uID, OrgID: oID})
		full, _ := app.Programs.Get(ctx, p.ID)

		// First update to advance the timestamp.
		body1 := mustJSON(t, map[string]any{"name": "New Name", "updated_at": full.UpdatedAt})
		r1 := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d", p.ID), body1, uID)
		w1 := app.Do(r1)
		if w1.Code != http.StatusOK {
			t.Fatalf("setup: got status %d, want %d", w1.Code, http.StatusOK)
		}

		// Second update using the old timestamp should conflict.
		body2 := mustJSON(t, map[string]any{"name": "Another Name", "updated_at": full.UpdatedAt})
		r2 := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d", p.ID), body2, uID)
		w2 := app.Do(r2)

		if w2.Code != http.StatusConflict {
			t.Errorf("stale: got status %d, want %d", w2.Code, http.StatusConflict)
		}
	})

	t.Run("RLS: cannot update another org's program", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")

		p1 := app.SeedProgramForUser(t.Context(), "U1 Program", u1, o1)

		// Request as U2
		body := mustJSON(t, map[string]any{"name": "Hacked", "updated_at": time.Now()})
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d", p1.ID), body, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}

		// Verify not changed
		ctx1 := domain.NewContext(t.Context(), &domain.Identity{UserID: u1, OrgID: o1})
		ex, _ := app.Programs.Get(ctx1, p1.ID)
		if ex.Name != "U1 Program" {
			t.Errorf("got name %q, want %q", ex.Name, "U1 Program")
		}
	})
}

func TestProgramHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "To Delete", uID, oID)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/programs/%d", p.ID), nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		ctx := domain.NewContext(t.Context(), &domain.Identity{UserID: uID, OrgID: oID})
		_, err := app.Programs.Get(ctx, p.ID)
		if err == nil {
			t.Error("expected error getting deleted program")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")
		r := app.AuthRequest(http.MethodDelete, "/api/programs/999", nil, uID)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("RLS: cannot delete another org's program", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")

		p1 := app.SeedProgramForUser(t.Context(), "To Delete", u1, o1)

		// Request as U2
		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/programs/%d", p1.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}

		// Verify still exists
		ctx1 := domain.NewContext(t.Context(), &domain.Identity{UserID: u1, OrgID: o1})
		_, err := app.Programs.Get(ctx1, p1.ID)
		if err != nil {
			t.Errorf("expected program to still exist, got err: %v", err)
		}
	})
}

func TestProgramHandler_ReorderStructure(t *testing.T) {
	t.Run("reorder sets", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Strength", uID, oID)
		s1 := app.SeedProgramSet(p.ID, 3)
		s2 := app.SeedProgramSet(p.ID, 4)
		e := app.SeedExercise("Squat", nil)
		pe1 := app.SeedProgramExercise(p.ID, s1.ID, e.ID)

		// Swap s1 and s2.
		body := []map[string]any{
			{"set_id": s2.ID, "exercise_ids": []int64{}},
			{"set_id": s1.ID, "exercise_ids": []int64{pe1.ID}},
		}
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d/structure", p.ID), body, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
		}

		ctx := domain.NewContext(t.Context(), &domain.Identity{UserID: uID, OrgID: oID})
		got, err := app.Programs.Get(ctx, p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Sets[0].ID != s2.ID || got.Sets[1].ID != s1.ID {
			t.Errorf("expected sets in order [%d, %d], got [%d, %d]", s2.ID, s1.ID, got.Sets[0].ID, got.Sets[1].ID)
		}
	})

	t.Run("reorder exercises within set", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Strength", uID, oID)
		s := app.SeedProgramSet(p.ID, 3)
		e1 := app.SeedExercise("Squat", nil)
		e2 := app.SeedExercise("Deadlift", nil)
		pe1 := app.SeedProgramExercise(p.ID, s.ID, e1.ID)
		pe2 := app.SeedProgramExercise(p.ID, s.ID, e2.ID)

		// Swap pe1 and pe2.
		body := []map[string]any{
			{"set_id": s.ID, "exercise_ids": []int64{pe2.ID, pe1.ID}},
		}
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d/structure", p.ID), body, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
		}

		ctx := domain.NewContext(t.Context(), &domain.Identity{UserID: uID, OrgID: oID})
		got, err := app.Programs.Get(ctx, p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got.Sets[0].Exercises[0].ID != pe2.ID || got.Sets[0].Exercises[1].ID != pe1.ID {
			t.Errorf("expected exercises in order [%d, %d], got [%d, %d]", pe2.ID, pe1.ID, got.Sets[0].Exercises[0].ID, got.Sets[0].Exercises[1].ID)
		}
	})

	t.Run("cross-set exercise move", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Strength", uID, oID)
		s1 := app.SeedProgramSet(p.ID, 3)
		s2 := app.SeedProgramSet(p.ID, 3)
		e := app.SeedExercise("Squat", nil)
		pe := app.SeedProgramExercise(p.ID, s1.ID, e.ID)

		// Move pe from s1 to s2.
		body := []map[string]any{
			{"set_id": s1.ID, "exercise_ids": []int64{}},
			{"set_id": s2.ID, "exercise_ids": []int64{pe.ID}},
		}
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d/structure", p.ID), body, uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
		}

		ctx := domain.NewContext(t.Context(), &domain.Identity{UserID: uID, OrgID: oID})
		got, err := app.Programs.Get(ctx, p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Sets[0].Exercises) != 0 {
			t.Errorf("expected s1 to have 0 exercises, got %d", len(got.Sets[0].Exercises))
		}
		if len(got.Sets[1].Exercises) != 1 || got.Sets[1].Exercises[0].ID != pe.ID {
			t.Errorf("expected s2 to have exercise %d, got %+v", pe.ID, got.Sets[1].Exercises)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Strength", uID, oID)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d/structure", p.ID), strings.NewReader("not json"), uID)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing set", func(t *testing.T) {
		app := NewTestApp(t)
		uID, oID := app.SeedUserWithOrg("test@test.com", "sub")
		p := app.SeedProgramForUser(t.Context(), "Strength", uID, oID)
		app.SeedProgramSet(p.ID, 3)

		// Send wrong set_id.
		body := []map[string]any{
			{"set_id": 999, "exercise_ids": []int64{}},
		}
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d/structure", p.ID), body, uID)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")

		body := []map[string]any{}
		r := app.AuthRequest(http.MethodPut, "/api/programs/999/structure", body, uID)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("RLS: cannot reorder another org's program", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		p := app.SeedProgramForUser(t.Context(), "Strength", u1, o1)
		s := app.SeedProgramSet(p.ID, 3)

		body := []map[string]any{
			{"set_id": s.ID, "exercise_ids": []int64{}},
		}
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d/structure", p.ID), body, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
