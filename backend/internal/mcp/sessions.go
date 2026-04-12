// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerSessionTools(server *mcp.Server, sessions *service.WorkoutSessionService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_sessions",
		Description: "List workout sessions, optionally filtered by date range (from/to as YYYY-MM-DD). Returns activity, duration, RPE, program name, notes, and AI-generated summaries.",
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
