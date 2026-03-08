// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"

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
		if _, err := app.Programs.Create(ctx1, "Public Strength", nil, true); err != nil {
			t.Fatal(err)
		}

		// Private program for U1
		if _, err := app.Programs.Create(ctx1, "U1 Secret Bodyweight", nil, false); err != nil {
			t.Fatal(err)
		}

		// Private program for U2
		ctx2 := domain.NewContext(t.Context(), &domain.Identity{UserID: u2, OrgID: o2})
		if _, err := app.Programs.Create(ctx2, "U2 Secret Yoga", nil, false); err != nil {
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
		app.SeedProgramExercise(ps.ID, e.ID)

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
		p1, err := app.Programs.Create(ctx1, "Public Strength", nil, true)
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

		body := `{"name":"New Name"}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d", p.ID), strings.NewReader(body), uID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		// Verify change
		ctx := domain.NewContext(t.Context(), &domain.Identity{UserID: uID, OrgID: oID})
		ex, _ := app.Programs.GetDetail(ctx, p.ID)
		if ex.Name != "New Name" {
			t.Errorf("got name %q, want %q", ex.Name, "New Name")
		}
	})

	t.Run("not found", func(t *testing.T) {
		app := NewTestApp(t)
		uID, _ := app.SeedUserWithOrg("test@test.com", "sub")
		r := app.AuthRequest(http.MethodPut, "/api/programs/999", strings.NewReader(`{"name":"test"}`), uID)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("RLS: cannot update another org's program", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")

		p1 := app.SeedProgramForUser(t.Context(), "U1 Program", u1, o1)

		// Request as U2
		body := `{"name":"Hacked"}`
		r := app.AuthRequest(http.MethodPut, fmt.Sprintf("/api/programs/%d", p1.ID), strings.NewReader(body), u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}

		// Verify not changed
		ctx1 := domain.NewContext(t.Context(), &domain.Identity{UserID: u1, OrgID: o1})
		ex, _ := app.Programs.GetDetail(ctx1, p1.ID)
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
		_, err := app.Programs.GetDetail(ctx, p.ID)
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
		_, err := app.Programs.GetDetail(ctx1, p1.ID)
		if err != nil {
			t.Errorf("expected program to still exist, got err: %v", err)
		}
	})
}
