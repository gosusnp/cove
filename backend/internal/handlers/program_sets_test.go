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

	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
)

func newTestProgramSetServer(t *testing.T) (http.Handler, *service.ProgramSetService, int64) {
	t.Helper()
	db := newTestDB(t)
	svc := service.NewProgramSetService(store.NewProgramSetStore(db))
	p, err := store.NewProgramStore(db).Create("Test Program")
	if err != nil {
		t.Fatal(err)
	}
	mux := http.NewServeMux()
	NewProgramSetHandler(svc).RegisterRoutes(mux)
	return mux, svc, p.ID
}

func TestProgramSetHandler_List(t *testing.T) {
	t.Run("empty returns array not null", func(t *testing.T) {
		mux, _, programID := newTestProgramSetServer(t)

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets", programID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got == nil {
			t.Error("expected empty slice, got nil")
		}
	})

	t.Run("returns sets for program", func(t *testing.T) {
		mux, s, programID := newTestProgramSetServer(t)
		if _, err := s.Create(programID, nil, 1, nil, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create(programID, nil, 1, nil, nil); err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets", programID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 2 {
			t.Errorf("got %d sets, want 2", len(got))
		}
	})
}

func TestProgramSetHandler_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		mux, s, programID := newTestProgramSetServer(t)
		created, err := s.Create(programID, nil, 3, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets/%d", programID, created.ID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Rounds != 3 {
			t.Errorf("got rounds %d, want 3", got.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _, programID := newTestProgramSetServer(t)

		r := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/programs/%d/sets/999", programID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramSetHandler_Create(t *testing.T) {
	t.Run("creates with all fields", func(t *testing.T) {
		mux, _, programID := newTestProgramSetServer(t)

		body := `{"name":"Warmup","rounds":3,"rest_s":60}`
		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/programs/%d/sets", programID), strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name == nil || *got.Name != "Warmup" {
			t.Errorf("got name %v, want %q", got.Name, "Warmup")
		}
		if got.Rounds != 3 {
			t.Errorf("got rounds %d, want 3", got.Rounds)
		}
	})

	t.Run("defaults rounds to 1 when not provided", func(t *testing.T) {
		mux, _, programID := newTestProgramSetServer(t)

		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/programs/%d/sets", programID), strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("got status %d, want %d", w.Code, http.StatusCreated)
		}
		var got store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", got.Rounds)
		}
	})

	t.Run("invalid body returns 400", func(t *testing.T) {
		mux, _, programID := newTestProgramSetServer(t)

		r := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/programs/%d/sets", programID), strings.NewReader(`not json`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})
}

func TestProgramSetHandler_Update(t *testing.T) {
	t.Run("updates set", func(t *testing.T) {
		mux, s, programID := newTestProgramSetServer(t)
		created, err := s.Create(programID, nil, 1, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		body := `{"name":"Warmup","rounds":4}`
		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/programs/%d/sets/%d", programID, created.ID), strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got store.ProgramSet
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Name == nil || *got.Name != "Warmup" {
			t.Errorf("got name %v, want %q", got.Name, "Warmup")
		}
		if got.Rounds != 4 {
			t.Errorf("got rounds %d, want 4", got.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _, programID := newTestProgramSetServer(t)

		r := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/programs/%d/sets/999", programID), strings.NewReader(`{"rounds":1}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}

func TestProgramSetHandler_Delete(t *testing.T) {
	t.Run("deletes set", func(t *testing.T) {
		mux, s, programID := newTestProgramSetServer(t)
		created, err := s.Create(programID, nil, 1, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/programs/%d/sets/%d", programID, created.ID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("not found", func(t *testing.T) {
		mux, _, programID := newTestProgramSetServer(t)

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/programs/%d/sets/999", programID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
