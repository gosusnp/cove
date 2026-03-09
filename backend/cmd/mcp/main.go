// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"context"
	"log"
	"os"

	"github.com/gosusnp/cove/backend/internal/db"
	covemcp "github.com/gosusnp/cove/backend/internal/mcp"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	database := db.Open(dbURL)
	defer database.Close()

	pSvc := service.NewProgramService(database)
	svcs := covemcp.Services{
		Exercises:        service.NewExerciseService(database, store.NewExerciseStore()),
		Programs:         pSvc,
		ProgramSets:      service.NewProgramSetService(pSvc),
		ProgramExercises: service.NewProgramExerciseService(database, store.NewProgramExerciseStore()),
	}

	server := covemcp.NewServer(svcs)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
