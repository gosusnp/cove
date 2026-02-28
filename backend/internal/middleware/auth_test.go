// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testdb"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testdb.New(t, containerDSN, db.MigrationsFS)
}

func TestAPIKey(t *testing.T) {
	const key = "test-key"

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	handler := APIKey(key, next)

	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{"valid key", "Bearer test-key", http.StatusOK},
		{"wrong key", "Bearer wrong-key", http.StatusUnauthorized},
		{"missing header", "", http.StatusUnauthorized},
		{"bearer prefix only", "Bearer ", http.StatusUnauthorized},
		{"no bearer prefix", "test-key", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				r.Header.Set("Authorization", tt.authHeader)
			}
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, r)

			if w.Code != tt.wantStatus {
				t.Errorf("got status %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestOAuth(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("missing header returns 401", func(t *testing.T) {
		us := store.NewUserStore(newTestDB(t))
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		OAuth(us, next).ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		us := store.NewUserStore(newTestDB(t))
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer notavalidtoken")
		w := httptest.NewRecorder()

		OAuth(us, next).ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("valid token calls next", func(t *testing.T) {
		us := store.NewUserStore(newTestDB(t))

		user, _, err := us.GetOrCreate("test@example.com", "sub-test")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token, err := us.CreateSession(user.ID)
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		OAuth(us, next).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
	})
}
