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

func newTestOAuthHandler(t *testing.T, allowed []string, tokenURL, userinfoURL string) (*OAuthHandler, *http.ServeMux) {
	t.Helper()
	us := store.NewUserStore(testdb.New(t, containerDSN, db.MigrationsFS))
	cfg := &oauth2.Config{
		ClientID:     "test-client-id",
		ClientSecret: "test-client-secret",
		RedirectURL:  "http://localhost/auth/callback",
		Endpoint: oauth2.Endpoint{
			AuthURL:  "http://localhost/auth",
			TokenURL: tokenURL,
		},
	}
	h := NewOAuthHandler(cfg, us, allowed)
	h.userinfoURL = userinfoURL
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return h, mux
}

func TestOAuthHandler_Login(t *testing.T) {
	t.Run("redirects to provider", func(t *testing.T) {
		_, mux := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

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
		_, mux := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

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
		_, mux := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=somestate&code=somecode", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want %d", w.Code, http.StatusBadRequest)
		}
	})

	t.Run("state mismatch returns 400", func(t *testing.T) {
		_, mux := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")

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
		_, mux := newTestOAuthHandler(t, nil, tokenURL, userinfoURL)

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
		_, mux := newTestOAuthHandler(t, []string{"allowed@example.com"}, tokenURL, userinfoURL)

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
		_, mux := newTestOAuthHandler(t, []string{"allowed@example.com"}, tokenURL, userinfoURL)

		r := httptest.NewRequest(http.MethodGet, "/auth/callback?state=mystate&code=mycode", nil)
		r.AddCookie(&http.Cookie{Name: "oauth_state", Value: "mystate"})
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusFound {
			t.Errorf("got status %d, want %d: %s", w.Code, http.StatusFound, w.Body.String())
		}
	})
}
