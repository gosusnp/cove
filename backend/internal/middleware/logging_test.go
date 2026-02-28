// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogging(t *testing.T) {
	tests := []struct {
		name       string
		wantStatus int
	}{
		{"calls next and captures 200", http.StatusOK},
		{"captures non-200 status", http.StatusNotFound},
		{"captures 500 status", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.wantStatus)
			})

			r := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()

			Logging(next).ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestResponseWriter_DefaultStatus(t *testing.T) {
	// If WriteHeader is never called, status should default to 200.
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	Logging(next).ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
	}
}
