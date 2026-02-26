package service

import (
	"testing"

	"github.com/gosusnp/cove/api/store"
)

func newTestProgramSetService(t *testing.T) (*ProgramSetService, int64) {
	t.Helper()
	db := newTestDB(t)
	p, err := store.NewProgramStore(db).Create("Test Program")
	if err != nil {
		t.Fatal(err)
	}
	return NewProgramSetService(store.NewProgramSetStore(db)), p.ID
}

func TestProgramSetService_Create(t *testing.T) {
	t.Run("rounds below 1 defaults to 1", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)

		ps, err := svc.Create(programID, nil, 0, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", ps.Rounds)
		}
	})

	t.Run("negative rounds defaults to 1", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)

		ps, err := svc.Create(programID, nil, -5, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", ps.Rounds)
		}
	})

	t.Run("valid rounds are preserved", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)

		ps, err := svc.Create(programID, nil, 4, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.Rounds != 4 {
			t.Errorf("got rounds %d, want 4", ps.Rounds)
		}
	})
}

func TestProgramSetService_Update(t *testing.T) {
	t.Run("rounds below 1 defaults to 1", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		created, err := svc.Create(programID, nil, 3, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		updated, err := svc.Update(programID, created.ID, nil, 0, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", updated.Rounds)
		}
	})
}
