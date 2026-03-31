// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"testing"

	"github.com/gosusnp/cove/backend/internal/testutil"
)

func TestScopedQuerier_setSession(t *testing.T) {
	t.Run("user path sets org and user session variables", func(t *testing.T) {
		db := testutil.NewDB(t)
		ctx := t.Context()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		q := NewScopedQuerier(tx, "org-123", "user-456")

		var orgID, userID, isService string
		if err := q.QueryRowContext(ctx,
			`SELECT current_setting('app.current_org_id', true),
			        current_setting('app.current_user_id', true),
			        COALESCE(current_setting('app.current_is_service', true), '')`,
		).Scan(&orgID, &userID, &isService); err != nil {
			t.Fatalf("query session vars: %v", err)
		}

		if orgID != "org-123" {
			t.Errorf("current_org_id = %q, want org-123", orgID)
		}
		if userID != "user-456" {
			t.Errorf("current_user_id = %q, want user-456", userID)
		}
		if isService == "true" {
			t.Error("current_is_service should not be set for user path")
		}
	})

	t.Run("service account path sets is_service and user, not org", func(t *testing.T) {
		db := testutil.NewDB(t)
		ctx := t.Context()

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			t.Fatalf("begin tx: %v", err)
		}
		defer func() { _ = tx.Rollback() }()

		q := NewServiceScopedQuerier(tx, "svc-user-789")

		var userID, isService, orgID string
		if err := q.QueryRowContext(ctx,
			`SELECT current_setting('app.current_user_id', true),
			        current_setting('app.current_is_service', true),
			        COALESCE(current_setting('app.current_org_id', true), '')`,
		).Scan(&userID, &isService, &orgID); err != nil {
			t.Fatalf("query session vars: %v", err)
		}

		if userID != "svc-user-789" {
			t.Errorf("current_user_id = %q, want svc-user-789", userID)
		}
		if isService != "true" {
			t.Errorf("current_is_service = %q, want true", isService)
		}
		if orgID != "" {
			t.Errorf("current_org_id should not be set for service path, got %q", orgID)
		}
	})
}
