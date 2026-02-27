// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"encoding/json"

	"github.com/gosusnp/cove/backend/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerProgramExerciseTools(server *mcp.Server, exercises *service.ProgramExerciseService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_program_exercise",
		Description: "Add an exercise to a program set",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		SetID                 int64    `json:"set_id"`
		ExerciseID            int64    `json:"exercise_id"`
		Laterality            *string  `json:"laterality,omitempty"`
		TargetReps            *int     `json:"reps,omitempty"`
		TargetDurationSeconds *int     `json:"duration_s,omitempty"`
		TargetWeightKg        *float64 `json:"weight_kg,omitempty"`
		SortOrder             *int     `json:"sort_order,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		pe, err := exercises.Create(params.SetID, params.ExerciseID, params.Laterality, params.TargetReps, params.TargetDurationSeconds, params.TargetWeightKg, params.SortOrder)
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
		Description: "Update a program exercise's targets or laterality",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		SetID                 int64    `json:"set_id"`
		ID                    int64    `json:"id"`
		ExerciseID            int64    `json:"exercise_id"`
		Laterality            *string  `json:"laterality,omitempty"`
		TargetReps            *int     `json:"reps,omitempty"`
		TargetDurationSeconds *int     `json:"duration_s,omitempty"`
		TargetWeightKg        *float64 `json:"weight_kg,omitempty"`
		SortOrder             *int     `json:"sort_order,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		pe, err := exercises.Update(params.SetID, params.ID, params.ExerciseID, params.Laterality, params.TargetReps, params.TargetDurationSeconds, params.TargetWeightKg, params.SortOrder)
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
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		SetID int64 `json:"set_id"`
		ID    int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := exercises.Delete(params.SetID, params.ID); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})
}
