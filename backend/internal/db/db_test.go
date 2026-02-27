// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package db

import (
	"context"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/gosusnp/cove/backend/internal/testdb"
)

var containerDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()

	dsn, cleanup, err := testdb.StartContainer(ctx)
	if err != nil {
		panic(err)
	}
	containerDSN = dsn

	code := m.Run()
	cleanup()
	os.Exit(code)
}

func TestMigrations_Roundtrip(t *testing.T) {
	db := testdb.NewEmpty(t, containerDSN)

	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
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
	if err := db.QueryRow(
		`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name != 'schema_migrations'`,
	).Scan(&count); err != nil {
		t.Fatalf("query after Up(): %v", err)
	}
	if count == 0 {
		t.Fatal("expected tables after Up()")
	}

	if err := m.Down(); err != nil {
		t.Fatalf("down: %v", err)
	}

	if err := db.QueryRow(
		`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name != 'schema_migrations'`,
	).Scan(&count); err != nil {
		t.Fatalf("query after Down(): %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no tables after Down(), got %d", count)
	}
}
