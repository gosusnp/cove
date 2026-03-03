// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
)

func newUsersServer(t *testing.T) (context.Context, http.Handler, *store.UserStore, *service.UserService) {
	t.Helper()
	db := newTestDB(t)
	us := store.NewUserStore(db)
	orgs := store.NewOrgStore()
	svc := service.NewUserService(db, us, orgs)
	mux := http.NewServeMux()
	NewUserHandler(svc).RegisterRoutes(mux)
	return t.Context(), middleware.OAuth(us, mux), us, svc
}

func TestUserHandler_Logout(t *testing.T) {
	t.Run("deletes current session", func(t *testing.T) {
		ctx, handler, us, svc := newUsersServer(t)
		user, _, err := svc.GetOrCreate(ctx, "logout@example.com", "sub-logout")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token, err := us.CreateSession(user.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		// Verify session exists.
		sessions, _ := us.ListSessions(user.ID)
		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}

		// Logout.
		r := httptest.NewRequest(http.MethodPost, "/users/logout", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify session is gone.
		sessions, _ = us.ListSessions(user.ID)
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions after logout, got %d", len(sessions))
		}
	})

	t.Run("returns 401 for missing token", func(t *testing.T) {
		_, handler, _, _ := newUsersServer(t)
		r := httptest.NewRequest(http.MethodPost, "/users/logout", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
func TestUserHandler_Me(t *testing.T) {
	t.Run("returns current user", func(t *testing.T) {
		ctx, handler, us, svc := newUsersServer(t)

		user, _, err := svc.GetOrCreate(ctx, "me@example.com", "sub-me")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		token, err := us.CreateSession(user.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		var got struct {
			ID    domain.UserID `json:"id"`
			Email string        `json:"email"`
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
		_, handler, _, _ := newUsersServer(t)

		r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		_, handler, _, _ := newUsersServer(t)

		r := httptest.NewRequest(http.MethodGet, "/users/me", nil)
		r.Header.Set("Authorization", "Bearer notavalidtoken")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestUserHandler_Sessions(t *testing.T) {
	setup := func(t *testing.T) (http.Handler, *store.UserStore, string) {
		t.Helper()
		ctx, handler, us, svc := newUsersServer(t)
		user, _, err := svc.GetOrCreate(ctx, "sess-h@example.com", "sub-sess-h")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		session, err := us.CreateSession(user.ID, "1.2.3.4", "Chrome", "macOS")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}
		return handler, us, session
	}

	t.Run("list returns active sessions with is_current", func(t *testing.T) {
		handler, _, session := setup(t)

		r := httptest.NewRequest(http.MethodGet, "/users/sessions", nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 session, got %d", len(got))
		}
		if got[0]["is_current"] != true {
			t.Errorf("expected is_current=true")
		}
		if got[0]["initial_ip_masked"] != "1.2.3.4" {
			t.Errorf("got ip %v, want %q", got[0]["initial_ip_masked"], "1.2.3.4")
		}
	})

	t.Run("delete removes session", func(t *testing.T) {
		handler, _, session := setup(t)

		// List to get session ID.
		r := httptest.NewRequest(http.MethodGet, "/users/sessions", nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		var sessions []map[string]any
		if err := json.NewDecoder(w.Body).Decode(&sessions); err != nil {
			t.Fatalf("decode: %v", err)
		}
		id := sessions[0]["id"].(string)

		// Delete.
		r = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/sessions/%s", id), nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}
	})

	t.Run("cannot list another user's sessions", func(t *testing.T) {
		ctx, handler, us, svc := newUsersServer(t)

		// User A has a session.
		userA, _, err := svc.GetOrCreate(ctx, "a-sess@example.com", "sub-a-sess")
		if err != nil {
			t.Fatalf("GetOrCreate A: %v", err)
		}
		if _, err := us.CreateSession(userA.ID, "1.1.1.1", "", ""); err != nil {
			t.Fatalf("CreateSession A: %v", err)
		}

		// User B lists sessions — should only see their own.
		userB, _, err := svc.GetOrCreate(ctx, "b-sess@example.com", "sub-b-sess")
		if err != nil {
			t.Fatalf("GetOrCreate B: %v", err)
		}
		sessionB, err := us.CreateSession(userB.ID, "2.2.2.2", "", "")
		if err != nil {
			t.Fatalf("CreateSession B: %v", err)
		}

		r := httptest.NewRequest(http.MethodGet, "/users/sessions", nil)
		r.Header.Set("Authorization", "Bearer "+sessionB)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		var got []map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 1 {
			t.Errorf("expected user B to see 1 session (theirs), got %d", len(got))
		}
		if got[0]["initial_ip_masked"] != "2.2.2.2" {
			t.Errorf("expected user B to see their own session")
		}
	})

	t.Run("cannot delete another user's session", func(t *testing.T) {
		ctx, handler, us, svc := newUsersServer(t)

		// User A has a session.
		userA, _, err := svc.GetOrCreate(ctx, "a-sess-del@example.com", "sub-a-sess-del")
		if err != nil {
			t.Fatalf("GetOrCreate A: %v", err)
		}
		if _, err := us.CreateSession(userA.ID, "1.1.1.1", "", ""); err != nil {
			t.Fatalf("CreateSession A: %v", err)
		}
		sessionsA, err := us.ListSessions(userA.ID)
		if err != nil {
			t.Fatalf("ListSessions A: %v", err)
		}
		idA := sessionsA[0].ID

		// User B attempts to delete User A's session.
		userB, _, err := svc.GetOrCreate(ctx, "b-sess-del@example.com", "sub-b-sess-del")
		if err != nil {
			t.Fatalf("GetOrCreate B: %v", err)
		}
		sessionB, err := us.CreateSession(userB.ID, "2.2.2.2", "", "")
		if err != nil {
			t.Fatalf("CreateSession B: %v", err)
		}

		r := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/sessions/%s", idA), nil)
		r.Header.Set("Authorization", "Bearer "+sessionB)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}

		// Verify User A still has their session.
		sessionsA, _ = us.ListSessions(userA.ID)
		if len(sessionsA) != 1 {
			t.Errorf("user A session was deleted by user B")
		}
	})
}

func TestUserHandler_Tokens(t *testing.T) {
	setup := func(t *testing.T) (http.Handler, *store.UserStore, string) {
		t.Helper()
		ctx, handler, us, svc := newUsersServer(t)
		user, _, err := svc.GetOrCreate(ctx, "tok@example.com", "sub-tok")
		if err != nil {
			t.Fatalf("GetOrCreate: %v", err)
		}
		session, err := us.CreateSession(user.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}
		return handler, us, session
	}

	t.Run("list returns empty array initially", func(t *testing.T) {
		handler, _, session := setup(t)

		r := httptest.NewRequest(http.MethodGet, "/users/tokens", nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got []map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected empty list, got %d items", len(got))
		}
	})

	t.Run("create returns 201 with token", func(t *testing.T) {
		handler, _, session := setup(t)

		body := strings.NewReader(`{"name":"ci"}`)
		r := httptest.NewRequest(http.MethodPost, "/users/tokens", body)
		r.Header.Set("Authorization", "Bearer "+session)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Fatalf("got status %d, want %d: %s", w.Code, http.StatusCreated, w.Body.String())
		}
		var got map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		tok, _ := got["token"].(string)
		if !strings.HasPrefix(tok, "pat_") {
			t.Errorf("expected pat_ prefix, got %q", tok)
		}
		if got["name"] != "ci" {
			t.Errorf("got name %v, want %q", got["name"], "ci")
		}
	})

	t.Run("create with empty name returns 400", func(t *testing.T) {
		handler, _, session := setup(t)

		body := strings.NewReader(`{"name":""}`)
		r := httptest.NewRequest(http.MethodPost, "/users/tokens", body)
		r.Header.Set("Authorization", "Bearer "+session)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("list returns created token", func(t *testing.T) {
		handler, _, session := setup(t)

		body := strings.NewReader(`{"name":"listed"}`)
		r := httptest.NewRequest(http.MethodPost, "/users/tokens", body)
		r.Header.Set("Authorization", "Bearer "+session)
		r.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(httptest.NewRecorder(), r)

		r = httptest.NewRequest(http.MethodGet, "/users/tokens", nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		var got []map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 token, got %d", len(got))
		}
		if got[0]["name"] != "listed" {
			t.Errorf("got name %v, want %q", got[0]["name"], "listed")
		}
	})

	t.Run("delete removes token", func(t *testing.T) {
		handler, _, session := setup(t)

		// Create a token.
		body := strings.NewReader(`{"name":"to-delete"}`)
		r := httptest.NewRequest(http.MethodPost, "/users/tokens", body)
		r.Header.Set("Authorization", "Bearer "+session)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		var created map[string]any
		if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
			t.Fatalf("decode: %v", err)
		}
		id := created["id"].(string)

		// Delete it.
		r = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/tokens/%s", id), nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Confirm list is empty.
		r = httptest.NewRequest(http.MethodGet, "/users/tokens", nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		var listed []map[string]any
		if err := json.NewDecoder(w.Body).Decode(&listed); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(listed) != 0 {
			t.Errorf("expected 0 tokens after delete, got %d", len(listed))
		}
	})

	t.Run("delete unknown id returns 404", func(t *testing.T) {
		handler, _, session := setup(t)

		r := httptest.NewRequest(http.MethodDelete, "/users/tokens/00000000-0000-0000-0000-000000000000", nil)
		r.Header.Set("Authorization", "Bearer "+session)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("cannot list another user's tokens", func(t *testing.T) {
		ctx, handler, us, svc := newUsersServer(t)

		// User A creates a token.
		userA, _, err := svc.GetOrCreate(ctx, "a@example.com", "sub-a")
		if err != nil {
			t.Fatalf("GetOrCreate A: %v", err)
		}
		sessionA, err := us.CreateSession(userA.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession A: %v", err)
		}
		body := strings.NewReader(`{"name":"a-key"}`)
		r := httptest.NewRequest(http.MethodPost, "/users/tokens", body)
		r.Header.Set("Authorization", "Bearer "+sessionA)
		r.Header.Set("Content-Type", "application/json")
		handler.ServeHTTP(httptest.NewRecorder(), r)

		// User B lists tokens — should see none.
		userB, _, err := svc.GetOrCreate(ctx, "b@example.com", "sub-b")
		if err != nil {
			t.Fatalf("GetOrCreate B: %v", err)
		}
		sessionB, err := us.CreateSession(userB.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession B: %v", err)
		}
		r = httptest.NewRequest(http.MethodGet, "/users/tokens", nil)
		r.Header.Set("Authorization", "Bearer "+sessionB)
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		var got []map[string]any
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 0 {
			t.Errorf("expected user B to see 0 tokens, got %d", len(got))
		}
	})

	t.Run("cannot delete another user's token", func(t *testing.T) {
		ctx, handler, us, svc := newUsersServer(t)

		// User A creates a token.
		userA, _, err := svc.GetOrCreate(ctx, "a2@example.com", "sub-a2")
		if err != nil {
			t.Fatalf("GetOrCreate A: %v", err)
		}
		sessionA, err := us.CreateSession(userA.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession A: %v", err)
		}
		body := strings.NewReader(`{"name":"a-key"}`)
		r := httptest.NewRequest(http.MethodPost, "/users/tokens", body)
		r.Header.Set("Authorization", "Bearer "+sessionA)
		r.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, r)
		var created map[string]any
		if err := json.NewDecoder(w.Body).Decode(&created); err != nil {
			t.Fatalf("decode: %v", err)
		}
		tokenID := created["id"].(string)

		// User B attempts to delete A's token.
		userB, _, err := svc.GetOrCreate(ctx, "b2@example.com", "sub-b2")
		if err != nil {
			t.Fatalf("GetOrCreate B: %v", err)
		}
		sessionB, err := us.CreateSession(userB.ID, "", "", "")
		if err != nil {
			t.Fatalf("CreateSession B: %v", err)
		}
		r = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/users/tokens/%s", tokenID), nil)
		r.Header.Set("Authorization", "Bearer "+sessionB)
		w = httptest.NewRecorder()
		handler.ServeHTTP(w, r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})
}
