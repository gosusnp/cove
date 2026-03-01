// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
)

func newUsersServer(t *testing.T) (http.Handler, *store.UserStore) {
	t.Helper()
	us := store.NewUserStore(newTestDB(t))
	mux := http.NewServeMux()
	NewUserHandler(service.NewUserService(us)).RegisterRoutes(mux)
	return middleware.OAuth(us, mux), us
}

func TestUserHandler_Me(t *testing.T) {
	t.Run("returns current user", func(t *testing.T) {
		handler, us := newUsersServer(t)

		user, _, err := us.GetOrCreate("me@example.com", "sub-me")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token, err := us.CreateSession(user.ID)
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		r := httptest.NewRequest(http.MethodGet, "/me", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		var got struct {
			ID    uuid.UUID `json:"id"`
			Email string    `json:"email"`
		}
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Email != "me@example.com" {
			t.Errorf("got email %q, want %q", got.Email, "me@example.com")
		}
		if got.ID != user.ID {
			t.Errorf("got id %v, want %v", got.ID, user.ID)
		}
	})

	t.Run("missing token returns 401", func(t *testing.T) {
		handler, _ := newUsersServer(t)

		r := httptest.NewRequest(http.MethodGet, "/me", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		handler, _ := newUsersServer(t)

		r := httptest.NewRequest(http.MethodGet, "/me", nil)
		r.Header.Set("Authorization", "Bearer notavalidtoken")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
