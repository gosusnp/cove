// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"errors"

	"github.com/gosusnp/cove/backend/internal/llm/prompts"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerPrompts(server *mcp.Server, profiles *service.TrainingProfileService) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "build-training-profile",
		Description: "A conversational onboarding prompt to gather and classify the user's training goals, history, and constraints.",
	}, func(ctx context.Context, _ *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		// Fetch current profile to provide context to the agent
		tp, err := profiles.Get(ctx)
		if err != nil && !errors.Is(err, service.ErrNotFound) {
			return nil, err
		}

		promptText, err := prompts.BuildTrainingProfile(ctx, tp)
		if err != nil {
			return nil, err
		}

		return &mcp.GetPromptResult{
			Messages: []*mcp.PromptMessage{
				{
					Role: "user",
					Content: &mcp.TextContent{
						Text: promptText,
					},
				},
			},
		}, nil
	})
}
