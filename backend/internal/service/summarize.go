// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/llm"
	"github.com/gosusnp/cove/backend/internal/llm/prompts"
)

// SessionSummary is the result of summarizing a single workout session.
type SessionSummary struct {
	Summary string `json:"summary"`
}

// SummarizeService calls an LLM to generate summaries for fitness data.
type SummarizeService struct {
	router llm.Router
}

// NewSummarizeService returns a SummarizeService backed by the given Router.
// Pass nil to create a preview-only service.
func NewSummarizeService(r llm.Router) *SummarizeService {
	return &SummarizeService{router: r}
}

// PreviewSession renders the prompt that would be sent to the LLM and writes
// it to w. No LLM call is made. Safe to call on a nil-client service.
func (s *SummarizeService) PreviewSession(w io.Writer, ws *domain.WorkoutSession, sd domain.SessionSensitiveData) error {
	req, err := prompts.SessionSummary(ws, sd)
	if err != nil {
		return fmt.Errorf("build session summary prompt: %w", err)
	}
	sep := strings.Repeat("─", 60)
	for _, m := range req.Messages {
		fmt.Fprintf(w, "%s\n[%s]\n%s\n%s\n\n", sep, strings.ToUpper(m.Role), sep, m.Content)
	}
	return nil
}

// SummarizeSession generates a markdown summary of a single workout session.
//
// Must be called inside a WorkoutSession.UseSensitiveData callback. sd contains
// the decrypted sensitive fields and will be zeroed by the caller after this
// method returns, limiting plaintext exposure to the duration of the LLM call.
//
// Returns an error if the service was constructed without an LLM client (e.g.
// preview-only mode via NewSummarizeService(nil)).
func (s *SummarizeService) SummarizeSession(ctx context.Context, ws *domain.WorkoutSession, sd domain.SessionSensitiveData) (*SessionSummary, error) {
	if s.router == nil {
		return nil, fmt.Errorf("summarize session: no LLM router configured")
	}
	req, err := prompts.SessionSummary(ws, sd)
	if err != nil {
		return nil, fmt.Errorf("build session summary prompt: %w", err)
	}
	text, err := s.router.Complete(ctx, llm.TaskSummarize, req)
	if err != nil {
		return nil, fmt.Errorf("summarize session: %w", err)
	}
	return &SessionSummary{Summary: text}, nil
}
