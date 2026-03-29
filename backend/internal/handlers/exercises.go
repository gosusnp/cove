// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
)

type ExerciseHandler struct {
	svc *service.ExerciseService
}

func NewExerciseHandler(s *service.ExerciseService) *ExerciseHandler {
	return &ExerciseHandler{svc: s}
}

func (h *ExerciseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /exercises", h.list)
	mux.HandleFunc("POST /exercises", h.create)
	mux.HandleFunc("GET /exercises/{id}", h.get)
	mux.HandleFunc("PUT /exercises/{id}", h.update)
	mux.HandleFunc("DELETE /exercises/{id}", h.delete)
}

type exerciseRequest struct {
	UpdatedAt   *time.Time `json:"updated_at,omitempty"`
	Name        string     `json:"name"`
	Progression *string    `json:"progression,omitempty"`
	Description *string    `json:"description,omitempty"`
	IsPublic    bool       `json:"is_public"`
}

func (h *ExerciseHandler) list(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.svc.List(r.Context())
	if err != nil {
		handleServiceError(w, r, err, "exercise not found")
		return
	}
	jsonOK(w, exercises)
}

func (h *ExerciseHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.ExerciseID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	exercise, err := h.svc.Get(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err, "exercise not found")
		return
	}
	jsonOK(w, exercise)
}

func (h *ExerciseHandler) create(w http.ResponseWriter, r *http.Request) {
	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	exercise, err := h.svc.Create(r.Context(), req.Name, req.Progression, req.Description, req.IsPublic)
	if err != nil {
		handleServiceError(w, r, err, "exercise not found")
		return
	}
	jsonResponse(w, exercise, http.StatusCreated)
}

func (h *ExerciseHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.ExerciseID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	exercise, err := h.svc.Update(r.Context(), id, req.UpdatedAt, req.Name, req.Progression, req.Description, req.IsPublic)
	if errors.Is(err, service.ErrConflict) {
		current, err := h.svc.Get(r.Context(), id)
		if err != nil {
			handleServiceError(w, r, err, "exercise not found")
			return
		}
		jsonResponse(w, current, http.StatusConflict)
		return
	}
	if err != nil {
		handleServiceError(w, r, err, "exercise not found")
		return
	}
	jsonOK(w, exercise)
}

func (h *ExerciseHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.ExerciseID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err, "exercise not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
