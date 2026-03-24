// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package prompts

import (
	"text/template"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
)

// promptFuncs contains template functions available in all prompt templates.
var promptFuncs = template.FuncMap{
	"sensitive":  sensitiveString,
	"formatDate": formatDate,
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
