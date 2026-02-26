package service

import (
	"errors"
	"testing"

	"github.com/gosusnp/cove/api/store"
)

func newTestProgramService(t *testing.T) *ProgramService {
	t.Helper()
	return NewProgramService(newTestDB(t))
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
		db := newTestDB(t)
		svc := NewProgramService(db)
		e, err := store.NewExerciseStore(db).Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}
		reps := 8

		detail, err := svc.CreateFull("Strength", []ProgramSetInput{
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
}
