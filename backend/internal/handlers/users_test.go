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

	"github.com/google/uuid"
)

func TestUserHandler_Logout(t *testing.T) {
	t.Run("success deletes current session", func(t *testing.T) {
		app := NewTestApp(t)
		userID := app.SeedUser("logout@example.com", "sub-logout")

		// We need a session token.
		token, _, err := app.Users.CreateSession(context.Background(), userID, "1.2.3.4", "agent", "OS")
		if err != nil {
			t.Fatalf("CreateSession: %v", err)
		}

		// Verify session exists.
		sessions, _ := app.Users.ListSessions(context.Background(), userID)
		if len(sessions) != 1 {
			t.Fatalf("expected 1 session, got %d", len(sessions))
		}

		// Logout via Mux (with OAuth middleware)
		r := httptest.NewRequest(http.MethodPost, "/api/users/logout", nil)
		r.Header.Set("Authorization", "Bearer "+token)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify session is gone.
		sessions, _ = app.Users.ListSessions(context.Background(), userID)
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions after logout, got %d", len(sessions))
		}
	})

	t.Run("returns 401 for missing token", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodPost, "/api/users/logout", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}

func TestUserHandler_Me(t *testing.T) {
	t.Run("returns current user", func(t *testing.T) {
		app := NewTestApp(t)
		userID := app.SeedUser("me@example.com", "sub-me")

		r := app.AuthRequest(http.MethodGet, "/api/users/me", nil, userID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}

		var got userResponse
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.ID != userID {
			t.Errorf("got ID %v, want %v", got.ID, userID)
		}
		if got.Email != "me@example.com" {
			t.Errorf("got email %q, want %q", got.Email, "me@example.com")
		}
	})

	t.Run("missing token returns 401", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
		w := app.Do(r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("invalid token returns 401", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/users/me", nil)
		r.Header.Set("Authorization", "Bearer invalid")
		w := app.Do(r)
		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})
}

func TestUserHandler_Sessions(t *testing.T) {
	t.Run("list returns active sessions with is_current", func(t *testing.T) {
		app := NewTestApp(t)
		userID := app.SeedUser("sess@example.com", "sub-sess")

		r := app.AuthRequest(http.MethodGet, "/api/users/sessions", nil, userID)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Fatalf("got status %d, want 200", w.Code)
		}
		var got []sessionResponse
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if len(got) != 1 {
			t.Fatalf("expected 1 session, got %d", len(got))
		}
		if !got[0].IsCurrent {
			t.Errorf("expected is_current=true")
		}
	})

	t.Run("delete removes session", func(t *testing.T) {
		app := NewTestApp(t)
		userID := app.SeedUser("sess-del@example.com", "sub-sess-del")

		// Get current session ID
		r := app.AuthRequest(http.MethodGet, "/api/users/sessions", nil, userID)
		w := app.Do(r)
		var sessions []sessionResponse
		_ = json.NewDecoder(w.Body).Decode(&sessions)
		id := sessions[0].ID

		// Delete it
		r = app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/users/sessions/%s", id.UUID), nil, userID)
		w = app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want 204", w.Code)
		}
	})

	t.Run("cannot list another user's sessions", func(t *testing.T) {
		app := NewTestApp(t)
		userA := app.SeedUser("a-sess@example.com", "sub-a-sess")
		userB := app.SeedUser("b-sess@example.com", "sub-b-sess")

		// User A creates an extra session (outside of AuthRequest).
		_, _, _ = app.Users.CreateSession(context.Background(), userA, "1.1.1.1", "agent", "OS")

		// User B lists sessions — should only see their own (1 from AuthRequest), not A's.
		r := app.AuthRequest(http.MethodGet, "/api/users/sessions", nil, userB)
		w := app.Do(r)

		var got []sessionResponse
		_ = json.NewDecoder(w.Body).Decode(&got)
		if len(got) != 1 {
			t.Errorf("expected 1 session (own), got %d", len(got))
		}
	})

	t.Run("cannot delete another user's session", func(t *testing.T) {
		app := NewTestApp(t)
		userA := app.SeedUser("a@example.com", "sub-a")
		userB := app.SeedUser("b@example.com", "sub-b")

		// User A has a session
		tokenA, _, _ := app.Users.CreateSession(context.Background(), userA, "1.1.1.1", "agent", "OS")
		sessionsA, _ := app.Users.ListSessions(context.Background(), userA)
		idA := sessionsA[0].ID

		// User B attempts to delete A's session
		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/users/sessions/%s", idA.UUID), nil, userB)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404", w.Code)
		}

		// Verify A's session still exists
		sessionsA, _ = app.Users.ListSessions(context.Background(), userA)
		if len(sessionsA) != 1 {
			t.Error("user A session was deleted by user B")
		}
		_ = tokenA
	})
}

func TestUserHandler_Tokens(t *testing.T) {
	t.Run("full lifecycle", func(t *testing.T) {
		app := NewTestApp(t)
		userID := app.SeedUser("tok@example.com", "sub-tok")

		// 1. List empty
		r := app.AuthRequest(http.MethodGet, "/api/users/tokens", nil, userID)
		w := app.Do(r)
		var got []tokenResponse
		_ = json.NewDecoder(w.Body).Decode(&got)
		if len(got) != 0 {
			t.Errorf("expected 0 tokens, got %d", len(got))
		}

		// 2. Create
		body := `{"name":"test-token"}`
		r = app.AuthRequest(http.MethodPost, "/api/users/tokens", strings.NewReader(body), userID)
		w = app.Do(r)
		if w.Code != http.StatusCreated {
			t.Fatalf("got status %d, want 201: %s", w.Code, w.Body.String())
		}
		var created createTokenResponse
		_ = json.NewDecoder(w.Body).Decode(&created)
		if !strings.HasPrefix(created.Token, "pat_") {
			t.Errorf("expected pat_ prefix, got %q", created.Token)
		}

		// 3. List contains created
		r = app.AuthRequest(http.MethodGet, "/api/users/tokens", nil, userID)
		w = app.Do(r)
		_ = json.NewDecoder(w.Body).Decode(&got)
		if len(got) != 1 || got[0].Name != "test-token" {
			t.Errorf("expected token in list, got %v", got)
		}

		// 4. Delete
		r = app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/users/tokens/%s", created.ID.UUID), nil, userID)
		w = app.Do(r)
		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want 204", w.Code)
		}

		// 5. List empty again
		r = app.AuthRequest(http.MethodGet, "/api/users/tokens", nil, userID)
		w = app.Do(r)
		_ = json.NewDecoder(w.Body).Decode(&got)
		if len(got) != 0 {
			t.Errorf("expected 0 tokens, got %d", len(got))
		}
	})

	t.Run("cannot list another user's tokens", func(t *testing.T) {
		app := NewTestApp(t)
		userA := app.SeedUser("a-tok2@example.com", "sub-a-tok2")
		userB := app.SeedUser("b-tok2@example.com", "sub-b-tok2")

		// User A creates a token.
		r := app.AuthRequest(http.MethodPost, "/api/users/tokens", strings.NewReader(`{"name":"a-key"}`), userA)
		w := app.Do(r)
		if w.Code != http.StatusCreated {
			t.Fatalf("setup: got %d", w.Code)
		}

		// User B lists tokens — should see none.
		r = app.AuthRequest(http.MethodGet, "/api/users/tokens", nil, userB)
		w = app.Do(r)

		var got []tokenResponse
		_ = json.NewDecoder(w.Body).Decode(&got)
		if len(got) != 0 {
			t.Errorf("expected 0 tokens, got %d", len(got))
		}
	})

	t.Run("delete unknown id returns 404", func(t *testing.T) {
		app := NewTestApp(t)
		userID := app.SeedUser("tok-del-404@example.com", "sub-tok-del-404")

		r := app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/users/tokens/%s", uuid.New()), nil, userID)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404", w.Code)
		}
	})

	t.Run("create with empty name returns 400", func(t *testing.T) {
		app := NewTestApp(t)
		userID := app.SeedUser("tok-empty@example.com", "sub-tok-empty")

		r := app.AuthRequest(http.MethodPost, "/api/users/tokens", strings.NewReader(`{"name":""}`), userID)
		w := app.Do(r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}
	})

	t.Run("cannot delete another user's token", func(t *testing.T) {
		app := NewTestApp(t)
		userA := app.SeedUser("a-tok@example.com", "sub-a-tok")
		userB := app.SeedUser("b-tok@example.com", "sub-b-tok")

		// User A creates a token
		r := app.AuthRequest(http.MethodPost, "/api/users/tokens", strings.NewReader(`{"name":"a-key"}`), userA)
		w := app.Do(r)
		var created createTokenResponse
		_ = json.NewDecoder(w.Body).Decode(&created)

		// User B attempts to delete A's token
		r = app.AuthRequest(http.MethodDelete, fmt.Sprintf("/api/users/tokens/%s", created.ID.UUID), nil, userB)
		w = app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want 404", w.Code)
		}
	})
}
