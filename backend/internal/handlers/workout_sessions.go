// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
)

type WorkoutSessionHandler struct {
	svc *service.WorkoutSessionService
}

func NewWorkoutSessionHandler(s *service.WorkoutSessionService) *WorkoutSessionHandler {
	return &WorkoutSessionHandler{svc: s}
}

func (h *WorkoutSessionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /sessions", h.list)
	mux.HandleFunc("POST /sessions", h.create)
	mux.HandleFunc("GET /sessions/{id}", h.get)
	mux.HandleFunc("PUT /sessions/{id}", h.replace)
	mux.HandleFunc("PATCH /sessions/{id}", h.patch)
	mux.HandleFunc("DELETE /sessions/{id}", h.delete)
}

type workoutSessionRequest struct {
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
	ProgramID        *int64     `json:"program_id,omitempty"`
	ProgramName      *string    `json:"program_name,omitempty"`
	ProgramStructure *string    `json:"program_structure,omitempty"`
	Activity         *string    `json:"activity,omitempty"`
	DurationS        *int       `json:"duration_s,omitempty"`
	PerceivedEffort  *int       `json:"perceived_effort,omitempty"`
	SessionNotes     *string    `json:"session_notes,omitempty"`
	Summary          *string    `json:"summary,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

func sensitiveStringPtr(s *string) *crypto.SensitiveString {
	if s == nil {
		return nil
	}
	ss := crypto.NewSensitiveString(*s)
	return &ss
}

func stringPtr(s *crypto.SensitiveString) *string {
	if s == nil {
		return nil
	}
	v := s.String()
	return &v
}

func (req *workoutSessionRequest) toParams() store.WorkoutSessionParams {
	p := store.WorkoutSessionParams{
		Activity:    req.Activity,
		DurationS:   req.DurationS,
		StartedAt:   req.StartedAt,
		CompletedAt: req.CompletedAt,
		SensitiveData: domain.SessionSensitiveData{
			PerceivedEffort:  req.PerceivedEffort,
			SessionNotes:     sensitiveStringPtr(req.SessionNotes),
			ProgramName:      sensitiveStringPtr(req.ProgramName),
			ProgramStructure: sensitiveStringPtr(req.ProgramStructure),
			Summary:          sensitiveStringPtr(req.Summary),
		},
	}
	if req.ProgramID != nil {
		id := domain.ProgramID(*req.ProgramID)
		p.ProgramID = &id
	}
	return p
}

// workoutSessionResponse is the JSON shape returned to clients.
// Sensitive fields are decrypted inline and flattened here; they go out of scope
// immediately after the response is encoded.
type workoutSessionResponse struct {
	ID                 domain.WorkoutSessionID `json:"id"`
	OrgID              domain.OrgID            `json:"org_id"`
	UserID             domain.UserID           `json:"user_id"`
	ProgramID          *domain.ProgramID       `json:"program_id,omitempty"`
	Activity           *string                 `json:"activity,omitempty"`
	DurationS          *int                    `json:"duration_s,omitempty"`
	StartedAt          *time.Time              `json:"started_at,omitempty"`
	CompletedAt        *time.Time              `json:"completed_at,omitempty"`
	SummaryGeneratedAt *time.Time              `json:"summary_generated_at,omitempty"`
	CreatedBy          domain.UserID           `json:"created_by"`
	CreatedAt          time.Time               `json:"created_at"`
	UpdatedBy          *domain.UserID          `json:"updated_by,omitempty"`
	UpdatedAt          time.Time               `json:"updated_at"`
	// Sensitive fields flattened:
	PerceivedEffort  *int    `json:"perceived_effort,omitempty"`
	SessionNotes     *string `json:"session_notes,omitempty"`
	ProgramName      *string `json:"program_name,omitempty"`
	ProgramStructure *string `json:"program_structure,omitempty"`
	Summary          *string `json:"summary,omitempty"`
}

func toResponse(r *http.Request, ws *domain.WorkoutSession) (*workoutSessionResponse, error) {
	resp := &workoutSessionResponse{
		ID:                 ws.ID,
		OrgID:              ws.OrgID,
		UserID:             ws.UserID,
		ProgramID:          ws.ProgramID,
		Activity:           ws.Activity,
		DurationS:          ws.DurationS,
		StartedAt:          ws.StartedAt,
		CompletedAt:        ws.CompletedAt,
		SummaryGeneratedAt: ws.SummaryGeneratedAt,
		CreatedBy:          ws.CreatedBy,
		CreatedAt:          ws.CreatedAt,
		UpdatedBy:          ws.UpdatedBy,
		UpdatedAt:          ws.UpdatedAt,
	}
	err := ws.UseSensitiveData(r.Context(), func(private domain.SessionSensitiveData) error {
		resp.PerceivedEffort = private.PerceivedEffort
		resp.SessionNotes = stringPtr(private.SessionNotes)
		resp.ProgramName = stringPtr(private.ProgramName)
		resp.ProgramStructure = stringPtr(private.ProgramStructure)
		resp.Summary = stringPtr(private.Summary)
		return nil
	})
	return resp, err
}

func (h *WorkoutSessionHandler) create(w http.ResponseWriter, r *http.Request) {
	var req workoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ws, err := h.svc.Create(r.Context(), req.toParams())
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	resp, err := toResponse(r, ws)
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonResponse(w, resp, http.StatusCreated)
}

func (h *WorkoutSessionHandler) replace(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var req workoutSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ws, err := h.svc.Update(r.Context(), id, req.UpdatedAt, req.toParams())
	if errors.Is(err, service.ErrConflict) {
		current, err := h.svc.Get(r.Context(), id)
		if err != nil {
			handleServiceError(w, r, err, "session not found")
			return
		}
		resp, err := toResponse(r, current)
		if err != nil {
			internalError(w, r, err)
			return
		}
		jsonResponse(w, resp, http.StatusConflict)
		return
	}
	if err != nil {
		handleServiceError(w, r, err, "session not found")
		return
	}
	resp, err := toResponse(r, ws)
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, resp)
}

func (h *WorkoutSessionHandler) patch(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	var patch service.WorkoutSessionPatch
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		jsonError(w, "invalid request body", http.StatusBadRequest)
		return
	}
	ws, err := h.svc.Patch(r.Context(), id, patch)
	if errors.Is(err, service.ErrConflict) {
		current, err := h.svc.Get(r.Context(), id)
		if err != nil {
			handleServiceError(w, r, err, "session not found")
			return
		}
		resp, err := toResponse(r, current)
		if err != nil {
			internalError(w, r, err)
			return
		}
		jsonResponse(w, resp, http.StatusConflict)
		return
	}
	if err != nil {
		handleServiceError(w, r, err, "session not found")
		return
	}
	resp, err := toResponse(r, ws)
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, resp)
}

func (h *WorkoutSessionHandler) list(w http.ResponseWriter, r *http.Request) {
	sessions, err := h.svc.List(r.Context())
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	resps := make([]*workoutSessionResponse, 0, len(sessions))
	for _, ws := range sessions {
		resp, err := toResponse(r, ws)
		if err != nil {
			internalError(w, r, err)
			return
		}
		resps = append(resps, resp)
	}
	jsonOK(w, resps)
}

func (h *WorkoutSessionHandler) get(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	session, err := h.svc.Get(r.Context(), id)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "session not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	resp, err := toResponse(r, session)
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, resp)
}

func (h *WorkoutSessionHandler) delete(w http.ResponseWriter, r *http.Request) {
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = h.svc.Delete(r.Context(), id)
	if errors.Is(err, service.ErrUnauthorized) {
		jsonError(w, "unauthorized", http.StatusUnauthorized)
		return
	}
	if errors.Is(err, service.ErrNotFound) {
		jsonError(w, "session not found", http.StatusNotFound)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
