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

func registerExerciseTools(server *mcp.Server, exercises *service.ExerciseService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_exercises",
		Description: "List all exercises",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
		list, err := exercises.List(ctx)
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
		Name:        "get_exercise",
		Description: "Get an exercise by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		exercise, err := exercises.Get(ctx, domain.ExerciseID(params.ID))
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(exercise)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_exercise",
		Description: "Create a new exercise",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		Name        string  `json:"name"`
		Progression *string `json:"progression,omitempty"`
		Description *string `json:"description,omitempty"`
		IsPublic    bool    `json:"is_public"`
	}) (*mcp.CallToolResult, struct{}, error) {
		exercise, err := exercises.Create(ctx, params.Name, params.Progression, params.Description, params.IsPublic)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(exercise)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_exercise",
		Description: "Update an exercise's name, progression, description or visibility. updated_at is the version token for optimistic locking; omit for last-write-wins.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ID          int64      `json:"id"`
		UpdatedAt   *time.Time `json:"updated_at,omitempty"`
		Name        string     `json:"name"`
		Progression *string    `json:"progression,omitempty"`
		Description *string    `json:"description,omitempty"`
		IsPublic    bool       `json:"is_public"`
	}) (*mcp.CallToolResult, struct{}, error) {
		exercise, err := exercises.Update(ctx, domain.ExerciseID(params.ID), params.UpdatedAt, params.Name, params.Progression, params.Description, params.IsPublic)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(exercise)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_exercise",
		Description: "Delete an exercise by ID",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := exercises.Delete(ctx, domain.ExerciseID(params.ID)); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})
}
