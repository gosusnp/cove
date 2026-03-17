// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"golang.org/x/oauth2"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/httputil"
	"github.com/gosusnp/cove/backend/internal/service"
)

const sessionCookieName = "cove_session"

// setSessionCookie writes an HttpOnly session cookie to the response.
func setSessionCookie(w http.ResponseWriter, token string, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		Path:     "/",
	})
}

// clearSessionCookie expires the session cookie.
func clearSessionCookie(w http.ResponseWriter, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
		Path:     "/",
	})
}

type OAuthHandler struct {
	config        *oauth2.Config
	userSvc       *service.UserService
	allowedEmails map[string]struct{}
	userinfoURL   string
	secureCookies bool
}

func (h *OAuthHandler) createSession(r *http.Request, userID domain.UserID) (string, error) {
	ip, browser, os := httputil.FromRequest(r)
	token, _, err := h.userSvc.CreateSession(r.Context(), userID, ip, browser, os)
	return token, err
}

func NewOAuthHandler(cfg *oauth2.Config, svc *service.UserService, allowed []string, secureCookies bool) *OAuthHandler {
	m := make(map[string]struct{}, len(allowed))
	for _, e := range allowed {
		m[e] = struct{}{}
	}
	return &OAuthHandler{
		config:        cfg,
		userSvc:       svc,
		allowedEmails: m,
		userinfoURL:   "https://www.googleapis.com/oauth2/v3/userinfo",
		secureCookies: secureCookies,
	}
}

func (h *OAuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /auth/login", h.login)
	mux.HandleFunc("GET /auth/callback", h.callback)
}

// RegisterDevRoutes adds a dev-only login endpoint.
// Only call this when COVE_DEV is set — do not expose in production.
func (h *OAuthHandler) RegisterDevRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/dev-login", h.devLogin)
}

func (h *OAuthHandler) devLogin(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Email == "" {
		jsonError(w, "email is required", http.StatusBadRequest)
		return
	}
	user, _, err := h.userSvc.GetOrCreate(
		r.Context(),
		domain.Email(req.Email),
		domain.GoogleSub("dev:"+req.Email),
	)
	if err != nil {
		internalError(w, r, fmt.Errorf("get or create user: %w", err))
		return
	}
	token, err := h.createSession(r, user.ID)
	if err != nil {
		internalError(w, r, fmt.Errorf("create session: %w", err))
		return
	}
	setSessionCookie(w, token, h.secureCookies)
	w.WriteHeader(http.StatusNoContent)
}

func (h *OAuthHandler) login(w http.ResponseWriter, r *http.Request) {
	state, err := randomHex(16)
	if err != nil {
		internalError(w, r, err)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     "oauth_state",
		Value:    state,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
		Path:     "/",
	})
	http.Redirect(w, r, h.config.AuthCodeURL(state), http.StatusFound)
}

func (h *OAuthHandler) callback(w http.ResponseWriter, r *http.Request) {
	// Validate state cookie
	cookie, err := r.Cookie("oauth_state")
	if err != nil || cookie.Value != r.URL.Query().Get("state") {
		jsonError(w, "invalid state", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	code := r.URL.Query().Get("code")
	oauthToken, err := h.config.Exchange(r.Context(), code)
	if err != nil {
		jsonError(w, "token exchange failed", http.StatusBadRequest)
		return
	}

	// Fetch user info from Google
	info, err := fetchUserInfo(r.Context(), h.config, oauthToken, h.userinfoURL)
	if err != nil {
		internalError(w, r, fmt.Errorf("fetch user info: %w", err))
		return
	}

	// Whitelist check
	if len(h.allowedEmails) > 0 {
		if _, ok := h.allowedEmails[info.Email]; !ok {
			jsonError(w, "not allowed", http.StatusForbidden)
			return
		}
	}

	// Get or create user + org
	user, _, err := h.userSvc.GetOrCreate(r.Context(), domain.Email(info.Email), domain.GoogleSub(info.Sub))
	if err != nil {
		internalError(w, r, fmt.Errorf("get or create user: %w", err))
		return
	}

	// Issue session
	sessionToken, err := h.createSession(r, user.ID)
	if err != nil {
		internalError(w, r, fmt.Errorf("create session: %w", err))
		return
	}

	setSessionCookie(w, sessionToken, h.secureCookies)

	// After OAuth login, complete a pending MCP authorization flow if one was started.
	if c, err := r.Cookie(oauthReturnToCookieName); err == nil {
		returnTo := c.Value
		if strings.HasPrefix(returnTo, "/oauth/authorize?") {
			http.SetCookie(w, &http.Cookie{
				Name:     oauthReturnToCookieName,
				Value:    "",
				HttpOnly: true,
				Secure:   h.secureCookies,
				SameSite: http.SameSiteLaxMode,
				MaxAge:   -1,
				Path:     "/",
			})
			http.Redirect(w, r, returnTo, http.StatusFound)
			return
		}
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

type googleUserInfo struct {
	Sub   string `json:"sub"`
	Email string `json:"email"`
}

func fetchUserInfo(ctx context.Context, cfg *oauth2.Config, token *oauth2.Token, userinfoURL string) (*googleUserInfo, error) {
	client := cfg.Client(ctx, token)
	resp, err := client.Get(userinfoURL)
	if err != nil {
		return nil, fmt.Errorf("get userinfo: %w", err)
	}
	defer resp.Body.Close()

	var info googleUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode userinfo: %w", err)
	}
	return &info, nil
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
