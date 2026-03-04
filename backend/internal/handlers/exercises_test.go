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

func TestExerciseHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/exercises", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.ExerciseLite
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("got nil, want empty slice")
		}
	})

	t.Run("returns exercises", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		app.SeedExerciseForUser(context.Background(), "Push-up", nil, u1, o1)
		app.SeedExerciseForUser(context.Background(), "Air Squat", nil, u1, o1)

		r := app.AuthRequest(http.MethodGet, "/api/exercises", nil, u1)
		w := app.Do(r)

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

	t.Run("RLS: list only returns own or public", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, o2 := app.SeedUserWithOrg("u2@test.com", "sub2")

		// Public exercise owned by U1
		id1 := &domain.Identity{UserID: u1, OrgID: o1}
		ctx1 := domain.NewContext(context.Background(), id1)
		if _, err := app.Exercises.Create(ctx1, "Public Squat", nil, nil, true); err != nil {
			t.Fatal(err)
		}

		// Private exercise for U1
		if _, err := app.Exercises.Create(ctx1, "U1 Secret Press", nil, nil, false); err != nil {
			t.Fatal(err)
		}

		// Private exercise for U2
		id2 := &domain.Identity{UserID: u2, OrgID: o2}
		ctx2 := domain.NewContext(context.Background(), id2)
		if _, err := app.Exercises.Create(ctx2, "U2 Secret Pull", nil, nil, false); err != nil {
			t.Fatal(err)
		}

		// Request as U1
		r1 := app.AuthRequest(http.MethodGet, "/api/exercises", nil, u1)
		w1 := app.Do(r1)

		if w1.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w1.Code, http.StatusOK)
		}
		var got1 []domain.ExerciseLite
		if err := json.NewDecoder(w1.Body).Decode(&got1); err != nil {
			t.Fatalf("decode U1: %v", err)
		}

		if len(got1) != 2 {
			t.Errorf("U1 should see 2 exercises (public + own), got %d: %+v", len(got1), got1)
		}

		// Request as U2
		r2 := app.AuthRequest(http.MethodGet, "/api/exercises", nil, u2)
		w2 := app.Do(r2)
		if w2.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w2.Code, http.StatusOK)
		}
		var got2 []domain.ExerciseLite
		if err := json.NewDecoder(w2.Body).Decode(&got2); err != nil {
			t.Fatalf("decode U2: %v", err)
		}

		if len(got2) != 2 {
			t.Errorf("U2 should see 2 exercises (public + own), got %d: %+v", len(got2), got2)
		}
	})
}

func TestExerciseHandler_List_Unauthorized(t *testing.T) {
	app := NewTestApp(t)
	r := httptest.NewRequest(http.MethodGet, "/api/exercises", nil)
	w := app.Do(r)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
	}
}

func TestExerciseHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prog := "Add 1 rep"
		e := app.SeedExerciseForUser(context.Background(), "Push-up", &prog, u1, o1)

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/exercises/%d", e.ID), nil, u1)
		w := app.Do(r)

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
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/exercises/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/exercises/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/exercises/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestExerciseHandler_Create(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")

		body := `{"name":"New Exercise", "progression":"Test"}`
		r := app.AuthRequest(http.MethodPost, "/api/exercises", strings.NewReader(body), u1)
		w := app.Do(r)

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

		// Verify bookkeeping via trigger
		if got.CreatedBy.UUID != u1.UUID {
			t.Errorf("got created_by %v, want %v", got.CreatedBy, u1)
		}
		if got.CreatedAt.IsZero() {
			t.Error("expected created_at to be populated")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/exercises", strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/exercises", strings.NewReader(`{}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/exercises", strings.NewReader(`{"name":"test"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestExerciseHandler_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		e := app.SeedExerciseForUser(context.Background(), "Old Name", nil, u1, o1)

		body := `{"name":"New Name", "progression":"Updated"}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/exercises/%d", e.ID), strings.NewReader(body), u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		ex, _ := app.Exercises.Get(domain.NewContext(context.Background(), &domain.Identity{UserID: u1, OrgID: o1}), e.ID)
		if ex.Name != "New Name" {
			t.Errorf("got name %q, want %q", ex.Name, "New Name")
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		e := app.SeedExerciseForUser(context.Background(), "Push-up", nil, u1, o1)

		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/exercises/%d", e.ID), strings.NewReader(`not json`), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPut, "/api/exercises/999", strings.NewReader(`{"name":"test"}`), u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPut, "/api/exercises/1", strings.NewReader(`{"name":"test"}`))
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestExerciseHandler_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		e := app.SeedExerciseForUser(context.Background(), "To Delete", nil, u1, o1)

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/exercises/%d", e.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		_, err := app.Exercises.Get(domain.NewContext(context.Background(), &domain.Identity{UserID: u1, OrgID: o1}), e.ID)
		if err == nil {
			t.Error("expected error getting deleted exercise")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/exercises/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/api/exercises/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
