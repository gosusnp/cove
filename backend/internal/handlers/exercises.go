// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

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
}

func (h *ExerciseHandler) list(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.svc.List()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, exercises)
}

func (h *ExerciseHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[int64](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	exercise, err := h.svc.Get(id)
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
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
	exercise, err := h.svc.Create(req.Name, req.Progression)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, exercise, http.StatusCreated)
}

func (h *ExerciseHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[int64](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	exercise, err := h.svc.Update(id, req.Name, req.Progression)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, exercise)
}

func (h *ExerciseHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[int64](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.Delete(id)
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
