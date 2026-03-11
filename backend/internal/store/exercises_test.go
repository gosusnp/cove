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

func newTestExerciseStore(t *testing.T) (*ExerciseStore, Querier, context.Context) {
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

	return NewExerciseStore(), q, ctx
}

func TestExerciseStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		exercises, err := s.List(ctx, db, id.OrgID)
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

	t.Run("returns all exercises ordered by name", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		prog := "Add 1 rep"
		desc := "Bodyweight exercise"
		if _, err := s.Create(ctx, db, "Push-up", &prog, &desc, true); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create(ctx, db, "Air Squat", nil, nil, true); err != nil {
			t.Fatal(err)
		}

		exercises, err := s.List(ctx, db, id.OrgID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(exercises) != 2 {
			t.Fatalf("expected 2 exercises, got %d", len(exercises))
		}
		// Ordered by name: Air Squat, Push-up
		if exercises[0].Name != "Air Squat" {
			t.Errorf("expected first exercise to be Air Squat, got %q", exercises[0].Name)
		}
		if exercises[1].Name != "Push-up" {
			t.Errorf("expected second exercise to be Push-up, got %q", exercises[1].Name)
		}
		if exercises[1].Progression == nil || *exercises[1].Progression != prog {
			t.Errorf("expected progression %q, got %v", prog, exercises[1].Progression)
		}
		if exercises[1].Description == nil || *exercises[1].Description != desc {
			t.Errorf("expected description %q, got %v", desc, exercises[1].Description)
		}
	})
}

func TestExerciseStore_GetByIDs(t *testing.T) {
	t.Run("empty returns empty slices", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		found, missing, err := s.GetByIDs(ctx, db, id.OrgID, []domain.ExerciseID{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if found == nil || missing == nil {
			t.Fatal("expected empty slices, got nil")
		}
		if len(found) != 0 || len(missing) != 0 {
			t.Errorf("expected empty results, got found=%d missing=%d", len(found), len(missing))
		}
	})

	t.Run("returns found and missing", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		e1, _ := s.Create(ctx, db, "A", nil, nil, true)
		e2, _ := s.Create(ctx, db, "B", nil, nil, true)
		_, _ = s.Create(ctx, db, "C", nil, nil, true)

		var ghost domain.ExerciseID = 99999
		ids := []domain.ExerciseID{e1.ID, e2.ID, ghost}
		found, missing, err := s.GetByIDs(ctx, db, id.OrgID, ids)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(found) != 2 {
			t.Fatalf("expected 2 found, got %d", len(found))
		}
		if found[0].Name != "A" || found[1].Name != "B" {
			t.Errorf("got %v, want exercises A and B", found)
		}
		if len(missing) != 1 || missing[0] != ghost {
			t.Errorf("expected missing=[%v], got %v", ghost, missing)
		}
	})
}

func TestExerciseStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		prog := "Add 1 rep each session"
		created, err := s.Create(ctx, db, "Push-up", &prog, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.Get(ctx, db, id.OrgID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Name != "Push-up" {
			t.Errorf("got name %q, want %q", got.Name, "Push-up")
		}
		if got.Progression == nil || *got.Progression != prog {
			t.Errorf("got progression %v, want %q", got.Progression, prog)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.Get(ctx, db, id.OrgID, domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseStore_Create(t *testing.T) {
	t.Run("creates with progression", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		prog := "Bodyweight"

		e, err := s.Create(ctx, db, "Push-up", &prog, nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Name != "Push-up" {
			t.Errorf("got name %q, want %q", e.Name, "Push-up")
		}
		if e.Progression == nil || *e.Progression != prog {
			t.Errorf("got progression %v, want %q", e.Progression, prog)
		}
		if e.ID == domain.ExerciseID(0) {
			t.Error("expected non-zero ID")
		}
	})

	t.Run("creates without progression", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)

		e, err := s.Create(ctx, db, "Push-up", nil, nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if e.Progression != nil {
			t.Errorf("expected nil progression, got %v", e.Progression)
		}
	})
}

func TestExerciseStore_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, db, "Push-up", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		newProg := "Weight vest"
		updated, err := s.Update(ctx, db, id.OrgID, created.ID, "Hard Push-up", &newProg, nil, true)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name != "Hard Push-up" {
			t.Errorf("got name %q, want %q", updated.Name, "Hard Push-up")
		}
		if updated.Progression == nil || *updated.Progression != newProg {
			t.Errorf("got progression %v, want %q", updated.Progression, newProg)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.Update(ctx, db, id.OrgID, domain.ExerciseID(999), "Push-up", nil, nil, true)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestExerciseStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, db, "Push-up", nil, nil, true)
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
		s, db, ctx := newTestExerciseStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		err := s.Delete(ctx, db, id.OrgID, domain.ExerciseID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
