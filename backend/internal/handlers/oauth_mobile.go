// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// RegisterMobileRoutes adds the mobile OAuth endpoint used by the Android/iOS apps.
func (h *OAuthHandler) RegisterMobileRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /auth/google/token", h.mobileLogin)
}

type mobileLoginRequest struct {
	IDToken string `json:"id_token"`
}

type googleTokenInfo struct {
	Aud           string `json:"aud"`
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	EmailVerified string `json:"email_verified"`
}

// mobileLogin accepts a Google ID token from the native app, verifies it,
// and issues a session cookie identical to the web OAuth flow.
func (h *OAuthHandler) mobileLogin(w http.ResponseWriter, r *http.Request) {
	var req mobileLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		jsonError(w, "id_token is required", http.StatusBadRequest)
		return
	}

	info, err := h.verifyIDToken(r.Context(), req.IDToken)
	if err != nil {
		jsonError(w, "invalid id_token", http.StatusUnauthorized)
		return
	}

	if info.Aud != h.config.ClientID {
		jsonError(w, "invalid id_token audience", http.StatusUnauthorized)
		return
	}

	if info.EmailVerified != "true" {
		jsonError(w, "email not verified", http.StatusUnauthorized)
		return
	}

	if len(h.allowedEmails) > 0 {
		if _, ok := h.allowedEmails[info.Email]; !ok {
			jsonError(w, "not allowed", http.StatusForbidden)
			return
		}
	}

	user, _, err := h.userSvc.GetOrCreate(r.Context(), domain.Email(info.Email), domain.GoogleSub(info.Sub))
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

func (h *OAuthHandler) verifyIDToken(ctx context.Context, idToken string) (*googleTokenInfo, error) {
	params := url.Values{"id_token": {idToken}}
	reqURL := h.tokenInfoURL + "?" + params.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tokeninfo request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tokeninfo returned %d", resp.StatusCode)
	}

	var info googleTokenInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, fmt.Errorf("decode tokeninfo: %w", err)
	}
	return &info, nil
}
