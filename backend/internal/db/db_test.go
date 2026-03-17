// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package db

import (
	"database/sql"
	"net/url"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

var containerDSN string

func TestMain(m *testing.M) {
	testutil.RunMain(m, &containerDSN, MigrationsFS)
}

func TestRoleGrants(t *testing.T) {
	// Get a migrated per-test database (pgtestdb clone with all migrations applied).
	testDB := testutil.NewDB(t)

	var dbName string
	if err := testDB.QueryRow(`SELECT current_database()`).Scan(&dbName); err != nil {
		t.Fatalf("get database name: %v", err)
	}

	// Roles are cluster-wide — create them via the superuser (containerDSN) connection.
	superDB, err := sql.Open("pgx", containerDSN)
	if err != nil {
		t.Fatalf("open superuser db: %v", err)
	}
	defer superDB.Close()

	if _, err := superDB.Exec(`CREATE ROLE cove_app_role`); err != nil {
		t.Fatalf("create cove_app_role: %v", err)
	}
	t.Cleanup(func() {
		_, _ = superDB.Exec(`DROP ROLE IF EXISTS cove_app_test`)
		_, _ = superDB.Exec(`DROP ROLE IF EXISTS cove_app_role`)
	})
	if _, err := superDB.Exec(`CREATE ROLE cove_app_test LOGIN PASSWORD 'test'`); err != nil {
		t.Fatalf("create cove_app_test: %v", err)
	}
	if _, err := superDB.Exec(`GRANT cove_app_role TO cove_app_test`); err != nil {
		t.Fatalf("grant role: %v", err)
	}

	// Apply grants via testDB — the pgtestdb per-test user owns the schema and
	// tables, so it has the GRANT OPTION needed to issue the grants.
	if err := applyRoleGrants(testDB); err != nil {
		t.Fatalf("applyRoleGrants: %v", err)
	}

	// Connect as the restricted user to the same per-test database and verify
	// it can query the schema.
	u, _ := url.Parse(containerDSN)
	u.User = url.UserPassword("cove_app_test", "test")
	u.Path = "/" + dbName
	appDB, err := sql.Open("pgx", u.String())
	if err != nil {
		t.Fatalf("open restricted db: %v", err)
	}
	defer appDB.Close()

	if _, err := appDB.Exec(`SELECT 1 FROM cove.users LIMIT 0`); err != nil {
		t.Fatalf("restricted user cannot query cove.users: %v", err)
	}
}

func TestMigrations_Roundtrip(t *testing.T) {
	testDB := testutil.NewEmptyDB(t)

	if _, err := testDB.Exec("CREATE SCHEMA IF NOT EXISTS " + Schema); err != nil {
		t.Fatalf("create schema: %v", err)
	}

	driver, err := migratepostgres.WithInstance(testDB, &migratepostgres.Config{SchemaName: Schema})
	if err != nil {
		t.Fatal(err)
	}

	source, err := iofs.New(MigrationsFS, "migrations")
	if err != nil {
		t.Fatal(err)
	}

	m, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })

	if err := m.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}

	var count int
	if err := testDB.QueryRow(
		`SELECT count(*) FROM information_schema.tables WHERE table_schema = $1 AND table_name != 'schema_migrations'`,
		Schema,
	).Scan(&count); err != nil {
		t.Fatalf("query after Up(): %v", err)
	}
	if count == 0 {
		t.Fatal("expected tables after Up()")
	}

	if err := m.Down(); err != nil {
		t.Fatalf("down: %v", err)
	}

	if err := testDB.QueryRow(
		`SELECT count(*) FROM information_schema.tables WHERE table_schema = $1 AND table_name != 'schema_migrations'`,
		Schema,
	).Scan(&count); err != nil {
		t.Fatalf("query after Down(): %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no tables after Down(), got %d", count)
	}
}
