// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/httputil"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
)

// UserHandler handles user-related HTTP routes.
type UserHandler struct {
	svc           *service.UserService
	secureCookies bool
}

// NewUserHandler returns a new UserHandler.
func NewUserHandler(s *service.UserService, secureCookies bool) *UserHandler {
	return &UserHandler{svc: s, secureCookies: secureCookies}
}

// RegisterRoutes registers user routes on mux.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/me", h.me)
	mux.HandleFunc("POST /users/logout", h.logout)
	mux.HandleFunc("GET /users/tokens", h.listTokens)
	mux.HandleFunc("POST /users/tokens", h.createToken)
	mux.HandleFunc("DELETE /users/tokens/{id}", h.deleteToken)
	mux.HandleFunc("GET /users/sessions", h.listSessions)
	mux.HandleFunc("DELETE /users/sessions/{id}", h.deleteSession)
}

func (h *UserHandler) logout(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID.UUID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	tokenID := middleware.TokenIDFromContext(r.Context())
	if err := h.svc.DeleteSession(r.Context(), userID, domain.SessionID{UUID: tokenID}); err != nil && !errors.Is(err, service.ErrNotFound) {
		internalError(w, r, err)
		return
	}
	clearSessionCookie(w, h.secureCookies)
	w.WriteHeader(http.StatusNoContent)
}

type userResponse struct {
	ID        domain.UserID `json:"id"`
	Email     domain.Email  `json:"email"`
	CreatedAt time.Time     `json:"created_at"`
}

func (h *UserHandler) me(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID.UUID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.svc.Get(r.Context(), userID)
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, userResponse{
		ID:        user.ID,
		Email:     user.Email,
		CreatedAt: user.CreatedAt,
	})
}

type createTokenRequest struct {
	Name string `json:"name"`
}

type createTokenResponse struct {
	ID        domain.TokenID `json:"id"`
	Name      string         `json:"name"`
	Token     string         `json:"token"`
	CreatedAt time.Time      `json:"created_at"`
}

type tokenResponse struct {
	ID         domain.TokenID `json:"id"`
	Name       string         `json:"name"`
	CreatedAt  time.Time      `json:"created_at"`
	LastUsedAt *time.Time     `json:"last_used_at"`
}

type sessionResponse struct {
	ID              domain.SessionID `json:"id"`
	CreatedAt       time.Time        `json:"created_at"`
	LastUsedAt      *time.Time       `json:"last_used_at"`
	InitialIPMasked *domain.MaskedIP `json:"initial_ip_masked,omitempty"`
	InitialBrowser  *string          `json:"initial_browser,omitempty"`
	InitialOS       *string          `json:"initial_os,omitempty"`
	LastIPMasked    *domain.MaskedIP `json:"last_ip_masked,omitempty"`
	LastBrowser     *string          `json:"last_browser,omitempty"`
	LastOS          *string          `json:"last_os,omitempty"`
	IsCurrent       bool             `json:"is_current"`
}

func (h *UserHandler) listTokens(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID.UUID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	pats, err := h.svc.ListPATs(r.Context(), userID)
	if err != nil {
		internalError(w, r, err)
		return
	}
	resp := make([]tokenResponse, len(pats))
	for i, p := range pats {
		resp[i] = tokenResponse{ID: p.ID, Name: p.Name, CreatedAt: p.CreatedAt, LastUsedAt: p.LastUsedAt}
	}
	jsonOK(w, resp)
}

func (h *UserHandler) createToken(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	orgID := middleware.OrgIDFromContext(r.Context())
	if userID.UUID == uuid.Nil || orgID.UUID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req createTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ip, browser, os := httputil.FromRequest(r)
	token, pat, err := h.svc.CreatePAT(r.Context(), userID, orgID, req.Name, ip, browser, os)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Msg, http.StatusBadRequest)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonResponse(w, createTokenResponse{
		ID:        pat.ID,
		Name:      pat.Name,
		Token:     token,
		CreatedAt: pat.CreatedAt,
	}, http.StatusCreated)
}

func (h *UserHandler) deleteToken(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID.UUID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	tokenID := domain.NewTokenID(id)
	if err := h.svc.DeletePAT(r.Context(), userID, tokenID); errors.Is(err, service.ErrNotFound) {
		jsonError(w, "token not found", http.StatusNotFound)
		return
	} else if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) listSessions(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID.UUID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	tokenID := middleware.TokenIDFromContext(r.Context())
	sessions, err := h.svc.ListSessions(r.Context(), userID)
	if err != nil {
		internalError(w, r, err)
		return
	}
	resp := make([]sessionResponse, len(sessions))
	for i, s := range sessions {
		resp[i] = sessionResponse{
			ID:              s.ID,
			CreatedAt:       s.CreatedAt,
			LastUsedAt:      s.LastUsedAt,
			InitialIPMasked: s.InitialIPMasked,
			InitialBrowser:  s.InitialBrowser,
			InitialOS:       s.InitialOS,
			LastIPMasked:    s.LastIPMasked,
			LastBrowser:     s.LastBrowser,
			LastOS:          s.LastOS,
			IsCurrent:       s.ID.UUID == tokenID,
		}
	}
	jsonOK(w, resp)
}

func (h *UserHandler) deleteSession(w http.ResponseWriter, r *http.Request) {
	userID := middleware.UserIDFromContext(r.Context())
	if userID.UUID == uuid.Nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	sessionID := domain.SessionID{UUID: id}
	if err := h.svc.DeleteSession(r.Context(), userID, sessionID); errors.Is(err, service.ErrNotFound) {
		jsonError(w, "session not found", http.StatusNotFound)
		return
	} else if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
