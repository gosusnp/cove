// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"net/http"

	"github.com/gosusnp/cove/backend/internal/fdc"
)

// fdcSearcher is satisfied by *fdc.Client and any test double.
type fdcSearcher interface {
	Search(ctx context.Context, query string) ([]fdc.FoodResult, error)
}

// FDCHandler proxies ingredient searches to the USDA Food Data Central API.
type FDCHandler struct {
	client fdcSearcher
}

// NewFDCHandler returns a FDCHandler backed by the given client.
func NewFDCHandler(c fdcSearcher) *FDCHandler {
	return &FDCHandler{client: c}
}

func (h *FDCHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /fdc/search", h.search)
}

func (h *FDCHandler) search(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query().Get("q")
	if q == "" {
		jsonError(w, "q is required", http.StatusBadRequest)
		return
	}

	foods, err := h.client.Search(r.Context(), q)
	if err != nil {
		internalError(w, r, err)
		return
	}

	type response struct {
		Foods []fdc.FoodResult `json:"foods"`
	}
	jsonOK(w, response{Foods: foods})
}
