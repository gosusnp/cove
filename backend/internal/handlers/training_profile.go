// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
)

type TrainingProfileHandler struct {
	svc *service.TrainingProfileService
}

func NewTrainingProfileHandler(s *service.TrainingProfileService) *TrainingProfileHandler {
	return &TrainingProfileHandler{svc: s}
}

func (h *TrainingProfileHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /users/me/training-profile", h.get)
	mux.HandleFunc("PUT /users/me/training-profile", h.upsert)
	mux.HandleFunc("PATCH /users/me/training-profile", h.patch)
	mux.HandleFunc("DELETE /users/me/training-profile", h.delete)
}

type trainingProfileRequest struct {
	Motivation  *string                            `json:"motivation,omitempty"`
	Disciplines []trainingProfileDisciplineRequest `json:"disciplines,omitempty"`
	Constraints *string                            `json:"constraints,omitempty"`
}

type trainingProfileDisciplineRequest struct {
	Name          *string  `json:"name,omitempty"`
	YearsPractice *float64 `json:"years_practice,omitempty"`
	Level         *string  `json:"level,omitempty"`
	Notes         *string  `json:"notes,omitempty"`
}

func (req *trainingProfileRequest) toDomain() domain.TrainingProfileSensitiveData {
	disciplines := make([]domain.TrainingProfileDiscipline, len(req.Disciplines))
	for i, d := range req.Disciplines {
		disciplines[i] = domain.TrainingProfileDiscipline{
			Name:          crypto.NewSensitiveStringFromPtr(d.Name),
			YearsPractice: d.YearsPractice,
			Level:         crypto.NewSensitiveStringFromPtr(d.Level),
			Notes:         crypto.NewSensitiveStringFromPtr(d.Notes),
		}
	}
	return domain.TrainingProfileSensitiveData{
		Motivation:  crypto.NewSensitiveStringFromPtr(req.Motivation),
		Disciplines: disciplines,
		Constraints: crypto.NewSensitiveStringFromPtr(req.Constraints),
	}
}

type trainingProfileResponse struct {
	UserID    domain.UserID  `json:"user_id"`
	OrgID     domain.OrgID   `json:"org_id"`
	CreatedBy domain.UserID  `json:"created_by"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedBy *domain.UserID `json:"updated_by,omitempty"`
	UpdatedAt time.Time      `json:"updated_at"`

	Motivation  *string                            `json:"motivation,omitempty"`
	Disciplines []trainingProfileDisciplineRequest `json:"disciplines,omitempty"`
	Constraints *string                            `json:"constraints,omitempty"`
}

func toTrainingProfileResponse(r *http.Request, tp *domain.UserTrainingProfile) (*trainingProfileResponse, error) {
	resp := &trainingProfileResponse{
		UserID:    tp.UserID,
		OrgID:     tp.OrgID,
		CreatedBy: tp.CreatedBy,
		CreatedAt: tp.CreatedAt,
		UpdatedBy: tp.UpdatedBy,
		UpdatedAt: tp.UpdatedAt,
	}
	err := tp.UseSensitiveData(r.Context(), func(private domain.TrainingProfileSensitiveData) error {
		resp.Motivation = stringPtr(private.Motivation)
		resp.Constraints = stringPtr(private.Constraints)
		resp.Disciplines = make([]trainingProfileDisciplineRequest, len(private.Disciplines))
		for i, d := range private.Disciplines {
			resp.Disciplines[i] = trainingProfileDisciplineRequest{
				Name:          stringPtr(d.Name),
				YearsPractice: d.YearsPractice,
				Level:         stringPtr(d.Level),
				Notes:         stringPtr(d.Notes),
			}
		}
		return nil
	})
	return resp, err
}

func (h *TrainingProfileHandler) get(w http.ResponseWriter, r *http.Request) {
	tp, err := h.svc.Get(r.Context())
	if err != nil {
		handleServiceError(w, r, err, "training profile not found")
		return
	}
	resp, err := toTrainingProfileResponse(r, tp)
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, resp)
}

func (h *TrainingProfileHandler) upsert(w http.ResponseWriter, r *http.Request) {
	var req trainingProfileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tp, err := h.svc.Upsert(r.Context(), req.toDomain())
	if err != nil {
		handleServiceError(w, r, err, "failed to update training profile")
		return
	}
	resp, err := toTrainingProfileResponse(r, tp)
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, resp)
}

func (h *TrainingProfileHandler) patch(w http.ResponseWriter, r *http.Request) {
	var patch service.TrainingProfilePatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	tp, err := h.svc.Patch(r.Context(), patch)
	if err != nil {
		handleServiceError(w, r, err, "failed to update training profile")
		return
	}
	resp, err := toTrainingProfileResponse(r, tp)
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, resp)
}

func (h *TrainingProfileHandler) delete(w http.ResponseWriter, r *http.Request) {
	err := h.svc.Delete(r.Context())
	if err != nil {
		handleServiceError(w, r, err, "training profile not found")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
