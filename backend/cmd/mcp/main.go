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

	exSvc := service.NewExerciseService(database, store.NewExerciseStore())
	pSvc := service.NewProgramService(database, exSvc)
	svcs := covemcp.Services{
		Exercises: exSvc,
		Programs:  pSvc,
	}

	server := covemcp.NewServer(svcs)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
