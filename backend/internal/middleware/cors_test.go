// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func newCORSHandler(origins []string) http.Handler {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	return CORS(origins, next)
}

func TestCORS(t *testing.T) {
	allowed := []string{"capacitor://localhost", "https://example.com"}

	t.Run("allowed origin sets CORS headers", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "capacitor://localhost")
		w := httptest.NewRecorder()

		newCORSHandler(allowed).ServeHTTP(w, r)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "capacitor://localhost" {
			t.Errorf("got %q, want %q", got, "capacitor://localhost")
		}
		if got := w.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Errorf("got %q, want %q", got, "true")
		}
		if got := w.Header().Get("Vary"); got != "Origin" {
			t.Errorf("got %q, want %q", got, "Origin")
		}
		if w.Code != http.StatusOK {
			t.Errorf("got code %d, want 200", w.Code)
		}
	})

	t.Run("disallowed origin does not set CORS headers", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		newCORSHandler(allowed).ServeHTTP(w, r)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
		if w.Code != http.StatusOK {
			t.Errorf("got code %d, want 200", w.Code)
		}
	})

	t.Run("no origin does not set CORS headers", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		newCORSHandler(allowed).ServeHTTP(w, r)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("preflight OPTIONS returns 204 with allowed methods and headers", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		r.Header.Set("Origin", "capacitor://localhost")
		w := httptest.NewRecorder()

		newCORSHandler(allowed).ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got code %d, want 204", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Error("expected Access-Control-Allow-Methods to be set")
		}
		if got := w.Header().Get("Access-Control-Allow-Headers"); got == "" {
			t.Error("expected Access-Control-Allow-Headers to be set")
		}
	})

	t.Run("preflight from disallowed origin still returns 204 but no allow-origin", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodOptions, "/", nil)
		r.Header.Set("Origin", "https://evil.com")
		w := httptest.NewRecorder()

		newCORSHandler(allowed).ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got code %d, want 204", w.Code)
		}
		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})

	t.Run("empty allowed origins blocks all", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Origin", "capacitor://localhost")
		w := httptest.NewRecorder()

		newCORSHandler(nil).ServeHTTP(w, r)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "" {
			t.Errorf("got %q, want empty", got)
		}
	})
}
