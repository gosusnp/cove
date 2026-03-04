// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package testutil

import (
	"context"
	"crypto/md5"
	"database/sql"
	"embed"
	"fmt"
	"io/fs"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	migratepostgres "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/peterldowns/pgtestdb"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	testDSN        string
	testMigrations embed.FS
)

// RunMain starts a PostgreSQL testcontainer, runs the tests, and shuts down the container.
// If TEST_DATABASE_URL is set, it skips container management and uses the provided DSN.
func RunMain(m *testing.M, dsnPtr *string, migrations embed.FS) {
	ctx := context.Background()
	testMigrations = migrations

	if dsn := os.Getenv("TEST_DATABASE_URL"); dsn != "" {
		*dsnPtr = dsn
		testDSN = dsn
		os.Exit(m.Run())
	}

	dsn, cleanup, err := StartContainer(ctx)
	if err != nil {
		panic(err)
	}
	*dsnPtr = dsn
	testDSN = dsn

	code := m.Run()
	cleanup()
	os.Exit(code)
}

// StartContainer starts a PostgreSQL testcontainer and returns the DSN and a cleanup function.
func StartContainer(ctx context.Context) (string, func(), error) {
	container, err := tcpostgres.Run(ctx,
		"postgres:18-alpine",
		tcpostgres.WithDatabase("testdb"),
		tcpostgres.WithUsername("test"),
		tcpostgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").WithOccurrence(2),
		),
	)
	if err != nil {
		return "", nil, fmt.Errorf("start postgres container: %w", err)
	}

	dsn, err := container.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		_ = container.Terminate(ctx)
		return "", nil, fmt.Errorf("get connection string: %w", err)
	}

	return dsn, func() { _ = container.Terminate(ctx) }, nil
}

// NewDB creates an isolated test database with the migrations provided to RunMain applied.
func NewDB(t *testing.T) *sql.DB {
	t.Helper()
	if testDSN == "" {
		t.Fatal("testutil.NewDB called before RunMain or with empty DSN")
	}
	return pgtestdb.New(t, parseConfig(testDSN), &fsMigrator{fs: testMigrations})
}

// NewEmptyDB creates an isolated empty test database with no migrations applied.
func NewEmptyDB(t *testing.T) *sql.DB {
	t.Helper()
	if testDSN == "" {
		t.Fatal("testutil.NewEmptyDB called before RunMain or with empty DSN")
	}
	return pgtestdb.New(t, parseConfig(testDSN), pgtestdb.NoopMigrator{})
}

type fsMigrator struct {
	fs embed.FS
}

func (m *fsMigrator) Hash() (string, error) {
	h := md5.New()
	err := fs.WalkDir(m.fs, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := m.fs.ReadFile(path)
		if err != nil {
			return err
		}
		_, _ = h.Write(data)
		return nil
	})
	return fmt.Sprintf("%x", h.Sum(nil)), err
}

func (m *fsMigrator) Migrate(_ context.Context, db *sql.DB, _ pgtestdb.Config) error {
	driver, err := migratepostgres.WithInstance(db, &migratepostgres.Config{})
	if err != nil {
		return err
	}

	source, err := iofs.New(m.fs, "migrations")
	if err != nil {
		return err
	}

	mg, err := migrate.NewWithInstance("iofs", source, "postgres", driver)
	if err != nil {
		return err
	}

	if err := mg.Up(); err != nil && err != migrate.ErrNoChange {
		return err
	}

	// Release the advisory lock so pgtestdb can clone the template database.
	mg.Close()
	return nil
}

func (m *fsMigrator) Verify(_ context.Context, db *sql.DB, _ pgtestdb.Config) error {
	var count int
	return db.QueryRow(
		`SELECT count(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name != 'schema_migrations'`,
	).Scan(&count)
}

func parseConfig(dsn string) pgtestdb.Config {
	u, _ := url.Parse(dsn)
	password, _ := u.User.Password()
	return pgtestdb.Config{
		DriverName: "pgx",
		Host:       u.Hostname(),
		Port:       u.Port(),
		User:       u.User.Username(),
		Password:   password,
		Database:   strings.TrimPrefix(u.Path, "/"),
		Options:    "sslmode=disable",
	}
}
