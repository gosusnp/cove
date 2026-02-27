// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package db

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func newTestMigrator(t *testing.T) (*sql.DB, *migrate.Migrate) {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	driver, err := sqlite.WithInstance(db, &sqlite.Config{})
	if err != nil {
		t.Fatal(err)
	}

	source, err := (&file.File{}).Open("file://migrations")
	if err != nil {
		t.Fatal(err)
	}

	m, err := migrate.NewWithInstance("file", source, "sqlite", driver)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })

	return db, m
}

func TestMigrations_Roundtrip(t *testing.T) {
	db, m := newTestMigrator(t)

	if err := m.Up(); err != nil {
		t.Fatalf("up: %v", err)
	}

	var count int
	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name != 'schema_migrations'`).Scan(&count); err != nil {
		t.Fatalf("query after Up(): %v", err)
	}
	if count == 0 {
		t.Fatal("expected tables after Up()")
	}

	if err := m.Down(); err != nil {
		t.Fatalf("down: %v", err)
	}

	if err := db.QueryRow(`SELECT count(*) FROM sqlite_master WHERE type='table' AND name != 'schema_migrations'`).Scan(&count); err != nil {
		t.Fatalf("query after Down(): %v", err)
	}
	if count != 0 {
		t.Fatalf("expected no tables after Down(), got %d", count)
	}
}
