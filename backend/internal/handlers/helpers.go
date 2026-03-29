// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gosusnp/cove/backend/internal/service"
)

func pathID[T ~int64](r *http.Request, key string) (T, error) {
	id, err := strconv.ParseInt(r.PathValue(key), 10, 64)
	return T(id), err
}

func jsonOK(w http.ResponseWriter, v any) {
	jsonResponse(w, v, http.StatusOK)
}

func jsonResponse(w http.ResponseWriter, v any, status int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("encode response: %v", err)
	}
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	jsonResponse(w, map[string]string{"error": msg}, status)
}

// internalError logs err and responds with 500. Use this instead of jsonError
// for unexpected server-side failures so the cause is always visible in logs.
func internalError(w http.ResponseWriter, r *http.Request, err error) {
	slog.Error("internal error", "method", r.Method, "path", r.URL.Path, "err", err)
	jsonError(w, "internal error", http.StatusInternalServerError)
}

// handleServiceError maps common service errors to HTTP responses. notFoundMsg
// is used for ErrNotFound; all other unrecognised errors fall through to
// internalError.
func handleServiceError(w http.ResponseWriter, r *http.Request, err error, notFoundMsg string) {
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, notFoundMsg, http.StatusNotFound)
		return
	}
	internalError(w, r, err)
}
