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

type PreparationHandler struct {
	svc *service.PreparationService
}

func NewPreparationHandler(s *service.PreparationService) *PreparationHandler {
	return &PreparationHandler{svc: s}
}

func (h *PreparationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /preparations", h.list)
	mux.HandleFunc("POST /preparations", h.create)
	mux.HandleFunc("GET /preparations/{id}", h.get)
	mux.HandleFunc("PUT /preparations/{id}", h.update)
	mux.HandleFunc("DELETE /preparations/{id}", h.delete)
	mux.HandleFunc("POST /preparations/{id}/ingredients", h.addIngredient)
	mux.HandleFunc("PUT /preparations/{id}/ingredients/{ingId}", h.updateIngredient)
	mux.HandleFunc("DELETE /preparations/{id}/ingredients/{ingId}", h.deleteIngredient)
}

type preparationStepRequest struct {
	Description string `json:"description"`
}

type preparationRequest struct {
	Name        string                   `json:"name"`
	Description *string                  `json:"description,omitempty"`
	YieldAmount float64                  `json:"yield_amount"`
	YieldUnit   string                   `json:"yield_unit"`
	Steps       []preparationStepRequest `json:"steps"`
	IsPublic    bool                     `json:"is_public"`
}

type preparationIngredientRequest struct {
	IngredientID int64   `json:"ingredient_id"`
	Name         string  `json:"name"`
	Amount       float64 `json:"amount"`
	Unit         string  `json:"unit"`
	Prep         *string `json:"prep,omitempty"`
}

func (h *PreparationHandler) list(w http.ResponseWriter, r *http.Request) {
	preparations, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, preparations)
}

func (h *PreparationHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.PreparationID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	preparation, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, preparation)
}

func (h *PreparationHandler) create(w http.ResponseWriter, r *http.Request) {
	var req preparationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	steps := make([]domain.PreparationStep, len(req.Steps))
	for i, s := range req.Steps {
		steps[i] = domain.PreparationStep{Description: s.Description}
	}
	preparation, err := h.svc.Create(r.Context(), domain.PreparationParams{
		Name:        req.Name,
		Description: req.Description,
		YieldAmount: req.YieldAmount,
		YieldUnit:   req.YieldUnit,
		Steps:       steps,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonResponse(w, preparation, http.StatusCreated)
}

func (h *PreparationHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.PreparationID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req preparationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	steps := make([]domain.PreparationStep, len(req.Steps))
	for i, s := range req.Steps {
		steps[i] = domain.PreparationStep{Description: s.Description}
	}
	preparation, err := h.svc.Update(r.Context(), id, domain.PreparationParams{
		Name:        req.Name,
		Description: req.Description,
		YieldAmount: req.YieldAmount,
		YieldUnit:   req.YieldUnit,
		Steps:       steps,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, preparation)
}

func (h *PreparationHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.PreparationID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	if err := h.svc.Delete(r.Context(), id); err != nil {
		h.handleError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PreparationHandler) addIngredient(w http.ResponseWriter, r *http.Request) {
	prepID, err := pathID[domain.PreparationID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req preparationIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ingredient, err := h.svc.AddIngredient(r.Context(), prepID, domain.PreparationIngredientParams{
		IngredientID: domain.IngredientID(req.IngredientID),
		Name:         req.Name,
		Amount:       req.Amount,
		Unit:         req.Unit,
		Prep:         req.Prep,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonResponse(w, ingredient, http.StatusCreated)
}

func (h *PreparationHandler) updateIngredient(w http.ResponseWriter, r *http.Request) {
	prepID, err := pathID[domain.PreparationID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	ingID, err := pathID[domain.PreparationIngredientID](r, "ingId")
	if err != nil {
		jsonError(w, "invalid ingredient id", http.StatusBadRequest)
		return
	}
	var req preparationIngredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ingredient, err := h.svc.UpdateIngredient(r.Context(), prepID, ingID, domain.PreparationIngredientParams{
		IngredientID: domain.IngredientID(req.IngredientID),
		Name:         req.Name,
		Amount:       req.Amount,
		Unit:         req.Unit,
		Prep:         req.Prep,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, ingredient)
}

func (h *PreparationHandler) deleteIngredient(w http.ResponseWriter, r *http.Request) {
	prepID, err := pathID[domain.PreparationID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	ingID, err := pathID[domain.PreparationIngredientID](r, "ingId")
	if err != nil {
		jsonError(w, "invalid ingredient id", http.StatusBadRequest)
		return
	}
	if err := h.svc.DeleteIngredient(r.Context(), prepID, ingID); err != nil {
		h.handleError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *PreparationHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "preparation not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	internalError(w, r, err)
}
