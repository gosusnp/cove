// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"database/sql"
	"testing"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	return testutil.NewDB(t, containerDSN, db.MigrationsFS)
}
