package mcp

import (
	"context"
	"encoding/json"

	"github.com/gosusnp/cove/api/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

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
