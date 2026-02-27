// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/testdb"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testdb.New(t, containerDSN, db.MigrationsFS)
}

func newTestStore(t *testing.T) *ExerciseStore {
	t.Helper()
	return NewExerciseStore(newTestDB(t))
}

func TestExerciseStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		s := newTestStore(t)

		exercises, err := s.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exercises == nil {
			t.Error("expected empty slice, got nil")
		}
		if len(exercises) != 0 {
			t.Errorf("expected 0 exercises, got %d", len(exercises))
		}
	})

	t.Run("returns all exercises ordered by name", func(t *testing.T) {
		s := newTestStore(t)
		if _, err := s.Create("Push-up", nil); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create("Pull-up", nil); err != nil {
			t.Fatal(err)
		}

		exercises, err := s.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(exercises) != 2 {
			t.Fatalf("expected 2 exercises, got %d", len(exercises))
		}
		if exercises[0].Name != "Pull-up" || exercises[1].Name != "Push-up" {
			t.Errorf("unexpected order: %q, %q", exercises[0].Name, exercises[1].Name)
		}
	})
}

func TestExerciseStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s := newTestStore(t)
		created, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.Get(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Pull-up" {
			t.Errorf("got name %q, want %q", got.Name, "Pull-up")
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := newTestStore(t)

		_, err := s.Get(999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseStore_Create(t *testing.T) {
	t.Run("without progression", func(t *testing.T) {
		s := newTestStore(t)

		e, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Name != "Pull-up" {
			t.Errorf("got name %q, want %q", e.Name, "Pull-up")
		}
		if e.Progression != nil {
			t.Errorf("expected nil progression, got %v", e.Progression)
		}
	})

	t.Run("with progression", func(t *testing.T) {
		s := newTestStore(t)
		prog := "weighted"

		e, err := s.Create("Pull-up", &prog)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Progression == nil || *e.Progression != "weighted" {
			t.Errorf("got progression %v, want %q", e.Progression, "weighted")
		}
	})

	t.Run("duplicate name returns error", func(t *testing.T) {
		s := newTestStore(t)
		if _, err := s.Create("Pull-up", nil); err != nil {
			t.Fatal(err)
		}

		_, err := s.Create("Pull-up", nil)
		if err == nil {
			t.Error("expected error for duplicate name, got nil")
		}
	})
}

func TestExerciseStore_Update(t *testing.T) {
	t.Run("updates name and progression", func(t *testing.T) {
		s := newTestStore(t)
		created, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}
		prog := "weighted"

		updated, err := s.Update(created.ID, "Weighted Pull-up", &prog)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Weighted Pull-up" {
			t.Errorf("got name %q, want %q", updated.Name, "Weighted Pull-up")
		}
		if updated.Progression == nil || *updated.Progression != "weighted" {
			t.Errorf("got progression %v, want %q", updated.Progression, "weighted")
		}
	})

	t.Run("clears progression when nil", func(t *testing.T) {
		s := newTestStore(t)
		prog := "weighted"
		created, err := s.Create("Pull-up", &prog)
		if err != nil {
			t.Fatal(err)
		}

		updated, err := s.Update(created.ID, "Pull-up", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Progression != nil {
			t.Errorf("expected nil progression, got %v", updated.Progression)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := newTestStore(t)

		_, err := s.Update(999, "Pull-up", nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s := newTestStore(t)
		created, err := s.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := s.Delete(created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = s.Get(created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := newTestStore(t)

		err := s.Delete(999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
