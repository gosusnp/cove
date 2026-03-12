// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"errors"
	"net/http"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
)

type WorkoutSessionHandler struct {
	svc *service.WorkoutSessionService
}

func NewWorkoutSessionHandler(s *service.WorkoutSessionService) *WorkoutSessionHandler {
	return &WorkoutSessionHandler{svc: s}
}

func (h *WorkoutSessionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /sessions", h.list)
	mux.HandleFunc("GET /sessions/{id}", h.get)
}

func (h *WorkoutSessionHandler) list(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.svc.List(r.Context())
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, sessions)
}

func (h *WorkoutSessionHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	session, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "session not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonOK(w, session)
}
