// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerProgramExerciseTools(server *mcp.Server, svc *service.ProgramService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_program_exercise",
		Description: "Add an exercise to a program set",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID             int64        `json:"program_id"`
		SetID                 int64        `json:"set_id"`
		ExerciseID            int64        `json:"exercise_id"`
		Laterality            *string      `json:"laterality,omitempty"`
		TargetReps            *int         `json:"reps,omitempty"`
		TargetDurationSeconds *int         `json:"duration_s,omitempty"`
		TargetWeight          *float64     `json:"weight,omitempty"`
		WeightUnit            *domain.Unit `json:"weight_unit,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		pe, err := svc.CreateExercise(ctx, domain.ProgramID(params.ProgramID), params.SetID, domain.ExerciseID(params.ExerciseID), params.Laterality, params.TargetReps, params.TargetDurationSeconds, params.TargetWeight, params.WeightUnit)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(pe)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_program_exercise",
		Description: "Partially update a program exercise. Only provided fields are changed; omitted fields retain their current values. updated_at is the version token for optimistic locking; omit for last-write-wins.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID             int64                              `json:"program_id"`
		SetID                 int64                              `json:"set_id"`
		ID                    int64                              `json:"id"`
		UpdatedAt             *time.Time                         `json:"updated_at,omitempty"`
		ExerciseID            domain.Optional[domain.ExerciseID] `json:"exercise_id"`
		Laterality            domain.Optional[*string]           `json:"laterality"`
		TargetReps            domain.Optional[*int]              `json:"reps"`
		TargetDurationSeconds domain.Optional[*int]              `json:"duration_s"`
		TargetWeight          domain.Optional[*float64]          `json:"weight"`
		WeightUnit            domain.Optional[*domain.Unit]      `json:"weight_unit"`
	}) (*mcp.CallToolResult, struct{}, error) {
		patch := service.ProgramExercisePatch{
			UpdatedAt:             params.UpdatedAt,
			ExerciseID:            params.ExerciseID,
			Laterality:            params.Laterality,
			TargetReps:            params.TargetReps,
			TargetDurationSeconds: params.TargetDurationSeconds,
			TargetWeight:          params.TargetWeight,
			WeightUnit:            params.WeightUnit,
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

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_program_exercise",
		Description: "Remove an exercise from a program set",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID int64 `json:"program_id"`
		SetID     int64 `json:"set_id"`
		ID        int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := svc.DeleteExercise(ctx, domain.ProgramID(params.ProgramID), params.SetID, params.ID); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})
}
