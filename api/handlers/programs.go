package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/gosusnp/cove/api/service"
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
}

type programRequest struct {
	Name string `json:"name"`
}

func (h *ProgramHandler) list(w http.ResponseWriter, r *http.Request) {
	programs, err := h.svc.List()
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
	program, err := h.svc.GetDetail(id)
	if errors.Is(err, service.ErrNotFound) {
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
	program, err := h.svc.Create(req.Name)
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
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
	program, err := h.svc.Update(id, req.Name)
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
	err = h.svc.Delete(id)
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
