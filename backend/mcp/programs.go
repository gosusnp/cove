// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"encoding/json"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/gosusnp/cove/backend/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// createProgramFullSchema returns an explicit JSON schema for create_program_full.
// This is required because jsonschema-go unconditionally emits ["null","array"] for
// Go slices, which makes required array fields appear nullable and confuses callers.
func createProgramFullSchema() *jsonschema.Schema {
	// jsonschema-go requires a strict tree — no shared pointers. Each schema node
	// must appear exactly once, so we use a helper instead of a shared variable.
	noExtra := func() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }

	exerciseItem := &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Required:             []string{"exercise_id"},
		Properties: map[string]*jsonschema.Schema{
			"exercise_id": {
				Type:        "integer",
				Description: "ID of an existing exercise. Run list_exercises first.",
			},
			"laterality": {
				Type:        "string",
				Description: "bilateral | unilateral | left | right | alternating. Omit if not applicable.",
			},
			"reps": {
				Type:        "integer",
				Description: "Target reps per round. Omit for duration-based exercises.",
			},
			"duration_s": {
				Type:        "integer",
				Description: "Target duration in seconds per round. Omit for rep-based exercises.",
			},
			"weight_kg": {
				Type:        "number",
				Description: "Load in kg. Positive = added (e.g. 20), negative = assisted (e.g. -10), omit or 0 = bodyweight.",
			},
			"sort_order": {
				Type:        "integer",
				Description: "Position within the set. Inferred from array index if omitted.",
			},
		},
	}

	setItem := &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Required:             []string{"rounds", "exercises"},
		Properties: map[string]*jsonschema.Schema{
			"name": {
				Type:        "string",
				Description: "Label for the set, e.g. 'A', 'B', 'Superset 1'. Optional.",
			},
			"rounds": {
				Type:        "integer",
				Description: "Number of times to complete the full set.",
			},
			"rest_s": {
				Type:        "integer",
				Description: "Rest in seconds between rounds of this set. Optional.",
			},
			"sort_order": {
				Type:        "integer",
				Description: "Position within the program. Inferred from array index if omitted.",
			},
			"exercises": {
				Type:        "array",
				Description: "Exercises to perform within this set, in execution order.",
				Items:       exerciseItem,
			},
		},
	}

	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Required:             []string{"name", "sets"},
		Properties: map[string]*jsonschema.Schema{
			"name": {
				Type:        "string",
				Description: "Name of the program.",
			},
			"sets": {
				Type:        "array",
				Description: "Ordered list of training sets (blocks) that make up the program.",
				Items:       setItem,
			},
		},
	}
}

func registerProgramTools(server *mcp.Server, programs *service.ProgramService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_programs",
		Description: "List all programs",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
		list, err := programs.List()
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(list)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_program",
		Description: "Get a program by ID including its full set and exercise hierarchy",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.GetDetail(params.ID)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(program)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_program",
		Description: "Create a new program",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		Name string `json:"name"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.Create(params.Name)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(program)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_program",
		Description: "Update a program's name",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.Update(params.ID, params.Name)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(program)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_program",
		Description: "Delete a program by ID",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := programs.Delete(params.ID); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_program_full",
		Description: "Creates a complete program with all its sets and exercises in a single atomic operation. All exercise_ids must exist before calling — use list_exercises and create_exercise to prepare the exercise library first. If any part fails nothing is written.",
		InputSchema: createProgramFullSchema(),
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		Name string                    `json:"name"`
		Sets []service.ProgramSetInput `json:"sets"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.CreateFull(params.Name, params.Sets)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(program)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})
}
