// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package db

import (
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
