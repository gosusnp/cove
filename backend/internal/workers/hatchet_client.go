// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"

	hatchet "github.com/hatchet-dev/hatchet/sdks/go"
)

// HatchetClient wraps the Hatchet workflow engine client, providing job
// enqueueing and run status queries.
type HatchetClient struct {
	client *hatchet.Client
}

// NewHatchetClient creates a Hatchet client and returns a HatchetClient ready
// to enqueue jobs. token is the Hatchet API token (from HATCHET_CLIENT_TOKEN).
func NewHatchetClient(token string) (*HatchetClient, error) {
	// The Hatchet SDK reads HATCHET_CLIENT_TOKEN via viper env binding. The
	// v0 WithToken opt is deprecated, so we set the env var explicitly to
	// ensure the SDK sees the value resolved by getSecret in main.go.
	if err := os.Setenv("HATCHET_CLIENT_TOKEN", token); err != nil {
		return nil, fmt.Errorf("set hatchet token: %w", err)
	}
	client, err := hatchet.NewClient()
	if err != nil {
		return nil, fmt.Errorf("create hatchet client: %w", err)
	}
	return &HatchetClient{client: client}, nil
}

// RequestSummary enqueues a session-summary job and returns the
// Hatchet run ID so the caller can poll for status.
func (h *HatchetClient) RequestSummary(ctx context.Context, sessionID int64, orgID, userID string) (string, error) {
	ref, err := h.client.RunNoWait(ctx, "session-summary", SessionSummaryInput{
		SessionID: sessionID,
		OrgID:     orgID,
		UserID:    userID,
	}, hatchet.WithRunMetadata(map[string]string{
		"session_id": strconv.FormatInt(sessionID, 10),
	}))
	if err != nil {
		return "", fmt.Errorf("dispatch session summary: %w", err)
	}
	return ref.RunId, nil
}

// ErrRunNotFound is returned by GetSessionSummaryStatus when no run with the
// given ID exists in Hatchet.
var ErrRunNotFound = errors.New("run not found")

// ErrRunSessionMismatch is returned by GetSessionSummaryStatus when the run's
// recorded session_id does not match the expected session ID.
var ErrRunSessionMismatch = errors.New("run does not belong to session")

// GetSessionSummaryStatus fetches the status of a session-summary run and
// verifies that the run's input session_id matches expectedSessionID.
// Returns ErrRunNotFound if the run does not exist, ErrRunSessionMismatch if
// the session_id does not match.
func (h *HatchetClient) GetSessionSummaryStatus(ctx context.Context, runID string, expectedSessionID int64) (string, error) {
	details, err := h.client.Runs().Get(ctx, runID)
	if err != nil {
		return "", fmt.Errorf("get run details: %w", err)
	}
	if details == nil {
		return "", ErrRunNotFound
	}
	if len(details.Tasks) == 0 {
		return "", fmt.Errorf("get run details: no tasks in run")
	}
	task := details.Tasks[0]
	sessionIDVal, ok := task.Input["session_id"]
	if !ok {
		return "", fmt.Errorf("get run details: session_id missing from input")
	}
	// JSON numbers unmarshal into interface{} as float64.
	sessionIDFloat, ok := sessionIDVal.(float64)
	if !ok || int64(sessionIDFloat) != expectedSessionID {
		return "", ErrRunSessionMismatch
	}
	return string(task.Status), nil
}
