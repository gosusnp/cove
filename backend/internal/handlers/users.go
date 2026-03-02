// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/httputil"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
)

// UserHandler handles user-related HTTP routes.
type UserHandler struct {
	svc *service.UserService
}

// NewUserHandler returns a new UserHandler.
func NewUserHandler(s *service.UserService) *UserHandler {
	return &UserHandler{svc: s}
}

// RegisterRoutes registers user routes on mux.
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/me", h.me)
	mux.HandleFunc("GET /users/tokens", h.listTokens)
	mux.HandleFunc("POST /users/tokens", h.createToken)
	mux.HandleFunc("DELETE /users/tokens/{id}", h.deleteToken)
}

type userResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *UserHandler) me(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	if authUser == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	user, err := h.svc.Get(authUser.ID)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "user not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
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
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Token     string    `json:"token"`
	CreatedAt time.Time `json:"created_at"`
}

type tokenResponse struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

func (h *UserHandler) listTokens(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	if authUser == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	pats, err := h.svc.ListPATs(authUser.ID)
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	resp := make([]tokenResponse, len(pats))
	for i, p := range pats {
		resp[i] = tokenResponse{ID: p.ID, Name: p.Name, CreatedAt: p.CreatedAt, LastUsedAt: p.LastUsedAt}
	}
	jsonOK(w, resp)
}

func (h *UserHandler) createToken(w http.ResponseWriter, r *http.Request) {
	authUser := middleware.UserFromContext(r.Context())
	authOrg := middleware.OrgFromContext(r.Context())
	if authUser == nil || authOrg == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var req createTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	ip, browser, os := httputil.FromRequest(r)
	token, pat, err := h.svc.CreatePAT(authUser.ID, authOrg.ID, req.Name, ip, browser, os)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Msg, http.StatusBadRequest)
		return
	}
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
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
	authUser := middleware.UserFromContext(r.Context())
	if authUser == nil {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeletePAT(authUser.ID, id); errors.Is(err, store.ErrNotFound) {
		jsonError(w, "token not found", http.StatusNotFound)
		return
	} else if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
