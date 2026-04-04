// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// updateProgramExerciseSchema returns an explicit JSON schema for update_program_exercise.
// A custom schema is required because the params struct uses domain.Optional[T] fields,
// which jsonschema-go renders as {Value, Set} objects rather than plain primitives.
func updateProgramExerciseSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}},
		Required:             []string{"program_id", "set_id", "id"},
		Properties: map[string]*jsonschema.Schema{
			"program_id":  {Type: "integer", Description: "ID of the program that owns the exercise."},
			"set_id":      {Type: "integer", Description: "ID of the set that owns the exercise."},
			"id":          {Type: "integer", Description: "ID of the program exercise to update."},
			"updated_at":  {Type: "string", Format: "date-time", Description: "ISO 8601 version token for optimistic locking. Omit for last-write-wins."},
			"exercise_id": {Type: "integer", Description: "ID of the exercise definition. Omit to leave unchanged."},
			"laterality":  {Type: "string", Description: "bilateral | unilateral | left | right | alternating. Omit to leave unchanged."},
			"reps":        {Type: "integer", Description: "Target reps per round. Omit to leave unchanged."},
			"duration_s":  {Type: "integer", Description: "Target duration in seconds per round. Omit to leave unchanged."},
			"weight":      {Type: "number", Description: "Load in the specified unit. Positive = added, negative = assisted, 0 = bodyweight. Omit to leave unchanged."},
			"weight_unit": {Type: "string", Description: "Unit for the weight: kg or lb. Omit to leave unchanged."},
		},
	}
}

type updateProgramExerciseParams struct {
	ProgramID             int64        `json:"program_id"`
	SetID                 int64        `json:"set_id"`
	ID                    int64        `json:"id"`
	UpdatedAt             *time.Time   `json:"updated_at"`
	ExerciseID            *int64       `json:"exercise_id"`
	Laterality            *string      `json:"laterality"`
	TargetReps            *int         `json:"reps"`
	TargetDurationSeconds *int         `json:"duration_s"`
	TargetWeight          *float64     `json:"weight"`
	WeightUnit            *domain.Unit `json:"weight_unit"`
}

func registerProgramExerciseTools(server *mcp.Server, svc *service.ProgramService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_program_exercise",
		Description: "Partially update a program exercise. Only provided fields are changed; omitted fields retain their current values. updated_at is the version token for optimistic locking; omit for last-write-wins.",
		InputSchema: updateProgramExerciseSchema(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params updateProgramExerciseParams) (*mcp.CallToolResult, struct{}, error) {
		patch := service.ProgramExercisePatch{UpdatedAt: params.UpdatedAt}
		if params.ExerciseID != nil {
			patch.ExerciseID = domain.Optional[domain.ExerciseID]{Value: domain.ExerciseID(*params.ExerciseID), Set: true}
		}
		if params.Laterality != nil {
			patch.Laterality = domain.Optional[*string]{Value: params.Laterality, Set: true}
		}
		if params.TargetReps != nil {
			patch.TargetReps = domain.Optional[*int]{Value: params.TargetReps, Set: true}
		}
		if params.TargetDurationSeconds != nil {
			patch.TargetDurationSeconds = domain.Optional[*int]{Value: params.TargetDurationSeconds, Set: true}
		}
		if params.TargetWeight != nil {
			patch.TargetWeight = domain.Optional[*float64]{Value: params.TargetWeight, Set: true}
		}
		if params.WeightUnit != nil {
			patch.WeightUnit = domain.Optional[*domain.Unit]{Value: params.WeightUnit, Set: true}
		}
		pe, err := svc.PatchExercise(ctx, domain.ProgramID(params.ProgramID), params.SetID, params.ID, patch)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(pe)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

}
