// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"errors"
	"net/http"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
)

func (h *ProgramHandler) listVersions(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}

	versions, err := h.svc.ListVersions(r.Context(), programID)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "program not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}

	jsonOK(w, versions)
}

func (h *ProgramHandler) getVersion(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}

	versionID, err := pathID[domain.ProgramVersionID](r, "vid")
	if err != nil {
		jsonError(w, "invalid version id", http.StatusBadRequest)
		return
	}

	v, err := h.svc.GetVersion(r.Context(), programID, versionID)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "version not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}

	jsonOK(w, v)
}

func (h *ProgramHandler) rollbackVersion(w http.ResponseWriter, r *http.Request) {
	programID, err := pathID[domain.ProgramID](r, "id")
	if err != nil {
		jsonError(w, "invalid program id", http.StatusBadRequest)
		return
	}

	versionID, err := pathID[domain.ProgramVersionID](r, "vid")
	if err != nil {
		jsonError(w, "invalid version id", http.StatusBadRequest)
		return
	}

	if err := h.svc.Rollback(r.Context(), programID, versionID); err != nil {
		if errors.Is(err, service.ErrUnauthorized) {
			jsonError(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			jsonError(w, ve.Error(), http.StatusBadRequest)
			return
		}
		if errors.Is(err, service.ErrNotFound) {
			jsonError(w, "version not found", http.StatusNotFound)
			return
		}
		internalError(w, r, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
