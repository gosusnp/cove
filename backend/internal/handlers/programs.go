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

type ProgramHandler struct {
	svc *service.ProgramService
}

func NewProgramHandler(s *service.ProgramService) *ProgramHandler {
	return &ProgramHandler{svc: s}
}

func (h *ProgramHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /programs", h.list)
	mux.HandleFunc("POST /programs", h.create)
	mux.HandleFunc("GET /programs/{id}", h.get)
	mux.HandleFunc("PUT /programs/{id}", h.update)
	mux.HandleFunc("DELETE /programs/{id}", h.delete)
	mux.HandleFunc("PUT /programs/{id}/structure", h.reorderStructure)
	mux.HandleFunc("GET /programs/{id}/versions", h.listVersions)
	mux.HandleFunc("GET /programs/{id}/versions/{vid}", h.getVersion)
	// POST …/rollback is an intentional action sub-resource; no CRUD equivalent captures the intent.
	mux.HandleFunc("POST /programs/{id}/versions/{vid}/rollback", h.rollbackVersion)
}

type programRequest struct {
	Name        string  `json:"name"`
	Description *string `json:"description,omitempty"`
	Activity    *string `json:"activity,omitempty"`
	IsPublic    bool    `json:"is_public"`
}

func (h *ProgramHandler) list(w http.ResponseWriter, r *http.Request) {
	programs, err := h.svc.List(r.Context())
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, programs)
}

func (h *ProgramHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.ProgramID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	program, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, program)
}

func (h *ProgramHandler) create(w http.ResponseWriter, r *http.Request) {
	var req programRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	program, err := h.svc.Create(r.Context(), req.Name, req.Description, req.Activity, req.IsPublic)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonResponse(w, program, http.StatusCreated)
}

func (h *ProgramHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.ProgramID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req programRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	program, err := h.svc.Update(r.Context(), id, req.Name, req.Description, req.Activity, req.IsPublic)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, program)
}

type programStructureEntry struct {
	SetID       int64   `json:"set_id"`
	ExerciseIDs []int64 `json:"exercise_ids"`
}

func (h *ProgramHandler) reorderStructure(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.ProgramID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req []programStructureEntry
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	entries := make([]service.StructureEntry, len(req))
	for i, e := range req {
		entries[i] = service.StructureEntry{SetID: e.SetID, ExerciseIDs: e.ExerciseIDs}
	}
	err = h.svc.ReorderStructure(r.Context(), id, entries)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProgramHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.ProgramID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.Delete(r.Context(), id)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
