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

func strPtr(s string) *string { return &s }
