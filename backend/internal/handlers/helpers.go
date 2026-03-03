// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
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
