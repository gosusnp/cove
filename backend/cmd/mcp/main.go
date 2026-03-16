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

	schema := os.Getenv("COVE_DB_SCHEMA")
	if schema == "" {
		schema = db.DefaultSchema
	}

	database := db.Open(dbURL, schema)
	defer database.Close()

	exStore := store.NewExerciseStore()
	exSvc := service.NewExerciseService(database, exStore)
	pSvc := service.NewProgramService(database, exStore)
	svcs := covemcp.Services{
		Exercises: exSvc,
		Programs:  pSvc,
	}

	server := covemcp.NewServer(svcs)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
