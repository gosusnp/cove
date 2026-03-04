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

	svcs := covemcp.Services{
		Exercises:        service.NewExerciseService(database, store.NewExerciseStore()),
		Programs:         service.NewProgramService(database),
		ProgramSets:      service.NewProgramSetService(store.NewProgramSetStore(database)),
		ProgramExercises: service.NewProgramExerciseService(store.NewProgramExerciseStore(database)),
	}

	server := covemcp.NewServer(svcs)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
