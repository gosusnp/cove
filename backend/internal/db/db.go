// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Schema is the PostgreSQL schema used by Cove.
const Schema = "cove"

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// applyRoleGrants grants the cove_app_role runtime permissions on the cove schema,
// if the role exists. Each statement is run as a separate Exec call.
func applyRoleGrants(database *sql.DB) error {
	var exists bool
	err := database.QueryRow(`SELECT EXISTS (SELECT FROM pg_roles WHERE rolname = 'cove_app_role')`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check cove_app_role: %w", err)
	}
	if !exists {
		return nil
	}

	stmts := []string{
		`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA cove TO cove_app_role`,
		`GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA cove TO cove_app_role`,
		`ALTER DEFAULT PRIVILEGES IN SCHEMA cove GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO cove_app_role`,
		`ALTER DEFAULT PRIVILEGES IN SCHEMA cove GRANT USAGE, SELECT ON SEQUENCES TO cove_app_role`,
	}
	for _, stmt := range stmts {
		if _, err := database.Exec(stmt); err != nil {
			return fmt.Errorf("apply role grant: %w", err)
		}
	}
	return nil
}

// Migrate runs all pending migrations. It creates the cove schema if it does
// not exist, runs golang-migrate, then applies role grants. The dedicated
// connection is closed after migrations complete.
func Migrate(dsn string) {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open migration database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("failed to connect to migration database: %v", err)
	}

	if _, err := database.Exec(`CREATE SCHEMA IF NOT EXISTS cove`); err != nil {
		log.Fatalf("failed to create schema cove: %v", err)
	}

	driver, err := migratepostgres.WithInstance(database, &migratepostgres.Config{SchemaName: Schema})
	if err != nil {
		log.Fatalf("failed to create migration driver: %v", err)
	}

	source, err := iofs.New(MigrationsFS, "migrations")
	if err != nil {
		log.Fatalf("failed to open migrations: %v", err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		log.Fatalf("failed to create migrator: %v", err)
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("failed to run migrations: %v", err)
	}

	if err := applyRoleGrants(database); err != nil {
		log.Fatalf("failed to apply role grants: %v", err)
	}

	log.Println("migrations up to date")
}

// Open opens and returns the application database connection pool.
// Migrations must be run separately via Migrate before calling Open.
func Open(dsn string) *sql.DB {
	database, err := sql.Open("pgx", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := database.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return database
}
