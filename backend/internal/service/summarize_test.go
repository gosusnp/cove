// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/llm"
)

// mockLLMClient is a test double for llm.Client.
type mockLLMClient struct {
	response string
	err      error
}

func (m *mockLLMClient) Complete(_ context.Context, _ llm.CompletionRequest) (string, error) {
	return m.response, m.err
}

func newTestWorkoutSession(activity string) *domain.WorkoutSession {
	return &domain.WorkoutSession{Activity: &activity}
}

func TestSummarizeService_SummarizeSession_nilClient(t *testing.T) {
	svc := NewSummarizeService(nil)
	_, err := svc.SummarizeSession(context.Background(), newTestWorkoutSession("running"), domain.SessionSensitiveData{})
	if err == nil {
		t.Fatal("expected error for nil client, got nil")
	}
}

func TestSummarizeService_SummarizeSession_success(t *testing.T) {
	svc := NewSummarizeService(&mockLLMClient{response: "**Overview** Solid run."})
	result, err := svc.SummarizeSession(context.Background(), newTestWorkoutSession("running"), domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Summary != "**Overview** Solid run." {
		t.Errorf("Summary = %q, want %q", result.Summary, "**Overview** Solid run.")
	}
}

func TestSummarizeService_SummarizeSession_clientError(t *testing.T) {
	svc := NewSummarizeService(&mockLLMClient{err: errors.New("timeout")})
	_, err := svc.SummarizeSession(context.Background(), newTestWorkoutSession("running"), domain.SessionSensitiveData{})
	if err == nil {
		t.Fatal("expected error from client, got nil")
	}
	if !strings.Contains(err.Error(), "timeout") {
		t.Errorf("error = %q, want it to contain %q", err.Error(), "timeout")
	}
}

func TestSummarizeService_PreviewSession_nilClient(t *testing.T) {
	svc := NewSummarizeService(nil)
	var buf bytes.Buffer
	if err := svc.PreviewSession(&buf, newTestWorkoutSession("running"), domain.SessionSensitiveData{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if buf.Len() == 0 {
		t.Error("expected non-empty output from PreviewSession")
	}
}

func TestSummarizeService_PreviewSession_writesSystemAndUser(t *testing.T) {
	svc := NewSummarizeService(nil)
	var buf bytes.Buffer
	if err := svc.PreviewSession(&buf, newTestWorkoutSession("cycling"), domain.SessionSensitiveData{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	out := buf.String()
	for _, want := range []string{"[SYSTEM]", "[USER]", "cycling"} {
		if !strings.Contains(out, want) {
			t.Errorf("preview output missing %q:\n%s", want, out)
		}
	}
}
