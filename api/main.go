package main

import (
	_ "embed"
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

//go:embed cove.html
var uiHTML []byte

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

	svcs := covemcp.Services{
		Exercises:        service.NewExerciseService(store.NewExerciseStore(database)),
		Programs:         service.NewProgramService(database),
		ProgramSets:      service.NewProgramSetService(store.NewProgramSetStore(database)),
		ProgramExercises: service.NewProgramExerciseService(store.NewProgramExerciseStore(database)),
	}

	// API sub-mux: handlers register routes without a prefix (e.g. /exercises).
	// Mounted at /api/ via StripPrefix so no handler files need changing.
	apiMux := http.NewServeMux()
	handlers.NewExerciseHandler(svcs.Exercises).RegisterRoutes(apiMux)
	handlers.NewProgramHandler(svcs.Programs).RegisterRoutes(apiMux)
	handlers.NewProgramSetHandler(svcs.ProgramSets).RegisterRoutes(apiMux)
	handlers.NewProgramExerciseHandler(svcs.ProgramExercises).RegisterRoutes(apiMux)

	mux := http.NewServeMux()
	mux.Handle("/api/", http.StripPrefix("/api", middleware.APIKey(apiKey, apiMux)))
	mux.Handle("/mcp/", middleware.APIKey(apiKey, covemcp.NewHTTPHandler(svcs)))

	// Outer mux: UI at / (no auth), everything else to mux.
	outer := http.NewServeMux()
	outer.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(uiHTML)
	})
	outer.Handle("/", mux)

	port := os.Getenv("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, outer))
}
