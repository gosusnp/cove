// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"context"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// WorkoutSessionSummaryPatch carries the fields the session-summary worker
// writes back after generating a summary. It is intentionally narrow — only
// the fields the worker ever sets — so a future remote adapter can translate
// directly to a single API call without leaking service internals.
//
// UpdatedAt is always required and used as an optimistic lock. Worker patches
// must never overwrite a session that was modified after it was fetched.
type WorkoutSessionSummaryPatch struct {
	Summary   string
	UpdatedAt time.Time
}

// WorkoutSessionPort is the data-tier boundary for workers that operate on
// workout sessions. It is the seam between worker orchestration logic and the
// persistence layer.
//
// In the current fat-binary deployment this interface is satisfied by
// LocalWorkoutSessionAdapter, which delegates directly to
// WorkoutSessionService. When the worker is split into a separate process the
// same interface is satisfied by an HTTP client, with no changes to the
// workflow code.
//
// Identity must be present in ctx (via domain.NewContext) before calling any
// method — the same requirement as the service layer.
type WorkoutSessionPort interface {
	// Get fetches a single workout session. The returned *WorkoutSession has
	// its encryptor attached; call UseSensitiveData to access sensitive fields.
	Get(ctx context.Context, id domain.WorkoutSessionID) (*domain.WorkoutSession, error)

	// PatchSummary writes a generated summary back to the session using an
	// optimistic lock. Returns service.ErrConflict if the session was modified
	// after it was fetched; callers should retry the full workflow in that case.
	PatchSummary(ctx context.Context, id domain.WorkoutSessionID, patch WorkoutSessionSummaryPatch) error
}
