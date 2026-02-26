package service

import (
	"errors"
	"testing"

	"github.com/gosusnp/cove/api/store"
)

type programExerciseFixture struct {
	svc        *ProgramExerciseService
	setID      int64
	exerciseID int64
}

func newTestProgramExerciseService(t *testing.T) programExerciseFixture {
	t.Helper()
	db := newTestDB(t)
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
}
