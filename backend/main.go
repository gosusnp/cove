// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"context"
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/fdc"
	"github.com/gosusnp/cove/backend/internal/handlers"
	"github.com/gosusnp/cove/backend/internal/llm"
	covemcp "github.com/gosusnp/cove/backend/internal/mcp"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/workers"
)

//go:embed ui
var uiFS embed.FS

// getSecret reads a config value using the following precedence:
//  1. $COVE_SECRETS_DIR/<key>  — mounted secrets directory
//  2. $<key>_FILE              — explicit file path (e.g. for CNPG-generated secrets)
//  3. $<key>                   — plain env var (local dev)
func getSecret(key string) string {
	if dir := os.Getenv("COVE_SECRETS_DIR"); dir != "" {
		data, err := os.ReadFile(filepath.Join(dir, key))
		if err == nil {
			return strings.TrimSpace(string(data))
		}
	}
	if path := os.Getenv(key + "_FILE"); path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			log.Fatalf("read secret file for %s: %v", key, err)
		}
		return strings.TrimSpace(string(data))
	}
	return os.Getenv(key)
}

// resolveMigrationDSN returns the DSN to use for migrations and whether it is valid.
// In dev mode, falls back to appDSN when MIGRATION_DATABASE_URL is unset.
// In production, returns ("", false) when MIGRATION_DATABASE_URL is unset.
func resolveMigrationDSN(appDSN string, dev bool) (string, bool) {
	if dsn := getSecret("MIGRATION_DATABASE_URL"); dsn != "" {
		return dsn, true
	}
	if dev {
		return appDSN, true
	}
	return "", false
}

func main() {
	dbURL := getSecret("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	migrationDBURL, ok := resolveMigrationDSN(dbURL, os.Getenv("COVE_DEV") != "")
	if !ok {
		log.Fatal("MIGRATION_DATABASE_URL is required in production")
	}

	googleClientID := getSecret("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		log.Fatal("GOOGLE_CLIENT_ID is required")
	}
	googleClientSecret := getSecret("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		log.Fatal("GOOGLE_CLIENT_SECRET is required")
	}
	googleRedirectURL := getSecret("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		log.Fatal("GOOGLE_REDIRECT_URL is required")
	}

	encryptionKey := getSecret("SESSION_ENCRYPTION_KEY")
	if encryptionKey == "" {
		log.Fatal("SESSION_ENCRYPTION_KEY is required")
	}
	enc, err := crypto.NewAESEncryptor(0, map[byte]string{0: encryptionKey})
	if err != nil {
		log.Fatalf("init encryptor: %v", err)
	}

	var allowedEmails []string
	if raw := getSecret("COVE_ALLOWED_EMAILS"); raw != "" {
		allowedEmails = strings.Split(raw, ",")
	}

	db.Migrate(migrationDBURL)
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

	fdcClient := fdc.NewClient(getSecret("FDC_API_KEY"))

	exStore := store.NewExerciseStore()
	exSvc := service.NewExerciseService(database, exStore)
	pSvc := service.NewProgramService(database, exStore)
	wsSvc := service.NewWorkoutSessionService(database, store.NewWorkoutSessionStore(), enc)
	tpSvc := service.NewTrainingProfileService(database, store.NewTrainingProfileStore(), enc)
	ingSvc := service.NewIngredientService(database, store.NewIngredientStore(), fdcClient)
	recipeSvc := service.NewRecipeService(database, store.NewRecipeStore())
	prepSvc := service.NewPreparationService(database, store.NewPreparationStore())
	oauthSvc := service.NewOAuthService(database, store.NewOAuthStore(), userStore)
	mcpSvcs := covemcp.Services{
		Exercises: exSvc,
		Programs:  pSvc,
		Profiles:  tpSvc,
	}

	apiSvcs := Services{
		Users:            userSvc,
		Exercises:        exSvc,
		Programs:         pSvc,
		WorkoutSessions:  wsSvc,
		TrainingProfiles: tpSvc,
		Ingredients:      ingSvc,
		Recipes:          recipeSvc,
		Preparations:     prepSvc,
		FDCClient:        fdcClient,
	}

	secureCookies := os.Getenv("COVE_DEV") == ""

	mux := http.NewServeMux()
	mux.Handle("/api/", NewAPIHandler(apiSvcs, secureCookies))
	mux.Handle("/mcp/", middleware.OAuth(userSvc, covemcp.NewHTTPHandler(mcpSvcs)))

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
	oauthHandler := handlers.NewOAuthHandler(oauthCfg, userSvc, allowedEmails, secureCookies)
	oauthHandler.RegisterRoutes(outer)
	oauthHandler.RegisterMobileRoutes(outer)
	handlers.NewOAuthServerHandler(oauthSvc, userSvc, secureCookies).RegisterRoutes(outer)
	if os.Getenv("COVE_DEV") != "" {
		oauthHandler.RegisterDevRoutes(outer)
	}
	outer.Handle("/api/", mux)
	outer.Handle("/mcp/", mux)
	outer.Handle("/", spaHandler)

	if os.Getenv("COVE_WORKER_ENABLED") == "true" {
		if os.Getenv("HATCHET_CLIENT_TOKEN") == "" {
			log.Fatal("HATCHET_CLIENT_TOKEN is required when COVE_WORKER_ENABLED=true")
		}
		llmClient := llm.NewOpenAICompatClient(llm.Config{
			BaseURL: getSecret("LLM_BASE_URL"),
			APIKey:  getSecret("LLM_API_KEY"),
			Model:   getSecret("LLM_MODEL"),
		})
		summarizeSvc := service.NewSummarizeService(llmClient)
		adapter := workers.NewLocalWorkoutSessionAdapter(wsSvc)
		workerCtx, workerStop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer workerStop()
		if err := workers.StartWorker(workerCtx, adapter, summarizeSvc); err != nil {
			log.Fatalf("start hatchet worker: %v", err)
		}
		log.Printf("hatchet worker started")
	}

	port := getSecret("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("cove listening on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, middleware.Logging(middleware.CORS([]string{"capacitor://localhost"}, outer))))
}
