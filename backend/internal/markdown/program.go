// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Package markdown provides converters from domain types to Markdown text.
package markdown

import (
	"fmt"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// Program converts a domain.Program to a human-readable Markdown string
// covering name, description, sets (name, rounds, rest) and exercises
// (name, laterality, reps, duration, weight).
func Program(p *domain.Program) string {
	var b strings.Builder

	fmt.Fprintf(&b, "# %s\n", p.Name)

	if p.Description != nil && *p.Description != "" {
		fmt.Fprintf(&b, "\n%s\n", *p.Description)
	}

	for _, set := range p.Sets {
		b.WriteString("\n")
		b.WriteString(setHeading(set))
		b.WriteString("\n")
		for _, ex := range set.Exercises {
			fmt.Fprintf(&b, "- %s\n", exerciseLine(ex))
		}
	}

	return b.String()
}

// setHeading produces a level-2 heading for a set, e.g. "## Set A (3× · 90s rest)".
func setHeading(s domain.ProgramSet) string {
	var label string
	if s.Name != nil && *s.Name != "" {
		label = *s.Name
	} else {
		label = fmt.Sprintf("Set %d", s.ID)
	}

	detail := roundsLabel(s.Rounds)
	if s.IntraSetRestSeconds != nil && *s.IntraSetRestSeconds > 0 {
		detail += fmt.Sprintf(" · %ds rest", *s.IntraSetRestSeconds)
	}

	return fmt.Sprintf("## %s (%s)", label, detail)
}

// exerciseLine produces a single list item for an exercise, e.g.
// "Bench Press — bilateral · 8 reps · +80kg".
func exerciseLine(e domain.ProgramExercise) string {
	var details []string
	if e.Laterality != nil {
		details = append(details, *e.Laterality)
	}
	if e.TargetReps != nil {
		details = append(details, fmt.Sprintf("%d reps", *e.TargetReps))
	}
	if e.TargetDurationSeconds != nil {
		details = append(details, fmt.Sprintf("%ds", *e.TargetDurationSeconds))
	}
	if w := weightLabel(e.TargetWeightKg); w != "" {
		details = append(details, w)
	}

	if len(details) == 0 {
		return e.Name
	}
	return e.Name + " — " + strings.Join(details, " · ")
}

// roundsLabel returns a human-readable round count, e.g. "1 round" or "3 rounds".
func roundsLabel(n int) string {
	if n == 1 {
		return "1 round"
	}
	return fmt.Sprintf("%d rounds", n)
}

// weightLabel returns a human-readable weight string, or empty string for bodyweight.
func weightLabel(kg *float64) string {
	if kg == nil || *kg == 0 {
		return ""
	}
	if *kg > 0 {
		return fmt.Sprintf("+%gkg", *kg)
	}
	return fmt.Sprintf("%gkg (assisted)", *kg)
}
