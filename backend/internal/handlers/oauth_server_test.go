// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestOAuthServerHandler(t *testing.T) (*OAuthServerHandler, *http.ServeMux) {
	t.Helper()
	db := testutil.NewDB(t)
	uStore := store.NewUserStore()
	oStore := store.NewOAuthStore()
	orgStore := store.NewOrgStore()
	userSvc := service.NewUserService(db, uStore, orgStore)
	oauthSvc := service.NewOAuthService(db, oStore, uStore)
	h := NewOAuthServerHandler(oauthSvc, userSvc, false)
	mux := http.NewServeMux()
	h.RegisterRoutes(mux)
	return h, mux
}

// registerClient is a helper that registers a client and returns its client_id.
func registerClient(t *testing.T, mux *http.ServeMux, redirectURIs []string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]any{
		"client_name":   "Test Client",
		"redirect_uris": redirectURIs,
	})
	r := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(string(body)))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)
	if w.Code != http.StatusCreated {
		t.Fatalf("register client: got %d, want 201: %s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode register response: %v", err)
	}
	return resp["client_id"].(string)
}

func TestOAuthServerHandler_Metadata(t *testing.T) {
	_, mux := newTestOAuthServerHandler(t)
	r := httptest.NewRequest(http.MethodGet, "/.well-known/oauth-authorization-server", nil)
	r.Host = "example.com"
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", w.Code)
	}
	var meta map[string]any
	if err := json.NewDecoder(w.Body).Decode(&meta); err != nil {
		t.Fatalf("decode metadata: %v", err)
	}
	if meta["issuer"] != "http://example.com" {
		t.Errorf("issuer = %q, want %q", meta["issuer"], "http://example.com")
	}
	if meta["authorization_endpoint"] != "http://example.com/oauth/authorize" {
		t.Errorf("authorization_endpoint = %q", meta["authorization_endpoint"])
	}
	if meta["token_endpoint"] != "http://example.com/oauth/token" {
		t.Errorf("token_endpoint = %q", meta["token_endpoint"])
	}
}

func TestOAuthServerHandler_Register(t *testing.T) {
	t.Run("valid https redirect uri", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		body := `{"client_name":"Claude","redirect_uris":["https://claude.ai/callback"]}`
		r := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want 201: %s", w.Code, w.Body.String())
		}
		var resp map[string]any
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["client_id"] == "" {
			t.Error("expected non-empty client_id")
		}
	})

	t.Run("valid localhost redirect uri", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		body := `{"client_name":"Desktop","redirect_uris":["http://localhost:12345/callback"]}`
		r := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusCreated {
			t.Errorf("got %d, want 201: %s", w.Code, w.Body.String())
		}
	})

	t.Run("http non-loopback is rejected", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		body := `{"client_name":"Bad","redirect_uris":["http://evil.com/callback"]}`
		r := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})

	t.Run("missing redirect_uris returns 400", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		body := `{"client_name":"NoURIs","redirect_uris":[]}`
		r := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})

	t.Run("invalid json returns 400", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		r := httptest.NewRequest(http.MethodPost, "/oauth/register", strings.NewReader("not-json"))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})
}

func TestOAuthServerHandler_Authorize(t *testing.T) {
	t.Run("unknown client_id returns 400", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		r := httptest.NewRequest(http.MethodGet,
			"/oauth/authorize?client_id=unknown&redirect_uri=https://example.com/cb&response_type=code&code_challenge=abc&code_challenge_method=S256", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})

	t.Run("unregistered redirect_uri returns 400", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		clientID := registerClient(t, mux, []string{"https://claude.ai/callback"})

		r := httptest.NewRequest(http.MethodGet,
			fmt.Sprintf("/oauth/authorize?client_id=%s&redirect_uri=https://evil.com/cb&response_type=code&code_challenge=abc&code_challenge_method=S256", clientID), nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})

	t.Run("no session redirects to login and sets oauth_return_to cookie", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		clientID := registerClient(t, mux, []string{"https://claude.ai/callback"})

		authorizeURL := fmt.Sprintf("/oauth/authorize?client_id=%s&redirect_uri=https://claude.ai/callback&response_type=code&code_challenge=abc&code_challenge_method=S256&state=mystate", clientID)
		r := httptest.NewRequest(http.MethodGet, authorizeURL, nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusFound {
			t.Fatalf("got %d, want 302", w.Code)
		}
		if loc := w.Header().Get("Location"); loc != "/auth/login" {
			t.Errorf("redirect location = %q, want /auth/login", loc)
		}
		var returnToCookie *http.Cookie
		for _, c := range w.Result().Cookies() {
			if c.Name == oauthReturnToCookieName {
				returnToCookie = c
				break
			}
		}
		if returnToCookie == nil {
			t.Fatal("expected oauth_return_to cookie to be set")
		}
		if !strings.HasPrefix(returnToCookie.Value, "/oauth/authorize?") {
			t.Errorf("oauth_return_to = %q, want prefix /oauth/authorize?", returnToCookie.Value)
		}
		if !returnToCookie.HttpOnly {
			t.Error("expected HttpOnly cookie")
		}
	})

	t.Run("unsupported response_type returns 400", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		r := httptest.NewRequest(http.MethodGet,
			"/oauth/authorize?client_id=x&redirect_uri=https://x.com&response_type=token&code_challenge=abc&code_challenge_method=S256", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})

	t.Run("unsupported code_challenge_method returns 400", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		r := httptest.NewRequest(http.MethodGet,
			"/oauth/authorize?client_id=x&redirect_uri=https://x.com&response_type=code&code_challenge=abc&code_challenge_method=plain", nil)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})
}

func TestOAuthServerHandler_Token(t *testing.T) {
	t.Run("unsupported grant_type returns error", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		form := url.Values{"grant_type": {"client_credentials"}}
		r := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
		var resp map[string]string
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["error"] != "unsupported_grant_type" {
			t.Errorf("error = %q, want unsupported_grant_type", resp["error"])
		}
	})

	t.Run("invalid code returns invalid_grant", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		clientID := registerClient(t, mux, []string{"https://claude.ai/callback"})

		form := url.Values{
			"grant_type":    {"authorization_code"},
			"code":          {"invalid-code"},
			"redirect_uri":  {"https://claude.ai/callback"},
			"client_id":     {clientID},
			"code_verifier": {"someverifier"},
		}
		r := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
		var resp map[string]string
		_ = json.NewDecoder(w.Body).Decode(&resp)
		if resp["error"] != "invalid_grant" {
			t.Errorf("error = %q, want invalid_grant", resp["error"])
		}
	})
}

func TestOAuthServerHandler_Revoke(t *testing.T) {
	t.Run("missing token returns 400", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		r := httptest.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(""))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got %d, want 400", w.Code)
		}
	})

	t.Run("unknown token is silently accepted per RFC 7009", func(t *testing.T) {
		_, mux := newTestOAuthServerHandler(t)
		form := url.Values{"token": {"pat_nonexistent"}}
		r := httptest.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(form.Encode()))
		r.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusOK {
			t.Errorf("got %d, want 200", w.Code)
		}
	})
}

func TestOAuthServerHandler_FullFlow(t *testing.T) {
	// End-to-end: register → authorize (with session) → token → revoke
	app := NewTestApp(t)
	mux := http.NewServeMux()

	oauthStore := store.NewOAuthStore()
	oauthSvc := service.NewOAuthService(app.DB, oauthStore, app.UserStore)
	h := NewOAuthServerHandler(oauthSvc, app.Users, false)
	h.RegisterRoutes(mux)

	// 1. Register a client.
	clientID := registerClient(t, mux, []string{"https://claude.ai/callback"})

	// 2. Create a user and obtain a session token.
	userID := app.SeedUser("flowtest@example.com", "sub-flow")
	sessionToken, _, _, err := app.Users.CreateSession(t.Context(), userID, "127.0.0.1", "test", "test")
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	// PKCE: code_verifier and code_challenge (S256).
	codeVerifier := "dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk"
	// SHA256("dBjftJeZ4CVP-mB92K27uhbUJU1p1r_wW1gFWFOEjXk") base64url-encoded.
	codeChallenge := "E9Melhoa2OwvFrEMTJguCHaoeK1t8URWbuGJSstw-cM"

	// 3. Call authorize with a valid session.
	authorizeURL := fmt.Sprintf(
		"/oauth/authorize?client_id=%s&redirect_uri=https://claude.ai/callback&response_type=code&code_challenge=%s&code_challenge_method=S256&state=xyz",
		clientID, codeChallenge,
	)
	r := httptest.NewRequest(http.MethodGet, authorizeURL, nil)
	r.AddCookie(&http.Cookie{Name: sessionCookieName, Value: sessionToken})
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, r)

	if w.Code != http.StatusFound {
		t.Fatalf("authorize: got %d, want 302: %s", w.Code, w.Body.String())
	}
	loc := w.Header().Get("Location")
	if !strings.HasPrefix(loc, "https://claude.ai/callback?code=") {
		t.Fatalf("authorize redirect location = %q, expected code in redirect", loc)
	}
	parsed, err := url.Parse(loc)
	if err != nil {
		t.Fatalf("parse location: %v", err)
	}
	code := parsed.Query().Get("code")
	if code == "" {
		t.Fatal("expected code in redirect location")
	}
	if parsed.Query().Get("state") != "xyz" {
		t.Errorf("state = %q, want xyz", parsed.Query().Get("state"))
	}

	// 4. Exchange the code for a token.
	form := url.Values{
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {"https://claude.ai/callback"},
		"client_id":     {clientID},
		"code_verifier": {codeVerifier},
	}
	r2 := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	r2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w2 := httptest.NewRecorder()
	mux.ServeHTTP(w2, r2)

	if w2.Code != http.StatusOK {
		t.Fatalf("token: got %d, want 200: %s", w2.Code, w2.Body.String())
	}
	var tokenResp map[string]string
	if err := json.NewDecoder(w2.Body).Decode(&tokenResp); err != nil {
		t.Fatalf("decode token response: %v", err)
	}
	accessToken := tokenResp["access_token"]
	if !strings.HasPrefix(accessToken, "pat_") {
		t.Errorf("access_token = %q, expected pat_ prefix", accessToken)
	}
	if tokenResp["token_type"] != "Bearer" {
		t.Errorf("token_type = %q, want Bearer", tokenResp["token_type"])
	}
	if w2.Header().Get("Cache-Control") != "no-store" {
		t.Errorf("Cache-Control = %q, want no-store", w2.Header().Get("Cache-Control"))
	}

	// 5. Code replay is rejected.
	r3 := httptest.NewRequest(http.MethodPost, "/oauth/token", strings.NewReader(form.Encode()))
	r3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w3 := httptest.NewRecorder()
	mux.ServeHTTP(w3, r3)
	if w3.Code != http.StatusBadRequest {
		t.Errorf("replay: got %d, want 400", w3.Code)
	}

	// 6. Revoke the issued token.
	revokeForm := url.Values{"token": {accessToken}}
	r4 := httptest.NewRequest(http.MethodPost, "/oauth/revoke", strings.NewReader(revokeForm.Encode()))
	r4.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w4 := httptest.NewRecorder()
	mux.ServeHTTP(w4, r4)
	if w4.Code != http.StatusOK {
		t.Errorf("revoke: got %d, want 200", w4.Code)
	}

	// 7. Revoked token must not authenticate.
	_, _, _, authErr := app.Users.GetUserByToken(t.Context(), accessToken, "127.0.0.1", "test", "test")
	if authErr == nil {
		t.Error("revoked token should not authenticate")
	}
}
