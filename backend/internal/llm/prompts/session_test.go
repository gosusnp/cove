// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package prompts

import (
	"strings"
	"testing"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
)

func sensitivePtr(s string) *crypto.SensitiveString {
	ss := crypto.NewSensitiveString(s)
	return &ss
}

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }

func newTestSession(activity string) *domain.WorkoutSession {
	dur := 3600
	t := time.Date(2026, 3, 24, 9, 0, 0, 0, time.UTC)
	return &domain.WorkoutSession{
		Activity:  strPtr(activity),
		DurationS: &dur,
		StartedAt: &t,
	}
}

func TestSessionSummary_messageStructure(t *testing.T) {
	req, err := SessionSummary(newTestSession("running"), domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(req.Messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(req.Messages))
	}
	if req.Messages[0].Role != "system" {
		t.Errorf("messages[0].Role = %q, want system", req.Messages[0].Role)
	}
	if req.Messages[1].Role != "user" {
		t.Errorf("messages[1].Role = %q, want user", req.Messages[1].Role)
	}
}

func TestSessionSummary_systemPrompt(t *testing.T) {
	req, err := SessionSummary(newTestSession("running"), domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(req.Messages[0].Content, "strength and conditioning coach") {
		t.Errorf("system prompt does not contain expected content: %q", req.Messages[0].Content)
	}
}

func TestSessionSummary_publicFields(t *testing.T) {
	req, err := SessionSummary(newTestSession("cycling"), domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	user := req.Messages[1].Content
	if !strings.Contains(user, "cycling") {
		t.Errorf("user prompt missing activity: %q", user)
	}
	if !strings.Contains(user, "1h 0m") {
		t.Errorf("user prompt missing duration: %q", user)
	}
	if !strings.Contains(user, "Tue Mar 24, 2026") {
		t.Errorf("user prompt missing date: %q", user)
	}
}

func TestSessionSummary_sensitiveFieldsRendered(t *testing.T) {
	sd := domain.SessionSensitiveData{
		PerceivedEffort:  intPtr(8),
		SessionNotes:     sensitivePtr("felt strong"),
		ProgramName:      sensitivePtr("Strength Block"),
		ProgramStructure: sensitivePtr("5x5 squat, bench, deadlift"),
	}
	req, err := SessionSummary(newTestSession("lifting"), sd)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	user := req.Messages[1].Content
	for _, want := range []string{"8", "felt strong", "Strength Block", "5x5 squat"} {
		if !strings.Contains(user, want) {
			t.Errorf("user prompt missing %q:\n%s", want, user)
		}
	}
}

func TestSessionSummary_nilSensitiveFieldsOmitted(t *testing.T) {
	req, err := SessionSummary(newTestSession("running"), domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	user := req.Messages[1].Content
	// "RPE:" is the data line; "RPE" alone also appears in the format instructions so we check the colon form.
	for _, absent := range []string{"RPE:", "User Notes", "Program name", "Planned Structure"} {
		if strings.Contains(user, absent) {
			t.Errorf("user prompt should not contain %q when field is nil:\n%s", absent, user)
		}
	}
}

func TestSessionSummary_temperature(t *testing.T) {
	req, err := SessionSummary(newTestSession("running"), domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if req.Temperature == nil {
		t.Fatal("expected Temperature to be set")
	}
	if *req.Temperature != 0.1 {
		t.Errorf("Temperature = %v, want 0.1", *req.Temperature)
	}
}

func TestSessionSummary_nilDurationOmitted(t *testing.T) {
	ws := &domain.WorkoutSession{}
	req, err := SessionSummary(ws, domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(req.Messages[1].Content, "Duration:") {
		t.Errorf("user prompt should not contain Duration: when DurationS is nil:\n%s", req.Messages[1].Content)
	}
}

func TestSessionSummary_nilActivityFallsBackToDefault(t *testing.T) {
	ws := &domain.WorkoutSession{}
	req, err := SessionSummary(ws, domain.SessionSensitiveData{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(req.Messages[1].Content, "unspecified") {
		t.Errorf("expected fallback activity %q in prompt: %s", "unspecified", req.Messages[1].Content)
	}
}
