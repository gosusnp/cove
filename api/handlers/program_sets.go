package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gosusnp/cove/api/store"
)

type ProgramSetHandler struct {
	store *store.ProgramSetStore
}

func NewProgramSetHandler(s *store.ProgramSetStore) *ProgramSetHandler {
	return &ProgramSetHandler{store: s}
}

func (h *ProgramSetHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /programs/{program_id}/sets", h.list)
	mux.HandleFunc("POST /programs/{program_id}/sets", h.create)
	mux.HandleFunc("GET /programs/{program_id}/sets/{id}", h.get)
	mux.HandleFunc("PUT /programs/{program_id}/sets/{id}", h.update)
	mux.HandleFunc("DELETE /programs/{program_id}/sets/{id}", h.delete)
}

type programSetRequest struct {
	Name                *string `json:"name"`
	Rounds              int     `json:"rounds"`
	IntraSetRestSeconds *int    `json:"intra_set_rest_seconds"`
	SortOrder           *int    `json:"sort_order"`
}

func (h *ProgramSetHandler) list(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID(r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	sets, err := h.store.List(programID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, sets)
}

func (h *ProgramSetHandler) get(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID(r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	ps, err := h.store.Get(programID, id)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program set not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, ps)
}

func (h *ProgramSetHandler) create(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID(r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	var req programSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Rounds < 1 {
		req.Rounds = 1
	}
	ps, err := h.store.Create(programID, req.Name, req.Rounds, req.IntraSetRestSeconds, req.SortOrder)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonResponse(w, ps, http.StatusCreated)
}

func (h *ProgramSetHandler) update(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID(r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req programSetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	if req.Rounds < 1 {
		req.Rounds = 1
	}
	ps, err := h.store.Update(programID, id, req.Name, req.Rounds, req.IntraSetRestSeconds, req.SortOrder)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program set not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, ps)
}

func (h *ProgramSetHandler) delete(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID(r, "program_id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}
	id, err := pathID(r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.store.Delete(programID, id)
	if errors.Is(err, store.ErrNotFound) {
		jsonError(w, "program set not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
