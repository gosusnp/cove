package store

import (
	"errors"
	"testing"
)

func newTestProgramStore(t *testing.T) *ProgramStore {
	t.Helper()
	return NewProgramStore(newTestDB(t))
}

func TestProgramStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		s := newTestProgramStore(t)

		programs, err := s.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if programs == nil {
			t.Error("expected empty slice, got nil")
		}
		if len(programs) != 0 {
			t.Errorf("expected 0 programs, got %d", len(programs))
		}
	})

	t.Run("returns all programs ordered by name", func(t *testing.T) {
		s := newTestProgramStore(t)
		if _, err := s.Create("Strength"); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create("Hypertrophy"); err != nil {
			t.Fatal(err)
		}

		programs, err := s.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(programs) != 2 {
			t.Fatalf("expected 2 programs, got %d", len(programs))
		}
		if programs[0].Name != "Hypertrophy" || programs[1].Name != "Strength" {
			t.Errorf("unexpected order: %q, %q", programs[0].Name, programs[1].Name)
		}
	})
}

func TestProgramStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s := newTestProgramStore(t)
		created, err := s.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.Get(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got name %q, want %q", got.Name, "Strength")
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := newTestProgramStore(t)

		_, err := s.Get(999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_Create(t *testing.T) {
	t.Run("creates program", func(t *testing.T) {
		s := newTestProgramStore(t)

		p, err := s.Create("Strength")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "Strength" {
			t.Errorf("got name %q, want %q", p.Name, "Strength")
		}
		if p.ID == 0 {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("duplicate name returns error", func(t *testing.T) {
		s := newTestProgramStore(t)
		if _, err := s.Create("Strength"); err != nil {
			t.Fatal(err)
		}

		_, err := s.Create("Strength")
		if err == nil {
			t.Error("expected error for duplicate name, got nil")
		}
	})
}

func TestProgramStore_Update(t *testing.T) {
	t.Run("updates name", func(t *testing.T) {
		s := newTestProgramStore(t)
		created, err := s.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}

		updated, err := s.Update(created.ID, "Max Strength")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Max Strength" {
			t.Errorf("got name %q, want %q", updated.Name, "Max Strength")
		}
	})

	t.Run("not found", func(t *testing.T) {
		s := newTestProgramStore(t)

		_, err := s.Update(999, "Strength")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s := newTestProgramStore(t)
		created, err := s.Create("Strength")
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
		s := newTestProgramStore(t)

		err := s.Delete(999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
