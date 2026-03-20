// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"net/http"

	"github.com/gosusnp/cove/backend/internal/domain"
)

type ActivityHandler struct{}

func NewActivityHandler() *ActivityHandler {
	return &ActivityHandler{}
}

func (h *ActivityHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /activities", h.list)
}

func (h *ActivityHandler) list(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, domain.KnownActivities)
}
