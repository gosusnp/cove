// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

func TestWorkoutSessionHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/sessions", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.WorkoutSession
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("got nil, want empty slice")
		}
	})

	t.Run("returns own sessions", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		activity := "Run"
		app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{Activity: &activity})
		app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{Activity: &activity})

		r := app.AuthRequest(http.MethodGet, "/api/sessions", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.WorkoutSession
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d sessions, want 2", len(got))
		}
	})

	t.Run("does not return other users sessions", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		activity := "Run"
		app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{Activity: &activity})

		r := app.AuthRequest(http.MethodGet, "/api/sessions", nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []domain.WorkoutSession
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("got %d sessions, want 0", len(got))
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestWorkoutSessionHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		activity := "Swim"
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{Activity: &activity})

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/sessions/%d", ws.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got domain.WorkoutSession
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Activity == nil || *got.Activity != activity {
			t.Errorf("got activity %v, want %q", got.Activity, activity)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/sessions/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/sessions/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cannot access another user session", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})

		r := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/sessions/%d", ws.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/sessions/1", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestWorkoutSessionHandler_Create(t *testing.T) {
	t.Run("creates session with activity", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := mustJSON(t, map[string]any{"activity": "Yoga"})
		r := app.AuthRequest(http.MethodPost, "/api/sessions", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.WorkoutSession
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Activity == nil || *got.Activity != "Yoga" {
			t.Errorf("got activity %v, want %q", got.Activity, "Yoga")
		}
	})

	t.Run("creates session linked to a program", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		prog := app.SeedProgramForUser(context.Background(), "Strength", u1, o1)
		body := mustJSON(t, map[string]any{"program_id": int64(prog.ID), "program_name": prog.Name})
		r := app.AuthRequest(http.MethodPost, "/api/sessions", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got domain.WorkoutSession
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ProgramID == nil || *got.ProgramID != prog.ID {
			t.Errorf("got program_id %v, want %v", got.ProgramID, prog.ID)
		}
		if got.ProgramName == nil || *got.ProgramName != prog.Name {
			t.Errorf("got program_name %v, want %q", got.ProgramName, prog.Name)
		}
	})

	t.Run("creates freeform session with no body fields", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := mustJSON(t, map[string]any{})
		r := app.AuthRequest(http.MethodPost, "/api/sessions", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodPost, "/api/sessions", bytes.NewBufferString("not json"), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		body := mustJSON(t, map[string]any{})
		r := httptest.NewRequest(http.MethodPost, "/api/sessions", body)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestWorkoutSessionHandler_Update(t *testing.T) {
	t.Run("updates session notes", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		activity := "Swim"
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{Activity: &activity})
		notes := "Great session"
		body := mustJSON(t, map[string]any{"activity": "Swim", "session_notes": notes})
		r := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), body, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got domain.WorkoutSession
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.SessionNotes == nil || *got.SessionNotes != notes {
			t.Errorf("got notes %v, want %q", got.SessionNotes, notes)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := mustJSON(t, map[string]any{})
		r := app.AuthRequest(http.MethodPatch, "/api/sessions/999", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := mustJSON(t, map[string]any{})
		r := app.AuthRequest(http.MethodPatch, "/api/sessions/abc", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})
		r := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), bytes.NewBufferString("not json"), u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cannot update another user session", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})
		body := mustJSON(t, map[string]any{})
		r := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), body, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		body := mustJSON(t, map[string]any{})
		r := httptest.NewRequest(http.MethodPatch, "/api/sessions/1", body)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func mustJSON(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	return bytes.NewBuffer(b)
}
