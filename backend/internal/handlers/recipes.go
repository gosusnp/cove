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

type RecipeHandler struct {
	svc *service.RecipeService
}

func NewRecipeHandler(s *service.RecipeService) *RecipeHandler {
	return &RecipeHandler{svc: s}
}

func (h *RecipeHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /recipes", h.list)
	mux.HandleFunc("POST /recipes", h.create)
	mux.HandleFunc("GET /recipes/{id}", h.get)
	mux.HandleFunc("PUT /recipes/{id}", h.update)
	mux.HandleFunc("DELETE /recipes/{id}", h.delete)
}

type recipeRequest struct {
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	YieldAmount *float64 `json:"yield_amount,omitempty"`
	YieldUnit   *string  `json:"yield_unit,omitempty"`
	Servings    int      `json:"servings"`
	IsPublic    bool     `json:"is_public"`
}

func (h *RecipeHandler) list(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.svc.List(r.Context())
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, recipes)
}

func (h *RecipeHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.RecipeID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	recipe, err := h.svc.Get(r.Context(), id)
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, recipe)
}

func (h *RecipeHandler) create(w http.ResponseWriter, r *http.Request) {
	var req recipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	recipe, err := h.svc.Create(r.Context(), domain.RecipeParams{
		Name:        req.Name,
		Description: req.Description,
		YieldAmount: req.YieldAmount,
		YieldUnit:   req.YieldUnit,
		Servings:    req.Servings,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonResponse(w, recipe, http.StatusCreated)
}

func (h *RecipeHandler) update(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.RecipeID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req recipeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	recipe, err := h.svc.Update(r.Context(), id, domain.RecipeParams{
		Name:        req.Name,
		Description: req.Description,
		YieldAmount: req.YieldAmount,
		YieldUnit:   req.YieldUnit,
		Servings:    req.Servings,
		IsPublic:    req.IsPublic,
	})
	if err != nil {
		h.handleError(w, r, err)
		return
	}
	jsonOK(w, recipe)
}

func (h *RecipeHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.RecipeID](r, "id")
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

func (h *RecipeHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "recipe not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	internalError(w, r, err)
}
