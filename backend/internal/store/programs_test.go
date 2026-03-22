// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"errors"
	"sync"
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

	_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'test@test.com', 'sub')`, uID)
	_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'test-org')`, oID)
	_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)

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
		if _, err := s.Create(ctx, db, id.OrgID, "Strength", nil, nil, true); err != nil {
			t.Fatal(err)
		}
		if _, err := s.Create(ctx, db, id.OrgID, "Hypertrophy", nil, nil, true); err != nil {
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

func TestProgramStore_GetLite(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, db, id.OrgID, "Strength", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		got, err := s.GetLite(ctx, db, id.OrgID, created.ID)
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

		_, err := s.GetLite(ctx, db, id.OrgID, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_Create(t *testing.T) {
	t.Run("creates program", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		p, err := s.Create(ctx, db, id.OrgID, "Strength", nil, nil, true)
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
		created, err := s.Create(ctx, db, id.OrgID, "Strength", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		updated, err := s.Update(ctx, db, id.OrgID, created.ID, "Max Strength", nil, nil, true)
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

		_, err := s.Update(ctx, db, id.OrgID, domain.ProgramID(999), "Strength", nil, nil, true)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_Get(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)

		_, err := s.Get(ctx, db, id.OrgID, domain.ProgramID(999))
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("empty program has empty sets slice", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		p, err := s.Create(ctx, db, id.OrgID, "Strength", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		detail, err := s.Get(ctx, db, id.OrgID, p.ID)
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

		p, err := s.Create(ctx, db, id.OrgID, "Full Program", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		setName := "Set 1"
		restSecs := 60
		ps, err := s.CreateSet(ctx, db, id.OrgID, p.ID, &setName, 3, &restSecs)
		if err != nil {
			t.Fatal(err)
		}

		e, err := NewExerciseStore().Create(ctx, db, "Pull-up", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		reps := 8
		_, err = s.CreateExercise(ctx, db, id.OrgID, p.ID, ps.ID, e.ID, "Pull-up", nil, &reps, nil, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		detail, err := s.Get(ctx, db, id.OrgID, p.ID)

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

// seedTwoOrgs seeds two users/orgs, creates a program owned by org1, and returns a
// store, the program ID, and a committed ScopedQuerier for org2 ready for cross-org attempts.
func seedTwoOrgs(t *testing.T) (s *ProgramStore, programID domain.ProgramID, ctx2 context.Context, q2 Querier) {
	t.Helper()
	db := testutil.NewDB(t)

	u1ID := domain.UserID{UUID: uuid.MustParse("019cb68a-0000-0000-0000-000000000001")}
	o1ID := domain.OrgID{UUID: uuid.MustParse("019cb68a-0000-0000-0000-000000000002")}
	u2ID := domain.UserID{UUID: uuid.MustParse("019cb68a-0000-0000-0000-000000000003")}
	o2ID := domain.OrgID{UUID: uuid.MustParse("019cb68a-0000-0000-0000-000000000004")}

	_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'u1@test.com', 'sub1')`, u1ID)
	_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'u2@test.com', 'sub2')`, u2ID)
	_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'org1')`, o1ID)
	_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'org2')`, o2ID)
	_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, o1ID, u1ID)
	_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, o2ID, u2ID)

	ctx1 := domain.NewContext(context.Background(), &domain.Identity{UserID: u1ID, OrgID: o1ID})
	tx1, err := db.BeginTx(ctx1, nil)
	if err != nil {
		t.Fatal(err)
	}
	ps := NewProgramStore()
	p, err := ps.Create(ctx1, NewScopedQuerier(tx1, o1ID.String(), u1ID.String()), o1ID, "Org1 Program", nil, nil, false)
	if err != nil {
		_ = tx1.Rollback()
		t.Fatal(err)
	}
	if err := tx1.Commit(); err != nil {
		t.Fatal(err)
	}

	ctx2 = domain.NewContext(context.Background(), &domain.Identity{UserID: u2ID, OrgID: o2ID})
	tx2, err := db.BeginTx(ctx2, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = tx2.Rollback() })

	return ps, p.ID, ctx2, NewScopedQuerier(tx2, o2ID.String(), u2ID.String())
}

func TestProgramStore_CreateSet(t *testing.T) {
	t.Run("cannot create set in another org's program", func(t *testing.T) {
		s, programID, ctx2, q2 := seedTwoOrgs(t)
		o2ID := domain.OrgID{UUID: uuid.MustParse("019cb68a-0000-0000-0000-000000000004")}

		_, err := s.CreateSet(ctx2, q2, o2ID, programID, nil, 1, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})

	t.Run("concurrent creates produce distinct IDs", func(t *testing.T) {
		db := testutil.NewDB(t)

		uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")}
		oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")}

		_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'conc@test.com', 'sub-conc')`, uID)
		_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'conc-org')`, oID)
		_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)

		// Create a program outside any concurrent transaction.
		ctx := domain.NewContext(context.Background(), &domain.Identity{UserID: uID, OrgID: oID})
		setupTx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatal(err)
		}
		q := NewScopedQuerier(setupTx, oID.String(), uID.String())
		ps := NewProgramStore()
		p, err := ps.Create(ctx, q, oID, "Concurrent Program", nil, nil, false)
		if err != nil {
			_ = setupTx.Rollback()
			t.Fatal(err)
		}
		if err := setupTx.Commit(); err != nil {
			t.Fatal(err)
		}
		programID := p.ID

		const n = 5
		ids := make([]int64, n)
		errs := make([]error, n)

		var wg sync.WaitGroup
		wg.Add(n)
		for i := 0; i < n; i++ {
			i := i
			go func() {
				defer wg.Done()
				tx, err := db.BeginTx(ctx, nil)
				if err != nil {
					errs[i] = err
					return
				}
				q := NewScopedQuerier(tx, oID.String(), uID.String())
				set, err := ps.CreateSet(ctx, q, oID, programID, nil, 3, nil)
				if err != nil {
					_ = tx.Rollback()
					errs[i] = err
					return
				}
				ids[i] = set.ID
				errs[i] = tx.Commit()
			}()
		}
		wg.Wait()

		for i, err := range errs {
			if err != nil {
				t.Errorf("goroutine %d: %v", i, err)
			}
		}

		seen := make(map[int64]bool)
		for _, id := range ids {
			if seen[id] {
				t.Errorf("duplicate set ID %d produced by concurrent creates", id)
			}
			seen[id] = true
		}
	})
}

func TestProgramStore_UpdateSet(t *testing.T) {
	t.Run("cannot update set in another org's program", func(t *testing.T) {
		s, programID, ctx2, q2 := seedTwoOrgs(t)
		o2ID := domain.OrgID{UUID: uuid.MustParse("019cb68a-0000-0000-0000-000000000004")}

		_, err := s.UpdateSet(ctx2, q2, o2ID, programID, 999, nil, 1, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_DeleteSet(t *testing.T) {
	t.Run("cannot delete set in another org's program", func(t *testing.T) {
		s, programID, ctx2, q2 := seedTwoOrgs(t)
		o2ID := domain.OrgID{UUID: uuid.MustParse("019cb68a-0000-0000-0000-000000000004")}

		err := s.DeleteSet(ctx2, q2, o2ID, programID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s, db, ctx := newTestProgramStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		created, err := s.Create(ctx, db, id.OrgID, "Strength", nil, nil, true)
		if err != nil {
			t.Fatal(err)
		}

		if err := s.Delete(ctx, db, id.OrgID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err = s.GetLite(ctx, db, id.OrgID, created.ID)
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
