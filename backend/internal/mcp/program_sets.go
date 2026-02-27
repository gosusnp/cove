// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"encoding/json"

	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func registerProgramSetTools(server *mcp.Server, sets *service.ProgramSetService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_program_set",
		Description: "Add a set (block) to a program",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID           int64   `json:"program_id"`
		Name                *string `json:"name,omitempty"`
		Rounds              int     `json:"rounds"`
		IntraSetRestSeconds *int    `json:"rest_s,omitempty"`
		SortOrder           *int    `json:"sort_order,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		ps, err := sets.Create(params.ProgramID, params.Name, params.Rounds, params.IntraSetRestSeconds, params.SortOrder)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(ps)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_program_set",
		Description: "Update a program set's name, rounds, or rest period",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID           int64   `json:"program_id"`
		ID                  int64   `json:"id"`
		Name                *string `json:"name,omitempty"`
		Rounds              int     `json:"rounds"`
		IntraSetRestSeconds *int    `json:"rest_s,omitempty"`
		SortOrder           *int    `json:"sort_order,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		ps, err := sets.Update(params.ProgramID, params.ID, params.Name, params.Rounds, params.IntraSetRestSeconds, params.SortOrder)
		if err != nil {
			return nil, struct{}{}, err
		}
		b, err := json.Marshal(ps)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(b)}}}, struct{}{}, nil
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "delete_program_set",
		Description: "Delete a set from a program",
	}, func(_ context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID int64 `json:"program_id"`
		ID        int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := sets.Delete(params.ProgramID, params.ID); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})
}
