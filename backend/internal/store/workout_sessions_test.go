// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestWorkoutSessionStore(t *testing.T) (*WorkoutSessionStore, Querier, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)

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

	return NewWorkoutSessionStore(), q, ctx
}

// secureBytes returns the encrypted form of p.SensitiveData using the test
// encryptor, bound to the user_id in ctx as GCM additional data.
func secureBytes(t *testing.T, ctx context.Context, p WorkoutSessionParams) []byte {
	t.Helper()
	id, _ := domain.IdentityFromContext(ctx)
	enc := crypto.NewTestEncryptor()
	field := crypto.NewEncryptedField[domain.SessionSensitiveData](enc)
	if err := field.Set(ctx, p.SensitiveData, id.UserID.UUID[:]); err != nil {
		t.Fatalf("encrypt sensitive data: %v", err)
	}
	return field.Value()
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

		_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'u1@test.com', 'sub1')`, uID1)
		_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'u2@test.com', 'sub2')`, uID2)
		_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'test-org')`, oID)
		_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID1)
		_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'member')`, oID, uID2)

		ctx1 := domain.NewContext(context.Background(), &domain.Identity{UserID: uID1, OrgID: oID})
		ctx2 := domain.NewContext(context.Background(), &domain.Identity{UserID: uID2, OrgID: oID})

		tx1, _ := db.BeginTx(ctx1, nil)
		t.Cleanup(func() { _ = tx1.Rollback() })
		q1 := NewScopedQuerier(tx1, oID.String(), uID1.String())

		tx2, _ := db.BeginTx(ctx2, nil)
		t.Cleanup(func() { _ = tx2.Rollback() })
		q2 := NewScopedQuerier(tx2, oID.String(), uID2.String())

		s := NewWorkoutSessionStore()
		p := WorkoutSessionParams{Activity: strPtr("Run")}
		if _, err := s.Create(ctx1, q1, p, secureBytes(t, ctx1, p)); err != nil {
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
		p := WorkoutSessionParams{Activity: strPtr("Swim")}
		created, err := s.Create(ctx, q, p, secureBytes(t, ctx, p))
		if err != nil {
			t.Fatal(err)
		}
		id, _ := domain.IdentityFromContext(ctx)

		got, err := s.Get(ctx, q, id.OrgID, created.ID)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got.Activity == nil || *got.Activity != "Swim" {
			t.Errorf("got activity %v, want %q", got.Activity, "Swim")
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
		effort := 7
		duration := 3600
		notes := "felt great"
		notesSS := crypto.NewSensitiveString(notes)
		p := WorkoutSessionParams{
			Activity:  strPtr("Bike"),
			DurationS: &duration,
			SensitiveData: domain.SessionSensitiveData{
				PerceivedEffort: &effort,
				SessionNotes:    &notesSS,
			},
		}
		ws, err := s.Create(ctx, q, p, secureBytes(t, ctx, p))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.ID == domain.WorkoutSessionID(0) {
			t.Error("expected non-zero ID")
		}
		if ws.Activity == nil || *ws.Activity != "Bike" {
			t.Errorf("got activity %v, want %q", ws.Activity, "Bike")
		}
		if ws.DurationS == nil || *ws.DurationS != duration {
			t.Errorf("got duration %v, want %d", ws.DurationS, duration)
		}
		// Verify the ciphertext was stored and can round-trip.
		ws.SetEncryptor(crypto.NewTestEncryptor())
		if err := ws.UseSensitiveData(ctx, func(private domain.SessionSensitiveData) error {
			if private.PerceivedEffort == nil || *private.PerceivedEffort != effort {
				t.Errorf("got effort %v, want %d", private.PerceivedEffort, effort)
			}
			if private.SessionNotes == nil || private.SessionNotes.String() != notes {
				t.Errorf("got notes %v, want %q", private.SessionNotes, notes)
			}
			return nil
		}); err != nil {
			t.Fatalf("UseSensitiveData: %v", err)
		}
	})

	t.Run("creates with no optional fields", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		p := WorkoutSessionParams{}
		ws, err := s.Create(ctx, q, p, secureBytes(t, ctx, p))
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ws.Activity != nil {
			t.Errorf("expected nil activity, got %v", ws.Activity)
		}
		// Private data should decrypt to empty struct.
		ws.SetEncryptor(crypto.NewTestEncryptor())
		if err := ws.UseSensitiveData(ctx, func(private domain.SessionSensitiveData) error {
			if private.ProgramStructure != nil {
				t.Errorf("expected nil program_structure, got %v", private.ProgramStructure)
			}
			return nil
		}); err != nil {
			t.Fatalf("UseSensitiveData: %v", err)
		}
	})
}

func TestWorkoutSessionStore_Update(t *testing.T) {
	t.Run("updates fields", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		p0 := WorkoutSessionParams{Activity: strPtr("Run")}
		created, err := s.Create(ctx, q, p0, secureBytes(t, ctx, p0))
		if err != nil {
			t.Fatal(err)
		}

		effort := 8
		p1 := WorkoutSessionParams{
			Activity:      strPtr("Swim"),
			SensitiveData: domain.SessionSensitiveData{PerceivedEffort: &effort},
		}
		updated, err := s.Update(ctx, q, id.OrgID, created.ID, p1, secureBytes(t, ctx, p1), false, &created.UpdatedAt)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if updated.Activity == nil || *updated.Activity != "Swim" {
			t.Errorf("got activity %v, want %q", updated.Activity, "Swim")
		}
		updated.SetEncryptor(crypto.NewTestEncryptor())
		if err := updated.UseSensitiveData(ctx, func(private domain.SessionSensitiveData) error {
			if private.PerceivedEffort == nil || *private.PerceivedEffort != effort {
				t.Errorf("got effort %v, want %d", private.PerceivedEffort, effort)
			}
			return nil
		}); err != nil {
			t.Fatalf("UseSensitiveData: %v", err)
		}
	})

	t.Run("not found", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		p := WorkoutSessionParams{}
		_, err := s.Update(ctx, q, id.OrgID, domain.WorkoutSessionID(999), p, secureBytes(t, ctx, p), false, nil)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("got %v, want ErrNotFound", err)
		}
	})
}

func TestWorkoutSessionStore_Delete(t *testing.T) {
	t.Run("deletes existing", func(t *testing.T) {
		s, q, ctx := newTestWorkoutSessionStore(t)
		id, _ := domain.IdentityFromContext(ctx)
		p := WorkoutSessionParams{}
		created, err := s.Create(ctx, q, p, secureBytes(t, ctx, p))
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

// TestWorkoutSessionStore_ServiceAccountRLS verifies that migration 023 exposes exactly
// the right surface to service accounts: SELECT and UPDATE pass; INSERT and DELETE are blocked.
func TestWorkoutSessionStore_ServiceAccountRLS(t *testing.T) {
	db := testutil.NewDB(t)

	uID := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")}
	oID := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")}
	svcUID := domain.NewUserID()

	_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'rls@test.com', 'sub-rls')`, uID)
	_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'rls-org')`, oID)
	_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID, uID)
	_, _ = db.Exec(`INSERT INTO cove.users (id, is_service_account) VALUES ($1, true)`, svcUID)

	// Seed a session as the regular user (committed so service account tx can see it).
	userCtx := domain.NewContext(context.Background(), &domain.Identity{UserID: uID, OrgID: oID})
	seedTx, _ := db.BeginTx(userCtx, nil)
	seedQ := NewScopedQuerier(seedTx, oID.String(), uID.String())
	s := NewWorkoutSessionStore()
	p := WorkoutSessionParams{Activity: strPtr("Run")}
	created, err := s.Create(userCtx, seedQ, p, secureBytes(t, userCtx, p))
	if err != nil {
		t.Fatalf("seed session: %v", err)
	}
	if err := seedTx.Commit(); err != nil {
		t.Fatalf("seed commit: %v", err)
	}

	svcCtx := domain.NewContext(context.Background(), &domain.Identity{
		UserID:         svcUID,
		ServiceAccount: true,
	})

	newSvcTx := func(t *testing.T) Querier {
		t.Helper()
		tx, err := db.BeginTx(svcCtx, nil)
		if err != nil {
			t.Fatalf("begin svc tx: %v", err)
		}
		t.Cleanup(func() { _ = tx.Rollback() })
		return NewServiceScopedQuerier(tx, svcUID.String())
	}

	t.Run("service account can SELECT workout sessions", func(t *testing.T) {
		q := newSvcTx(t)
		got, err := s.Get(svcCtx, q, oID, created.ID)
		if err != nil {
			t.Fatalf("expected SELECT to succeed, got: %v", err)
		}
		if got.ID != created.ID {
			t.Errorf("got ID %v, want %v", got.ID, created.ID)
		}
	})

	t.Run("service account can UPDATE workout sessions", func(t *testing.T) {
		q := newSvcTx(t)
		p := WorkoutSessionParams{Activity: strPtr("Bike")}
		updated, err := s.Update(svcCtx, q, oID, created.ID, p, nil, true, nil)
		if err != nil {
			t.Fatalf("expected UPDATE to succeed, got: %v", err)
		}
		if updated.Activity == nil || *updated.Activity != "Bike" {
			t.Errorf("got activity %v, want Bike", updated.Activity)
		}
	})

	t.Run("service account cannot INSERT workout sessions", func(t *testing.T) {
		q := newSvcTx(t)
		// Use a real (empty) sensitive-data blob so the failure is unambiguously
		// RLS, not a null-dereference or encoding error.
		p := WorkoutSessionParams{Activity: strPtr("Swim")}
		if _, err := s.Create(svcCtx, q, p, secureBytes(t, userCtx, p)); err == nil {
			t.Error("expected INSERT to be blocked for service account, got nil error")
		}
	})

	t.Run("service account cannot DELETE workout sessions", func(t *testing.T) {
		q := newSvcTx(t)
		// RLS makes the row invisible to the DELETE rather than returning a
		// permission error, so RowsAffected == 0 → ErrNotFound. This is the
		// expected RLS-via-invisibility behaviour for unextended USING clauses.
		if err := s.Delete(svcCtx, q, oID, created.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected DELETE to be invisible via RLS (ErrNotFound), got: %v", err)
		}
	})

	t.Run("service account is constrained by explicit org_id parameter", func(t *testing.T) {
		// Seed a second org with its own session.
		uID2 := domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe03")}
		oID2 := domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe04")}
		_, _ = db.Exec(`INSERT INTO cove.users (id, email, google_sub) VALUES ($1, 'other@test.com', 'sub-other')`, uID2)
		_, _ = db.Exec(`INSERT INTO cove.orgs (id, name) VALUES ($1, 'other-org')`, oID2)
		_, _ = db.Exec(`INSERT INTO cove.org_members (org_id, user_id, role) VALUES ($1, $2, 'admin')`, oID2, uID2)

		userCtx2 := domain.NewContext(context.Background(), &domain.Identity{UserID: uID2, OrgID: oID2})
		seedTx2, _ := db.BeginTx(userCtx2, nil)
		seedQ2 := NewScopedQuerier(seedTx2, oID2.String(), uID2.String())
		p2 := WorkoutSessionParams{Activity: strPtr("Yoga")}
		session2, err := s.Create(userCtx2, seedQ2, p2, secureBytes(t, userCtx2, p2))
		if err != nil {
			t.Fatalf("seed org2 session: %v", err)
		}
		if err := seedTx2.Commit(); err != nil {
			t.Fatalf("seed org2 commit: %v", err)
		}

		q := newSvcTx(t)

		// Service account can read the session when given the correct org_id.
		if _, err := s.Get(svcCtx, q, oID2, session2.ID); err != nil {
			t.Fatalf("expected service account to read session with correct org_id, got: %v", err)
		}

		// Service account cannot read across orgs: passing the wrong org_id returns not found.
		if _, err := s.Get(svcCtx, q, oID, session2.ID); !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound when org_id does not match, got: %v", err)
		}
	})
}

func strPtr(s string) *string { return &s }
