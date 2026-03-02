// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"testing"

	"github.com/google/uuid"
)

func newTestOrgStore(t *testing.T) (*OrgStore, *UserStore) {
	t.Helper()
	db := newTestDB(t)
	return NewOrgStore(db), NewUserStore(db)
}

func TestOrgStore_CreateOrg(t *testing.T) {
	t.Run("creates org successfully", func(t *testing.T) {
		os, _ := newTestOrgStore(t)

		id, _ := uuid.NewV7()
		if err := os.CreateOrg(id, "test@example.com"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var count int
		if err := os.db.QueryRow(`SELECT count(*) FROM orgs WHERE id = $1`, id).Scan(&count); err != nil {
			t.Fatalf("query: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 org, got %d", count)
		}
	})
}

func TestOrgStore_CreateOrgMember(t *testing.T) {
	t.Run("creates membership successfully", func(t *testing.T) {
		os, us := newTestOrgStore(t)

		userID, _ := uuid.NewV7()
		user, _, err := us.UpsertUser(userID, "member@example.com", "sub-member")
		if err != nil {
			t.Fatalf("UpsertUser: %v", err)
		}

		orgID, _ := uuid.NewV7()
		if err := os.CreateOrg(orgID, "member@example.com"); err != nil {
			t.Fatalf("CreateOrg: %v", err)
		}
		if err := os.CreateOrgMember(orgID, user.ID, "owner"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var role string
		if err := os.db.QueryRow(
			`SELECT role FROM org_members WHERE org_id = $1 AND user_id = $2`,
			orgID, user.ID,
		).Scan(&role); err != nil {
			t.Fatalf("query: %v", err)
		}
		if role != "owner" {
			t.Errorf("got role %q, want %q", role, "owner")
		}
	})
}

func TestOrgStore_WithTx(t *testing.T) {
	t.Run("rolled back transaction leaves no data", func(t *testing.T) {
		db := newTestDB(t)
		os := NewOrgStore(db)

		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}

		id, _ := uuid.NewV7()
		if err := os.WithTx(tx).CreateOrg(id, "rollback@example.com"); err != nil {
			t.Fatalf("CreateOrg: %v", err)
		}

		if err := tx.Rollback(); err != nil {
			t.Fatalf("rollback: %v", err)
		}

		var count int
		if err := db.QueryRow(`SELECT count(*) FROM orgs WHERE id = $1`, id).Scan(&count); err != nil {
			t.Fatalf("query: %v", err)
		}
		if count != 0 {
			t.Errorf("expected 0 orgs after rollback, got %d", count)
		}
	})

	t.Run("committed transaction persists data", func(t *testing.T) {
		db := newTestDB(t)
		os := NewOrgStore(db)

		tx, err := db.Begin()
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}

		id, _ := uuid.NewV7()
		if err := os.WithTx(tx).CreateOrg(id, "commit@example.com"); err != nil {
			t.Fatalf("CreateOrg: %v", err)
		}

		if err := tx.Commit(); err != nil {
			t.Fatalf("commit: %v", err)
		}

		var count int
		if err := db.QueryRow(`SELECT count(*) FROM orgs WHERE id = $1`, id).Scan(&count); err != nil {
			t.Fatalf("query: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 org after commit, got %d", count)
		}
	})
}
