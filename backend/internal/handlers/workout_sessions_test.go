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
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

// sessionResp is used to decode workout session JSON responses in tests.
type sessionResp struct {
	ID               domain.WorkoutSessionID `json:"id"`
	Activity         *string                 `json:"activity,omitempty"`
	ProgramID        *domain.ProgramID       `json:"program_id,omitempty"`
	ProgramName      *string                 `json:"program_name,omitempty"`
	PerceivedEffort  *int                    `json:"perceived_effort,omitempty"`
	SessionNotes     *string                 `json:"session_notes,omitempty"`
	ProgramStructure *string                 `json:"program_structure,omitempty"`
}

func TestWorkoutSessionHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/sessions", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []sessionResp
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
		var got []sessionResp
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
		var got []sessionResp
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
		var got sessionResp
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
		var got sessionResp
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
		var got sessionResp
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

func TestWorkoutSessionHandler_Replace(t *testing.T) {
	t.Run("replaces session fields", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		activity := "Swim"
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{Activity: &activity})
		notes := "Great session"
		body := mustJSON(t, map[string]any{"activity": "Swim", "session_notes": notes, "updated_at": ws.UpdatedAt})
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/sessions/%d", ws.ID), body, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got sessionResp
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
		body := mustJSON(t, map[string]any{"updated_at": time.Now()})
		r := app.AuthRequest(http.MethodPut, "/api/sessions/999", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := mustJSON(t, map[string]any{"updated_at": time.Now()})
		r := app.AuthRequest(http.MethodPut, "/api/sessions/abc", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/sessions/%d", ws.ID), bytes.NewBufferString("not json"), u1)
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
		body := mustJSON(t, map[string]any{"updated_at": ws.UpdatedAt})
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/sessions/%d", ws.ID), body, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		body := mustJSON(t, map[string]any{})
		r := httptest.NewRequest(http.MethodPut, "/api/sessions/1", body)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestWorkoutSessionHandler_Patch(t *testing.T) {
	t.Run("updates a single field without affecting others", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		activity := "Run"
		effort := 7
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{
			Activity: &activity,
			SensitiveData: domain.SessionSensitiveData{
				PerceivedEffort: &effort,
			},
		})
		notes := "Felt great"
		body := mustJSON(t, map[string]any{"session_notes": notes, "updated_at": ws.UpdatedAt})
		r := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), body, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got sessionResp
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.SessionNotes == nil || *got.SessionNotes != notes {
			t.Errorf("got notes %v, want %q", got.SessionNotes, notes)
		}
		if got.Activity == nil || *got.Activity != activity {
			t.Errorf("got activity %v, want %q — field should be preserved", got.Activity, activity)
		}
		if got.PerceivedEffort == nil || *got.PerceivedEffort != effort {
			t.Errorf("got perceived_effort %v, want %d — field should be preserved", got.PerceivedEffort, effort)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := mustJSON(t, map[string]any{"updated_at": time.Now()})
		r := app.AuthRequest(http.MethodPatch, "/api/sessions/999", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		body := mustJSON(t, map[string]any{"updated_at": time.Now()})
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

	t.Run("cannot patch another user session", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})
		body := mustJSON(t, map[string]any{"updated_at": ws.UpdatedAt})
		r := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), body, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("optimistic locking is optional", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})

		// Omitting updated_at should succeed ("last write wins").
		body := mustJSON(t, map[string]any{"activity": "Yoga"})
		r := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), body, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("omitted: got status %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("conflict when updated_at is provided but stale", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})

		// First update to move the timestamp.
		body1 := mustJSON(t, map[string]any{"activity": "Yoga", "updated_at": ws.UpdatedAt})
		r1 := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), body1, u1)
		w1 := app.Do(r1)
		if w1.Code != http.StatusOK {
			t.Fatalf("setup: got status %d, want %d", w1.Code, http.StatusOK)
		}

		// Second update using the OLD timestamp should fail.
		body2 := mustJSON(t, map[string]any{"activity": "Swim", "updated_at": ws.UpdatedAt})
		r2 := app.AuthRequest(http.MethodPatch, fmt.Sprintf("/api/sessions/%d", ws.ID), body2, u1)
		w2 := app.Do(r2)

		if w2.Code != http.StatusConflict {
			t.Errorf("stale: got status %d, want %d", w2.Code, http.StatusConflict)
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

func TestWorkoutSessionHandler_Delete(t *testing.T) {
	t.Run("deletes own session", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/sessions/%d", ws.ID), nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify the session is gone.
		r2 := app.AuthRequest(http.MethodGet, fmt.Sprintf("/api/sessions/%d", ws.ID), nil, u1)
		w2 := app.Do(r2)
		if w2.Code != http.StatusNotFound {
			t.Errorf("after delete: got status %d, want %d", w2.Code, http.StatusNotFound)
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/sessions/999", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodDelete, "/api/sessions/abc", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("cannot delete another user session", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")
		ws := app.SeedWorkoutSession(context.Background(), u1, o1, store.WorkoutSessionParams{})

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/sessions/%d", ws.ID), nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodDelete, "/api/sessions/1", nil)
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
