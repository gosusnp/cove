// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func createExpiredSession(t *testing.T, db *sql.DB, userID domain.UserID) string {
	t.Helper()
	token := "expiredtoken"
	hash := hex.EncodeToString(sha256.New().Sum([]byte(token)))

	var orgID domain.OrgID
	if err := db.QueryRow(
		`SELECT org_id FROM cove.org_members WHERE user_id = $1 LIMIT 1`, userID,
	).Scan(&orgID); err != nil {
		t.Fatalf("get org: %v", err)
	}

	expiresAt := time.Now().Add(-time.Hour)
	if _, err := db.Exec(
		`INSERT INTO cove.user_tokens (id, user_id, org_id, kind, token, expires_at, initial_ip_masked, initial_browser, initial_os) VALUES ($1, $2, $3, 'session', $4, $5, '', '', '')`,
		domain.NewSessionID(), userID, orgID, hash, expiresAt,
	); err != nil {
		t.Fatalf("insert expired session: %v", err)
	}
	return token
}

func newOAuth(t *testing.T, next http.Handler) http.Handler {
	db := testutil.NewDB(t)
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
		handler := newOAuth(t, next)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got code %d, want 401", w.Code)
		}
	})

	t.Run("invalid token format returns 401", func(t *testing.T) {
		handler := newOAuth(t, next)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "InvalidFormat")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got code %d, want 401", w.Code)
		}
	})

	t.Run("valid session token succeeds and sets context", func(t *testing.T) {
		database := testutil.NewDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		user, _, _ := svc.GetOrCreate(context.Background(), "test@example.com", "sub123")
		token, _, _ := svc.CreateSession(context.Background(), user.ID, "1.2.3.4", "Chrome", "macOS")

		handler := OAuth(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := UserIDFromContext(r.Context())
			if got != user.ID {
				t.Errorf("got userID %v, want %v", got, user.ID)
			}
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got code %d, want 200", w.Code)
		}
	})

	t.Run("valid pat token succeeds and sets context", func(t *testing.T) {
		database := testutil.NewDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		user, _, _ := svc.GetOrCreate(context.Background(), "test@example.com", "sub123")
		var orgID domain.OrgID
		err := database.QueryRowContext(context.Background(), "SELECT org_id FROM cove.org_members WHERE user_id = $1 LIMIT 1", user.ID).Scan(&orgID)
		if err != nil {
			t.Fatalf("get org for pat: %v", err)
		}
		token, _, _ := svc.CreatePAT(context.Background(), user.ID, orgID, "My PAT", "1.2.3.4", "Chrome", "macOS")

		handler := OAuth(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := UserIDFromContext(r.Context())
			if got != user.ID {
				t.Errorf("got userID %v, want %v", got, user.ID)
			}
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got code %d, want 200", w.Code)
		}
	})

	t.Run("non-Bearer authorization header returns 401", func(t *testing.T) {
		handler := newOAuth(t, next)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "SomeArbitraryToken")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got code %d, want 401", w.Code)
		}
	})

	t.Run("valid session token via cookie succeeds", func(t *testing.T) {
		database := testutil.NewDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		user, _, _ := svc.GetOrCreate(context.Background(), "cookie@example.com", "sub-cookie")
		token, _, _ := svc.CreateSession(context.Background(), user.ID, "1.2.3.4", "Chrome", "macOS")

		handler := OAuth(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			got := UserIDFromContext(r.Context())
			if got != user.ID {
				t.Errorf("got userID %v, want %v", got, user.ID)
			}
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.AddCookie(&http.Cookie{Name: "cove_session", Value: token})
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got code %d, want 200", w.Code)
		}
	})

	t.Run("service account token sets IsServiceAccount and no OrgID", func(t *testing.T) {
		database := testutil.NewDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		svcUserID := domain.NewUserID()
		if _, err := database.Exec(
			`INSERT INTO cove.users (id, is_service_account) VALUES ($1, true)`, svcUserID,
		); err != nil {
			t.Fatalf("insert service account: %v", err)
		}
		rawToken := "svc_middleware_test"
		h := sha256.Sum256([]byte(rawToken))
		hash := hex.EncodeToString(h[:])
		if _, err := database.Exec(
			`INSERT INTO cove.user_tokens (id, user_id, org_id, kind, token, initial_ip_masked, initial_browser, initial_os)
			 VALUES ($1, $2, NULL, 'pat', $3, '', '', '')`,
			domain.NewSessionID(), svcUserID, hash,
		); err != nil {
			t.Fatalf("insert service account token: %v", err)
		}

		handler := OAuth(svc, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			id, ok := IdentityFromContext(r.Context())
			if !ok {
				t.Fatal("expected identity in context")
			}
			if id.UserID != svcUserID {
				t.Errorf("got userID %v, want %v", id.UserID, svcUserID)
			}
			if !id.IsServiceAccount() {
				t.Error("expected IsServiceAccount=true")
			}
			if id.OrgID != (domain.OrgID{}) {
				t.Errorf("expected zero OrgID for service account, got %v", id.OrgID)
			}
			w.WriteHeader(http.StatusOK)
		}))

		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+rawToken)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got code %d, want 200", w.Code)
		}
	})

	t.Run("expired token returns 401", func(t *testing.T) {
		database := testutil.NewDB(t)
		us := store.NewUserStore()
		orgs := store.NewOrgStore()
		svc := service.NewUserService(database, us, orgs)

		user, _, _ := svc.GetOrCreate(context.Background(), "test@example.com", "sub123")
		token := createExpiredSession(t, database, user.ID)

		handler := OAuth(svc, next)
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got code %d, want 401", w.Code)
		}
	})
}
