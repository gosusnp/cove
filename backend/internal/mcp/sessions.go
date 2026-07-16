// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createSessionParams struct {
	Activity         *string  `json:"activity,omitempty"`
	StartedAt        *string  `json:"started_at,omitempty"`
	CompletedAt      *string  `json:"completed_at,omitempty"`
	DurationS        *int     `json:"duration_s,omitempty"`
	ProgramID        *int64   `json:"program_id,omitempty"`
	ProgramName      *string  `json:"program_name,omitempty"`
	ProgramStructure *string  `json:"program_structure,omitempty"`
	PerceivedEffort  *int     `json:"perceived_effort,omitempty"`
	SessionNotes     *string  `json:"session_notes,omitempty"`
	Labels           []string `json:"labels,omitempty"`
}

func buildSessionParams(params createSessionParams) (store.WorkoutSessionParams, error) {
	p := store.WorkoutSessionParams{
		Activity:  params.Activity,
		DurationS: params.DurationS,
	}

	if params.ProgramID != nil {
		pid := domain.ProgramID(*params.ProgramID)
		p.ProgramID = &pid
	}

	if params.StartedAt != nil {
		t, err := time.Parse(time.RFC3339, *params.StartedAt)
		if err != nil {
			return p, fmt.Errorf("invalid started_at: use RFC3339 format, e.g. 2026-06-01T09:00:00Z")
		}
		p.StartedAt = &t
	}

	if params.CompletedAt != nil {
		t, err := time.Parse(time.RFC3339, *params.CompletedAt)
		if err != nil {
			return p, fmt.Errorf("invalid completed_at: use RFC3339 format, e.g. 2026-06-01T09:00:00Z")
		}
		p.CompletedAt = &t
	}

	labels := make([]crypto.SensitiveString, len(params.Labels))
	for i, l := range params.Labels {
		labels[i] = crypto.NewSensitiveString(l)
	}
	p.SensitiveData = domain.SessionSensitiveData{
		PerceivedEffort:  params.PerceivedEffort,
		SessionNotes:     crypto.NewSensitiveStringFromPtr(params.SessionNotes),
		ProgramName:      crypto.NewSensitiveStringFromPtr(params.ProgramName),
		ProgramStructure: crypto.NewSensitiveStringFromPtr(params.ProgramStructure),
		Labels:           labels,
	}

	return p, nil
}

func registerSessionTools(server *mcp.Server, sessions *service.WorkoutSessionService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_session",
		Description: "Create a historical workout session. All fields are optional — supply only what you know. labels is an optional array of classification tags (e.g. [\"deload\", \"recovery\"]).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params createSessionParams) (*mcp.CallToolResult, struct{}, error) {
		p, err := buildSessionParams(params)
		if err != nil {
			return nil, struct{}{}, err
		}

		ws, err := sessions.Create(ctx, p)
		if err != nil {
			return nil, struct{}{}, fmt.Errorf("create session: %w", err)
		}

		var text string
		if err := ws.UseSensitiveData(ctx, func(sd domain.SessionSensitiveData) error {
			text = fmt.Sprintf("**ID:** %d\n\n%s", ws.ID, markdown.SessionEntry(ws, sd))
			return nil
		}); err != nil {
			return nil, struct{}{}, fmt.Errorf("decrypt session: %w", err)
		}

		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List workout sessions, optionally filtered by date range (from/to as YYYY-MM-DD). Returns activity, duration, RPE, program name, notes, AI-generated summaries, and labels.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		From *string `json:"from,omitempty"`
		To   *string `json:"to,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		f, err := service.NewSessionFilter(params.From, params.To)
		if err != nil {
			return nil, struct{}{}, err
		}

		list, err := sessions.List(ctx, f)
		if err != nil {
			return nil, struct{}{}, err
		}

		entries := make([]string, 0, len(list))
		for _, ws := range list {
			var entry string
			if err := ws.UseSensitiveData(ctx, func(sd domain.SessionSensitiveData) error {
				entry = markdown.SessionEntry(ws, sd)
				return nil
			}); err != nil {
				return nil, struct{}{}, fmt.Errorf("decrypt session %d: %w", ws.ID, err)
			}
			entries = append(entries, entry)
		}

		text := markdown.SessionList(entries)
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: text}}}, struct{}{}, nil
	})
}
