// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
)

func TestRequireAdmin(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	withIdentity := func(r *http.Request, id *domain.Identity) *http.Request {
		return r.WithContext(domain.NewContext(r.Context(), id))
	}

	t.Run("admin identity is allowed through", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = withIdentity(r, &domain.Identity{Admin: true})
		w := httptest.NewRecorder()

		RequireAdmin(ok).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want 200", w.Code)
		}
	})

	t.Run("non-admin identity is forbidden", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r = withIdentity(r, &domain.Identity{Admin: false})
		w := httptest.NewRecorder()

		RequireAdmin(ok).ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Errorf("got %d, want 403", w.Code)
		}
	})

	t.Run("missing identity is forbidden", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		RequireAdmin(ok).ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Errorf("got %d, want 403", w.Code)
		}
	})
}
