// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func seedProgram(t *testing.T, db *sql.DB, name string) domain.ProgramID {
	t.Helper()
	uID := uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")
	oID := uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")

	// Ensure user and org exist
	_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'test@test.com', 'sub') ON CONFLICT DO NOTHING`, uID)
	_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org') ON CONFLICT DO NOTHING`, oID)

	// Set session variables for RLS
	_, err := db.Exec(fmt.Sprintf("SELECT set_config('app.current_org_id', '%s', false), set_config('app.current_user_id', '%s', false)", oID, uID))
	if err != nil {
		t.Fatal(err)
	}

	var id int64
	err = db.QueryRowContext(context.Background(), `INSERT INTO programs (name, org_id, created_by) VALUES ($1, $2, $3) RETURNING id`, name, oID, uID).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}
	return domain.ProgramID(id)
}

func newTestProgramSetStore(t *testing.T) (*ProgramSetStore, domain.ProgramID) {
	t.Helper()
	db := testutil.NewDB(t)
	ps := NewProgramSetStore(db)

	// create a parent program for tests
	pID := seedProgram(t, db, "Test Program")
	return ps, pID
}

func TestProgramSetStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)

		sets, err := s.List(programID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sets == nil {
			t.Error("expected empty slice, got nil")
		}
		if len(sets) != 0 {
			t.Errorf("expected 0 sets, got %d", len(sets))
		}
	})

	t.Run("returns sets for program only", func(t *testing.T) {
		db := testutil.NewDB(t)
		ps := NewProgramSetStore(db)
		p1 := seedProgram(t, db, "Program 1")
		p2 := seedProgram(t, db, "Program 2")

		if _, err := ps.Create(p1, nil, 1, nil, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := ps.Create(p1, nil, 1, nil, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := ps.Create(p2, nil, 1, nil, nil); err != nil {
			t.Fatal(err)
		}

		sets, err := ps.List(p1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sets) != 2 {
			t.Errorf("expected 2 sets for program 1, got %d", len(sets))
		}
	})
}

func TestProgramSetStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)
		created, err := s.Create(programID, nil, 3, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.Get(programID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Rounds != 3 {
			t.Errorf("got rounds %d, want 3", got.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)

		_, err := s.Get(programID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("wrong program returns not found", func(t *testing.T) {
		db := testutil.NewDB(t)
		ps := NewProgramSetStore(db)
		p1 := seedProgram(t, db, "Program 1")
		p2 := seedProgram(t, db, "Program 2")

		created, err := ps.Create(p1, nil, 1, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = ps.Get(p2, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramSetStore_Create(t *testing.T) {
	t.Run("creates with all fields", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)
		name := "Warmup"
		rest := 60
		order := 1

		ps, err := s.Create(programID, &name, 3, &rest, &order)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.ProgramID != programID {
			t.Errorf("got program_id %d, want %d", ps.ProgramID, programID)
		}
		if ps.Name == nil || *ps.Name != "Warmup" {
			t.Errorf("got name %v, want %q", ps.Name, "Warmup")
		}
		if ps.Rounds != 3 {
			t.Errorf("got rounds %d, want 3", ps.Rounds)
		}
		if ps.IntraSetRestSeconds == nil || *ps.IntraSetRestSeconds != 60 {
			t.Errorf("got intra_set_rest_seconds %v, want 60", ps.IntraSetRestSeconds)
		}
	})

	t.Run("creates with minimal fields", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)

		ps, err := s.Create(programID, nil, 1, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.Name != nil {
			t.Errorf("expected nil name, got %v", ps.Name)
		}
		if ps.IntraSetRestSeconds != nil {
			t.Errorf("expected nil rest, got %v", ps.IntraSetRestSeconds)
		}
	})
}

func TestProgramSetStore_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)
		created, err := s.Create(programID, nil, 1, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		name := "Warmup"
		rest := 90

		updated, err := s.Update(programID, created.ID, &name, 4, &rest, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name == nil || *updated.Name != "Warmup" {
			t.Errorf("got name %v, want %q", updated.Name, "Warmup")
		}
		if updated.Rounds != 4 {
			t.Errorf("got rounds %d, want 4", updated.Rounds)
		}
		if updated.IntraSetRestSeconds == nil || *updated.IntraSetRestSeconds != 90 {
			t.Errorf("got rest %v, want 90", updated.IntraSetRestSeconds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)

		_, err := s.Update(programID, 999, nil, 1, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramSetStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)
		created, err := s.Create(programID, nil, 1, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := s.Delete(programID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = s.Get(programID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, programID := newTestProgramSetStore(t)

		err := s.Delete(programID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
