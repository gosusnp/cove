// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package middleware

import (
	"context"
	"os"
	"testing"

	"github.com/gosusnp/cove/backend/internal/testdb"
)

var containerDSN string

func TestMain(m *testing.M) {
	ctx := context.Background()

	dsn, cleanup, err := testdb.StartContainer(ctx)
	if err != nil {
		panic(err)
	}
	containerDSN = dsn

	code := m.Run()
	cleanup()
	os.Exit(code)
}
