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
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testdb"
)

// createExpiredSession inserts a session that is already expired and returns the raw token.
func createExpiredSession(t *testing.T, database *sql.DB, userID domain.UserID) string {
	t.Helper()
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		t.Fatalf("rand: %v", err)
	}
	token := "sess_" + hex.EncodeToString(buf)
	h := sha256.Sum256([]byte(token))
	hash := hex.EncodeToString(h[:])

	var orgID uuid.UUID
	if err := database.QueryRow(
		`SELECT org_id FROM org_members WHERE user_id = $1 LIMIT 1`, userID,
	).Scan(&orgID); err != nil {
		t.Fatalf("get org: %v", err)
	}

	expiresAt := time.Now().Add(-time.Hour)
	if _, err := database.Exec(
		`INSERT INTO user_tokens (user_id, org_id, kind, token, expires_at, initial_ip_masked, initial_browser, initial_os) VALUES ($1, $2, 'session', $3, $4, '', '', '')`,
		userID, orgID, hash, expiresAt,
	); err != nil {
		t.Fatalf("insert expired session: %v", err)
	}
	return token
}

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testdb.New(t, containerDSN, db.MigrationsFS)
}

func newOAuth(t *testing.T, next http.Handler) http.Handler {
	db := newTestDB(t)
	us := store.NewUserStore()
	orgs := store.NewOrgStore()
	svc := service.NewUserService(db, us, orgs)
	return OAuth(svc, next)
}

func TestOAuth(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	t.Run("missing header returns 401", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		oauth := newOAuth(t, next)

		oauth.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer notavalidtoken")
		w := httptest.NewRecorder()
		oauth := newOAuth(t, next)

		oauth.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		ctx := t.Context()
		database := newTestDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		user, _, err := svc.GetOrCreate(ctx, "expired@example.com", "sub-expired")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token := createExpiredSession(t, database, user.ID)

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		OAuth(svc, next).ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("valid token calls next", func(t *testing.T) {
		ctx := t.Context()
		database := newTestDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		user, _, err := svc.GetOrCreate(ctx, "test@example.com", "sub-test")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token, err := svc.CreateSession(ctx, user.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		OAuth(svc, next).ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
	})

	t.Run("valid token stores user in context", func(t *testing.T) {
		ctx := t.Context()
		database := newTestDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		user, _, err := svc.GetOrCreate(ctx, "ctx@example.com", "sub-ctx")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token, err := svc.CreateSession(ctx, user.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		var gotUser *domain.User
		capture := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			gotUser = UserFromContext(r.Context())
			w.WriteHeader(http.StatusOK)
		})

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		OAuth(svc, capture).ServeHTTP(w, r)

		if gotUser == nil {
			t.Fatal("expected user in context, got nil")
		}
		if gotUser.Email != "ctx@example.com" {
			t.Errorf("got email %q, want %q", gotUser.Email, "ctx@example.com")
		}
	})
}
