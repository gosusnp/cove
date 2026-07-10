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

//go:embed mobile.zip
var mobileBundleZip []byte

// appVersion is set via -ldflags "-X main.appVersion=<git-sha>" at build time.
// Empty in local dev; the update endpoint returns "no update" when empty.
var appVersion string

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	enc, err := crypto.NewAESEncryptor(0, map[byte]string{0: cfg.EncryptionKey})
	if err != nil {
		log.Fatalf("init encryptor: %v", err)
	}

	db.Migrate(cfg.MigrationDatabaseURL)
	database := db.Open(cfg.DatabaseURL)
	defer database.Close()

	userStore := store.NewUserStore()
	orgStore := store.NewOrgStore()
	userSvc := service.NewUserService(database, userStore, orgStore)
	oauthCfg := &oauth2.Config{
		ClientID:     cfg.GoogleClientID,
		ClientSecret: cfg.GoogleClientSecret,
		RedirectURL:  cfg.GoogleRedirectURL,
		Scopes:       []string{"openid", "email"},
		Endpoint:     google.Endpoint,
	}

	fdcClient := fdc.NewClient(cfg.FDCAPIKey)

	exStore := store.NewExerciseStore()
	exSvc := service.NewExerciseService(database, exStore)
	pSvc := service.NewProgramService(database, exStore)
	wsSvc := service.NewWorkoutSessionService(database, store.NewWorkoutSessionStore(), enc)
	tpSvc := service.NewTrainingProfileService(database, store.NewTrainingProfileStore(), enc)
	ingSvc := service.NewIngredientService(database, store.NewIngredientStore(), fdcClient)
	recipeSvc := service.NewRecipeService(database, store.NewRecipeStore())
	prepSvc := service.NewPreparationService(database, store.NewPreparationStore())
	oauthSvc := service.NewOAuthService(database, store.NewOAuthStore(), userStore)
	llmRouter := llm.NewStaticRouter(llm.NewOpenAICompatClient(llm.Config{
		BaseURL: cfg.LLMBaseURL,
		APIKey:  cfg.LLMAPIKey,
		Model:   cfg.LLMModel,
	}))
	summarizeSvc := service.NewSummarizeService(llmRouter)

	mcpSvcs := covemcp.Services{
		Exercises: exSvc,
		Programs:  pSvc,
		Profiles:  tpSvc,
		Sessions:  wsSvc,
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

	var hatchetClient *workers.HatchetClient
	if cfg.HatchetToken != "" {
		hatchetClient, err = workers.NewHatchetClient(cfg.HatchetToken)
		if err != nil {
			log.Fatalf("create hatchet client: %v", err)
		}
		apiSvcs.HatchetClient = hatchetClient
	}

	secureCookies := !cfg.Dev

	mux := http.NewServeMux()
	mux.Handle("/api/", NewAPIHandler(apiSvcs, secureCookies))
	mux.Handle("/mcp/", middleware.OAuth(userSvc, covemcp.NewHTTPHandler(mcpSvcs)))

	var staticFS fs.FS
	if cfg.Dev {
		log.Printf("serving local ui assets")
		staticFS = os.DirFS("ui")
	} else {
		log.Printf("serving embedded ui assets")
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
	oauthHandler := handlers.NewOAuthHandler(oauthCfg, userSvc, cfg.AllowedEmails, secureCookies)
	oauthHandler.RegisterRoutes(outer)
	oauthHandler.RegisterMobileRoutes(outer)
	handlers.NewOAuthServerHandler(oauthSvc, userSvc, secureCookies).RegisterRoutes(outer)
	if cfg.Dev {
		oauthHandler.RegisterDevRoutes(outer)
	}
	handlers.NewOTAHandler(appVersion, mobileBundleZip).RegisterRoutes(outer)
	outer.Handle("/api/", mux)
	outer.Handle("/mcp/", mux)
	outer.Handle("/", spaHandler)

	if cfg.WorkerEnabled {
		if hatchetClient == nil {
			log.Fatalf("COVE_WORKER_ENABLED is true but HATCHET_CLIENT_TOKEN is not set")
		}
		adapter := workers.NewLocalWorkoutSessionAdapter(wsSvc)
		workerCtx, workerStop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		defer workerStop()
		if err := workers.StartWorker(workerCtx, hatchetClient, adapter, summarizeSvc); err != nil {
			log.Fatalf("start hatchet worker: %v", err)
		}
		log.Printf("hatchet worker started")
	}

	log.Printf("cove listening on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, middleware.Logging(middleware.CORS([]string{"capacitor://localhost"}, outer))))
}
