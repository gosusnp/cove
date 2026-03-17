// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/gosusnp/cove/backend/internal/httputil"
	"github.com/gosusnp/cove/backend/internal/service"
)

const oauthReturnToCookieName = "oauth_return_to"

// OAuthServerHandler implements an OAuth 2.0 authorization server for MCP clients.
// It supports Authorization Code flow with PKCE and dynamic client registration.
type OAuthServerHandler struct {
	svc           *service.OAuthService
	userSvc       *service.UserService
	secureCookies bool
}

// NewOAuthServerHandler returns a new OAuthServerHandler.
func NewOAuthServerHandler(svc *service.OAuthService, userSvc *service.UserService, secureCookies bool) *OAuthServerHandler {
	return &OAuthServerHandler{svc: svc, userSvc: userSvc, secureCookies: secureCookies}
}

// RegisterRoutes registers all OAuth 2.0 authorization server endpoints on mux.
// All routes are unauthenticated (the authorization endpoint validates the session itself).
func (h *OAuthServerHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /.well-known/oauth-authorization-server", h.metadata)
	mux.HandleFunc("POST /oauth/register", h.register)
	mux.HandleFunc("GET /oauth/authorize", h.authorize)
	mux.HandleFunc("POST /oauth/token", h.token)
	mux.HandleFunc("POST /oauth/revoke", h.revoke)
}

// metadata returns RFC 8414 server metadata.
func (h *OAuthServerHandler) metadata(w http.ResponseWriter, r *http.Request) {
	base := requestBase(r)
	jsonOK(w, map[string]any{
		"issuer":                                base,
		"authorization_endpoint":                base + "/oauth/authorize",
		"token_endpoint":                        base + "/oauth/token",
		"registration_endpoint":                 base + "/oauth/register",
		"revocation_endpoint":                   base + "/oauth/revoke",
		"response_types_supported":              []string{"code"},
		"grant_types_supported":                 []string{"authorization_code"},
		"code_challenge_methods_supported":      []string{"S256"},
		"token_endpoint_auth_methods_supported": []string{"none"},
	})
}

// registerRequest is the RFC 7591 dynamic registration request body.
type registerRequest struct {
	ClientName   string   `json:"client_name"`
	RedirectURIs []string `json:"redirect_uris"`
}

// register handles POST /oauth/register (RFC 7591 dynamic client registration).
func (h *OAuthServerHandler) register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	clientID, err := h.svc.RegisterClient(r.Context(), req.ClientName, req.RedirectURIs)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			jsonError(w, ve.Msg, http.StatusBadRequest)
			return
		}
		internalError(w, r, fmt.Errorf("register client: %w", err))
		return
	}

	jsonResponse(w, map[string]any{
		"client_id":     clientID,
		"client_name":   req.ClientName,
		"redirect_uris": req.RedirectURIs,
	}, http.StatusCreated)
}

// authorize handles GET /oauth/authorize (RFC 6749 §3.1).
func (h *OAuthServerHandler) authorize(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	clientID := q.Get("client_id")
	redirectURI := q.Get("redirect_uri")
	responseType := q.Get("response_type")
	state := q.Get("state")
	codeChallenge := q.Get("code_challenge")
	codeChallengeMethod := q.Get("code_challenge_method")

	// Validate required params before touching the redirect_uri to avoid open redirects.
	if clientID == "" || redirectURI == "" || codeChallenge == "" {
		jsonError(w, "missing required parameters", http.StatusBadRequest)
		return
	}
	if responseType != "code" {
		jsonError(w, "unsupported response_type", http.StatusBadRequest)
		return
	}
	if codeChallengeMethod != "S256" {
		jsonError(w, "unsupported code_challenge_method: only S256 is supported", http.StatusBadRequest)
		return
	}

	// Validate client and redirect_uri before checking session.
	// Per RFC 6749 §4.1.2.1, errors in client or redirect_uri must not redirect.
	if err := h.svc.ValidateClientRedirectURI(r.Context(), clientID, redirectURI); err != nil {
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "unknown client_id", http.StatusBadRequest)
			return
		}
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			jsonError(w, ve.Msg, http.StatusBadRequest)
			return
		}
		internalError(w, r, fmt.Errorf("validate client: %w", err))
		return
	}

	// Check for an existing session.
	var sessionToken string
	if c, err := r.Cookie(sessionCookieName); err == nil {
		sessionToken = c.Value
	}

	if sessionToken == "" {
		h.redirectToLogin(w, r)
		return
	}

	ip, browser, os := httputil.FromRequest(r)
	user, org, _, err := h.userSvc.GetUserByToken(r.Context(), sessionToken, ip, browser, os)
	if err != nil {
		// Session invalid or expired — send user to login.
		h.redirectToLogin(w, r)
		return
	}

	code, err := h.svc.Authorize(r.Context(), clientID, redirectURI, codeChallenge, user.ID, org.ID)
	if err != nil {
		// client/redirect_uri validation already passed above; any error here is internal.
		internalError(w, r, fmt.Errorf("authorize: %w", err))
		return
	}

	params := url.Values{"code": {code}}
	if state != "" {
		params.Set("state", state)
	}
	http.Redirect(w, r, redirectURI+"?"+params.Encode(), http.StatusFound)
}

// token handles POST /oauth/token (RFC 6749 §4.1.3).
func (h *OAuthServerHandler) token(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		jsonError(w, "invalid form body", http.StatusBadRequest)
		return
	}

	if r.FormValue("grant_type") != "authorization_code" {
		oauthError(w, "unsupported_grant_type", http.StatusBadRequest)
		return
	}

	code := r.FormValue("code")
	redirectURI := r.FormValue("redirect_uri")
	clientID := r.FormValue("client_id")
	codeVerifier := r.FormValue("code_verifier")

	if code == "" || redirectURI == "" || clientID == "" || codeVerifier == "" {
		oauthError(w, "invalid_request", http.StatusBadRequest)
		return
	}

	accessToken, err := h.svc.Exchange(r.Context(), code, redirectURI, clientID, codeVerifier)
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			oauthError(w, "invalid_grant", http.StatusBadRequest)
			return
		}
		internalError(w, r, fmt.Errorf("exchange: %w", err))
		return
	}

	w.Header().Set("Cache-Control", "no-store")
	jsonOK(w, map[string]string{
		"access_token": accessToken,
		"token_type":   "Bearer",
	})
}

// revoke handles POST /oauth/revoke (RFC 7009).
func (h *OAuthServerHandler) revoke(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		jsonError(w, "invalid form body", http.StatusBadRequest)
		return
	}

	token := r.FormValue("token")
	if token == "" {
		jsonError(w, "token is required", http.StatusBadRequest)
		return
	}

	if err := h.svc.Revoke(r.Context(), token); err != nil {
		internalError(w, r, fmt.Errorf("revoke: %w", err))
		return
	}

	w.WriteHeader(http.StatusOK)
}

// redirectToLogin stores the current authorize URL and redirects to /auth/login.
func (h *OAuthServerHandler) redirectToLogin(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:     oauthReturnToCookieName,
		Value:    r.URL.RequestURI(),
		HttpOnly: true,
		Secure:   h.secureCookies,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   600,
		Path:     "/",
	})
	http.Redirect(w, r, "/auth/login", http.StatusFound)
}

// oauthError responds with an RFC 6749 §5.2 token error response.
func oauthError(w http.ResponseWriter, code string, status int) {
	jsonResponse(w, map[string]string{"error": code}, status)
}

// requestBase derives the base URL (scheme + host) from an incoming request.
// It checks X-Forwarded-Proto for reverse proxy deployments.
func requestBase(r *http.Request) string {
	scheme := "https"
	if r.TLS == nil {
		if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
			scheme = proto
		} else {
			scheme = "http"
		}
	}
	return scheme + "://" + r.Host
}
