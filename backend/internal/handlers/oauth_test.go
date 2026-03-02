// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"golang.org/x/oauth2"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testdb"
)

// fakeOAuthServer returns a test server that handles both the token endpoint
// and the userinfo endpoint, along with the URLs for each.
func fakeOAuthServer(t *testing.T, email, sub string) (tokenURL, userinfoURL string) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"test-token","token_type":"Bearer","expires_in":3600}`))
		case "/userinfo":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]string{"sub": sub, "email": email})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL + "/token", srv.URL + "/userinfo"
}

func newTestOAuthHandler(t *testing.T, allowed []string, tokenURL, userinfoURL string) (*OAuthHandler, *http.ServeMux, *store.UserStore) {
	t.Helper()
	dbConn := testdb.New(t, containerDSN, db.MigrationsFS)
	us := store.NewUserStore(dbConn)
	orgs := store.NewOrgStore()
	svc := service.NewUserService(dbConn, us, orgs)
	cfg := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/auth/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost/auth",
			TokenURL: tokenURL,
		},
	}
	h := NewOAuthHandler(cfg, svc, allowed)
	h.userinfoURL = userinfoURL
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return h, mux, us
}

func TestOAuthHandler_Login(t *testing.T) {
	t.Run("redirects to provider", func(t *testing.T) {
		_, mux, _ := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

		r := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusFound)
		}
		if w.Header().Get("Location") == "" {
			t.Error("expected non-empty Location header")
		}
	})

	t.Run("sets oauth_state cookie", func(t *testing.T) {
		_, mux, _ := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

		r := httptest.NewRequest(http.MethodGet, "/auth/login", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		var stateCookie *http.Cookie
		for _, c := range w.Result().Cookies() {
			if c.Name == "oauth_state" {
				stateCookie = c
				break
			}
		}
		if stateCookie == nil {
			t.Fatal("oauth_state cookie not set")
		}
		if stateCookie.Value == "" {
			t.Error("expected non-empty state value")
		}
		if !stateCookie.HttpOnly {
			t.Error("expected HttpOnly cookie")
		}
	})
}

func TestOAuthHandler_Callback(t *testing.T) {
	t.Run("missing state cookie returns 400", func(t *testing.T) {
		_, mux, _ := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=somestate&code=somecode", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("state mismatch returns 400", func(t *testing.T) {
		_, mux, _ := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=wrong&code=somecode", nil)
		r.AddCookie(&http.Cookie{Name: "oauth_state", Value: "correct"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("valid flow redirects with session token", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "user@example.com", "sub-123")
		_, mux, _ := newTestOAuthHandler(t, nil, tokenURL, userinfoURL)

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=mystate&code=mycode", nil)
		r.AddCookie(&http.Cookie{Name: "oauth_state", Value: "mystate"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusFound {
			t.Errorf("got status %d, want %d: %s", w.Code, http.StatusFound, w.Body.String())
		}
		loc := w.Header().Get("Location")
		if loc == "" {
			t.Fatal("expected non-empty Location header")
		}
		if len(loc) <= len("/?token=") {
			t.Errorf("expected token in Location, got %q", loc)
		}
	})

	t.Run("email not in whitelist returns 403", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "stranger@example.com", "sub-stranger")
		_, mux, _ := newTestOAuthHandler(t, []string{"allowed@example.com"}, tokenURL, userinfoURL)

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=mystate&code=mycode", nil)
		r.AddCookie(&http.Cookie{Name: "oauth_state", Value: "mystate"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Errorf("got status %d, want %d", w.Code, http.StatusForbidden)
		}
	})

	t.Run("whitelisted email redirects with token", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "allowed@example.com", "sub-allowed")
		_, mux, _ := newTestOAuthHandler(t, []string{"allowed@example.com"}, tokenURL, userinfoURL)

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=mystate&code=mycode", nil)
		r.AddCookie(&http.Cookie{Name: "oauth_state", Value: "mystate"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusFound {
			t.Errorf("got status %d, want %d: %s", w.Code, http.StatusFound, w.Body.String())
		}
	})

	t.Run("captures session info", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "user-info@example.com", "sub-info")
		_, mux, us := newTestOAuthHandler(t, nil, tokenURL, userinfoURL)

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=mystate&code=mycode", nil)
		r.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
		r.RemoteAddr = "1.2.3.4:1234"
		r.AddCookie(&http.Cookie{Name: "oauth_state", Value: "mystate"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusFound {
			t.Fatalf("got status %d, want %d", w.Code, http.StatusFound)
		}

		var ip, browser, os string
		err := us.DB().QueryRow(`
			SELECT initial_ip_masked, initial_browser, initial_os
			FROM user_tokens t
			JOIN users u ON u.id = t.user_id
			WHERE u.email = 'user-info@example.com' AND t.kind = 'session'
		`).Scan(&ip, &browser, &os)
		if err != nil {
			t.Fatalf("query session info: %v", err)
		}

		if ip != "1.2.3.0" {
			t.Errorf("got ip %q, want %q", ip, "1.2.3.0")
		}
		if browser != "Chrome" {
			t.Errorf("got browser %q, want %q", browser, "Chrome")
		}
		if os != "macOS" {
			t.Errorf("got os %q, want %q", os, "macOS")
		}
	})
}
