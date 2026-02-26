package service

import (
	"database/sql"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
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

	source, err := (&file.File{}).Open("file://../db/migrations")
	if err != nil {
		t.Fatal(err)
	}

	m, err := migrate.NewWithInstance("file", source, "sqlite", driver)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { m.Close() })

	if err := m.Up(); err != nil {
		t.Fatal(err)
	}

	return db
}
