package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gosusnp/cove/api/db"
	"github.com/gosusnp/cove/api/handlers"
	"github.com/gosusnp/cove/api/middleware"
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

	handlers.NewExerciseHandler(store.NewExerciseStore(database)).RegisterRoutes(mux)
	handlers.NewProgramHandler(store.NewProgramStore(database)).RegisterRoutes(mux)
	handlers.NewProgramSetHandler(store.NewProgramSetStore(database)).RegisterRoutes(mux)

	port := os.Getenv("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.APIKey(apiKey, mux)))
}
