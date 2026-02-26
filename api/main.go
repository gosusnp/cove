package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gosusnp/cove/api/db"
	"github.com/gosusnp/cove/api/handlers"
	covemcp "github.com/gosusnp/cove/api/mcp"
	"github.com/gosusnp/cove/api/middleware"
	"github.com/gosusnp/cove/api/service"
	"github.com/gosusnp/cove/api/store"
)

func main() {
	apiKey := os.Getenv("COVE_API_KEY")
	if apiKey == "" {
		log.Fatal("COVE_API_KEY is required")
	}

	dbPath := os.Getenv("COVE_DB_PATH")
	if dbPath == "" {
		dbPath = "cove.db"
	}

	database := db.Open(dbPath)
	defer database.Close()

	mux := http.NewServeMux()

	svcs := covemcp.Services{
		Exercises:        service.NewExerciseService(store.NewExerciseStore(database)),
		Programs:         service.NewProgramService(store.NewProgramStore(database)),
		ProgramSets:      service.NewProgramSetService(store.NewProgramSetStore(database)),
		ProgramExercises: service.NewProgramExerciseService(store.NewProgramExerciseStore(database)),
	}

	handlers.NewExerciseHandler(svcs.Exercises).RegisterRoutes(mux)
	handlers.NewProgramHandler(svcs.Programs).RegisterRoutes(mux)
	handlers.NewProgramSetHandler(svcs.ProgramSets).RegisterRoutes(mux)
	handlers.NewProgramExerciseHandler(svcs.ProgramExercises).RegisterRoutes(mux)

	mux.Handle("/mcp/", covemcp.NewHTTPHandler(svcs))

	port := os.Getenv("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.APIKey(apiKey, mux)))
}
