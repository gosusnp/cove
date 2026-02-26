package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gosusnp/cove/api/store"
)

type ProgramHandler struct {
	store *store.ProgramStore
}

func NewProgramHandler(s *store.ProgramStore) *ProgramHandler {
	return &ProgramHandler{store: s}
}

func (h *ProgramHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /programs", h.list)
	mux.HandleFunc("POST /programs", h.create)
	mux.HandleFunc("GET /programs/{id}", h.get)
	mux.HandleFunc("PUT /programs/{id}", h.update)
	mux.HandleFunc("DELETE /programs/{id}", h.delete)
}

type programRequest struct {
	Name string `json:"name"`
}

func (h *ProgramHandler) list(w http.ResponseWriter, r *http.Request) {
	programs, err := h.store.List()
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, programs)
}

func (h *ProgramHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	program, err := h.store.Get(id)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
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
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	program, err := h.store.Create(req.Name)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, program, http.StatusCreated)
}

func (h *ProgramHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req programRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		jsonError(w, "name is required", http.StatusBadRequest)
		return
	}
	program, err := h.store.Update(id, req.Name)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, program)
}

func (h *ProgramHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.store.Delete(id)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
