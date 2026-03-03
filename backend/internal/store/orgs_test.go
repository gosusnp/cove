// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
)

func setupTestOrgStore(t *testing.T) (context.Context, *sql.DB, *OrgStore) {
	t.Helper()
	return t.Context(), newTestDB(t), NewOrgStore()
}

func TestOrgStore_CreateOrg(t *testing.T) {

	t.Run("creates org successfully", func(t *testing.T) {
		ctx, db, os := setupTestOrgStore(t)

		id := domain.NewOrgID()
		if err := os.CreateOrg(ctx, db, id, "test@example.com"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var count int
		if err := db.QueryRowContext(ctx, `SELECT count(*) FROM orgs WHERE id = $1`, id).Scan(&count); err != nil {
			t.Fatalf("query: %v", err)
		}
		if count != 1 {
			t.Errorf("expected 1 org, got %d", count)
		}
	})
}

func TestOrgStore_CreateOrgMember(t *testing.T) {
	t.Run("creates membership successfully", func(t *testing.T) {
		ctx, db, os := setupTestOrgStore(t)
		us := NewUserStore()

		userID := domain.NewUserID()
		user, _, err := us.UpsertUser(ctx, db, userID, "member@example.com", "sub-member")
		if err != nil {
			t.Fatalf("UpsertUser: %v", err)
		}

		orgID := domain.NewOrgID()
		if err := os.CreateOrg(ctx, db, orgID, "member@example.com"); err != nil {
			t.Fatalf("CreateOrg: %v", err)
		}
		// TODO remove userID wrapping
		if err := os.CreateOrgMember(ctx, db, orgID, domain.UserID(user.ID), "owner"); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		var role string
		if err := db.QueryRowContext(ctx, `SELECT role FROM org_members WHERE org_id = $1 AND user_id = $2`, orgID, user.ID).Scan(&role); err != nil {
			t.Fatalf("query: %v", err)
		}
		if role != "owner" {
			t.Errorf("got role %q, want %q", role, "owner")
		}
	})
}
