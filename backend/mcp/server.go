// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"net/http"

	"github.com/gosusnp/cove/backend/service"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Services struct {
	Exercises        *service.ExerciseService
	Programs         *service.ProgramService
	ProgramSets      *service.ProgramSetService
	ProgramExercises *service.ProgramExerciseService
}

func NewServer(svcs Services) *mcp.Server {
	server := mcp.NewServer(&mcp.Implementation{Name: "cove", Version: "1.0.0"}, nil)
	registerExerciseTools(server, svcs.Exercises)
	registerProgramTools(server, svcs.Programs)
	registerProgramSetTools(server, svcs.ProgramSets)
	registerProgramExerciseTools(server, svcs.ProgramExercises)
	return server
}

func NewHTTPHandler(svcs Services) http.Handler {
	server := NewServer(svcs)
	return mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server {
		return server
	}, nil)
}
