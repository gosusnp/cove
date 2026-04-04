// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package markdown

import (
	"fmt"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// Exercise converts a domain.Exercise to a human-readable Markdown string.
func Exercise(e *domain.Exercise) string {
	var b strings.Builder
	fmt.Fprintf(&b, "**%s** (id %d) — %s\n", e.Name, e.ID, visibilityLabel(e.IsPublic))
	if e.Progression != nil && *e.Progression != "" {
		fmt.Fprintf(&b, "Progression: %s\n", *e.Progression)
	}
	if e.Description != nil && *e.Description != "" {
		fmt.Fprintf(&b, "Description: %s\n", *e.Description)
	}
	return b.String()
}

// ExerciseList converts a slice of domain.Exercise to a Markdown list.
func ExerciseList(es []domain.Exercise) string {
	if len(es) == 0 {
		return "No exercises found.\n"
	}
	var b strings.Builder
	for _, e := range es {
		meta := []string{visibilityLabel(e.IsPublic)}
		if e.Progression != nil && *e.Progression != "" {
			meta = append(meta, "progression: "+*e.Progression)
		}
		fmt.Fprintf(&b, "- **%s** (id %d) — %s\n", e.Name, e.ID, strings.Join(meta, " · "))
	}
	return b.String()
}
