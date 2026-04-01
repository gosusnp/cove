// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// seedAdminUser creates a user and grants them admin via direct DB update.
func seedAdminUser(t *testing.T, app *TestApp) domain.UserID {
	t.Helper()
	uid := app.SeedUser(fmt.Sprintf("admin-%s@test.com", t.Name()), fmt.Sprintf("admin-sub-%s", t.Name()))
	_, err := app.DB.Exec(`UPDATE cove.users SET is_admin = true WHERE id = $1`, uid)
	if err != nil {
		t.Fatalf("grant admin: %v", err)
	}
	return uid
}

func TestServiceAccounts_Create(t *testing.T) {
	app := NewTestApp(t)
	adminID := seedAdminUser(t, app)

	t.Run("creates a service account", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "My Bot"}, adminID)
		w := app.Do(r)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp serviceAccountResponse
		DecodeJSON(t, w, &resp)
		if resp.Name != "My Bot" {
			t.Errorf("expected name %q, got %q", "My Bot", resp.Name)
		}
		if resp.ID.String() == "" {
			t.Error("expected non-empty ID")
		}
	})

	t.Run("returns 400 for empty name", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "  "}, adminID)
		w := app.Do(r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("returns 403 for non-admin", func(t *testing.T) {
		userID := app.SeedUser("regular@test.com", "regular-sub")
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "Bot"}, userID)
		w := app.Do(r)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})

	t.Run("returns 401 for unauthenticated", func(t *testing.T) {
		r, _ := http.NewRequest("POST", "/api/admin/service-accounts", nil)
		w := app.Do(r)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})
}

func TestServiceAccounts_List(t *testing.T) {
	app := NewTestApp(t)
	adminID := seedAdminUser(t, app)

	t.Run("returns empty list initially", func(t *testing.T) {
		r := app.AuthRequest("GET", "/api/admin/service-accounts", nil, adminID)
		w := app.Do(r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var resp []serviceAccountResponse
		DecodeJSON(t, w, &resp)
		if len(resp) != 0 {
			t.Errorf("expected empty list, got %d items", len(resp))
		}
	})

	t.Run("lists created service accounts", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "CI Bot"}, adminID)
		w := app.Do(r)
		if w.Code != http.StatusCreated {
			t.Fatalf("create failed: %d %s", w.Code, w.Body.String())
		}

		r = app.AuthRequest("GET", "/api/admin/service-accounts", nil, adminID)
		w = app.Do(r)
		var resp []serviceAccountResponse
		DecodeJSON(t, w, &resp)
		if len(resp) != 1 {
			t.Fatalf("expected 1 account, got %d", len(resp))
		}
		if resp[0].Name != "CI Bot" {
			t.Errorf("expected %q, got %q", "CI Bot", resp[0].Name)
		}
	})

	t.Run("returns 403 for non-admin", func(t *testing.T) {
		userID := app.SeedUser("regular2@test.com", "regular-sub-2")
		r := app.AuthRequest("GET", "/api/admin/service-accounts", nil, userID)
		w := app.Do(r)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}

func TestServiceAccounts_Delete(t *testing.T) {
	app := NewTestApp(t)
	adminID := seedAdminUser(t, app)

	t.Run("deletes a service account", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "To Delete"}, adminID)
		w := app.Do(r)
		var created serviceAccountResponse
		DecodeJSON(t, w, &created)

		r = app.AuthRequest("DELETE", fmt.Sprintf("/api/admin/service-accounts/%s", created.ID), nil, adminID)
		w = app.Do(r)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}

		// Verify gone from list
		r = app.AuthRequest("GET", "/api/admin/service-accounts", nil, adminID)
		w = app.Do(r)
		var list []serviceAccountResponse
		DecodeJSON(t, w, &list)
		for _, a := range list {
			if a.ID == created.ID {
				t.Error("service account still in list after deletion")
			}
		}
	})

	t.Run("returns 404 for unknown id", func(t *testing.T) {
		r := app.AuthRequest("DELETE", "/api/admin/service-accounts/00000000-0000-0000-0000-000000000001", nil, adminID)
		w := app.Do(r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("also deletes associated tokens", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "SA With Token"}, adminID)
		w := app.Do(r)
		var sa serviceAccountResponse
		DecodeJSON(t, w, &sa)

		r = app.AuthRequest("POST", fmt.Sprintf("/api/admin/service-accounts/%s/tokens", sa.ID), map[string]string{"name": "my-token"}, adminID)
		w = app.Do(r)
		if w.Code != http.StatusCreated {
			t.Fatalf("create token failed: %d", w.Code)
		}

		r = app.AuthRequest("DELETE", fmt.Sprintf("/api/admin/service-accounts/%s", sa.ID), nil, adminID)
		w = app.Do(r)
		if w.Code != http.StatusNoContent {
			t.Fatalf("delete SA failed: %d %s", w.Code, w.Body.String())
		}
	})
}

func TestServiceAccounts_Tokens(t *testing.T) {
	app := NewTestApp(t)
	adminID := seedAdminUser(t, app)

	// Create a service account to use across token tests
	r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "Token Test SA"}, adminID)
	w := app.Do(r)
	var sa serviceAccountResponse
	DecodeJSON(t, w, &sa)
	saPath := fmt.Sprintf("/api/admin/service-accounts/%s", sa.ID)

	t.Run("creates a token and returns raw value once", func(t *testing.T) {
		r := app.AuthRequest("POST", saPath+"/tokens", map[string]string{"name": "deploy-key"}, adminID)
		w := app.Do(r)
		if w.Code != http.StatusCreated {
			t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
		}
		var resp createServiceAccountTokenResponse
		DecodeJSON(t, w, &resp)
		if resp.Name != "deploy-key" {
			t.Errorf("expected name %q, got %q", "deploy-key", resp.Name)
		}
		if len(resp.Token) == 0 {
			t.Error("expected non-empty token")
		}
		if resp.Token[:4] != "pat_" {
			t.Errorf("expected pat_ prefix, got %q", resp.Token[:4])
		}
	})

	t.Run("returns 400 for empty token name", func(t *testing.T) {
		r := app.AuthRequest("POST", saPath+"/tokens", map[string]string{"name": ""}, adminID)
		w := app.Do(r)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("lists tokens without raw values", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "List Token SA"}, adminID)
		w := app.Do(r)
		var freshSA serviceAccountResponse
		DecodeJSON(t, w, &freshSA)
		freshPath := fmt.Sprintf("/api/admin/service-accounts/%s", freshSA.ID)

		r = app.AuthRequest("POST", freshPath+"/tokens", map[string]string{"name": "ci-token"}, adminID)
		w = app.Do(r)
		if w.Code != http.StatusCreated {
			t.Fatalf("create token failed: %d", w.Code)
		}

		r = app.AuthRequest("GET", freshPath+"/tokens", nil, adminID)
		w = app.Do(r)
		if w.Code != http.StatusOK {
			t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
		}
		var tokens []tokenResponse
		DecodeJSON(t, w, &tokens)
		if len(tokens) != 1 {
			t.Fatalf("expected 1 token, got %d", len(tokens))
		}
		if tokens[0].Name != "ci-token" {
			t.Errorf("expected %q, got %q", "ci-token", tokens[0].Name)
		}
	})

	t.Run("deletes a token", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "Del Token SA"}, adminID)
		w := app.Do(r)
		var dSA serviceAccountResponse
		DecodeJSON(t, w, &dSA)
		dPath := fmt.Sprintf("/api/admin/service-accounts/%s", dSA.ID)

		r = app.AuthRequest("POST", dPath+"/tokens", map[string]string{"name": "temp"}, adminID)
		w = app.Do(r)
		var tok createServiceAccountTokenResponse
		DecodeJSON(t, w, &tok)

		r = app.AuthRequest("DELETE", fmt.Sprintf("%s/tokens/%s", dPath, tok.ID), nil, adminID)
		w = app.Do(r)
		if w.Code != http.StatusNoContent {
			t.Fatalf("expected 204, got %d: %s", w.Code, w.Body.String())
		}

		r = app.AuthRequest("GET", dPath+"/tokens", nil, adminID)
		w = app.Do(r)
		var tokens []tokenResponse
		DecodeJSON(t, w, &tokens)
		if len(tokens) != 0 {
			t.Errorf("expected 0 tokens after delete, got %d", len(tokens))
		}
	})

	t.Run("token delete returns 404 for unknown token", func(t *testing.T) {
		r := app.AuthRequest("DELETE", saPath+"/tokens/00000000-0000-0000-0000-000000000001", nil, adminID)
		w := app.Do(r)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("generated token authenticates as service account", func(t *testing.T) {
		r := app.AuthRequest("POST", "/api/admin/service-accounts", map[string]string{"name": "Auth SA"}, adminID)
		w := app.Do(r)
		var authSA serviceAccountResponse
		DecodeJSON(t, w, &authSA)

		r = app.AuthRequest("POST", fmt.Sprintf("/api/admin/service-accounts/%s/tokens", authSA.ID), map[string]string{"name": "auth-token"}, adminID)
		w = app.Do(r)
		var tok createServiceAccountTokenResponse
		DecodeJSON(t, w, &tok)

		// Use the raw token to call /api/users/me
		req, _ := http.NewRequest("GET", "/api/users/me", nil)
		req.Header.Set("Authorization", "Bearer "+tok.Token)
		resp := app.Do(req)
		if resp.Code != http.StatusOK {
			t.Fatalf("expected 200 from /api/users/me with SA token, got %d: %s", resp.Code, resp.Body.String())
		}
	})
}
