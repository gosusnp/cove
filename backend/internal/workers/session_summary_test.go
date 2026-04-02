// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package workers_test

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/llm"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/workers"
)

// ─── Helpers ──────────────────────────────────────────────────────────────────

// newTestSession returns a WorkoutSession with a known UpdatedAt and a real
// encryptor attached. Sensitive data is nil (no ciphertext), so UseSensitiveData
// calls its callback with a zero-value SessionSensitiveData — sufficient for
// testing orchestration wiring without requiring a real database.
func newTestSession(updatedAt time.Time) *domain.WorkoutSession {
	ws := &domain.WorkoutSession{
		ID:        domain.WorkoutSessionID(1),
		UserID:    domain.UserID{UUID: uuid.MustParse("019cb68a-cfcb-76db-9003-87bbcaaebe01")},
		OrgID:     domain.OrgID{UUID: uuid.MustParse("019cb68a-cfce-7aa3-bdfb-9700ccaebe02")},
		UpdatedAt: updatedAt,
	}
	ws.SetEncryptor(crypto.NewTestEncryptor())
	return ws
}

// ─── Mocks ────────────────────────────────────────────────────────────────────

type mockPort struct {
	ws       *domain.WorkoutSession
	getErr   error
	patchErr error
	patches  []workers.WorkoutSessionSummaryPatch
}

func (m *mockPort) Get(_ context.Context, _ domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	return m.ws, m.getErr
}

func (m *mockPort) PatchSummary(_ context.Context, _ domain.WorkoutSessionID, p workers.WorkoutSessionSummaryPatch) error {
	m.patches = append(m.patches, p)
	return m.patchErr
}

type fakeLLM struct {
	resp string
	err  error
}

func (f *fakeLLM) Complete(_ context.Context, _ llm.CompletionRequest) (string, error) {
	return f.resp, f.err
}

// ─── Tests ────────────────────────────────────────────────────────────────────

func TestSessionSummaryWorker_Run_WiresSummaryAndUpdatedAt(t *testing.T) {
	updatedAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	ws := newTestSession(updatedAt)

	port := &mockPort{ws: ws}
	worker := workers.NewSessionSummaryWorker(port, service.NewSummarizeService(&fakeLLM{resp: "great session"}))

	if err := worker.Run(context.Background(), ws.ID); err != nil {
		t.Fatalf("Run: %v", err)
	}

	if len(port.patches) != 1 {
		t.Fatalf("want 1 PatchSummary call, got %d", len(port.patches))
	}
	got := port.patches[0]
	if got.Summary != "great session" {
		t.Errorf("Summary = %q, want %q", got.Summary, "great session")
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, updatedAt)
	}
}

func TestSessionSummaryWorker_Run_PropagatesGetError(t *testing.T) {
	port := &mockPort{getErr: errors.New("db unavailable")}
	worker := workers.NewSessionSummaryWorker(port, service.NewSummarizeService(&fakeLLM{}))

	if err := worker.Run(context.Background(), 1); err == nil {
		t.Fatal("want error from Get, got nil")
	}
	if len(port.patches) > 0 {
		t.Error("PatchSummary must not be called when Get fails")
	}
}

func TestSessionSummaryWorker_Run_PropagatesSummarizeError(t *testing.T) {
	ws := newTestSession(time.Now())
	port := &mockPort{ws: ws}
	worker := workers.NewSessionSummaryWorker(port, service.NewSummarizeService(&fakeLLM{err: errors.New("rate limited")}))

	if err := worker.Run(context.Background(), ws.ID); err == nil {
		t.Fatal("want error from SummarizeSession, got nil")
	}
	if len(port.patches) > 0 {
		t.Error("PatchSummary must not be called when summarization fails")
	}
}

func TestSessionSummaryWorker_Run_PropagatesConflict(t *testing.T) {
	ws := newTestSession(time.Now())
	port := &mockPort{ws: ws, patchErr: service.ErrConflict}
	worker := workers.NewSessionSummaryWorker(port, service.NewSummarizeService(&fakeLLM{resp: "ok"}))

	err := worker.Run(context.Background(), ws.ID)
	if !errors.Is(err, service.ErrConflict) {
		t.Errorf("want ErrConflict, got %v", err)
	}
}

func TestSessionSummaryWorker_Preview_DoesNotPatch(t *testing.T) {
	ws := newTestSession(time.Now())
	port := &mockPort{ws: ws}
	// nil LLM client — preview-only mode, no LLM call is made
	worker := workers.NewSessionSummaryWorker(port, service.NewSummarizeService(nil))

	if err := worker.Preview(context.Background(), ws.ID, io.Discard); err != nil {
		t.Fatalf("Preview: %v", err)
	}
	if len(port.patches) > 0 {
		t.Error("Preview must not call PatchSummary")
	}
}
