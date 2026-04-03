// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"errors"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func trainingProfileUpdateSchema() *jsonschema.Schema {
	noExtra := func() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }

	disciplineItem := &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Properties: map[string]*jsonschema.Schema{
			"name": {
				Type:        "string",
				Description: "Name of the sport or discipline (e.g. 'Olympic Weightlifting').",
			},
			"years_practice": {
				Type:        "number",
				Description: "Years of experience in this discipline.",
			},
			"level": {
				Type:        "string",
				Description: "Experience level: beginner | intermediate | advanced | expert.",
			},
			"notes": {
				Type:        "string",
				Description: "Additional details about your experience in this discipline.",
			},
		},
	}

	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Properties: map[string]*jsonschema.Schema{
			"motivation": {
				Type:        "string",
				Description: "Your primary goals and reasons for training.",
			},
			"constraints": {
				Type:        "string",
				Description: "Any physical limitations, injuries, or scheduling constraints.",
			},
			"disciplines": {
				Type:        "array",
				Description: "List of training disciplines. Replaces existing list if provided.",
				Items:       disciplineItem,
			},
		},
	}
}

func registerTrainingProfileTools(server *mcp.Server, svc *service.TrainingProfileService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_training_profile",
		Description: "Get your training profile (motivation, disciplines, constraints) as Markdown.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
		tp, err := svc.Get(ctx)
		if errors.Is(err, service.ErrNotFound) {
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: "No training profile found."}},
			}, struct{}{}, nil
		}
		if err != nil {
			return nil, struct{}{}, err
		}

		var md string
		err = tp.UseSensitiveData(ctx, func(data domain.TrainingProfileSensitiveData) error {
			md = markdown.TrainingProfile(data)
			return nil
		})
		if err != nil {
			return nil, struct{}{}, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: md}},
		}, struct{}{}, nil
	})

	type disciplineParams struct {
		Name          *string  `json:"name"`
		YearsPractice *float64 `json:"years_practice"`
		Level         *string  `json:"level"`
		Notes         *string  `json:"notes"`
	}

	type updateParams struct {
		Motivation  *string             `json:"motivation"`
		Constraints *string             `json:"constraints"`
		Disciplines *[]disciplineParams `json:"disciplines"`
	}

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_training_profile",
		Description: "Update your training profile (motivation, constraints, disciplines). All fields are optional. Motivation and constraints replace the existing value when provided. Disciplines is a full array replacement — pass the entire list of sports/activities you want to keep. Valid levels: beginner, intermediate, advanced, expert.",
		InputSchema: trainingProfileUpdateSchema(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params updateParams) (*mcp.CallToolResult, struct{}, error) {
		patch := service.TrainingProfilePatch{}
		if params.Motivation != nil {
			patch.Motivation = domain.Optional[*string]{Value: params.Motivation, Set: true}
		}
		if params.Constraints != nil {
			patch.Constraints = domain.Optional[*string]{Value: params.Constraints, Set: true}
		}
		if params.Disciplines != nil {
			ds := make([]service.TrainingProfileDisciplinePatch, len(*params.Disciplines))
			for i, d := range *params.Disciplines {
				ds[i] = service.TrainingProfileDisciplinePatch{
					Name:          d.Name,
					YearsPractice: d.YearsPractice,
					Level:         d.Level,
					Notes:         d.Notes,
				}
			}
			patch.Disciplines = domain.Optional[[]service.TrainingProfileDisciplinePatch]{Value: ds, Set: true}
		}

		tp, err := svc.Patch(ctx, patch)
		if err != nil {
			return nil, struct{}{}, err
		}

		var md string
		err = tp.UseSensitiveData(ctx, func(data domain.TrainingProfileSensitiveData) error {
			md = markdown.TrainingProfile(data)
			return nil
		})
		if err != nil {
			return nil, struct{}{}, err
		}

		return &mcp.CallToolResult{
			Content: []mcp.Content{&mcp.TextContent{Text: md}},
		}, struct{}{}, nil
	})
}
