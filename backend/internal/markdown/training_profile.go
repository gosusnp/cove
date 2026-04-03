// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package markdown

import (
	"fmt"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// TrainingProfile converts TrainingProfileSensitiveData to a human-readable Markdown string.
// It is intended to be called within a UseSensitiveData callback to ensure
// plaintext is handled securely and briefly.
func TrainingProfile(data domain.TrainingProfileSensitiveData) string {
	var b strings.Builder

	b.WriteString("# Training Profile\n")

	if data.Motivation != nil && data.Motivation.String() != "" {
		b.WriteString("\n## Motivation\n")
		b.WriteString(data.Motivation.String())
		b.WriteString("\n")
	}

	if data.Constraints != nil && data.Constraints.String() != "" {
		b.WriteString("\n## Constraints\n")
		b.WriteString(data.Constraints.String())
		b.WriteString("\n")
	}

	if len(data.Disciplines) > 0 {
		b.WriteString("\n## Disciplines\n")
		for _, d := range data.Disciplines {
			fmt.Fprintf(&b, "- %s\n", disciplineLine(d))
		}
	}

	if data.IsEmpty() {
		b.WriteString("\nYour training profile is currently empty. Use update_training_profile to add details about your goals and experience.")
	}

	return b.String()
}

func disciplineLine(d domain.TrainingProfileDiscipline) string {
	name := "Unknown Discipline"
	if d.Name != nil {
		name = d.Name.String()
	}

	var details []string
	if d.YearsPractice != nil {
		details = append(details, fmt.Sprintf("%g years", *d.YearsPractice))
	}
	if d.Level != nil && d.Level.String() != "" {
		details = append(details, d.Level.String())
	}

	line := "**" + name + "**"
	if len(details) > 0 {
		line += " (" + strings.Join(details, " · ") + ")"
	}

	if d.Notes != nil && d.Notes.String() != "" {
		line += " — " + d.Notes.String()
	}

	return line
}
