package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gosusnp/cove/api/store"
)

type ExerciseHandler struct {
	store *store.ExerciseStore
}

func NewExerciseHandler(s *store.ExerciseStore) *ExerciseHandler {
	return &ExerciseHandler{store: s}
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
	Progression *string `json:"progression"`
}

func (h *ExerciseHandler) list(w http.ResponseWriter, r *http.Request) {
	exercises, err := h.store.List()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, exercises)
}

func (h *ExerciseHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	exercise, err := h.store.Get(id)
	if errors.Is(err, store.ErrNotFound) {
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
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	exercise, err := h.store.Create(req.Name, req.Progression)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, exercise, http.StatusCreated)
}

func (h *ExerciseHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req exerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	exercise, err := h.store.Update(id, req.Name, req.Progression)
	if errors.Is(err, store.ErrNotFound) {
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
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.store.Delete(id)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "exercise not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
