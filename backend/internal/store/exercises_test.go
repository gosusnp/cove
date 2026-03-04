// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"errors"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestExerciseStore(t *testing.T) *ExerciseStore {
	t.Helper()
	return NewExerciseStore(testutil.NewDB(t))
}

func TestExerciseStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		s := newTestExerciseStore(t)

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
		s := newTestExerciseStore(t)
		if _, err := s.Create("Push-up", nil); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create("Air Squat", nil); err != nil {
			t.Fatal(err)
		}

		exercises, err := s.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(exercises) != 2 {
			t.Fatalf("expected 2 exercises, got %d", len(exercises))
		}
		if exercises[0].Name != "Air Squat" || exercises[1].Name != "Push-up" {
			t.Errorf("unexpected order: %q, %q", exercises[0].Name, exercises[1].Name)
		}
	})
}

func TestExerciseStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s := newTestExerciseStore(t)
		prog := "Add 1 rep each session"
		created, err := s.Create("Push-up", &prog)
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.Get(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Push-up" {
			t.Errorf("got name %q, want %q", got.Name, "Push-up")
		}
		if got.Progression == nil || *got.Progression != prog {
			t.Errorf("got progression %v, want %q", got.Progression, prog)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := newTestExerciseStore(t)

		_, err := s.Get(domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseStore_Create(t *testing.T) {
	t.Run("creates with progression", func(t *testing.T) {
		s := newTestExerciseStore(t)
		prog := "Bodyweight"

		e, err := s.Create("Push-up", &prog)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Name != "Push-up" {
			t.Errorf("got name %q, want %q", e.Name, "Push-up")
		}
		if e.Progression == nil || *e.Progression != prog {
			t.Errorf("got progression %v, want %q", e.Progression, prog)
		}
		if e.ID == domain.ExerciseID(0) {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("creates without progression", func(t *testing.T) {
		s := newTestExerciseStore(t)

		e, err := s.Create("Push-up", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Progression != nil {
			t.Errorf("expected nil progression, got %v", e.Progression)
		}
	})

	t.Run("duplicate name returns ErrDuplicate", func(t *testing.T) {
		s := newTestExerciseStore(t)
		if _, err := s.Create("Push-up", nil); err != nil {
			t.Fatal(err)
		}

		_, err := s.Create("Push-up", nil)
		if !errors.Is(err, ErrDuplicate) {
			t.Errorf("got %v, want ErrDuplicate", err)
		}
	})
}

func TestExerciseStore_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		s := newTestExerciseStore(t)
		created, err := s.Create("Push-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		newProg := "Weight vest"
		updated, err := s.Update(created.ID, "Hard Push-up", &newProg)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Hard Push-up" {
			t.Errorf("got name %q, want %q", updated.Name, "Hard Push-up")
		}
		if updated.Progression == nil || *updated.Progression != newProg {
			t.Errorf("got progression %v, want %q", updated.Progression, newProg)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := newTestExerciseStore(t)

		_, err := s.Update(domain.ExerciseID(999), "Push-up", nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s := newTestExerciseStore(t)
		created, err := s.Create("Push-up", nil)
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
		s := newTestExerciseStore(t)

		err := s.Delete(domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
