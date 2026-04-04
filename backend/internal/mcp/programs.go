// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// programExerciseItemSchema returns the JSON schema for a single exercise within a set.
// Returns a new value on each call — jsonschema-go requires a strict tree with no shared pointers.
func programExerciseItemSchema() *jsonschema.Schema {
	noExtra := func() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }
	return &jsonschema.Schema{
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
		},
	}
}

// programSetItemSchema returns the JSON schema for a single set within a program.
// Returns a new value on each call — jsonschema-go requires a strict tree with no shared pointers.
func programSetItemSchema() *jsonschema.Schema {
	noExtra := func() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }
	return &jsonschema.Schema{
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
			"exercises": {
				Type:        "array",
				Description: "Exercises to perform within this set, in execution order.",
				Items:       programExerciseItemSchema(),
			},
		},
	}
}

// createProgramFullSchema returns an explicit JSON schema for create_program_full.
// This is required because jsonschema-go unconditionally emits ["null","array"] for
// Go slices, which makes required array fields appear nullable and confuses callers.
func createProgramFullSchema() *jsonschema.Schema {
	noExtra := func() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Required:             []string{"name", "sets"},
		Properties: map[string]*jsonschema.Schema{
			"name": {
				Type:        "string",
				Description: "Name of the program.",
			},
			"description": {
				Type:        "string",
				Description: "Detailed description of the program. Optional.",
			},
			"is_public": {
				Type:        "boolean",
				Description: "Whether the program is visible to all users. Defaults to false.",
			},
			"sets": {
				Type:        "array",
				Description: "Ordered list of training sets (blocks) that make up the program.",
				Items:       programSetItemSchema(),
			},
		},
	}
}

// replaceProgramFullSchema returns the schema for replace_program_full.
func replaceProgramFullSchema() *jsonschema.Schema {
	noExtra := func() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Required:             []string{"program_id", "sets"},
		Properties: map[string]*jsonschema.Schema{
			"program_id": {
				Type:        "integer",
				Description: "ID of the program whose structure will be replaced.",
			},
			"sets": {
				Type:        "array",
				Description: "New ordered list of training sets. Replaces all existing sets and exercises.",
				Items:       programSetItemSchema(),
			},
		},
	}
}

// reorderProgramStructureSchema returns the schema for reorder_program_structure.
func reorderProgramStructureSchema() *jsonschema.Schema {
	noExtra := func() *jsonschema.Schema { return &jsonschema.Schema{Not: &jsonschema.Schema{}} }

	setEntry := &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Required:             []string{"set_id", "exercise_ids"},
		Properties: map[string]*jsonschema.Schema{
			"set_id": {
				Type:        "integer",
				Description: "ID of the set.",
			},
			"exercise_ids": {
				Type:        "array",
				Description: "Exercise IDs in their new order within this set. Cross-set moves are allowed.",
				Items:       &jsonschema.Schema{Type: "integer"},
			},
		},
	}

	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: noExtra(),
		Required:             []string{"program_id", "sets"},
		Properties: map[string]*jsonschema.Schema{
			"program_id": {
				Type:        "integer",
				Description: "ID of the program to reorder.",
			},
			"sets": {
				Type:        "array",
				Description: "All sets in their new order, each with its exercises in their new order. Must include every existing set and exercise exactly once.",
				Items:       setEntry,
			},
		},
	}
}

// updateProgramSchema returns the schema for the patch-semantics update_program tool.
func updateProgramSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}},
		Required:             []string{"id"},
		Properties: map[string]*jsonschema.Schema{
			"id":          {Type: "integer", Description: "Program ID."},
			"updated_at":  {Type: "string", Format: "date-time", Description: "ISO 8601 version token for optimistic locking. Omit for last-write-wins."},
			"name":        {Type: "string", Description: "New name. Omit to leave unchanged."},
			"description": {Type: "string", Description: "New description. Omit to leave unchanged."},
			"activity":    {Type: "string", Description: "Primary sport or activity. Omit to leave unchanged."},
			"is_public":   {Type: "boolean", Description: "Visibility flag. Omit to leave unchanged."},
		},
	}
}

type updateProgramParams struct {
	ID          int64      `json:"id"`
	UpdatedAt   *time.Time `json:"updated_at"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Activity    *string    `json:"activity"`
	IsPublic    *bool      `json:"is_public"`
}

func registerProgramTools(server *mcp.Server, programs *service.ProgramService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_programs",
		Description: "List all programs",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
		list, err := programs.List(ctx)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: markdown.ProgramList(list)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_program",
		Description: "Get a program by ID including its full set and exercise hierarchy",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.Get(ctx, domain.ProgramID(params.ID))
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: markdown.ProgramFull(program.Program)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_program",
		Description: "Partially update a program's metadata. Only provided fields are changed; omitted fields retain their current values. updated_at is the version token for optimistic locking; omit for last-write-wins.",
		InputSchema: updateProgramSchema(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params updateProgramParams) (*mcp.CallToolResult, struct{}, error) {
		patch := service.ProgramPatch{UpdatedAt: params.UpdatedAt}
		if params.Name != nil {
			patch.Name = domain.Optional[string]{Value: *params.Name, Set: true}
		}
		if params.Description != nil {
			patch.Description = domain.Optional[*string]{Value: params.Description, Set: true}
		}
		if params.Activity != nil {
			patch.Activity = domain.Optional[*string]{Value: params.Activity, Set: true}
		}
		if params.IsPublic != nil {
			patch.IsPublic = domain.Optional[bool]{Value: *params.IsPublic, Set: true}
		}
		program, err := programs.Patch(ctx, domain.ProgramID(params.ID), patch)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: markdown.ProgramLiteResult(program)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_program",
		Description: "Delete a program by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := programs.Delete(ctx, domain.ProgramID(params.ID)); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_program_full",
		Description: "Creates a complete program with all its sets and exercises in a single atomic operation. All exercise_ids must exist before calling — use list_exercises and create_exercise to prepare the exercise library first. If any part fails nothing is written.",
		InputSchema: createProgramFullSchema(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		Name        string                    `json:"name"`
		Description *string                   `json:"description,omitempty"`
		Activity    *string                   `json:"activity,omitempty"`
		IsPublic    bool                      `json:"is_public"`
		Sets        []service.ProgramSetInput `json:"sets"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.CreateFull(ctx, params.Name, params.Description, params.Activity, params.IsPublic, params.Sets)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: markdown.ProgramLiteResult(program)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "replace_program_full",
		Description: "Atomically replaces all sets and exercises in an existing program. Use this to restructure a program in a single operation. All exercise_ids must already exist. Existing set/exercise IDs are discarded and new ones are assigned.",
		InputSchema: replaceProgramFullSchema(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID int64                     `json:"program_id"`
		Sets      []service.ProgramSetInput `json:"sets"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.ReplaceFull(ctx, domain.ProgramID(params.ProgramID), params.Sets)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: markdown.ProgramFull(program.Program)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "reorder_program_structure",
		Description: "Reorders sets and exercises within a program without changing their content. Every existing set and exercise must appear exactly once. Cross-set exercise moves are allowed.",
		InputSchema: reorderProgramStructureSchema(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID int64                    `json:"program_id"`
		Sets      []service.StructureEntry `json:"sets"`
	}) (*mcp.CallToolResult, struct{}, error) {
		program, err := programs.ReorderStructure(ctx, domain.ProgramID(params.ProgramID), params.Sets)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: markdown.ProgramFull(program.Program)}}}, struct{}{}, nil
	})
}
