// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"testing"

	"github.com/gosusnp/cove/backend/internal/testutil"
)

var containerDSN string

func TestMain(m *testing.M) {
	testutil.RunMain(m, &containerDSN)
}
