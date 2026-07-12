// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package markdown

import (
	"strings"
	"testing"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
)

func sensitiveStr(s string) *crypto.SensitiveString {
	ss := crypto.NewSensitiveString(s)
	return &ss
}

func intPtr(i int) *int { return &i }

func strPtr(s string) *string { return &s }

func timePtr(t time.Time) *time.Time { return &t }

func TestSessionEntry_basicHeader(t *testing.T) {
	started := time.Date(2026, 4, 8, 10, 30, 0, 0, time.UTC)
	ws := &domain.WorkoutSession{
		StartedAt: timePtr(started),
		Activity:  strPtr("Strength"),
		DurationS: intPtr(3720), // 1h 2m
	}
	sd := domain.SessionSensitiveData{
		PerceivedEffort: intPtr(7),
	}

	got := SessionEntry(ws, sd)

	for _, want := range []string{"2026-04-08 10:30", "Strength", "1h 2m", "RPE 7"} {
		if !strings.Contains(got, want) {
			t.Errorf("missing %q in output:\n%s", want, got)
		}
	}
}

func TestSessionEntry_fallsBackToCreatedAt(t *testing.T) {
	ws := &domain.WorkoutSession{
		CreatedAt: time.Date(2026, 4, 8, 0, 0, 0, 0, time.UTC),
	}
	got := SessionEntry(ws, domain.SessionSensitiveData{})
	if !strings.Contains(got, "2026-04-08") {
		t.Errorf("expected created_at date in output, got: %q", got)
	}
}

func TestSessionEntry_summaryBeforeNotes(t *testing.T) {
	ws := &domain.WorkoutSession{}
	sd := domain.SessionSensitiveData{
		Summary:      sensitiveStr("great session"),
		SessionNotes: sensitiveStr("raw notes"),
	}
	got := SessionEntry(ws, sd)

	summaryIdx := strings.Index(got, "great session")
	notesIdx := strings.Index(got, "raw notes")
	if summaryIdx == -1 || notesIdx == -1 {
		t.Fatalf("missing summary or notes in output:\n%s", got)
	}
	if summaryIdx > notesIdx {
		t.Errorf("expected summary before notes")
	}
}

func TestSessionEntry_structureFallback(t *testing.T) {
	ws := &domain.WorkoutSession{}
	sd := domain.SessionSensitiveData{
		ProgramStructure: sensitiveStr("5x5 squat, bench, deadlift"),
	}
	got := SessionEntry(ws, sd)
	if !strings.Contains(got, "5x5 squat") {
		t.Errorf("expected structure in output when no summary/notes, got: %q", got)
	}
}

func TestSessionEntry_structureHiddenWhenNotesPresent(t *testing.T) {
	ws := &domain.WorkoutSession{}
	sd := domain.SessionSensitiveData{
		SessionNotes:     sensitiveStr("felt good"),
		ProgramStructure: sensitiveStr("5x5 squat"),
	}
	got := SessionEntry(ws, sd)
	if strings.Contains(got, "5x5 squat") {
		t.Errorf("structure should be hidden when notes are present")
	}
}

func TestSessionEntry_labels(t *testing.T) {
	ws := &domain.WorkoutSession{
		Labels: []string{"deload", "recovery"},
	}
	got := SessionEntry(ws, domain.SessionSensitiveData{})
	for _, label := range []string{"deload", "recovery"} {
		if !strings.Contains(got, label) {
			t.Errorf("expected %q in output:\n%s", label, got)
		}
	}
}

func TestSessionEntry_noLabelsLine(t *testing.T) {
	ws := &domain.WorkoutSession{
		Labels: []string{},
	}
	got := SessionEntry(ws, domain.SessionSensitiveData{})
	if strings.Contains(got, "Labels:") {
		t.Errorf("expected no Labels line for empty labels, got: %q", got)
	}
}

func TestSessionEntry_programName(t *testing.T) {
	ws := &domain.WorkoutSession{}
	sd := domain.SessionSensitiveData{
		ProgramName: sensitiveStr("Upper/Lower Split"),
	}
	got := SessionEntry(ws, sd)
	if !strings.Contains(got, "Upper/Lower Split") {
		t.Errorf("expected program name in output, got: %q", got)
	}
}

func TestSessionList_empty(t *testing.T) {
	got := SessionList(nil)
	if !strings.Contains(got, "No sessions found") {
		t.Errorf("expected empty message, got: %q", got)
	}
}

func TestSessionList_separatorBetweenEntries(t *testing.T) {
	got := SessionList([]string{"entry one", "entry two"})
	if !strings.Contains(got, "---") {
		t.Errorf("expected separator between entries, got: %q", got)
	}
	if !strings.Contains(got, "entry one") || !strings.Contains(got, "entry two") {
		t.Errorf("expected both entries in output, got: %q", got)
	}
}

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		seconds int
		want    string
	}{
		{3720, "1h 2m"},
		{3600, "1h"},
		{2700, "45m"},
		{2730, "45m 30s"},
		{45, "45s"},
	}
	for _, c := range cases {
		got := formatDuration(c.seconds)
		if got != c.want {
			t.Errorf("formatDuration(%d) = %q, want %q", c.seconds, got, c.want)
		}
	}
}
