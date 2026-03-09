// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

type programExerciseFixture struct {
	svc        *ProgramExerciseService
	ctx        context.Context
	setID      int64
	exerciseID domain.ExerciseID
}

func newTestProgramExerciseService(t *testing.T) programExerciseFixture {
	t.Helper()
	db := testutil.NewDB(t)

	// Seed user/org for required fields
	uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
	user, _, _ := uSvc.GetOrCreate(context.Background(), domain.Email("test@test.com"), domain.GoogleSub("sub"))
	var orgID domain.OrgID
	_ = db.QueryRow(`SELECT org_id FROM org_members WHERE user_id = $1`, user.ID).Scan(&orgID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})

	// Set session variables for RLS
	_, _ = db.Exec(`SELECT set_config('app.current_org_id', $1, false), set_config('app.current_user_id', $2, false)`, orgID.String(), user.ID.String())

	var pID int64
	err := db.QueryRowContext(ctx, `INSERT INTO programs (name, org_id, created_by) VALUES ($1, $2, $3) RETURNING id`, "Test Program", orgID, user.ID).Scan(&pID)
	if err != nil {
		t.Fatal(err)
	}
	var psID int64
	err = db.QueryRowContext(ctx, `INSERT INTO program_sets (program_id, rounds) VALUES ($1, 1) RETURNING id`, pID).Scan(&psID)
	if err != nil {
		t.Fatal(err)
	}

	tx, err := db.Begin()
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx.Rollback() })

	q := store.NewScopedQuerier(tx, orgID.String(), user.ID.String())

	// Create exercise with identity context and scoped transaction
	e, err := store.NewExerciseStore().Create(ctx, q, "Pull-up", nil, nil, true)
	if err != nil {
		t.Fatal(err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}

	return programExerciseFixture{
		svc:        NewProgramExerciseService(db, store.NewProgramExerciseStore()),
		ctx:        ctx,
		setID:      psID,
		exerciseID: e.ID,
	}
}

func TestProgramExerciseService_List(t *testing.T) {
	t.Run("returns all exercises for set", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		_, _ = f.svc.Create(f.ctx, f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		list, err := f.svc.List(f.ctx, f.setID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 exercise, got %d", len(list))
		}
	})
}

func TestProgramExerciseService_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		created, _ := f.svc.Create(f.ctx, f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		got, err := f.svc.Get(f.ctx, f.setID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.ExerciseID != f.exerciseID {
			t.Errorf("got %d, want %d", got.ExerciseID, f.exerciseID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		_, err := f.svc.Get(f.ctx, f.setID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramExerciseService_Create(t *testing.T) {
	t.Run("invalid exercise_id returns ValidationError", func(t *testing.T) {
		f := newTestProgramExerciseService(t)

		_, err := f.svc.Create(f.ctx, f.setID, 0, nil, nil, nil, nil, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		pe, err := f.svc.Create(f.ctx, f.setID, f.exerciseID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if pe.ExerciseID != f.exerciseID {
			t.Errorf("got %d, want %d", pe.ExerciseID, f.exerciseID)
		}
	})
}

func TestProgramExerciseService_Update(t *testing.T) {
	t.Run("invalid exercise_id returns ValidationError", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		pe, _ := f.svc.Create(f.ctx, f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		_, err := f.svc.Update(f.ctx, f.setID, pe.ID, 0, nil, nil, nil, nil, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		f := newTestProgramExerciseService(t)

		_, err := f.svc.Update(f.ctx, f.setID, 999, f.exerciseID, nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramExerciseService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		f := newTestProgramExerciseService(t)
		created, _ := f.svc.Create(f.ctx, f.setID, f.exerciseID, nil, nil, nil, nil, nil)

		if err := f.svc.Delete(f.ctx, f.setID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := f.svc.Get(f.ctx, f.setID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		f := newTestProgramExerciseService(t)

		err := f.svc.Delete(f.ctx, f.setID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
