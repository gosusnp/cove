// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func fakeTokenInfoServer(t *testing.T, clientID, email, sub, emailVerified string, statusCode int) string {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("id_token") == "" {
			http.Error(w, "missing id_token", http.StatusBadRequest)
			return
		}
		w.WriteHeader(statusCode)
		if statusCode == http.StatusOK {
			_ = json.NewEncoder(w).Encode(map[string]string{
				"aud":            clientID,
				"sub":            sub,
				"email":          email,
				"email_verified": emailVerified,
			})
		}
	}))
	t.Cleanup(srv.Close)
	return srv.URL
}

func TestOAuthHandler_MobileLogin(t *testing.T) {
	t.Run("valid id_token sets session cookie", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "mobile@example.com", "sub-mobile")
		h, mux, _ := newTestOAuthHandler(t, nil, tokenURL, userinfoURL)
		h.tokenInfoURL = fakeTokenInfoServer(t, "test-client-id", "mobile@example.com", "sub-mobile", "true", http.StatusOK)
		h.RegisterMobileRoutes(mux)

		body := `{"id_token":"valid-token"}`
		r := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(body))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want 204: %s", w.Code, w.Body.String())
		}
		var sessionCookie *http.Cookie
		for _, c := range w.Result().Cookies() {
			if c.Name == sessionCookieName {
				sessionCookie = c
				break
			}
		}
		if sessionCookie == nil {
			t.Fatal("expected cove_session cookie to be set")
		}
		if !strings.HasPrefix(sessionCookie.Value, "sess_") {
			t.Errorf("expected sess_ prefix, got %q", sessionCookie.Value)
		}
		if !sessionCookie.HttpOnly {
			t.Error("expected HttpOnly cookie")
		}
	})

	t.Run("missing id_token returns 400", func(t *testing.T) {
		h, mux, _ := newTestOAuthHandler(t, nil, "http://unused/token", "http://unused/userinfo")
		h.RegisterMobileRoutes(mux)

		r := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(`{}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusBadRequest {
			t.Errorf("got status %d, want 400", w.Code)
		}
	})

	t.Run("invalid id_token returns 401", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "mobile@example.com", "sub-mobile")
		h, mux, _ := newTestOAuthHandler(t, nil, tokenURL, userinfoURL)
		h.tokenInfoURL = fakeTokenInfoServer(t, "test-client-id", "", "", "", http.StatusBadRequest)
		h.RegisterMobileRoutes(mux)

		r := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(`{"id_token":"bad-token"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("audience mismatch returns 401", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "mobile@example.com", "sub-mobile")
		h, mux, _ := newTestOAuthHandler(t, nil, tokenURL, userinfoURL)
		h.tokenInfoURL = fakeTokenInfoServer(t, "wrong-client-id", "mobile@example.com", "sub-mobile", "true", http.StatusOK)
		h.RegisterMobileRoutes(mux)

		r := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(`{"id_token":"valid-token"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("unverified email returns 401", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "mobile@example.com", "sub-mobile")
		h, mux, _ := newTestOAuthHandler(t, nil, tokenURL, userinfoURL)
		h.tokenInfoURL = fakeTokenInfoServer(t, "test-client-id", "mobile@example.com", "sub-mobile", "false", http.StatusOK)
		h.RegisterMobileRoutes(mux)

		r := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(`{"id_token":"valid-token"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want 401", w.Code)
		}
	})

	t.Run("email not in whitelist returns 403", func(t *testing.T) {
		tokenURL, userinfoURL := fakeOAuthServer(t, "stranger@example.com", "sub-stranger")
		h, mux, _ := newTestOAuthHandler(t, []string{"allowed@example.com"}, tokenURL, userinfoURL)
		h.tokenInfoURL = fakeTokenInfoServer(t, "test-client-id", "stranger@example.com", "sub-stranger", "true", http.StatusOK)
		h.RegisterMobileRoutes(mux)

		r := httptest.NewRequest(http.MethodPost, "/auth/google/token", strings.NewReader(`{"id_token":"valid-token"}`))
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, r)

		if w.Code != http.StatusForbidden {
			t.Errorf("got status %d, want 403", w.Code)
		}
	})
}
