// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

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
	Name        string  `json:"name"`
	Progression *string `json:"progression,omitempty"`
	Description *string `json:"description,omitempty"`
	IsPublic    bool    `json:"is_public"`
}

func (h *ExerciseHandler) list(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, err)
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
		h.handleError(w, err)
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
		h.handleError(w, err)
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
	exercise, err := h.svc.Update(r.Context(), id, req.Name, req.Progression, req.Description, req.IsPublic)
	if err != nil {
		h.handleError(w, err)
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
		h.handleError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ExerciseHandler) handleError(w http.ResponseWriter, err error) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "exercise not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	jsonError(w, err.Error(), http.StatusInternalServerError)
}
