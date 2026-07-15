// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/workers"
)

// SummaryJobClient enqueues and queries session summary background jobs.
type SummaryJobClient interface {
	RequestSummary(ctx context.Context, sessionID int64, orgID, userID string) (string, error)
	GetSessionSummaryStatus(ctx context.Context, runID string, expectedSessionID int64) (string, error)
}

type WorkoutSessionHandler struct {
	svc       *service.WorkoutSessionService
	jobClient SummaryJobClient // nil when Hatchet is not configured
}

func NewWorkoutSessionHandler(s *service.WorkoutSessionService, d SummaryJobClient) *WorkoutSessionHandler {
	return &WorkoutSessionHandler{svc: s, jobClient: d}
}

func (h *WorkoutSessionHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /sessions", h.list)
	mux.HandleFunc("POST /sessions", h.create)
	mux.HandleFunc("GET /sessions/labels", h.labels)
	mux.HandleFunc("GET /sessions/{id}", h.get)
	mux.HandleFunc("PUT /sessions/{id}", h.replace)
	mux.HandleFunc("PATCH /sessions/{id}", h.patch)
	mux.HandleFunc("DELETE /sessions/{id}", h.delete)
	mux.HandleFunc("POST /sessions/{id}/summarize", h.triggerSummary)
	mux.HandleFunc("GET /sessions/{id}/summarize", h.getSummaryStatus)
}

type workoutSessionRequest struct {
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
	ProgramID        *int64     `json:"program_id,omitempty"`
	ProgramName      *string    `json:"program_name,omitempty"`
	ProgramStructure *string    `json:"program_structure,omitempty"`
	Activity         *string    `json:"activity,omitempty"`
	DurationS        *int       `json:"duration_s,omitempty"`
	Labels           []string   `json:"labels,omitempty"`
	PerceivedEffort  *int       `json:"perceived_effort,omitempty"`
	SessionNotes     *string    `json:"session_notes,omitempty"`
	Summary          *string    `json:"summary,omitempty"`
	StartedAt        *time.Time `json:"started_at,omitempty"`
	CompletedAt      *time.Time `json:"completed_at,omitempty"`
}

func parseSessionFilter(r *http.Request) (service.SessionFilter, error) {
	var from, to *string
	if v := r.URL.Query().Get("from"); v != "" {
		from = &v
	}
	if v := r.URL.Query().Get("to"); v != "" {
		to = &v
	}
	return service.NewSessionFilter(from, to)
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
			SessionNotes:     crypto.NewSensitiveStringFromPtr(req.SessionNotes),
			ProgramName:      crypto.NewSensitiveStringFromPtr(req.ProgramName),
			ProgramStructure: crypto.NewSensitiveStringFromPtr(req.ProgramStructure),
			Summary:          crypto.NewSensitiveStringFromPtr(req.Summary),
			Labels:           req.Labels,
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
	Labels             []string                `json:"labels"`
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

type sessionLabelsResponse struct {
	Labels []string `json:"labels"`
}

type summarizeResponse struct {
	RunID string `json:"run_id"`
}

type summaryStatusResponse struct {
	Status string `json:"status"`
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
		Labels:             []string{},
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
		if private.Labels != nil {
			resp.Labels = private.Labels
		}
		return nil
	})
	return resp, err
}

func (h *WorkoutSessionHandler) labels(w http.ResponseWriter, r *http.Request) {
	jsonOK(w, sessionLabelsResponse{Labels: service.AllowedSessionLabels()})
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
	var ve *service.ValidationError
	if errors.As(err, &ve) {
		jsonError(w, ve.Error(), http.StatusBadRequest)
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
	f, err := parseSessionFilter(r)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	sessions, err := h.svc.List(r.Context(), f)
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

func (h *WorkoutSessionHandler) triggerSummary(w http.ResponseWriter, r *http.Request) {
	if h.jobClient == nil {
		jsonError(w, "summarize not available", http.StatusServiceUnavailable)
		return
	}
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	ws, err := h.svc.Get(r.Context(), id)
	if err != nil {
		handleServiceError(w, r, err, "session not found")
		return
	}
	runID, err := h.jobClient.RequestSummary(r.Context(), int64(ws.ID), ws.OrgID.String(), ws.UserID.String())
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonResponse(w, summarizeResponse{RunID: runID}, http.StatusAccepted)
}

func (h *WorkoutSessionHandler) getSummaryStatus(w http.ResponseWriter, r *http.Request) {
	if h.jobClient == nil {
		jsonError(w, "summarize not available", http.StatusServiceUnavailable)
		return
	}
	id, err := pathID[domain.WorkoutSessionID](r, "id")
	if err != nil {
		jsonError(w, "invalid id", http.StatusBadRequest)
		return
	}
	runID := r.URL.Query().Get("run_id")
	if runID == "" {
		jsonError(w, "run_id is required", http.StatusBadRequest)
		return
	}
	status, err := h.jobClient.GetSessionSummaryStatus(r.Context(), runID, int64(id))
	if errors.Is(err, workers.ErrRunNotFound) {
		jsonError(w, "run not found", http.StatusNotFound)
		return
	}
	if errors.Is(err, workers.ErrRunSessionMismatch) {
		jsonError(w, "run does not belong to session", http.StatusForbidden)
		return
	}
	if err != nil {
		internalError(w, r, err)
		return
	}
	jsonOK(w, summaryStatusResponse{Status: status})
}
