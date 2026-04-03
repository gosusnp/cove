// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package prompts

import (
	"fmt"
	"text/template"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
)

// promptFuncs contains template functions available in all prompt templates.
var promptFuncs = template.FuncMap{
	"sensitive":      sensitiveString,
	"formatDate":     formatDate,
	"formatDuration": formatDuration,
}

// sensitiveString converts a *crypto.SensitiveString to a plain string at
// template render time. Returns an empty string for nil inputs.
func sensitiveString(s *crypto.SensitiveString) string {
	if s == nil {
		return ""
	}
	return s.String()
}

// formatDate formats a *time.Time as "Mon Jan 2, 2006". Returns an empty
// string for nil inputs so templates can use {{with formatDate .Field}}.
func formatDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("Mon Jan 2, 2006")
}

// formatDuration formats a duration in seconds as a human-readable string
// (e.g. "1h 0m", "45m"). Returns an empty string for nil inputs.
func formatDuration(s *int) string {
	if s == nil {
		return ""
	}
	total := *s
	h := total / 3600
	m := (total % 3600) / 60
	if h > 0 {
		return fmt.Sprintf("%dh %dm", h, m)
	}
	return fmt.Sprintf("%dm", m)
}
