// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Config holds all application configuration resolved from the environment.
type Config struct {
	// Database
	DatabaseURL          string
	MigrationDatabaseURL string

	// Google OAuth
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string
	AllowedEmails      []string

	// Session encryption
	EncryptionKey string

	// LLM
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string

	// USDA FoodData Central
	FDCAPIKey string

	// Hatchet worker — empty token disables the worker entirely
	HatchetToken  string
	WorkerEnabled bool

	// Server
	Port string
	Dev  bool
}

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

// loadConfig reads and validates all application configuration from the
// environment. Required fields that are missing or empty return an error.
func loadConfig() (Config, error) {
	dev := os.Getenv("COVE_DEV") != ""

	dbURL := getSecret("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	migrationURL, ok := resolveMigrationDSN(dbURL, dev)
	if !ok {
		return Config{}, fmt.Errorf("MIGRATION_DATABASE_URL is required in production")
	}

	googleClientID := getSecret("GOOGLE_CLIENT_ID")
	if googleClientID == "" {
		return Config{}, fmt.Errorf("GOOGLE_CLIENT_ID is required")
	}
	googleClientSecret := getSecret("GOOGLE_CLIENT_SECRET")
	if googleClientSecret == "" {
		return Config{}, fmt.Errorf("GOOGLE_CLIENT_SECRET is required")
	}
	googleRedirectURL := getSecret("GOOGLE_REDIRECT_URL")
	if googleRedirectURL == "" {
		return Config{}, fmt.Errorf("GOOGLE_REDIRECT_URL is required")
	}

	encryptionKey := getSecret("SESSION_ENCRYPTION_KEY")
	if encryptionKey == "" {
		return Config{}, fmt.Errorf("SESSION_ENCRYPTION_KEY is required")
	}

	var allowedEmails []string
	if raw := getSecret("COVE_ALLOWED_EMAILS"); raw != "" {
		allowedEmails = strings.Split(raw, ",")
	}

	port := getSecret("COVE_PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		DatabaseURL:          dbURL,
		MigrationDatabaseURL: migrationURL,
		GoogleClientID:       googleClientID,
		GoogleClientSecret:   googleClientSecret,
		GoogleRedirectURL:    googleRedirectURL,
		AllowedEmails:        allowedEmails,
		EncryptionKey:        encryptionKey,
		LLMBaseURL:           getSecret("LLM_BASE_URL"),
		LLMAPIKey:            getSecret("LLM_API_KEY"),
		LLMModel:             getSecret("LLM_MODEL"),
		FDCAPIKey:            getSecret("FDC_API_KEY"),
		HatchetToken:         getSecret("HATCHET_CLIENT_TOKEN"),
		WorkerEnabled:        getSecret("COVE_WORKER_ENABLED") == "true",
		Port:                 port,
		Dev:                  dev,
	}, nil
}
