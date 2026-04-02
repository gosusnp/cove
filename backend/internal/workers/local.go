// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"context"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
)

// LocalWorkoutSessionAdapter satisfies WorkoutSessionPort by delegating
// directly to WorkoutSessionService. Used in the fat-binary deployment where
// the worker and API server share the same process and database connection pool.
type LocalWorkoutSessionAdapter struct {
	svc *service.WorkoutSessionService
}

// NewLocalWorkoutSessionAdapter returns a WorkoutSessionPort backed by svc.
func NewLocalWorkoutSessionAdapter(svc *service.WorkoutSessionService) *LocalWorkoutSessionAdapter {
	return &LocalWorkoutSessionAdapter{svc: svc}
}

func (a *LocalWorkoutSessionAdapter) Get(ctx context.Context, id domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	return a.svc.Get(ctx, id)
}

func (a *LocalWorkoutSessionAdapter) PatchSummary(ctx context.Context, id domain.WorkoutSessionID, patch WorkoutSessionSummaryPatch) error {
	_, err := a.svc.Patch(ctx, id, service.WorkoutSessionPatch{
		UpdatedAt: &patch.UpdatedAt,
		Summary:   domain.Optional[*string]{Set: true, Value: &patch.Summary},
	})
	return err
}
