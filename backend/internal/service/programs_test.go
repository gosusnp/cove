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

func newTestProgramService(t *testing.T) *ProgramService {
	t.Helper()
	return NewProgramService(testutil.NewDB(t))
}

func TestProgramService_List(t *testing.T) {
	t.Run("returns all programs", func(t *testing.T) {
		svc := newTestProgramService(t)
		_, _ = svc.Create("Strength")
		_, _ = svc.Create("Hypertrophy")

		list, err := svc.List()
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 2 {
			t.Errorf("expected 2 programs, got %d", len(list))
		}
	})
}

func TestProgramService_GetDetail(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc := newTestProgramService(t)
		created, _ := svc.Create("Strength")

		got, err := svc.GetDetail(created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got %q, want Strength", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc := newTestProgramService(t)
		_, err := svc.GetDetail(domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
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

	t.Run("success", func(t *testing.T) {
		svc := newTestProgramService(t)

		p, err := svc.Create("Strength")
		if err != nil {
			t.Fatal(err)
		}
		if p.Name != "Strength" {
			t.Errorf("got name %q, want %q", p.Name, "Strength")
		}
	})
}

func TestProgramService_CreateFull(t *testing.T) {
	t.Run("creates full hierarchy atomically", func(t *testing.T) {
		db := testutil.NewDB(t)
		svc := NewProgramService(db)

		// Seed user/org for required fields
		uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
		user, _, _ := uSvc.GetOrCreate(context.Background(), "test@test.com", "sub")
		var orgID domain.OrgID
		_ = db.QueryRow(`SELECT org_id FROM org_members WHERE user_id = $1`, user.ID).Scan(&orgID)
		ctx := domain.NewContext(context.Background(), &domain.Identity{
			UserID: user.ID,
			OrgID:  orgID,
		})

		tx, err := db.Begin()
		if err != nil {
			t.Fatal(err)
		}
		q := store.NewScopedQuerier(tx, orgID.String(), user.ID.String())

		e, err := store.NewExerciseStore().Create(ctx, q, "Pull-up", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		if err := tx.Commit(); err != nil {
			t.Fatal(err)
		}
		reps := 8

		program, err := svc.CreateFull("Strength", []ProgramSetInput{
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

		// Verify program exists
		p, err := svc.GetDetail(program.ID)
		if err != nil {
			t.Fatal(err)
		}
		if len(p.Sets) != 1 {
			t.Errorf("expected 1 set, got %d", len(p.Sets))
		}
		if len(p.Sets[0].Exercises) != 1 {
			t.Errorf("expected 1 exercise in set, got %d", len(p.Sets[0].Exercises))
		}
		if p.Sets[0].Exercises[0].ExerciseID != e.ID {
			t.Errorf("got exercise_id %d, want %d", p.Sets[0].Exercises[0].ExerciseID, e.ID)
		}
	})

	t.Run("rolls back on error", func(t *testing.T) {
		db := testutil.NewDB(t)
		svc := NewProgramService(db)

		// Create with non-existent exercise ID
		_, err := svc.CreateFull("Fail", []ProgramSetInput{
			{
				Rounds: 1,
				Exercises: []ProgramExerciseInput{
					{ExerciseID: domain.ExerciseID(999)},
				},
			},
		})

		if err == nil {
			t.Fatal("expected error, got nil")
		}

		// Verify program was NOT created
		list, _ := svc.List()
		if len(list) != 0 {
			t.Errorf("expected 0 programs after rollback, got %d", len(list))
		}
	})

	t.Run("invalid exercise_id returns ValidationError", func(t *testing.T) {
		db := testutil.NewDB(t)
		svc := NewProgramService(db)

		_, err := svc.CreateFull("Fail", []ProgramSetInput{
			{
				Rounds: 1,
				Exercises: []ProgramExerciseInput{
					{ExerciseID: domain.ExerciseID(999)},
				},
			},
		})

		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "exercise_id 999 not found" {
			t.Errorf("unexpected message: %q", ve.Msg)
		}
	})
}

func TestProgramService_Update(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc := newTestProgramService(t)
		p, _ := svc.Create("Strength")

		_, err := svc.Update(p.ID, "")
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc := newTestProgramService(t)
		p, _ := svc.Create("Strength")

		updated, err := svc.Update(p.ID, "New Name")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("got %q, want New Name", updated.Name)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc := newTestProgramService(t)

		_, err := svc.Update(domain.ProgramID(999), "Name")
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc := newTestProgramService(t)
		p, _ := svc.Create("Strength")

		if err := svc.Delete(p.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.GetDetail(p.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc := newTestProgramService(t)

		err := svc.Delete(domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
