// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package db

import (
	"database/sql"
	"embed"
	"fmt"
	"log"
	"net/url"
	"regexp"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DefaultSchema is the default PostgreSQL schema used by Cove.
const DefaultSchema = "cove"

var validIdentifier = regexp.MustCompile(`^[a-z_][a-z0-9_]{0,62}$`)

// validateSchema rejects schema names that are not safe unquoted SQL identifiers.
func validateSchema(schema string) {
	if !validIdentifier.MatchString(schema) {
		log.Fatalf("invalid schema name %q: must match [a-z_][a-z0-9_]{0,62}", schema)
	}
}

//go:embed migrations/*.sql
var MigrationsFS embed.FS

// withSearchPath appends options=-c search_path=<schema> to the DSN query string,
// preserving any existing options parameter.
func withSearchPath(dsn, schema string) string {
	u, err := url.Parse(dsn)
	if err != nil {
		return dsn
	}
	q := u.Query()
	if existing := q.Get("options"); existing != "" {
		q.Set("options", existing+" -c search_path="+schema)
	} else {
		q.Set("options", "-c search_path="+schema)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

// applyRoleGrants grants the cove_app_role runtime permissions on the given schema,
// if the role exists. Each statement is run as a separate Exec call.
//
// This runs from Go rather than a SQL migration because the grants reference the
// schema name, which is runtime config (COVE_DB_SCHEMA). Static migration files
// cannot reference a variable schema name.
func applyRoleGrants(database *sql.DB, schema string) error {
	var exists bool
	err := database.QueryRow(`SELECT EXISTS (SELECT FROM pg_roles WHERE rolname = 'cove_app_role')`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check cove_app_role: %w", err)
	}
	if !exists {
		return nil
	}

	stmts := []string{
		fmt.Sprintf(`GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA %s TO cove_app_role`, schema),
		fmt.Sprintf(`GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA %s TO cove_app_role`, schema),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO cove_app_role`, schema),
		fmt.Sprintf(`ALTER DEFAULT PRIVILEGES IN SCHEMA %s GRANT USAGE, SELECT ON SEQUENCES TO cove_app_role`, schema),
	}
	for _, stmt := range stmts {
		if _, err := database.Exec(stmt); err != nil {
			return fmt.Errorf("apply role grant: %w", err)
		}
	}
	return nil
}

// Migrate runs all pending migrations using the provided DSN and schema.
// It creates the schema if it does not exist, runs golang-migrate, then applies
// role grants. The dedicated connection is closed after migrations complete.
func Migrate(dsn, schema string) {
	validateSchema(schema)
	dsnWithSchema := withSearchPath(dsn, schema)
	database, err := sql.Open("pgx", dsnWithSchema)
	if err != nil {
		log.Fatalf("failed to open migration database: %v", err)
	}
	defer database.Close()

	if err := database.Ping(); err != nil {
		log.Fatalf("failed to connect to migration database: %v", err)
	}

	if _, err := database.Exec(fmt.Sprintf(`CREATE SCHEMA IF NOT EXISTS %s`, schema)); err != nil {
		log.Fatalf("failed to create schema %s: %v", schema, err)
	}

	driver, err := migratepostgres.WithInstance(database, &migratepostgres.Config{SchemaName: schema})
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

	if err := applyRoleGrants(database, schema); err != nil {
		log.Fatalf("failed to apply role grants: %v", err)
	}

	log.Println("migrations up to date")
}

// Open opens and returns the application database connection pool with search_path
// set to the given schema. Migrations must be run separately via Migrate before calling Open.
func Open(dsn, schema string) *sql.DB {
	validateSchema(schema)
	database, err := sql.Open("pgx", withSearchPath(dsn, schema))
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}

	if err := database.Ping(); err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return database
}
