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

type programExerciseFixture struct {
	svc        *ProgramExerciseService
	setID      int64
	exerciseID domain.ExerciseID
}

func newTestProgramExerciseService(t *testing.T) programExerciseFixture {
	t.Helper()
	db := testutil.NewDB(t)
	p, err := store.NewProgramStore(db).Create("Test Program")
	if err != nil {
		t.Fatal(err)
	}
	ps, err := store.NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	e, err := store.NewExerciseStore(db).Create("Pull-up", nil)
	if err != nil {
		t.Fatal(err)
	}
	return programExerciseFixture{
		svc:        NewProgramExerciseService(store.NewProgramExerciseStore(db)),
		setID:      ps.ID,
		exerciseID: e.ID,
	}
}

func TestProgramExerciseService_List(t *testing.T) {
	t.Run("returns all exercises for set", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		_, _ = f.svc.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		list, err := f.svc.List(f.setID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 exercise, got %d", len(list))
		}
	})
}

func TestProgramExerciseService_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		created, _ := f.svc.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		got, err := f.svc.Get(f.setID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ExerciseID != f.exerciseID {
			t.Errorf("got exercise_id %d, want %d", got.ExerciseID, f.exerciseID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		_, err := f.svc.Get(f.setID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramExerciseService_Create(t *testing.T) {
	t.Run("zero exercise_id returns ValidationError", func(t *testing.T) {
		f := newTestProgramExerciseService(t)

		_, err := f.svc.Create(f.setID, 0, nil, nil, nil, nil, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "exercise_id is required" {
			t.Errorf("got msg %q, want %q", ve.Msg, "exercise_id is required")
		}
	})

	t.Run("valid exercise_id creates entry", func(t *testing.T) {
		f := newTestProgramExerciseService(t)

		pe, err := f.svc.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pe.ExerciseID != f.exerciseID {
			t.Errorf("got exercise_id %d, want %d", pe.ExerciseID, f.exerciseID)
		}
	})
}

func TestProgramExerciseService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		created, _ := f.svc.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		lat := "left"
		updated, err := f.svc.Update(f.setID, created.ID, f.exerciseID, &lat, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Laterality == nil || *updated.Laterality != "left" {
			t.Errorf("got %v, want left", updated.Laterality)
		}
	})

	t.Run("zero exercise_id returns ValidationError", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		created, err := f.svc.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = f.svc.Update(f.setID, created.ID, 0, nil, nil, nil, nil, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		_, err := f.svc.Update(f.setID, 999, f.exerciseID, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramExerciseService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		created, _ := f.svc.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		if err := f.svc.Delete(f.setID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := f.svc.Get(f.setID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		err := f.svc.Delete(f.setID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
