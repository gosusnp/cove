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

type ProgramExerciseHandler struct {
	svc *service.ProgramService
}

func NewProgramExerciseHandler(s *service.ProgramService) *ProgramExerciseHandler {
	return &ProgramExerciseHandler{svc: s}
}

func (h *ProgramExerciseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /programs/{program_id}/sets/{set_id}/exercises", h.list)
	mux.HandleFunc("POST /programs/{program_id}/sets/{set_id}/exercises", h.create)
	mux.HandleFunc("GET /programs/{program_id}/sets/{set_id}/exercises/{id}", h.get)
	mux.HandleFunc("PUT /programs/{program_id}/sets/{set_id}/exercises/{id}", h.update)
	mux.HandleFunc("DELETE /programs/{program_id}/sets/{set_id}/exercises/{id}", h.delete)
}

type programExerciseRequest struct {
	ExerciseID            domain.ExerciseID `json:"exercise_id"`
	Laterality            *string           `json:"laterality,omitempty"`
	TargetReps            *int              `json:"reps,omitempty"`
	TargetDurationSeconds *int              `json:"duration_s,omitempty"`
	TargetWeightKg        *float64          `json:"weight_kg,omitempty"`
}

func (h *ProgramExerciseHandler) list(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	setID, err := pathID[int64](r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	exercises, err := h.svc.ListExercises(r.Context(), programID, setID)
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program set not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, exercises)
}

func (h *ProgramExerciseHandler) get(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	setID, err := pathID[int64](r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	id, err := pathID[int64](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	pe, err := h.svc.GetExercise(r.Context(), programID, setID, id)
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, pe)
}

func (h *ProgramExerciseHandler) create(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	setID, err := pathID[int64](r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	var req programExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	pe, err := h.svc.CreateExercise(r.Context(), programID, setID, req.ExerciseID, req.Laterality, req.TargetReps, req.TargetDurationSeconds, req.TargetWeightKg)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program set not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonResponse(w, pe, http.StatusCreated)
}

func (h *ProgramExerciseHandler) update(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	setID, err := pathID[int64](r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	id, err := pathID[int64](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req programExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	pe, err := h.svc.UpdateExercise(r.Context(), programID, setID, id, req.ExerciseID, req.Laterality, req.TargetReps, req.TargetDurationSeconds, req.TargetWeightKg)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, pe)
}

func (h *ProgramExerciseHandler) delete(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	setID, err := pathID[int64](r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	id, err := pathID[int64](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.DeleteExercise(r.Context(), programID, setID, id)
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
