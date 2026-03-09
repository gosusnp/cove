// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

type programExerciseFixture struct {
	store      *ProgramExerciseStore
	ctx        context.Context
	db         Querier
	rawDB      *sql.DB
	orgID      domain.OrgID
	programID  domain.ProgramID
	setID      int64
	exerciseID domain.ExerciseID
}

func newProgramExerciseFixture(t *testing.T) programExerciseFixture {
	t.Helper()
	db := testutil.NewDB(t)

	// Seed user/org for required fields
	uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")}
	oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")}
	_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'test@test.com', 'sub') ON CONFLICT DO NOTHING`, uID)
	_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org') ON CONFLICT DO NOTHING`, oID)
	_, _ = db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'admin') ON CONFLICT DO NOTHING`, oID, uID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: uID,
		OrgID:  oID,
	})

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	sq := NewScopedQuerier(tx, oID.String(), uID.String())

	var pID int64
	err = sq.QueryRowContext(ctx, `INSERT INTO programs (name, org_id, created_by) VALUES ($1, $2, $3) RETURNING id`, "Test Program", oID, uID).Scan(&pID)
	if err != nil {
		t.Fatal(err)
	}

	var psID int64
	err = sq.QueryRowContext(ctx, `INSERT INTO program_sets (program_id, name, rounds, intra_set_rest_seconds, sort_order) VALUES ($1, $2, $3, $4, $5) RETURNING id`, pID, nil, 1, nil, nil).Scan(&psID)
	if err != nil {
		t.Fatal(err)
	}

	// Create exercise using ExerciseStore
	e, err := NewExerciseStore().Create(ctx, sq, "Pull-up", nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	return programExerciseFixture{
		store:      NewProgramExerciseStore(),
		ctx:        ctx,
		db:         sq,
		rawDB:      db,
		orgID:      oID,
		programID:  domain.ProgramID(pID),
		setID:      psID,
		exerciseID: e.ID,
	}
}

func TestProgramExerciseStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		f := newProgramExerciseFixture(t)

		exercises, err := f.store.List(f.ctx, f.db, f.orgID, f.setID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if exercises == nil {
			t.Error("expected empty slice, got nil")
		}
		if len(exercises) != 0 {
			t.Errorf("expected 0 exercises, got %d", len(exercises))
		}
	})

	t.Run("returns exercises for set only", func(t *testing.T) {
		f := newProgramExerciseFixture(t)
		// Add another set to the same program
		var psID2 int64
		_ = f.db.QueryRowContext(f.ctx, `INSERT INTO program_sets (program_id, rounds) VALUES ($1, 1) RETURNING id`, f.programID).Scan(&psID2)

		if _, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		// Add to other set
		if _, err := f.store.Create(f.ctx, f.db, f.orgID, psID2, f.exerciseID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}

		exercises, err := f.store.List(f.ctx, f.db, f.orgID, f.setID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(exercises) != 2 {
			t.Errorf("expected 2 exercises for set 1, got %d", len(exercises))
		}
	})
}

func TestProgramExerciseStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		f := newProgramExerciseFixture(t)
		created, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		got, err := f.store.Get(f.ctx, f.db, f.orgID, f.setID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ExerciseID != f.exerciseID {
			t.Errorf("got exercise_id %d, want %d", got.ExerciseID, f.exerciseID)
		}
	})

	t.Run("wrong set returns not found", func(t *testing.T) {
		f := newProgramExerciseFixture(t)
		// Add another set
		var psID2 int64
		_ = f.db.QueryRowContext(f.ctx, `INSERT INTO program_sets (program_id, rounds) VALUES ($1, 1) RETURNING id`, f.programID).Scan(&psID2)

		created, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = f.store.Get(f.ctx, f.db, f.orgID, psID2, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramExerciseStore_Create(t *testing.T) {
	t.Run("creates with all fields", func(t *testing.T) {
		f := newProgramExerciseFixture(t)
		lat := "bilateral"
		reps := 10
		dur := 30
		weight := 20.5
		order := 1

		pe, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, &lat, &reps, &dur, &weight, &order)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pe.Laterality == nil || *pe.Laterality != "bilateral" {
			t.Errorf("got laterality %v, want %q", pe.Laterality, "bilateral")
		}
		if pe.TargetReps == nil || *pe.TargetReps != 10 {
			t.Errorf("got target_reps %v, want 10", pe.TargetReps)
		}
		if pe.TargetWeightKg == nil || *pe.TargetWeightKg != 20.5 {
			t.Errorf("got target_weight_kg %v, want 20.5", pe.TargetWeightKg)
		}
	})

	t.Run("creates with minimal fields", func(t *testing.T) {
		f := newProgramExerciseFixture(t)

		pe, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pe.Laterality != nil {
			t.Errorf("expected nil laterality, got %v", pe.Laterality)
		}
		if pe.TargetReps != nil {
			t.Errorf("expected nil target_reps, got %v", pe.TargetReps)
		}
	})

	t.Run("enforces org access", func(t *testing.T) {
		f := newProgramExerciseFixture(t)

		// Create a program in another org
		// We can reuse the same DB connection, just need a different transaction/session
		otherOrgID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe09")}
		otherUserID := domain.UserID{UUID: uuid.New()}

		// The underlying DB for ScopedQuerier is a transaction, so we need the original DB
		rawDB := f.rawDB
		_, _ = rawDB.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'other-org') ON CONFLICT DO NOTHING`, otherOrgID)
		_, _ = rawDB.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'other@test.com', 'other-sub') ON CONFLICT DO NOTHING`, otherUserID)

		otherTx, _ := rawDB.Begin()
		defer func() { _ = otherTx.Rollback() }()
		otherSq := NewScopedQuerier(otherTx, otherOrgID.String(), otherUserID.String())

		var otherPID int64
		err := otherSq.QueryRowContext(f.ctx, `INSERT INTO programs (name, org_id, created_by) VALUES ($1, $2, $3) RETURNING id`, "Other Program", otherOrgID, otherUserID).Scan(&otherPID)
		if err != nil {
			t.Fatal(err)
		}
		var otherPSID int64
		err = otherSq.QueryRowContext(f.ctx, `INSERT INTO program_sets (program_id, rounds) VALUES ($1, 1) RETURNING id`, otherPID).Scan(&otherPSID)
		if err != nil {
			t.Fatal(err)
		}
		if err := otherTx.Commit(); err != nil {
			t.Fatal(err)
		}

		// Try to create exercise in other org's program set using my orgID
		_, err = f.store.Create(f.ctx, f.db, f.orgID, otherPSID, f.exerciseID, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound (due to org mismatch)", err)
		}
	})
}

func TestProgramExerciseStore_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		f := newProgramExerciseFixture(t)
		e2, _ := NewExerciseStore().Create(f.ctx, f.db, "Push-up", nil, nil, true)

		created, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		reps := 12

		updated, err := f.store.Update(f.ctx, f.db, f.orgID, f.setID, created.ID, e2.ID, nil, &reps, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.ExerciseID != e2.ID {
			t.Errorf("got exercise_id %d, want %d", updated.ExerciseID, e2.ID)
		}
		if updated.TargetReps == nil || *updated.TargetReps != 12 {
			t.Errorf("got target_reps %v, want 12", updated.TargetReps)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := newProgramExerciseFixture(t)

		_, err := f.store.Update(f.ctx, f.db, f.orgID, f.setID, 999, f.exerciseID, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramExerciseStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		f := newProgramExerciseFixture(t)
		created, err := f.store.Create(f.ctx, f.db, f.orgID, f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.store.Delete(f.ctx, f.db, f.orgID, f.setID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = f.store.Get(f.ctx, f.db, f.orgID, f.setID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := newProgramExerciseFixture(t)

		err := f.store.Delete(f.ctx, f.db, f.orgID, f.setID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
