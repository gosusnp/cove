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

func registerProgramSetTools(server *mcp.Server, svc *service.ProgramService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "create_program_set",
		Description: "Add a set (block) to a program",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID           int64   `json:"program_id"`
		Name                *string `json:"name,omitempty"`
		Rounds              int     `json:"rounds"`
		IntraSetRestSeconds *int    `json:"rest_s,omitempty"`
	}) (*mcp.CallToolResult, struct{}, error) {
		ps, err := svc.CreateSet(ctx, domain.ProgramID(params.ProgramID), params.Name, params.Rounds, params.IntraSetRestSeconds)
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
		Description: "Partially update a program set. Only provided fields are changed; omitted fields retain their current values. updated_at is the version token for optimistic locking; omit for last-write-wins.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID           int64                    `json:"program_id"`
		ID                  int64                    `json:"id"`
		UpdatedAt           *time.Time               `json:"updated_at,omitempty"`
		Name                domain.Optional[*string] `json:"name"`
		Rounds              domain.Optional[int]     `json:"rounds"`
		IntraSetRestSeconds domain.Optional[*int]    `json:"rest_s"`
	}) (*mcp.CallToolResult, struct{}, error) {
		patch := service.ProgramSetPatch{
			UpdatedAt:           params.UpdatedAt,
			Name:                params.Name,
			Rounds:              params.Rounds,
			IntraSetRestSeconds: params.IntraSetRestSeconds,
		}
		ps, err := svc.PatchSet(ctx, domain.ProgramID(params.ProgramID), params.ID, patch)
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
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params struct {
		ProgramID int64 `json:"program_id"`
		ID        int64 `json:"id"`
	}) (*mcp.CallToolResult, struct{}, error) {
		if err := svc.DeleteSet(ctx, domain.ProgramID(params.ProgramID), params.ID); err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: "deleted"}}}, struct{}{}, nil
	})
}
