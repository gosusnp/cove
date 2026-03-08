// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestProgramStore(t *testing.T) (*ProgramStore, Querier, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)

	// Seed user and org for required fields
	uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")}
	oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")}

	_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'test@test.com', 'sub')`, uID)
	_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org')`, oID)
	_, _ = db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: uID,
		OrgID:  oID,
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	q := NewScopedQuerier(tx, oID.String(), uID.String())

	return NewProgramStore(), q, ctx
}

func TestProgramStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		programs, err := s.List(ctx, db, id.OrgID)
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
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		if _, err := s.Create(ctx, db, "Strength", nil, true); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create(ctx, db, "Hypertrophy", nil, true); err != nil {
			t.Fatal(err)
		}

		programs, err := s.List(ctx, db, id.OrgID)
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
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, db, "Strength", nil, true)
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.Get(ctx, db, id.OrgID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got name %q, want %q", got.Name, "Strength")
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.Get(ctx, db, id.OrgID, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_Create(t *testing.T) {
	t.Run("creates program", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)

		p, err := s.Create(ctx, db, "Strength", nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.Name != "Strength" {
			t.Errorf("got name %q, want %q", p.Name, "Strength")
		}
		if p.ID == domain.ProgramID(0) {
			t.Error("expected non-zero ID")
		}
	})
}

func TestProgramStore_Update(t *testing.T) {
	t.Run("updates name", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, db, "Strength", nil, true)
		if err != nil {
			t.Fatal(err)
		}

		updated, err := s.Update(ctx, db, id.OrgID, created.ID, "Max Strength", nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Max Strength" {
			t.Errorf("got name %q, want %q", updated.Name, "Max Strength")
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.Update(ctx, db, id.OrgID, domain.ProgramID(999), "Strength", nil, true)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_GetDetail(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.GetDetail(ctx, db, id.OrgID, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("empty program has empty sets slice", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		p, err := s.Create(ctx, db, "Strength", nil, true)
		if err != nil {
			t.Fatal(err)
		}

		detail, err := s.GetDetail(ctx, db, id.OrgID, p.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if detail.Sets == nil {
			t.Error("expected empty sets slice, got nil")
		}
		if len(detail.Sets) != 0 {
			t.Errorf("expected 0 sets, got %d", len(detail.Sets))
		}
	})

	t.Run("returns full hierarchy", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		// Create program and sets using direct SQL because we only have ProgramStore
		var programID int64
		err := db.QueryRowContext(ctx, `INSERT INTO programs (name, org_id, is_public, created_by) VALUES ($1, $2, $3, $4) RETURNING id`, "Full Program", id.OrgID, true, id.UserID).Scan(&programID)
		if err != nil {
			t.Fatal(err)
		}

		var setID int64
		err = db.QueryRowContext(ctx, `INSERT INTO program_sets (program_id, name, rounds, intra_set_rest_seconds, sort_order) VALUES ($1, $2, $3, $4, $5) RETURNING id`, programID, "Set 1", 3, 60, 1).Scan(&setID)
		if err != nil {
			t.Fatal(err)
		}

		e, err := NewExerciseStore().Create(ctx, db, "Pull-up", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		reps := 8
		_, err = db.ExecContext(ctx, `INSERT INTO program_exercises (program_set_id, exercise_id, laterality, target_reps, target_duration_seconds, target_weight_kg, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`, setID, e.ID, nil, reps, nil, nil, 1)
		if err != nil {
			t.Fatal(err)
		}

		detail, err := s.GetDetail(ctx, db, id.OrgID, domain.ProgramID(programID))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(detail.Sets) != 1 {
			t.Fatalf("expected 1 set, got %d", len(detail.Sets))
		}
		if detail.Sets[0].Rounds != 3 {
			t.Errorf("got rounds %d, want 3", detail.Sets[0].Rounds)
		}
		if detail.Sets[0].Exercises == nil {
			t.Error("expected empty exercises slice, got nil")
		}
		if len(detail.Sets[0].Exercises) != 1 {
			t.Fatalf("expected 1 exercise, got %d", len(detail.Sets[0].Exercises))
		}
		if detail.Sets[0].Exercises[0].Name != "Pull-up" {
			t.Errorf("got exercise name %q, want %q", detail.Sets[0].Exercises[0].Name, "Pull-up")
		}
		if detail.Sets[0].Exercises[0].TargetReps == nil || *detail.Sets[0].Exercises[0].TargetReps != 8 {
			t.Errorf("got target_reps %v, want 8", detail.Sets[0].Exercises[0].TargetReps)
		}
	})
}

func TestProgramStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, db, "Strength", nil, true)
		if err != nil {
			t.Fatal(err)
		}

		if err := s.Delete(ctx, db, id.OrgID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = s.Get(ctx, db, id.OrgID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		err := s.Delete(ctx, db, id.OrgID, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
