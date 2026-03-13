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
	"github.com/gosusnp/cove/backend/internal/store"
)

type WorkoutSessionHandler struct {
	svc *service.WorkoutSessionService
}

func NewWorkoutSessionHandler(s *service.WorkoutSessionService) *WorkoutSessionHandler {
	return &WorkoutSessionHandler{svc: s}
}

func (h *WorkoutSessionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /sessions", h.list)
	mux.HandleFunc("POST /sessions", h.create)
	mux.HandleFunc("GET /sessions/{id}", h.get)
	mux.HandleFunc("PATCH /sessions/{id}", h.update)
}

type workoutSessionRequest struct {
	ProgramID        *int64     `json:"program_id,omitempty"`
	ProgramName      *string    `json:"program_name,omitempty"`
	ProgramStructure *string    `json:"program_structure,omitempty"`
	Activity         *string    `json:"activity,omitempty"`
	DurationS        *int       `json:"duration_s,omitempty"`
	PerceivedEffort  *int       `json:"perceived_effort,omitempty"`
	SessionNotes     *string    `json:"session_notes,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

func (req *workoutSessionRequest) toParams() store.WorkoutSessionParams {
	p := store.WorkoutSessionParams{
		ProgramName:      req.ProgramName,
		ProgramStructure: req.ProgramStructure,
		Activity:         req.Activity,
		DurationS:        req.DurationS,
		PerceivedEffort:  req.PerceivedEffort,
		SessionNotes:     req.SessionNotes,
		StartedAt:        req.StartedAt,
		CompletedAt:      req.CompletedAt,
	}
	if req.ProgramID != nil {
		id := domain.ProgramID(*req.ProgramID)
		p.ProgramID = &id
	}
	return p
}

func (h *WorkoutSessionHandler) create(w http.ResponseWriter, r *http.Request) {
	var req workoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ws, err := h.svc.Create(r.Context(), req.toParams())
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		jsonError(w, "internal error", http.StatusInternalServerError)
		return
	}
	jsonResponse(w, ws, http.StatusCreated)
}

func (h *WorkoutSessionHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req workoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ws, err := h.svc.Update(r.Context(), id, req.toParams())
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
	jsonOK(w, ws)
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
