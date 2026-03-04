// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestOrgStore(t *testing.T) (context.Context, *sql.DB, *OrgStore) {
	t.Helper()
	return t.Context(), testutil.NewDB(t), NewOrgStore()
}

func TestOrgStore_CreateOrg(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, db, s := newTestOrgStore(t)
		id := domain.NewOrgID()

		err := s.CreateOrg(ctx, db, id, "test@example.com")
		if err != nil {
			t.Fatalf("CreateOrg: %v", err)
		}

		// Verify it was created
		var count int
		err = db.QueryRow("SELECT count(*) FROM orgs WHERE id = $1", id).Scan(&count)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 org, got %d", count)
		}
	})
}

func TestOrgStore_CreateOrgMember(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		ctx, db, s := newTestOrgStore(t)
		orgID := domain.NewOrgID()
		userID := domain.NewUserID()

		// Setup org first
		if err := s.CreateOrg(ctx, db, orgID, "test@example.com"); err != nil {
			t.Fatal(err)
		}

		// Setup user first (required by FK)
		_, _, err := NewUserStore().UpsertUser(ctx, db, userID, "test@example.com", "sub")
		if err != nil {
			t.Fatal(err)
		}

		err = s.CreateOrgMember(ctx, db, orgID, userID, "owner")
		if err != nil {
			t.Fatalf("CreateOrgMember: %v", err)
		}

		// Verify it was created
		var count int
		err = db.QueryRow("SELECT count(*) FROM org_members WHERE org_id = $1 AND user_id = $2", orgID, userID).Scan(&count)
		if err != nil {
			t.Fatalf("query: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 member, got %d", count)
		}
	})
}
