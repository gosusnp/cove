// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"net/http"
)

// OTAHandler serves the OTA version check and bundle download endpoints
// used by @capgo/capacitor-updater in the Android app.
type OTAHandler struct {
	version string
	bundle  []byte
}

// NewOTAHandler creates a handler with the embedded bundle and current version.
// When version is empty (local dev), the version check always returns "no update".
func NewOTAHandler(version string, bundle []byte) *OTAHandler {
	return &OTAHandler{version: version, bundle: bundle}
}

// RegisterRoutes registers the version check and bundle download endpoints on mux.
func (h *OTAHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /ota/mobile/version", h.checkVersion)
	mux.HandleFunc("GET /ota/mobile/bundle", h.serveBundle)
}

type versionCheckRequest struct {
	VersionName string `json:"version_name"`
}

func (h *OTAHandler) checkVersion(w http.ResponseWriter, r *http.Request) {
	if h.version == "" {
		jsonOK(w, map[string]string{"message": "no update"})
		return
	}

	var req versionCheckRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	if req.VersionName == h.version {
		jsonOK(w, map[string]string{"message": "no update"})
		return
	}

	jsonOK(w, map[string]string{
		"version": h.version,
		"url":     "https://" + r.Host + "/ota/mobile/bundle",
	})
}

func (h *OTAHandler) serveBundle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/zip")
	_, _ = w.Write(h.bundle)
}
