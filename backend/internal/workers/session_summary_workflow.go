// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"fmt"

	"github.com/google/uuid"

	"github.com/hatchet-dev/hatchet/pkg/client/types"
	hatchet "github.com/hatchet-dev/hatchet/sdks/go"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// SessionSummaryInput is the payload dispatched to the session-summary workflow.
// OrgID and UserID are required to build the identity context for RLS-gated
// service calls.
type SessionSummaryInput struct {
	SessionID int64  `json:"session_id"`
	OrgID     string `json:"org_id"`
	UserID    string `json:"user_id"`
}

// newSessionSummaryTask returns a Hatchet StandaloneTask that wraps
// SessionSummaryWorker.Run. Concurrency is keyed on session_id with
// CANCEL_NEWEST so duplicate triggers for the same session drop the older run
// and keep the freshest request in flight.
func newSessionSummaryTask(client *hatchet.Client, w *SessionSummaryWorker) *hatchet.StandaloneTask {
	strategy := types.CancelNewest
	var maxRuns int32 = 1

	return client.NewStandaloneTask(
		"session-summary",
		func(ctx hatchet.Context, input SessionSummaryInput) (any, error) {
			orgID, err := uuid.Parse(input.OrgID)
			if err != nil {
				return nil, fmt.Errorf("parse org_id: %w", err)
			}
			userID, err := uuid.Parse(input.UserID)
			if err != nil {
				return nil, fmt.Errorf("parse user_id: %w", err)
			}

			identity := &domain.Identity{
				OrgID:  domain.OrgID{UUID: orgID},
				UserID: domain.UserID{UUID: userID},
			}
			idCtx := domain.NewContext(ctx.GetContext(), identity)

			return nil, w.Run(idCtx, domain.WorkoutSessionID(input.SessionID))
		},
		hatchet.WithWorkflowConcurrency(types.Concurrency{
			Expression:    "input.session_id",
			MaxRuns:       &maxRuns,
			LimitStrategy: &strategy,
		}),
	)
}
