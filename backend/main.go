// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	_ "embed"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/handlers"
	covemcp "github.com/gosusnp/cove/backend/internal/mcp"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
)

//go:embed cove.html
var uiHTML []byte

func main() {
	apiKey := os.Getenv("COVE_API_KEY")
	if apiKey == "" {
		log.Fatal("COVE_API_KEY is required")
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}
	googleClientSecret := os.Getenv("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET is required")
	}
	googleRedirectURL := os.Getenv("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		log.Fatal("GOOGLE_REDIRECT_URL is required")
	}

	var allowedEmails []string
	if raw := os.Getenv("COVE_ALLOWED_EMAILS"); raw != "" {
		allowedEmails = strings.Split(raw, ",")
	}

	database := db.Open(dbURL)
	defer database.Close()

	userStore := store.NewUserStore(database)
	oauthCfg := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}

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
	mux.Handle("/mcp/", middleware.OAuth(userStore, covemcp.NewHTTPHandler(svcs)))

	// Outer mux: UI at / (no auth), everything else to mux.
	outer := http.NewServeMux()
	outer.HandleFunc("GET /{$}", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write(uiHTML)
	})
	handlers.NewOAuthHandler(oauthCfg, userStore, allowedEmails).RegisterRoutes(outer)
	outer.Handle("/", mux)

	port := os.Getenv("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, outer))
}
