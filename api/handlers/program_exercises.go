package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gosusnp/cove/api/store"
)

type ProgramExerciseHandler struct {
	store *store.ProgramExerciseStore
}

func NewProgramExerciseHandler(s *store.ProgramExerciseStore) *ProgramExerciseHandler {
	return &ProgramExerciseHandler{store: s}
}

func (h *ProgramExerciseHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /programs/{program_id}/sets/{set_id}/exercises", h.list)
	mux.HandleFunc("POST /programs/{program_id}/sets/{set_id}/exercises", h.create)
	mux.HandleFunc("GET /programs/{program_id}/sets/{set_id}/exercises/{id}", h.get)
	mux.HandleFunc("PUT /programs/{program_id}/sets/{set_id}/exercises/{id}", h.update)
	mux.HandleFunc("DELETE /programs/{program_id}/sets/{set_id}/exercises/{id}", h.delete)
}

type programExerciseRequest struct {
	ExerciseID            int64    `json:"exercise_id"`
	Laterality            *string  `json:"laterality"`
	TargetReps            *int     `json:"target_reps"`
	TargetDurationSeconds *int     `json:"target_duration_seconds"`
	TargetWeightKg        *float64 `json:"target_weight_kg"`
	SortOrder             *int     `json:"sort_order"`
}

func (h *ProgramExerciseHandler) list(w http.ResponseWriter, r *http.Request) {
	setID, err := pathID(r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	exercises, err := h.store.List(setID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, exercises)
}

func (h *ProgramExerciseHandler) get(w http.ResponseWriter, r *http.Request) {
	setID, err := pathID(r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	pe, err := h.store.Get(setID, id)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, pe)
}

func (h *ProgramExerciseHandler) create(w http.ResponseWriter, r *http.Request) {
	setID, err := pathID(r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	var req programExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.ExerciseID == 0 {
		jsonError(w, "exercise_id is required", http.StatusBadRequest)
		return
	}
	pe, err := h.store.Create(setID, req.ExerciseID, req.Laterality, req.TargetReps, req.TargetDurationSeconds, req.TargetWeightKg, req.SortOrder)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, pe, http.StatusCreated)
}

func (h *ProgramExerciseHandler) update(w http.ResponseWriter, r *http.Request) {
	setID, err := pathID(r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req programExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.ExerciseID == 0 {
		jsonError(w, "exercise_id is required", http.StatusBadRequest)
		return
	}
	pe, err := h.store.Update(setID, id, req.ExerciseID, req.Laterality, req.TargetReps, req.TargetDurationSeconds, req.TargetWeightKg, req.SortOrder)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, pe)
}

func (h *ProgramExerciseHandler) delete(w http.ResponseWriter, r *http.Request) {
	setID, err := pathID(r, "set_id")
	if err != nil {
		jsonError(w, "invalid set id", http.StatusBadRequest)
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.store.Delete(setID, id)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
