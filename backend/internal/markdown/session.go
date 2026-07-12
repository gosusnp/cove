// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package markdown

import (
	"fmt"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// SessionList renders a slice of pre-rendered session entries into a single
// Markdown string, separated by horizontal rules. Accepts the output of
// repeated SessionEntry calls. Returns a "no sessions" message for an empty slice.
func SessionList(entries []string) string {
	if len(entries) == 0 {
		return "No sessions found.\n"
	}
	return strings.Join(entries, "\n---\n\n")
}

// SessionEntry renders a single workout session as a Markdown string.
// Callers are responsible for decrypting sensitive data via
// WorkoutSession.UseSensitiveData and passing the result as sd.
func SessionEntry(ws *domain.WorkoutSession, sd domain.SessionSensitiveData) string {
	var b strings.Builder

	// Header line: date/time · activity · duration · RPE
	header := sessionHeader(ws, sd)
	fmt.Fprintf(&b, "### %s\n", header)

	// Labels
	if len(ws.Labels) > 0 {
		fmt.Fprintf(&b, "**Labels:** %s\n", strings.Join(ws.Labels, ", "))
	}

	// Program name
	if sd.ProgramName != nil {
		fmt.Fprintf(&b, "**Program:** %s\n", sd.ProgramName.String())
	}

	// Content: summary, then notes, then structure (fallback if both absent)
	hasSummary := sd.Summary != nil
	hasNotes := sd.SessionNotes != nil
	hasStructure := sd.ProgramStructure != nil

	if hasSummary {
		fmt.Fprintf(&b, "\n%s\n", sd.Summary.String())
	}
	if hasNotes {
		fmt.Fprintf(&b, "\n%s\n", sd.SessionNotes.String())
	}
	if !hasSummary && !hasNotes && hasStructure {
		fmt.Fprintf(&b, "\n%s\n", sd.ProgramStructure.String())
	}

	return b.String()
}

func sessionHeader(ws *domain.WorkoutSession, sd domain.SessionSensitiveData) string {
	var parts []string

	// Date and time — include time when started_at is set (disambiguates same-day sessions)
	if ws.StartedAt != nil {
		parts = append(parts, ws.StartedAt.UTC().Format("2006-01-02 15:04"))
	} else {
		parts = append(parts, ws.CreatedAt.UTC().Format("2006-01-02"))
	}

	if ws.Activity != nil {
		parts = append(parts, *ws.Activity)
	}

	if ws.DurationS != nil {
		parts = append(parts, formatDuration(*ws.DurationS))
	}

	if sd.PerceivedEffort != nil {
		parts = append(parts, fmt.Sprintf("RPE %d", *sd.PerceivedEffort))
	}

	return strings.Join(parts, " · ")
}

// formatDuration converts seconds to a human-readable string (e.g. "1h 15m", "45m", "30s").
func formatDuration(seconds int) string {
	h := seconds / 3600
	m := (seconds % 3600) / 60
	s := seconds % 60
	switch {
	case h > 0 && m > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case h > 0:
		return fmt.Sprintf("%dh", h)
	case m > 0 && s > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	case m > 0:
		return fmt.Sprintf("%dm", m)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
