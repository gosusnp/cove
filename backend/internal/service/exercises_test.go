// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"errors"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestExerciseService(t *testing.T) *ExerciseService {
	t.Helper()
	return NewExerciseService(store.NewExerciseStore(testutil.NewDB(t)))
}

func TestExerciseService_List(t *testing.T) {
	t.Run("returns all exercises", func(t *testing.T) {
		svc := newTestExerciseService(t)
		_, _ = svc.Create("Push-up", nil)
		_, _ = svc.Create("Pull-up", nil)

		list, err := svc.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 exercises, got %d", len(list))
		}
	})
}

func TestExerciseService_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc := newTestExerciseService(t)
		created, _ := svc.Create("Push-up", nil)

		got, err := svc.Get(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Push-up" {
			t.Errorf("got %q, want Push-up", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := newTestExerciseService(t)
		_, err := svc.Get(domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseService_Create(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)

		_, err := svc.Create("", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "name is required" {
			t.Errorf("got msg %q, want %q", ve.Msg, "name is required")
		}
	})

	t.Run("trims name", func(t *testing.T) {
		svc := newTestExerciseService(t)

		e, err := svc.Create("  Push-up  ", nil)
		if err != nil {
			t.Fatal(err)
		}
		if e.Name != "Push-up" {
			t.Errorf("got %q, want %q", e.Name, "Push-up")
		}
	})

	t.Run("duplicate name returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)
		if _, err := svc.Create("Push-up", nil); err != nil {
			t.Fatal(err)
		}

		_, err := svc.Create("Push-up", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "exercise with this name already exists" {
			t.Errorf("unexpected msg: %q", ve.Msg)
		}
	})
}

func TestExerciseService_Update(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)
		e, _ := svc.Create("Push-up", nil)

		_, err := svc.Update(e.ID, "", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("duplicate name returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)
		e1, _ := svc.Create("Push-up", nil)
		e2, _ := svc.Create("Air Squat", nil)

		_, err := svc.Update(e2.ID, e1.Name, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc := newTestExerciseService(t)

		_, err := svc.Update(domain.ExerciseID(999), "Valid Name", nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := newTestExerciseService(t)
		created, _ := svc.Create("Push-up", nil)

		if err := svc.Delete(created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.Get(created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc := newTestExerciseService(t)

		err := svc.Delete(domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
