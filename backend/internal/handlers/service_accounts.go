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
	"github.com/gosusnp/cove/backend/internal/service"
)

// ServiceAccountHandler handles admin routes for service account management.
type ServiceAccountHandler struct {
	svc *service.UserService
}

// NewServiceAccountHandler returns a new ServiceAccountHandler.
func NewServiceAccountHandler(s *service.UserService) *ServiceAccountHandler {
	return &ServiceAccountHandler{svc: s}
}

// RegisterRoutes registers service account routes on mux.
// All routes are expected to be served under an admin-gated mux.
func (h *ServiceAccountHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /admin/service-accounts", h.create)
	mux.HandleFunc("GET /admin/service-accounts", h.list)
	mux.HandleFunc("DELETE /admin/service-accounts/{id}", h.delete)
	mux.HandleFunc("POST /admin/service-accounts/{id}/tokens", h.createToken)
	mux.HandleFunc("GET /admin/service-accounts/{id}/tokens", h.listTokens)
	mux.HandleFunc("DELETE /admin/service-accounts/{id}/tokens/{tokenId}", h.deleteToken)
}

type serviceAccountRequest struct {
	Name string `json:"name"`
}

type serviceAccountResponse struct {
	ID        domain.UserID `json:"id"`
	Name      string        `json:"name"`
	CreatedAt time.Time     `json:"created_at"`
}

func saToResponse(u *domain.User) serviceAccountResponse {
	name := ""
	if u.DisplayName != nil {
		name = *u.DisplayName
	}
	return serviceAccountResponse{ID: u.ID, Name: name, CreatedAt: u.CreatedAt}
}

func (h *ServiceAccountHandler) create(w http.ResponseWriter, r *http.Request) {
	var req serviceAccountRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	user, err := h.svc.CreateServiceAccount(r.Context(), req.Name)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Msg, http.StatusBadRequest)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonResponse(w, saToResponse(user), http.StatusCreated)
}

func (h *ServiceAccountHandler) list(w http.ResponseWriter, r *http.Request) {
	accounts, err := h.svc.ListServiceAccounts(r.Context())
	if err != nil {
		internalError(w, r, err)
		return
	}
	resp := make([]serviceAccountResponse, len(accounts))
	for i := range accounts {
		resp[i] = saToResponse(&accounts[i])
	}
	jsonOK(w, resp)
}

func (h *ServiceAccountHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	userID := domain.UserID{UUID: id}
	if err := h.svc.DeleteServiceAccount(r.Context(), userID); errors.Is(err, service.ErrNotFound) {
		jsonError(w, "service account not found", http.StatusNotFound)
		return
	} else if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

type createServiceAccountTokenRequest struct {
	Name string `json:"name"`
}

type createServiceAccountTokenResponse struct {
	ID        domain.TokenID `json:"id"`
	Name      string         `json:"name"`
	Token     string         `json:"token"`
	CreatedAt time.Time      `json:"created_at"`
}

func (h *ServiceAccountHandler) createToken(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	userID := domain.UserID{UUID: id}

	var req createServiceAccountTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, pat, err := h.svc.CreateServiceAccountPAT(r.Context(), userID, req.Name)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Msg, http.StatusBadRequest)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonResponse(w, createServiceAccountTokenResponse{
		ID:        pat.ID,
		Name:      pat.Name,
		Token:     token,
		CreatedAt: pat.CreatedAt,
	}, http.StatusCreated)
}

func (h *ServiceAccountHandler) listTokens(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	userID := domain.UserID{UUID: id}

	pats, err := h.svc.ListServiceAccountPATs(r.Context(), userID)
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

func (h *ServiceAccountHandler) deleteToken(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	userID := domain.UserID{UUID: id}

	tokenUUID, err := uuid.Parse(r.PathValue("tokenId"))
	if err != nil {
		jsonError(w, "invalid token id", http.StatusBadRequest)
		return
	}
	tokenID := domain.NewTokenID(tokenUUID)

	if err := h.svc.DeleteServiceAccountPAT(r.Context(), userID, tokenID); errors.Is(err, service.ErrNotFound) {
		jsonError(w, "token not found", http.StatusNotFound)
		return
	} else if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
