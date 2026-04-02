// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"context"
	"fmt"
	"io"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
)

// SessionSummaryWorker orchestrates the fetch → summarize → patch workflow for
// a single workout session. It is the unit of work executed by the
// session-summary Hatchet job.
//
// All data access goes through WorkoutSessionPort so the orchestration logic is
// identical regardless of whether the worker runs in-process or as a separate
// pod.
type SessionSummaryWorker struct {
	sessions  WorkoutSessionPort
	summarize *service.SummarizeService
}

// NewSessionSummaryWorker returns a SessionSummaryWorker wired with the given
// port and summarization service.
func NewSessionSummaryWorker(sessions WorkoutSessionPort, summarize *service.SummarizeService) *SessionSummaryWorker {
	return &SessionSummaryWorker{sessions: sessions, summarize: summarize}
}

// Run fetches the session, generates an LLM summary, and patches the session
// with the result. The identity in ctx must be set via domain.NewContext before
// calling — it determines which session is accessible and is used for the
// RLS-gated read and write.
//
// Returns service.ErrConflict if the session was modified between fetch and
// patch. Callers should treat this as a retriable error and re-run the full
// workflow to pick up the latest state.
func (w *SessionSummaryWorker) Run(ctx context.Context, id domain.WorkoutSessionID) error {
	ws, err := w.sessions.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch session %d: %w", id, err)
	}

	var result *service.SessionSummary
	if err := ws.UseSensitiveData(ctx, func(sd domain.SessionSensitiveData) error {
		var err error
		result, err = w.summarize.SummarizeSession(ctx, ws, sd)
		return err
	}); err != nil {
		return fmt.Errorf("summarize session %d: %w", id, err)
	}

	if err := w.sessions.PatchSummary(ctx, id, WorkoutSessionSummaryPatch{
		Summary:   result.Summary,
		UpdatedAt: ws.UpdatedAt,
	}); err != nil {
		return fmt.Errorf("patch session %d summary: %w", id, err)
	}

	return nil
}

// Preview fetches the session and renders the prompt that would be sent to the
// LLM, writing it to w. No LLM call is made and the session is not modified.
// Intended for development and prompt iteration via cmd/llm.
func (w *SessionSummaryWorker) Preview(ctx context.Context, id domain.WorkoutSessionID, out io.Writer) error {
	ws, err := w.sessions.Get(ctx, id)
	if err != nil {
		return fmt.Errorf("fetch session %d: %w", id, err)
	}

	return ws.UseSensitiveData(ctx, func(sd domain.SessionSensitiveData) error {
		return w.summarize.PreviewSession(out, ws, sd)
	})
}
