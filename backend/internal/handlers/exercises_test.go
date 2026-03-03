// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.NewDB(t, containerDSN, db.MigrationsFS)
}

func newTestServer(t *testing.T) (http.Handler, *service.ExerciseService) {
	t.Helper()
	svc := service.NewExerciseService(store.NewExerciseStore(newTestDB(t)))
	mux := http.NewServeMux()
	NewExerciseHandler(svc).RegisterRoutes(mux)
	return mux, svc
}

func TestExerciseHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodGet, "/exercises", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.Exercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns exercises", func(t *testing.T) {
		mux, s := newTestServer(t)
		if _, err := s.Create("Pull-up", nil); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create("Push-up", nil); err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodGet, "/exercises", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.Exercise
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
		mux, s := newTestServer(t)
		created, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/exercises/%d", created.ID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.Exercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Pull-up" {
			t.Errorf("got name %q, want %q", got.Name, "Pull-up")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodGet, "/exercises/999", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("invalid id", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodGet, "/exercises/abc", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestExerciseHandler_Create(t *testing.T) {
	t.Run("creates exercise", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodPost, "/exercises", strings.NewReader(`{"name":"Pull-up"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got store.Exercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Pull-up" {
			t.Errorf("got name %q, want %q", got.Name, "Pull-up")
		}
		if got.ID == 0 {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodPost, "/exercises", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodPost, "/exercises", strings.NewReader(`not json`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestExerciseHandler_Update(t *testing.T) {
	t.Run("updates exercise", func(t *testing.T) {
		mux, s := newTestServer(t)
		created, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		body := `{"name":"Weighted Pull-up","progression":"weighted"}`
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/exercises/%d", created.ID), strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.Exercise
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name != "Weighted Pull-up" {
			t.Errorf("got name %q, want %q", got.Name, "Weighted Pull-up")
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodPut, "/exercises/999", strings.NewReader(`{"name":"Pull-up"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("missing name returns 400", func(t *testing.T) {
		mux, s := newTestServer(t)
		created, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/exercises/%d", created.ID), strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestExerciseHandler_Delete(t *testing.T) {
	t.Run("deletes exercise", func(t *testing.T) {
		mux, s := newTestServer(t)
		created, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/exercises/%d", created.ID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _ := newTestServer(t)

		r := httptest.NewRequest(http.MethodDelete, "/exercises/999", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
