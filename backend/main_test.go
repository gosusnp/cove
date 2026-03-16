// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestGetSecret(t *testing.T) {
	t.Run("returns env var when nothing else is set", func(t *testing.T) {
		t.Setenv("MY_KEY", "from-env")
		if got := getSecret("MY_KEY"); got != "from-env" {
			t.Errorf("got %q, want %q", got, "from-env")
		}
	})

	t.Run("prefers _FILE over env var", func(t *testing.T) {
		f := writeFile(t, "file-value")
		t.Setenv("MY_KEY", "from-env")
		t.Setenv("MY_KEY_FILE", f)
		if got := getSecret("MY_KEY"); got != "file-value" {
			t.Errorf("got %q, want %q", got, "file-value")
		}
	})

	t.Run("prefers COVE_SECRETS_DIR over _FILE and env var", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "MY_KEY"), []byte("from-dir"), 0600); err != nil {
			t.Fatal(err)
		}
		t.Setenv("COVE_SECRETS_DIR", dir)
		t.Setenv("MY_KEY_FILE", writeFile(t, "from-file"))
		t.Setenv("MY_KEY", "from-env")
		if got := getSecret("MY_KEY"); got != "from-dir" {
			t.Errorf("got %q, want %q", got, "from-dir")
		}
	})

	t.Run("falls through COVE_SECRETS_DIR to _FILE when key file absent", func(t *testing.T) {
		t.Setenv("COVE_SECRETS_DIR", t.TempDir()) // dir exists but key file does not
		t.Setenv("MY_KEY_FILE", writeFile(t, "from-file"))
		if got := getSecret("MY_KEY"); got != "from-file" {
			t.Errorf("got %q, want %q", got, "from-file")
		}
	})

	t.Run("trims whitespace from file contents", func(t *testing.T) {
		f := writeFile(t, "  trimmed\n")
		t.Setenv("MY_KEY_FILE", f)
		if got := getSecret("MY_KEY"); got != "trimmed" {
			t.Errorf("got %q, want %q", got, "trimmed")
		}
	})

	t.Run("returns empty string when nothing is set", func(t *testing.T) {
		if got := getSecret("MY_KEY"); got != "" {
			t.Errorf("got %q, want empty string", got)
		}
	})
}

func TestResolveMigrationDSN(t *testing.T) {
	t.Run("uses MIGRATION_DATABASE_URL when set", func(t *testing.T) {
		t.Setenv("MIGRATION_DATABASE_URL", "migration-dsn")
		dsn, ok := resolveMigrationDSN("app-dsn", false)
		if !ok || dsn != "migration-dsn" {
			t.Errorf("got (%q, %v), want (\"migration-dsn\", true)", dsn, ok)
		}
	})

	t.Run("falls back to appDSN in dev when MIGRATION_DATABASE_URL unset", func(t *testing.T) {
		dsn, ok := resolveMigrationDSN("app-dsn", true)
		if !ok || dsn != "app-dsn" {
			t.Errorf("got (%q, %v), want (\"app-dsn\", true)", dsn, ok)
		}
	})

	t.Run("returns false in production when MIGRATION_DATABASE_URL unset", func(t *testing.T) {
		_, ok := resolveMigrationDSN("app-dsn", false)
		if ok {
			t.Error("expected ok=false in production without MIGRATION_DATABASE_URL")
		}
	})

	t.Run("MIGRATION_DATABASE_URL via _FILE takes precedence in production", func(t *testing.T) {
		t.Setenv("MIGRATION_DATABASE_URL_FILE", writeFile(t, "migration-from-file"))
		dsn, ok := resolveMigrationDSN("app-dsn", false)
		if !ok || dsn != "migration-from-file" {
			t.Errorf("got (%q, %v), want (\"migration-from-file\", true)", dsn, ok)
		}
	})
}

func writeFile(t *testing.T, content string) string {
	t.Helper()
	f := filepath.Join(t.TempDir(), "secret")
	if err := os.WriteFile(f, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}
	return f
}
