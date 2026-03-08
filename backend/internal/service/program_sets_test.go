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

func newTestProgramSetService(t *testing.T) (*ProgramSetService, domain.ProgramID) {
	t.Helper()
	db := testutil.NewDB(t)

	// Seed user and org for required fields
	uID := uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")
	oID := uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")
	_, _ = db.Exec(`INSERT INTO users (id, email, google_sub) VALUES ($1, 'test@test.com', 'sub') ON CONFLICT DO NOTHING`, uID)
	_, _ = db.Exec(`INSERT INTO orgs (id, name) VALUES ($1, 'test-org') ON CONFLICT DO NOTHING`, oID)

	// Set session variables for RLS
	_, _ = db.Exec(`SELECT set_config('app.current_org_id', $1, false), set_config('app.current_user_id', $2, false)`, oID.String(), uID.String())

	var pID int64
	err := db.QueryRowContext(context.Background(), `INSERT INTO programs (name, org_id, created_by) VALUES ($1, $2, $3) RETURNING id`, "Test Program", oID, uID).Scan(&pID)
	if err != nil {
		t.Fatal(err)
	}
	return NewProgramSetService(store.NewProgramSetStore(db)), domain.ProgramID(pID)
}

func TestProgramSetService_List(t *testing.T) {
	t.Run("returns all sets for program", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		_, _ = svc.Create(programID, nil, 1, nil, nil)

		list, err := svc.List(programID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(list) != 1 {
			t.Errorf("expected 1 set, got %d", len(list))
		}
	})
}

func TestProgramSetService_Get(t *testing.T) {
	t.Run("found", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		created, _ := svc.Create(programID, nil, 3, nil, nil)

		got, err := svc.Get(programID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Rounds != 3 {
			t.Errorf("got %d rounds, want 3", got.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		_, err := svc.Get(programID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramSetService_Create(t *testing.T) {
	t.Run("rounds below 1 defaults to 1", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)

		ps, err := svc.Create(programID, nil, 0, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", ps.Rounds)
		}
	})

	t.Run("negative rounds defaults to 1", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)

		ps, err := svc.Create(programID, nil, -5, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", ps.Rounds)
		}
	})

	t.Run("valid rounds are preserved", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)

		ps, err := svc.Create(programID, nil, 4, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ps.Rounds != 4 {
			t.Errorf("got rounds %d, want 4", ps.Rounds)
		}
	})
}

func TestProgramSetService_Update(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		created, _ := svc.Create(programID, nil, 1, nil, nil)

		name := "Warmup"
		updated, err := svc.Update(programID, created.ID, &name, 2, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Name == nil || *updated.Name != "Warmup" {
			t.Errorf("got %v, want Warmup", updated.Name)
		}
	})

	t.Run("rounds below 1 defaults to 1", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		created, err := svc.Create(programID, nil, 3, nil, nil)
		if err != nil {
			t.Fatal(err)
		}

		updated, err := svc.Update(programID, created.ID, nil, 0, nil, nil)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Rounds != 1 {
			t.Errorf("got rounds %d, want 1", updated.Rounds)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		_, err := svc.Update(programID, 999, nil, 1, nil, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestProgramSetService_Delete(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		created, _ := svc.Create(programID, nil, 1, nil, nil)

		if err := svc.Delete(programID, created.ID); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		_, err := svc.Get(programID, created.ID)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		svc, programID := newTestProgramSetService(t)
		err := svc.Delete(programID, 999)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}
