// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testdb"
)

// createExpiredSession inserts a session that is already expired and returns the raw token.
func createExpiredSession(t *testing.T, database *sql.DB, userID uuid.UUID) string {
	t.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand: %v", err)
	}
	token := hex.EncodeToString(buf)
	h := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(h[:])
	expiresAt := time.Now().Add(-time.Hour)
	if _, err := database.Exec(
		`INSERT INTO user_sessions (user_id, token, expires_at) VALUES ($1, $2, $3)`,
		userID, hash, expiresAt,
	); err != nil {
		t.Fatalf("insert expired session: %v", err)
	}
	return token
}

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

	t.Run("expired token returns 401", func(t *testing.T) {
		database := newTestDB(t)
		us := store.NewUserStore(database)

		user, _, err := us.GetOrCreate("expired@example.com", "sub-expired")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token := createExpiredSession(t, database, user.ID)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
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

	t.Run("valid token stores user in context", func(t *testing.T) {
		us := store.NewUserStore(newTestDB(t))

		user, _, err := us.GetOrCreate("ctx@example.com", "sub-ctx")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token, err := us.CreateSession(user.ID)
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		var gotUser *store.User
		capture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotUser = UserFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		OAuth(us, capture).ServeHTTP(w, r)

		if gotUser == nil {
			t.Fatal("expected user in context, got nil")
		}
		if gotUser.Email != "ctx@example.com" {
			t.Errorf("got email %q, want %q", gotUser.Email, "ctx@example.com")
		}
	})
}
