// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"encoding/json"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerExerciseTools(server *mcp.Server, exercises *service.ExerciseService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_exercises",
		Description: "List all exercises",
	}, func(_ context.Context, _ *mcp.CallToolRequest, _ struct{}) (*mcp.CallToolResult, struct{}, error) {
		list, err := exercises.List()
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
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		exercise, err := exercises.Get(domain.ExerciseID(params.ID))
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
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		Name        string  `json:"name"`
		Progression *string `json:"progression,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		exercise, err := exercises.Create(params.Name, params.Progression)
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
		Description: "Update an exercise's name or progression",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ID          int64   `json:"id"`
		Name        string  `json:"name"`
		Progression *string `json:"progression,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		exercise, err := exercises.Update(domain.ExerciseID(params.ID), params.Name, params.Progression)
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
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ID int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := exercises.Delete(domain.ExerciseID(params.ID)); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})
}
