package main

import (
	"context"
	"log"
	"os"

	"github.com/gosusnp/cove/api/db"
	covemcp "github.com/gosusnp/cove/api/mcp"
	"github.com/gosusnp/cove/api/service"
	"github.com/gosusnp/cove/api/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func main() {
	dbPath := os.Getenv("COVE_DB_PATH")
	if dbPath == "" {
		dbPath = "cove.db"
	}

	database := db.Open(dbPath)
	defer database.Close()

	exerciseSvc := service.NewExerciseService(store.NewExerciseStore(database))
	server := covemcp.NewServer(exerciseSvc)

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		log.Fatal(err)
	}
}
