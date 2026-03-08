// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestProgramService(t *testing.T) (*ProgramService, context.Context) {
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

	return NewProgramService(db), ctx
}

func TestProgramService_List(t *testing.T) {
	t.Run("returns all programs", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		_, _ = svc.Create(ctx, "Strength", nil, true)
		_, _ = svc.Create(ctx, "Hypertrophy", nil, true)

		list, err := svc.List(ctx)
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
		svc, ctx := newTestProgramService(t)
		created, _ := svc.Create(ctx, "Strength", nil, true)

		got, err := svc.GetDetail(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got %q, want Strength", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		_, err := svc.GetDetail(ctx, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Create(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		_, err := svc.Create(ctx, "", nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
		if ve.Msg != "name is required" {
			t.Errorf("got msg %q, want %q", ve.Msg, "name is required")
		}
	})

	t.Run("success", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		p, err := svc.Create(ctx, "Strength", nil, true)
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
		svc, ctx := newTestProgramService(t)
		id, _ := domain.IdentityFromContext(ctx)

		// Create exercise using store directly to simplify (must use ScopedQuerier or commit)
		// Actually service.Create is easier
		exSvc := NewExerciseService(svc.db, store.NewExerciseStore())
		e, err := exSvc.Create(ctx, "Pull-up", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}
		reps := 8

		program, err := svc.CreateFull(ctx, "Strength", nil, true, []ProgramSetInput{
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
		p, err := svc.GetDetail(ctx, program.ID)
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
		if p.OrgID != id.OrgID {
			t.Errorf("got org_id %v, want %v", p.OrgID, id.OrgID)
		}
	})

	t.Run("rolls back on error", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		// Create with non-existent exercise ID
		_, err := svc.CreateFull(ctx, "Fail", nil, true, []ProgramSetInput{
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
		list, _ := svc.List(ctx)
		if len(list) != 0 {
			t.Errorf("expected 0 programs after rollback, got %d", len(list))
		}
	})

	t.Run("invalid exercise_id returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		_, err := svc.CreateFull(ctx, "Fail", nil, true, []ProgramSetInput{
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
		if ve.Msg != "exercise_id 999 not found or access denied" {
			t.Errorf("unexpected message: %q", ve.Msg)
		}
	})
}

func TestProgramService_Update(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Strength", nil, true)

		_, err := svc.Update(ctx, p.ID, "", nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Strength", nil, true)

		updated, err := svc.Update(ctx, p.ID, "New Name", nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("got %q, want New Name", updated.Name)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		_, err := svc.Update(ctx, domain.ProgramID(999), "Name", nil, true)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Strength", nil, true)

		if err := svc.Delete(ctx, p.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.GetDetail(ctx, p.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		err := svc.Delete(ctx, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Normalization(t *testing.T) {
	t.Run("Create normalizes name", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		// Test case 1: Whitespace only name should fail validation
		_, err := svc.Create(ctx, "   ", nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("Create: got error %v, want ValidationError for whitespace name", err)
		}

		// Test case 2: Name with whitespace should be trimmed
		p, err := svc.Create(ctx, "  Strength  ", nil, true)
		if err != nil {
			t.Fatalf("Create: unexpected error: %v", err)
		}
		if p.Name != "Strength" {
			t.Errorf("Create: got name %q, want %q", p.Name, "Strength")
		}
	})

	t.Run("Update normalizes name", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Original", nil, true)

		// Test case 1: Whitespace only name should fail validation
		_, err := svc.Update(ctx, p.ID, "   ", nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("Update: got error %v, want ValidationError for whitespace name", err)
		}

		// Test case 2: Name with whitespace should be trimmed
		updated, err := svc.Update(ctx, p.ID, "  New Name  ", nil, true)
		if err != nil {
			t.Fatalf("Update: unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("Update: got name %q, want %q", updated.Name, "New Name")
		}
	})

	t.Run("CreateFull normalizes name", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		// Test case 1: Whitespace only name should fail validation
		_, err := svc.CreateFull(ctx, "   ", nil, true, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("CreateFull: got error %v, want ValidationError for whitespace name", err)
		}

		// Test case 2: Name with whitespace should be trimmed
		p, err := svc.CreateFull(ctx, "  Full Program  ", nil, true, nil)
		if err != nil {
			t.Fatalf("CreateFull: unexpected error: %v", err)
		}
		if p.Name != "Full Program" {
			t.Errorf("CreateFull: got name %q, want %q", p.Name, "Full Program")
		}
	})
}
