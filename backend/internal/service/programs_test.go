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

func newTestProgramService(t *testing.T) *ProgramService {
	t.Helper()
	return NewProgramService(testutil.NewDB(t))
}

func TestProgramService_List(t *testing.T) {
	t.Run("returns all programs", func(t *testing.T) {
		svc := newTestProgramService(t)
		_, _ = svc.Create("Strength")
		_, _ = svc.Create("Hypertrophy")

		list, err := svc.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 programs, got %d", len(list))
		}
	})
}

func TestProgramService_GetDetail(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc := newTestProgramService(t)
		created, _ := svc.Create("Strength")

		got, err := svc.GetDetail(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got %q, want Strength", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := newTestProgramService(t)
		_, err := svc.GetDetail(domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Create(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestProgramService(t)

		_, err := svc.Create("")
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "name is required" {
			t.Errorf("got msg %q, want %q", ve.Msg, "name is required")
		}
	})

	t.Run("valid name creates program", func(t *testing.T) {
		svc := newTestProgramService(t)

		p, err := svc.Create("Strength")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "Strength" {
			t.Errorf("got name %q, want %q", p.Name, "Strength")
		}
	})
}

func TestProgramService_CreateFull(t *testing.T) {
	t.Run("creates full hierarchy atomically", func(t *testing.T) {
		db := testutil.NewDB(t)
		svc := NewProgramService(db)
		e, err := store.NewExerciseStore(db).Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}
		reps := 8

		program, err := svc.CreateFull("Strength", []ProgramSetInput{
			{
				Rounds: 3,
				Exercises: []ProgramExerciseInput{
					{ExerciseID: e.ID, TargetReps: &reps},
				},
			},
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if program.Name != "Strength" {
			t.Errorf("got name %q, want %q", program.Name, "Strength")
		}

		detail, err := svc.GetDetail(program.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if detail.Name != "Strength" {
			t.Errorf("got name %q, want %q", detail.Name, "Strength")
		}
		if len(detail.Sets) != 1 {
			t.Fatalf("expected 1 set, got %d", len(detail.Sets))
		}
		if detail.Sets[0].Rounds != 3 {
			t.Errorf("got rounds %d, want 3", detail.Sets[0].Rounds)
		}
		if len(detail.Sets[0].Exercises) != 1 {
			t.Fatalf("expected 1 exercise, got %d", len(detail.Sets[0].Exercises))
		}
		if detail.Sets[0].Exercises[0].TargetReps == nil || *detail.Sets[0].Exercises[0].TargetReps != 8 {
			t.Errorf("got target_reps %v, want 8", detail.Sets[0].Exercises[0].TargetReps)
		}
	})

	t.Run("unknown exercise_id returns ValidationError and nothing is written", func(t *testing.T) {
		svc := newTestProgramService(t)

		_, err := svc.CreateFull("Strength", []ProgramSetInput{
			{Rounds: 3, Exercises: []ProgramExerciseInput{{ExerciseID: 999}}},
		})
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "exercise_id 999 not found" {
			t.Errorf("got msg %q, want %q", ve.Msg, "exercise_id 999 not found")
		}

		programs, err := svc.List()
		if err != nil {
			t.Fatal(err)
		}
		if len(programs) != 0 {
			t.Errorf("expected no programs written, got %d", len(programs))
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestProgramService(t)

		_, err := svc.CreateFull("", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})
}

func TestProgramService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := newTestProgramService(t)
		p, _ := svc.Create("Strength")

		updated, err := svc.Update(p.ID, "New Name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("got %q, want New Name", updated.Name)
		}
	})

	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestProgramService(t)
		p, err := svc.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}

		_, err = svc.Update(p.ID, "")
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc := newTestProgramService(t)
		_, err := svc.Update(domain.ProgramID(999), "Valid Name")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := newTestProgramService(t)
		p, _ := svc.Create("Strength")

		if err := svc.Delete(p.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.GetDetail(p.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc := newTestProgramService(t)
		err := svc.Delete(domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
