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

type IngredientHandler struct {
	svc *service.IngredientService
}

func NewIngredientHandler(s *service.IngredientService) *IngredientHandler {
	return &IngredientHandler{svc: s}
}

func (h *IngredientHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /ingredients", h.list)
	mux.HandleFunc("POST /ingredients", h.create)
	mux.HandleFunc("GET /ingredients/{id}", h.get)
	mux.HandleFunc("PUT /ingredients/{id}", h.update)
	mux.HandleFunc("DELETE /ingredients/{id}", h.delete)
}

type ingredientRequest struct {
	Name            string   `json:"name"`
	FdcID           *int     `json:"fdc_id,omitempty"`
	CaloriesPer100g float64  `json:"calories_per_100g"`
	ProteinPer100g  float64  `json:"protein_per_100g"`
	FatPer100g      float64  `json:"fat_per_100g"`
	CarbsPer100g    float64  `json:"carbs_per_100g"`
	DensityGPerMl   *float64 `json:"density_g_per_ml,omitempty"`
	IsPublic        bool     `json:"is_public"`
}

func (h *IngredientHandler) list(w http.ResponseWriter, r *http.Request) {
	ingredients, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, ingredients)
}

func (h *IngredientHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.IngredientID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	ingredient, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, ingredient)
}

func (h *IngredientHandler) create(w http.ResponseWriter, r *http.Request) {
	var req ingredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ingredient, err := h.svc.Create(r.Context(), domain.IngredientParams{
		Name:            req.Name,
		FdcID:           req.FdcID,
		CaloriesPer100g: req.CaloriesPer100g,
		ProteinPer100g:  req.ProteinPer100g,
		FatPer100g:      req.FatPer100g,
		CarbsPer100g:    req.CarbsPer100g,
		DensityGPerMl:   req.DensityGPerMl,
		IsPublic:        req.IsPublic,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonResponse(w, ingredient, http.StatusCreated)
}

func (h *IngredientHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.IngredientID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req ingredientRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ingredient, err := h.svc.Update(r.Context(), id, domain.IngredientParams{
		Name:            req.Name,
		FdcID:           req.FdcID,
		CaloriesPer100g: req.CaloriesPer100g,
		ProteinPer100g:  req.ProteinPer100g,
		FatPer100g:      req.FatPer100g,
		CarbsPer100g:    req.CarbsPer100g,
		DensityGPerMl:   req.DensityGPerMl,
		IsPublic:        req.IsPublic,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, ingredient)
}

func (h *IngredientHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.IngredientID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.Delete(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *IngredientHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "ingredient not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	internalError(w, r, err)
}
