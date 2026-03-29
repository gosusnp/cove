// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"strings"
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

	_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'test@test.com', 'sub')`, uID)
	_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'test-org')`, oID)
	_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: uID,
		OrgID:  oID,
	})

	return NewProgramService(db, store.NewExerciseStore()), ctx
}

func TestProgramService_List(t *testing.T) {
	t.Run("returns all programs", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		_, _ = svc.Create(ctx, "Strength", nil, nil, true)
		_, _ = svc.Create(ctx, "Hypertrophy", nil, nil, true)

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
		created, _ := svc.Create(ctx, "Strength", nil, nil, true)

		got, err := svc.Get(ctx, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Strength" {
			t.Errorf("got %q, want Strength", got.Name)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		_, err := svc.Get(ctx, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Create(t *testing.T) {
	t.Run("empty name returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		_, err := svc.Create(ctx, "", nil, nil, true)
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

		p, err := svc.Create(ctx, "Strength", nil, nil, true)
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

		program, err := svc.CreateFull(ctx, "Strength", nil, nil, true, []ProgramSetInput{
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
		p, err := svc.Get(ctx, program.ID)
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
		_, err := svc.CreateFull(ctx, "Fail", nil, nil, true, []ProgramSetInput{
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

		_, err := svc.CreateFull(ctx, "Fail", nil, nil, true, []ProgramSetInput{
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
		p, _ := svc.Create(ctx, "Strength", nil, nil, true)

		_, err := svc.Update(ctx, p.ID, nil, "", nil, nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("success", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Strength", nil, nil, true)
		full, _ := svc.Get(ctx, p.ID)

		updated, err := svc.Update(ctx, p.ID, &full.UpdatedAt, "New Name", nil, nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "New Name" {
			t.Errorf("got %q, want New Name", updated.Name)
		}
	})

	t.Run("not found returns ErrNotFound", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		_, err := svc.Update(ctx, domain.ProgramID(999), nil, "Name", nil, nil, true)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Strength", nil, nil, true)

		if err := svc.Delete(ctx, p.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.Get(ctx, p.ID)
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

func TestProgramService_Sets(t *testing.T) {
	t.Run("CreateSet defaults rounds", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)

		ps, err := svc.CreateSet(ctx, p.ID, nil, 0, nil)
		if err != nil {
			t.Fatal(err)
		}
		if ps.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", ps.Rounds)
		}
	})

	t.Run("UpdateSet defaults rounds", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 3, nil)
		full, _ := svc.Get(ctx, p.ID)

		updated, err := svc.UpdateSet(ctx, p.ID, ps.ID, &full.UpdatedAt, nil, -5, nil)
		if err != nil {
			t.Fatal(err)
		}
		if updated.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", updated.Rounds)
		}
	})

	t.Run("Set CRUD found and not found", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)

		// List empty
		list, _ := svc.ListSets(ctx, p.ID)
		if len(list) != 0 {
			t.Errorf("expected 0 sets, got %d", len(list))
		}

		// Create
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 3, nil)

		// Get found
		got, err := svc.GetSet(ctx, p.ID, ps.ID)
		if err != nil || got.Rounds != 3 {
			t.Errorf("failed to get set: %v", err)
		}

		// Get not found
		_, err = svc.GetSet(ctx, p.ID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}

		// Delete
		if err := svc.DeleteSet(ctx, p.ID, ps.ID); err != nil {
			t.Fatal(err)
		}
		_, err = svc.GetSet(ctx, p.ID, ps.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})
}

func TestProgramService_Exercises(t *testing.T) {
	t.Run("CreateExercise validation", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)

		_, err := svc.CreateExercise(ctx, p.ID, ps.ID, 0, nil, nil, nil, nil, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("got %v, want ValidationError", err)
		}
	})

	t.Run("weight without unit returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)
		w := 80.0

		_, err := svc.CreateExercise(ctx, p.ID, ps.ID, 1, nil, nil, nil, &w, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("weight with nil unit: got %v, want ValidationError", err)
		}
	})

	t.Run("weight with non-mass unit returns ValidationError", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)
		w := 80.0
		u := domain.UnitMilliliter

		_, err := svc.CreateExercise(ctx, p.ID, ps.ID, 1, nil, nil, nil, &w, &u)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("weight with volume unit: got %v, want ValidationError", err)
		}
	})

	t.Run("Exercise CRUD found and not found", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)

		// Seed an exercise to reference
		exSvc := NewExerciseService(svc.db, store.NewExerciseStore())
		e, _ := exSvc.Create(ctx, "Squat", nil, nil, true)

		// Create
		pe, err := svc.CreateExercise(ctx, p.ID, ps.ID, e.ID, nil, nil, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		// Get found
		got, err := svc.GetExercise(ctx, p.ID, ps.ID, pe.ID)
		if err != nil || got.ExerciseID != e.ID {
			t.Errorf("failed to get exercise: %v", err)
		}

		// Update
		fullForUpdate, _ := svc.Get(ctx, p.ID)
		reps := 10
		updated, err := svc.UpdateExercise(ctx, p.ID, ps.ID, pe.ID, &fullForUpdate.UpdatedAt, e.ID, nil, &reps, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}
		if updated.TargetReps == nil || *updated.TargetReps != 10 {
			t.Errorf("got reps %v, want 10", updated.TargetReps)
		}

		// Delete
		if err := svc.DeleteExercise(ctx, p.ID, ps.ID, pe.ID); err != nil {
			t.Fatal(err)
		}
		_, err = svc.GetExercise(ctx, p.ID, ps.ID, pe.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})
}

func TestProgramService_NameResolution(t *testing.T) {
	t.Run("GetDetail uses live exercise name", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		exSvc := NewExerciseService(svc.db, store.NewExerciseStore())
		e, _ := exSvc.Create(ctx, "Squat", nil, nil, true)

		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)
		if _, err := svc.CreateExercise(ctx, p.ID, ps.ID, e.ID, nil, nil, nil, nil, nil); err != nil {
			t.Fatal(err)
		}

		detail, err := svc.Get(ctx, p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got := detail.Sets[0].Exercises[0].Name; got != "Squat" {
			t.Errorf("got name %q, want %q", got, "Squat")
		}
	})

	t.Run("GetDetail falls back to name_snapshot for inaccessible exercise", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		id, _ := domain.IdentityFromContext(ctx)

		// Seed a private exercise for a second org (invisible to org1) using raw SQL.
		uID2 := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe02")}
		oID2 := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe03")}
		_, _ = svc.db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'u2@fallback.test', 'sub-fb2')`, uID2)
		_, _ = svc.db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'fallback-org2')`, oID2)
		_, _ = svc.db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID2, uID2)
		ctx2 := domain.NewContext(t.Context(), &domain.Identity{UserID: uID2, OrgID: oID2})

		exSvc2 := NewExerciseService(svc.db, store.NewExerciseStore())
		e, _ := exSvc2.Create(ctx2, "Deadlift", nil, nil, false)

		// Create program for org1 and seed the exercise reference via store (bypasses visibility).
		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)

		pStore := store.NewProgramStore()
		if err := withScopedTx(ctx, svc.db, func(q store.Querier) error {
			_, err := pStore.CreateExercise(t.Context(), q, id.OrgID, p.ID, ps.ID, e.ID, "Deadlift", nil, nil, nil, nil, nil)
			return err
		}); err != nil {
			t.Fatal(err)
		}

		detail, err := svc.Get(ctx, p.ID)
		if err != nil {
			t.Fatal(err)
		}
		if got := detail.Sets[0].Exercises[0].Name; got != "Deadlift" {
			t.Errorf("got fallback name %q, want %q", got, "Deadlift")
		}
	})

	t.Run("CreateExercise rejects inaccessible exercise_id", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)

		_, err := svc.CreateExercise(ctx, p.ID, ps.ID, domain.ExerciseID(999), nil, nil, nil, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("name_snapshot preserved on UpdateExercise", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		exSvc := NewExerciseService(svc.db, store.NewExerciseStore())
		e, _ := exSvc.Create(ctx, "Bench Press", nil, nil, true)

		p, _ := svc.Create(ctx, "Test", nil, nil, true)
		ps, _ := svc.CreateSet(ctx, p.ID, nil, 1, nil)
		pe, _ := svc.CreateExercise(ctx, p.ID, ps.ID, e.ID, nil, nil, nil, nil, nil)

		fullForSnapshot, _ := svc.Get(ctx, p.ID)
		reps := 5
		if _, err := svc.UpdateExercise(ctx, p.ID, ps.ID, pe.ID, &fullForSnapshot.UpdatedAt, e.ID, nil, &reps, nil, nil, nil); err != nil {
			t.Fatal(err)
		}

		// Verify snapshot is preserved by reading the JSONB directly.
		var setsJSON []byte
		if err := svc.db.QueryRow(`SELECT sets FROM cove.programs WHERE id = $1`, p.ID).Scan(&setsJSON); err != nil {
			t.Fatal(err)
		}
		if !strings.Contains(string(setsJSON), `"name_snapshot"`) {
			t.Error("name_snapshot field missing from JSONB after update")
		}
		if !strings.Contains(string(setsJSON), "Bench Press") {
			t.Errorf("name_snapshot value lost after update; json: %s", setsJSON)
		}
	})
}

func TestProgramService_Normalization(t *testing.T) {
	t.Run("Create normalizes name", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)

		// Test case 1: Whitespace only name should fail validation
		_, err := svc.Create(ctx, "   ", nil, nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("Create: got error %v, want ValidationError for whitespace name", err)
		}

		// Test case 2: Name with whitespace should be trimmed
		p, err := svc.Create(ctx, "  Strength  ", nil, nil, true)
		if err != nil {
			t.Fatalf("Create: unexpected error: %v", err)
		}
		if p.Name != "Strength" {
			t.Errorf("Create: got name %q, want %q", p.Name, "Strength")
		}
	})

	t.Run("Update normalizes name", func(t *testing.T) {
		svc, ctx := newTestProgramService(t)
		p, _ := svc.Create(ctx, "Original", nil, nil, true)
		fullOrig, _ := svc.Get(ctx, p.ID)

		// Test case 1: Whitespace only name should fail validation
		_, err := svc.Update(ctx, p.ID, nil, "   ", nil, nil, true)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("Update: got error %v, want ValidationError for whitespace name", err)
		}

		// Test case 2: Name with whitespace should be trimmed
		updated, err := svc.Update(ctx, p.ID, &fullOrig.UpdatedAt, "  New Name  ", nil, nil, true)
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
		_, err := svc.CreateFull(ctx, "   ", nil, nil, true, nil)
		var ve *ValidationError
		if !errors.As(err, &ve) {
			t.Fatalf("CreateFull: got error %v, want ValidationError for whitespace name", err)
		}

		// Test case 2: Name with whitespace should be trimmed
		p, err := svc.CreateFull(ctx, "  Full Program  ", nil, nil, true, nil)
		if err != nil {
			t.Fatalf("CreateFull: unexpected error: %v", err)
		}
		if p.Name != "Full Program" {
			t.Errorf("CreateFull: got name %q, want %q", p.Name, "Full Program")
		}
	})
}
