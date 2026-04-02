// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"context"
	"fmt"
	"log"

	hatchet "github.com/hatchet-dev/hatchet/sdks/go"

	"github.com/gosusnp/cove/backend/internal/service"
)

// StartWorker creates a Hatchet client, registers the session-summary workflow,
// and starts the worker in a background goroutine. It returns an error if
// initialization fails; once started the worker runs until ctx is cancelled.
//
// HATCHET_CLIENT_TOKEN must be set in the environment before calling. Callers
// should validate its presence before invoking this function.
func StartWorker(ctx context.Context, sessions WorkoutSessionPort, summarize *service.SummarizeService) error {
	client, err := hatchet.NewClient()
	if err != nil {
		return fmt.Errorf("create hatchet client: %w", err)
	}

	summaryWorker := NewSessionSummaryWorker(sessions, summarize)
	task := newSessionSummaryTask(client, summaryWorker)

	w, err := client.NewWorker("cove-worker", hatchet.WithWorkflows(task))
	if err != nil {
		return fmt.Errorf("create hatchet worker: %w", err)
	}

	go func() {
		if err := w.StartBlocking(ctx); err != nil {
			log.Printf("hatchet worker stopped: %v", err)
		}
	}()

	return nil
}
