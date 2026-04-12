// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/google/uuid"

	"github.com/hatchet-dev/hatchet/pkg/client/types"
	hatchet "github.com/hatchet-dev/hatchet/sdks/go"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// heartbeatInterval is how often the task refreshes its Hatchet execution
// timeout while waiting for the LLM. Must be shorter than heartbeatTimeout.
const heartbeatInterval = 20 * time.Second

// heartbeatTimeout is the execution timeout set on the task and the value
// passed to each RefreshTimeout call. Hatchet will mark the task as failed if
// no refresh arrives within this window, which bounds how long a crashed worker
// goes undetected.
const heartbeatTimeout = 30 * time.Second

// timeoutRefresher is the subset of hatchet.Context needed by startHeartbeat.
type timeoutRefresher interface {
	RefreshTimeout(incrementTimeoutBy string) error
}

// startHeartbeat calls r.RefreshTimeout every interval until ctx is cancelled
// or the returned stop function is called. name is used in log lines.
// The returned done channel is closed when the goroutine exits, which allows
// callers to wait for a clean shutdown.
func startHeartbeat(ctx context.Context, r timeoutRefresher, interval, timeout time.Duration, name string) (stop func(), done <-chan struct{}) {
	heartbeatCtx, cancel := context.WithCancel(ctx)
	ch := make(chan struct{})
	go func() {
		defer close(ch)
		t := time.NewTicker(interval)
		defer t.Stop()
		for {
			select {
			case <-heartbeatCtx.Done():
				return
			case <-t.C:
				if err := r.RefreshTimeout(timeout.String()); err != nil {
					log.Printf("%s heartbeat: refresh timeout: %v", name, err)
				}
			}
		}
	}()
	return cancel, ch
}

// SessionSummaryInput is the payload dispatched to the session-summary workflow.
// OrgID and UserID are required to build the identity context for RLS-gated
// service calls. SessionID is a string so it remains opaque to the workflow
// engine and the concurrency key expression requires no type cast.
type SessionSummaryInput struct {
	SessionID string `json:"session_id"`
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
			sessionID, err := strconv.ParseInt(input.SessionID, 10, 64)
			if err != nil {
				return nil, fmt.Errorf("parse session_id: %w", err)
			}
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

			// The LLM call may take several minutes. Keep the task alive by
			// refreshing the Hatchet execution timeout on each tick. If this
			// process crashes the heartbeat stops and Hatchet marks the task
			// failed within heartbeatTimeout.
			stop, _ := startHeartbeat(idCtx, ctx, heartbeatInterval, heartbeatTimeout, "session-summary")
			defer stop()

			return nil, w.Run(idCtx, domain.WorkoutSessionID(sessionID))
		},
		hatchet.WithWorkflowConcurrency(types.Concurrency{
			Expression:    "input.session_id",
			MaxRuns:       &maxRuns,
			LimitStrategy: &strategy,
		}),
		hatchet.WithExecutionTimeout(heartbeatTimeout),
	)
}
