package service

import (
	"errors"
	"testing"

	"github.com/gosusnp/cove/api/store"
)

func newTestExerciseService(t *testing.T) *ExerciseService {
	t.Helper()
	return NewExerciseService(store.NewExerciseStore(newTestDB(t)))
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

	t.Run("valid name creates exercise", func(t *testing.T) {
		svc := newTestExerciseService(t)

		e, err := svc.Create("Pull-up", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Name != "Pull-up" {
			t.Errorf("got name %q, want %q", e.Name, "Pull-up")
		}
	})

	t.Run("duplicate name returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)

		if _, err := svc.Create("Pull-up", nil); err != nil {
			t.Fatal(err)
		}
		_, err := svc.Create("Pull-up", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("case-insensitive duplicate returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)

		if _, err := svc.Create("Pull-up", nil); err != nil {
			t.Fatal(err)
		}
		_, err := svc.Create("pull-up", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("whitespace is normalized", func(t *testing.T) {
		svc := newTestExerciseService(t)

		e, err := svc.Create("  Pull  up  ", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Name != "Pull up" {
			t.Errorf("got name %q, want %q", e.Name, "Pull up")
		}
	})

	t.Run("name is stored as entered", func(t *testing.T) {
		svc := newTestExerciseService(t)

		e, err := svc.Create("Romanian Deadlift", nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Name != "Romanian Deadlift" {
			t.Errorf("got name %q, want %q", e.Name, "Romanian Deadlift")
		}
	})
}

func TestExerciseService_Update(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)
		e, err := svc.Create("Pull-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = svc.Update(e.ID, "", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("rename to case-insensitive duplicate returns ValidationError", func(t *testing.T) {
		svc := newTestExerciseService(t)
		if _, err := svc.Create("Pull-up", nil); err != nil {
			t.Fatal(err)
		}
		e, err := svc.Create("Chin-up", nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = svc.Update(e.ID, "pull-up", nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})
}
