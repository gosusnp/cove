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

func newTestWorkoutSessionStore(t *testing.T) (*WorkoutSessionStore, Querier, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)

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

	return NewWorkoutSessionStore(), q, ctx
}

func TestWorkoutSessionStore_List(t *testing.T) {
	t.Run("empty returns empty slice not nil", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		sessions, err := s.List(ctx, q, id.OrgID, id.UserID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if sessions == nil {
			t.Error("expected empty slice, got nil")
		}
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions, got %d", len(sessions))
		}
	})

	t.Run("returns sessions for user only", func(t *testing.T) {
		db := testutil.NewDB(t)

		uID1 := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")}
		uID2 := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe02")}
		oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")}

		_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'u1@test.com', 'sub1')`, uID1)
		_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'u2@test.com', 'sub2')`, uID2)
		_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org')`, oID)
		_, _ = db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID1)
		_, _ = db.Exec(`INSERT INTO org_members (org_id, user_id, role) VALUES ($1, $2, 'member')`, oID, uID2)

		ctx1 := domain.NewContext(context.Background(), &domain.Identity{UserID: uID1, OrgID: oID})
		ctx2 := domain.NewContext(context.Background(), &domain.Identity{UserID: uID2, OrgID: oID})

		tx1, _ := db.BeginTx(ctx1, nil)
		t.Cleanup(func() { _ = tx1.Rollback() })
		q1 := NewScopedQuerier(tx1, oID.String(), uID1.String())

		tx2, _ := db.BeginTx(ctx2, nil)
		t.Cleanup(func() { _ = tx2.Rollback() })
		q2 := NewScopedQuerier(tx2, oID.String(), uID2.String())

		s := NewWorkoutSessionStore()
		activity := "Run"
		if _, err := s.Create(ctx1, q1, WorkoutSessionParams{Activity: &activity}); err != nil {
			t.Fatal(err)
		}
		_ = tx1.Commit()

		sessions, err := s.List(ctx2, q2, oID, uID2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(sessions) != 0 {
			t.Errorf("expected 0 sessions for user2, got %d", len(sessions))
		}
	})
}

func TestWorkoutSessionStore_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		activity := "Swim"
		created, err := s.Create(ctx, q, WorkoutSessionParams{Activity: &activity})
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.Get(ctx, q, id.OrgID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Activity == nil || *got.Activity != activity {
			t.Errorf("got activity %v, want %q", got.Activity, activity)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.Get(ctx, q, id.OrgID, domain.WorkoutSessionID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestWorkoutSessionStore_Create(t *testing.T) {
	t.Run("creates with all fields", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		activity := "Bike"
		notes := "felt great"
		effort := 7
		duration := 3600

		ws, err := s.Create(ctx, q, WorkoutSessionParams{
			Activity:        &activity,
			SessionNotes:    &notes,
			PerceivedEffort: &effort,
			DurationS:       &duration,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ID == domain.WorkoutSessionID(0) {
			t.Error("expected non-zero ID")
		}
		if ws.Activity == nil || *ws.Activity != activity {
			t.Errorf("got activity %v, want %q", ws.Activity, activity)
		}
		if ws.SessionNotes == nil || *ws.SessionNotes != notes {
			t.Errorf("got notes %v, want %q", ws.SessionNotes, notes)
		}
		if ws.PerceivedEffort == nil || *ws.PerceivedEffort != effort {
			t.Errorf("got effort %v, want %d", ws.PerceivedEffort, effort)
		}
		if ws.DurationS == nil || *ws.DurationS != duration {
			t.Errorf("got duration %v, want %d", ws.DurationS, duration)
		}
	})

	t.Run("creates with no optional fields", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)

		ws, err := s.Create(ctx, q, WorkoutSessionParams{})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.Activity != nil {
			t.Errorf("expected nil activity, got %v", ws.Activity)
		}
		if ws.ProgramStructure != nil {
			t.Errorf("expected nil program_structure, got %v", ws.ProgramStructure)
		}
	})
}

func TestWorkoutSessionStore_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		activity := "Run"
		created, err := s.Create(ctx, q, WorkoutSessionParams{Activity: &activity})
		if err != nil {
			t.Fatal(err)
		}

		newActivity := "Swim"
		effort := 8
		updated, err := s.Update(ctx, q, id.OrgID, created.ID, WorkoutSessionParams{
			Activity:        &newActivity,
			PerceivedEffort: &effort,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Activity == nil || *updated.Activity != newActivity {
			t.Errorf("got activity %v, want %q", updated.Activity, newActivity)
		}
		if updated.PerceivedEffort == nil || *updated.PerceivedEffort != effort {
			t.Errorf("got effort %v, want %d", updated.PerceivedEffort, effort)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.Update(ctx, q, id.OrgID, domain.WorkoutSessionID(999), WorkoutSessionParams{})
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestWorkoutSessionStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, q, WorkoutSessionParams{})
		if err != nil {
			t.Fatal(err)
		}

		if err := s.Delete(ctx, q, id.OrgID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = s.Get(ctx, q, id.OrgID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound after delete, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		err := s.Delete(ctx, q, id.OrgID, domain.WorkoutSessionID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
