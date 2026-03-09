// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"embed"
	"io/fs"
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

//go:embed ui
var uiFS embed.FS

func main() {
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

	userStore := store.NewUserStore()
	orgStore := store.NewOrgStore()
	userSvc := service.NewUserService(database, userStore, orgStore)
	oauthCfg := &oauth2.Config{
		ClientID:     googleClientID,
		ClientSecret: googleClientSecret,
		RedirectURL:  googleRedirectURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}

	pSvc := service.NewProgramService(database)
	svcs := covemcp.Services{
		Exercises:        service.NewExerciseService(database, store.NewExerciseStore()),
		Programs:         pSvc,
		ProgramSets:      service.NewProgramSetService(pSvc),
		ProgramExercises: service.NewProgramExerciseService(store.NewProgramExerciseStore(database)),
	}

	mux := http.NewServeMux()
	mux.Handle("/api/", NewAPIHandler(userStore, userSvc, svcs))
	mux.Handle("/mcp/", middleware.OAuth(userSvc, covemcp.NewHTTPHandler(svcs)))

	var staticFS fs.FS
	if os.Getenv("COVE_DEV") != "" {
		log.Printf("serving local ui assets")
		// Serve from disk so frontend rebuilds are visible without restarting.
		staticFS = os.DirFS("ui")
	} else {
		log.Printf("serving embedded ui assets")
		var err error
		staticFS, err = fs.Sub(uiFS, "ui")
		if err != nil {
			log.Fatal(err)
		}
	}
	fileServer := http.FileServerFS(staticFS)
	spaHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Try to serve the file; fall back to index.html for SPA routing.
		_, err := staticFS.Open(strings.TrimLeft(r.URL.Path, "/"))
		if err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})

	// Outer mux: UI at / (no auth), everything else to mux.
	outer := http.NewServeMux()
	oauthHandler := handlers.NewOAuthHandler(oauthCfg, userSvc, allowedEmails)
	oauthHandler.RegisterRoutes(outer)
	if os.Getenv("COVE_DEV") != "" {
		oauthHandler.RegisterDevRoutes(outer)
	}
	outer.Handle("/api/", mux)
	outer.Handle("/mcp/", mux)
	outer.Handle("/", spaHandler)

	port := os.Getenv("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.Logging(outer)))
}
