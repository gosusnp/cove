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

type programExerciseFixture struct {
	store      *ProgramExerciseStore
	ctx        context.Context
	db         Querier
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

	p, err := NewProgramStore(db).Create("Test Program")
	if err != nil {
		t.Fatal(err)
	}
	ps, err := NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// Create exercise with identity context and scoped querier
	e, err := NewExerciseStore().Create(ctx, q, "Pull-up", nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	return programExerciseFixture{
		store:      NewProgramExerciseStore(db),
		ctx:        ctx,
		db:         NewScopedQuerier(db, oID.String(), uID.String()),
		programID:  p.ID,
		setID:      ps.ID,
		exerciseID: e.ID,
	}
}

func TestProgramExerciseStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		f := newProgramExerciseFixture(t)

		exercises, err := f.store.List(f.setID)
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
		db := testutil.NewDB(t)

		// Seed user/org for required fields
		uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe07")}
		oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe08")}
		_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'test4@test.com', 'sub4')`, uID)
		_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org4')`, oID)
		_, _ = db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)

		ctx := domain.NewContext(context.Background(), &domain.Identity{
			UserID: uID,
			OrgID:  oID,
		})

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback() }()

		q := NewScopedQuerier(tx, oID.String(), uID.String())

		p, _ := NewProgramStore(db).Create("Program")
		set1, _ := NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
		set2, _ := NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
		// Create exercise with identity context and scoped transaction
		e, _ := NewExerciseStore().Create(ctx, q, "Pull-up", nil, nil, true)

		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}

		pes := NewProgramExerciseStore(db)

		if _, err := pes.Create(set1.ID, e.ID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := pes.Create(set1.ID, e.ID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}
		if _, err := pes.Create(set2.ID, e.ID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}

		exercises, err := pes.List(set1.ID)
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
		created, err := f.store.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		got, err := f.store.Get(f.setID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ExerciseID != f.exerciseID {
			t.Errorf("got exercise_id %d, want %d", got.ExerciseID, f.exerciseID)
		}
	})

	t.Run("wrong set returns not found", func(t *testing.T) {
		db := testutil.NewDB(t)

		// Seed user/org for required fields
		uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe05")}
		oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe06")}
		_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'test3@test.com', 'sub3')`, uID)
		_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org3')`, oID)
		_, _ = db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)

		ctx := domain.NewContext(context.Background(), &domain.Identity{
			UserID: uID,
			OrgID:  oID,
		})

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback() }()

		q := NewScopedQuerier(tx, oID.String(), uID.String())

		p, _ := NewProgramStore(db).Create("Program")
		set1, _ := NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
		set2, _ := NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
		// Create exercise with identity context and scoped transaction
		e, _ := NewExerciseStore().Create(ctx, q, "Pull-up", nil, nil, true)

		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}

		pes := NewProgramExerciseStore(db)

		created, err := pes.Create(set1.ID, e.ID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		_, err = pes.Get(set2.ID, created.ID)
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

		pe, err := f.store.Create(f.setID, f.exerciseID, &lat, &reps, &dur, &weight, &order)
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

		pe, err := f.store.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)
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
}

func TestProgramExerciseStore_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		db := testutil.NewDB(t)

		// Seed user/org for required fields
		uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe03")}
		oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe04")}
		_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'test2@test.com', 'sub2')`, uID)
		_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org2')`, oID)
		_, _ = db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)

		ctx := domain.NewContext(context.Background(), &domain.Identity{
			UserID: uID,
			OrgID:  oID,
		})

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback() }()

		q := NewScopedQuerier(tx, oID.String(), uID.String())

		p, _ := NewProgramStore(db).Create("Program")
		ps, _ := NewProgramSetStore(db).Create(p.ID, nil, 1, nil, nil)
		// Create exercise with identity context and scoped transaction
		e1, _ := NewExerciseStore().Create(ctx, q, "Pull-up", nil, nil, true)
		e2, _ := NewExerciseStore().Create(ctx, q, "Push-up", nil, nil, true)

		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}

		pes := NewProgramExerciseStore(db)

		created, err := pes.Create(ps.ID, e1.ID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		reps := 12

		updated, err := pes.Update(ps.ID, created.ID, e2.ID, nil, &reps, nil, nil, nil)
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

		_, err := f.store.Update(f.setID, 999, f.exerciseID, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramExerciseStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		f := newProgramExerciseFixture(t)
		created, err := f.store.Create(f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		if err := f.store.Delete(f.setID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = f.store.Get(f.setID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := newProgramExerciseFixture(t)

		err := f.store.Delete(f.setID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
