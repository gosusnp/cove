// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/fdc"
)

// mockFDCSearcher is a test double for fdcSearcher.
type mockFDCSearcher struct {
	results []fdc.FoodResult
	err     error
}

func (m *mockFDCSearcher) Search(_ context.Context, _ string) ([]fdc.FoodResult, error) {
	return m.results, m.err
}

func newFDCMux(client fdcSearcher) http.Handler {
	mux := http.NewServeMux()
	NewFDCHandler(client).RegisterRoutes(mux)
	return mux
}

func TestFDCHandler_search_missingQ(t *testing.T) {
	mux := newFDCMux(&mockFDCSearcher{})
	r := httptest.NewRequest(http.MethodGet, "/fdc/search", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", w.Code)
	}
	var body map[string]string
	_ = json.NewDecoder(w.Body).Decode(&body)
	if body["error"] == "" {
		t.Error("expected error message in response body")
	}
}

func TestFDCHandler_search_clientError(t *testing.T) {
	client := &mockFDCSearcher{err: errors.New("upstream unavailable")}
	mux := newFDCMux(client)
	r := httptest.NewRequest(http.MethodGet, "/fdc/search?q=egg", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusInternalServerError {
		t.Errorf("got %d, want 500", w.Code)
	}
}

func TestFDCHandler_search_success(t *testing.T) {
	foods := []fdc.FoodResult{
		{FDCID: 1, Name: "Egg, whole, raw", DataType: "Foundation", CaloriesPer100g: 143},
		{FDCID: 2, Name: "Eggs, Grade A", DataType: "SR Legacy", CaloriesPer100g: 140},
	}
	client := &mockFDCSearcher{results: foods}
	mux := newFDCMux(client)
	r := httptest.NewRequest(http.MethodGet, "/fdc/search?q=egg", nil)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("got %d, want 200", w.Code)
	}

	var body struct {
		Foods []fdc.FoodResult `json:"foods"`
	}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(body.Foods) != 2 {
		t.Errorf("got %d foods, want 2", len(body.Foods))
	}
	if body.Foods[0].FDCID != 1 {
		t.Errorf("got fdc_id %d, want 1", body.Foods[0].FDCID)
	}
}
