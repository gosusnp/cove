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

	"github.com/gosusnp/cove/api/service"
	"github.com/gosusnp/cove/api/store"
)

func newTestProgramServer(t *testing.T) (http.Handler, *service.ProgramService) {
	t.Helper()
	svc := service.NewProgramService(newTestDB(t))
	mux := http.NewServeMux()
	NewProgramHandler(svc).RegisterRoutes(mux)
	return mux, svc
}

func TestProgramHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodGet, "/programs", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

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
		mux, s := newTestProgramServer(t)
		if _, err := s.Create("Strength"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create("Hypertrophy"); err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodGet, "/programs", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

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
		db := newTestDB(t)
		p, err := store.NewProgramStore(db).Create("Strength")
		if err != nil {
			t.Fatal(err)
		}
		ps, err := store.NewProgramSetStore(db).Create(p.ID, nil, 3, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		e, err := store.NewExerciseStore(db).Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := store.NewProgramExerciseStore(db).Create(ps.ID, e.ID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		mux := http.NewServeMux()
		NewProgramHandler(service.NewProgramService(db)).RegisterRoutes(mux)

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d", p.ID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

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
			t.Fatalf("expected 1 set, got %d", len(got.Sets))
		}
		if len(got.Sets[0].Exercises) != 1 {
			t.Fatalf("expected 1 exercise, got %d", len(got.Sets[0].Exercises))
		}
		if got.Sets[0].Exercises[0].Name != "Pull-up" {
			t.Errorf("got exercise name %q, want %q", got.Sets[0].Exercises[0].Name, "Pull-up")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodGet, "/programs/999", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodGet, "/programs/abc", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramHandler_Create(t *testing.T) {
	t.Run("creates program", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodPost, "/programs", strings.NewReader(`{"name":"Strength"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got store.Program
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got name %q, want %q", got.Name, "Strength")
		}
		if got.ID == 0 {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodPost, "/programs", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodPost, "/programs", strings.NewReader(`not json`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramHandler_Update(t *testing.T) {
	t.Run("updates program", func(t *testing.T) {
		mux, s := newTestProgramServer(t)
		created, err := s.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/programs/%d", created.ID), strings.NewReader(`{"name":"Max Strength"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.Program
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Max Strength" {
			t.Errorf("got name %q, want %q", got.Name, "Max Strength")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodPut, "/programs/999", strings.NewReader(`{"name":"Strength"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		mux, s := newTestProgramServer(t)
		created, err := s.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/programs/%d", created.ID), strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramHandler_Delete(t *testing.T) {
	t.Run("deletes program", func(t *testing.T) {
		mux, s := newTestProgramServer(t)
		created, err := s.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/programs/%d", created.ID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _ := newTestProgramServer(t)

		r := httptest.NewRequest(http.MethodDelete, "/programs/999", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
