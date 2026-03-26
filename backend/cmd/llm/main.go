// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Command llm is a standalone CLI for iterating on LLM prompt workflows
// without running the full HTTP server. It wires the same service layer used
// by the application so that prompt changes tested here behave identically in
// production.
//
// Usage:
//
//	llm [--preview] session <id>
//
// --preview renders the prompt and prints it without calling the LLM.
// DATABASE_URL and SESSION_ENCRYPTION_KEY are always required.
// LLM_BASE_URL and LLM_MODEL are only required without --preview.
//
// Required environment variables:
//
//	DATABASE_URL            — PostgreSQL connection string
//	SESSION_ENCRYPTION_KEY  — base64-encoded 32-byte AES key
//
// Required without --preview:
//
//	LLM_BASE_URL            — OpenAI-compatible endpoint (e.g. https://api.openai.com/v1)
//	LLM_MODEL               — model name (e.g. gpt-4o, llama-3.1-70b)
//
// Optional:
//
//	LLM_API_KEY             — Bearer token; omit for unauthenticated endpoints (KubeAI)
package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/llm"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
)

func main() {
	preview := flag.Bool("preview", false, "render prompt and print without calling the LLM")
	flag.Parse()

	args := flag.Args()
	if len(args) < 2 {
		log.Fatalf("usage: llm [--preview] <command> <id>\n  commands: session")
	}
	cmd, rawID := args[0], args[1]

	enc := mustEncryptor()
	database := db.Open(mustEnv("DATABASE_URL"))
	defer database.Close()

	wsSvc := service.NewWorkoutSessionService(database, store.NewWorkoutSessionStore(), enc)

	var summarizeSvc *service.SummarizeService
	if *preview {
		summarizeSvc = service.NewSummarizeService(nil)
	} else {
		summarizeSvc = service.NewSummarizeService(llm.NewOpenAICompatClient(llm.Config{
			BaseURL: mustEnv("LLM_BASE_URL"),
			APIKey:  os.Getenv("LLM_API_KEY"),
			Model:   mustEnv("LLM_MODEL"),
			Debug:   os.Getenv("LLM_DEBUG") != "",
		}))
	}

	switch cmd {
	case "session":
		runSession(rawID, *preview, database, wsSvc, summarizeSvc)
	default:
		log.Fatalf("unknown command: %s", cmd)
	}
}

func runSession(rawID string, preview bool, database *sql.DB, wsSvc *service.WorkoutSessionService, summarizeSvc *service.SummarizeService) {
	n, err := strconv.ParseInt(rawID, 10, 64)
	if err != nil {
		log.Fatalf("invalid session id %q: %v", rawID, err)
	}
	id := domain.WorkoutSessionID(n)

	timeout := optDuration("LLM_TIMEOUT")
	if timeout == 0 {
		timeout = 5 * time.Minute
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	// Bootstrap an identity from the session row so we can call the service
	// with the correct org/user scope. The CLI is a trusted dev tool — it has
	// direct DB access and does not go through OAuth.
	identity, err := sessionIdentity(ctx, database, id)
	if err != nil {
		log.Fatalf("resolve identity for session %d: %v", id, err)
	}
	ctx = domain.NewContext(ctx, identity)

	ws, err := wsSvc.Get(ctx, id)
	if err != nil {
		log.Fatalf("get session %d: %v", id, err)
	}

	if preview {
		if err := ws.UseSensitiveData(ctx, func(sd domain.SessionSensitiveData) error {
			return summarizeSvc.PreviewSession(os.Stdout, ws, sd)
		}); err != nil {
			log.Fatalf("preview session %d: %v", id, err)
		}
		return
	}

	var result *service.SessionSummary
	if err := ws.UseSensitiveData(ctx, func(sd domain.SessionSensitiveData) error {
		var err error
		result, err = summarizeSvc.SummarizeSession(ctx, ws, sd)
		return err
	}); err != nil {
		log.Fatalf("summarize session %d: %v", id, err)
	}

	out, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		log.Fatalf("marshal result: %v", err)
	}
	fmt.Println(string(out))
}

// sessionIdentity queries the user_id and org_id for a session so the CLI can
// build a real identity context and call the service layer normally.
func sessionIdentity(ctx context.Context, database *sql.DB, id domain.WorkoutSessionID) (*domain.Identity, error) {
	var userID domain.UserID
	var orgID domain.OrgID
	err := database.QueryRowContext(ctx,
		`SELECT user_id, org_id FROM cove.workout_sessions WHERE id = $1`, id,
	).Scan(&userID, &orgID)
	if err != nil {
		return nil, fmt.Errorf("query: %w", err)
	}
	return &domain.Identity{
		UserID:  userID,
		OrgID:   orgID,
		TokenID: uuid.Nil,
	}, nil
}

// optDuration reads an env var as a duration (e.g. "10m", "90s").
// Returns zero if unset, which lets the client fall back to its default.
func optDuration(key string) time.Duration {
	v := os.Getenv(key)
	if v == "" {
		return 0
	}
	d, err := time.ParseDuration(v)
	if err != nil {
		log.Fatalf("invalid %s %q: %v", key, v, err)
	}
	return d
}

func mustEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("%s is required", key)
	}
	return v
}

func mustEncryptor() crypto.Encryptor {
	key := mustEnv("SESSION_ENCRYPTION_KEY")
	enc, err := crypto.NewAESEncryptor(0, map[byte]string{0: key})
	if err != nil {
		log.Fatalf("init encryptor: %v", err)
	}
	return enc
}
