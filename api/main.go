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

	exerciseSvc := service.NewExerciseService(store.NewExerciseStore(database))
	handlers.NewExerciseHandler(exerciseSvc).RegisterRoutes(mux)
	handlers.NewProgramHandler(service.NewProgramService(store.NewProgramStore(database))).RegisterRoutes(mux)
	handlers.NewProgramSetHandler(service.NewProgramSetService(store.NewProgramSetStore(database))).RegisterRoutes(mux)
	handlers.NewProgramExerciseHandler(service.NewProgramExerciseService(store.NewProgramExerciseStore(database))).RegisterRoutes(mux)

	mux.Handle("/mcp/", covemcp.NewHTTPHandler(exerciseSvc))

	port := os.Getenv("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.APIKey(apiKey, mux)))
}
