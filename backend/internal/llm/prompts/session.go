// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package prompts

import (
	"bytes"
	_ "embed"
	"fmt"
	"text/template"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/llm"
)

//go:embed fitness_coach_system.md
var fitnessCoachSystem string

//go:embed session_summary_user.tmpl
var sessionSummaryUserTmpl string

var sessionSummaryUserTemplate = template.Must(
	template.New("session_summary_user").
		Funcs(promptFuncs).
		Parse(sessionSummaryUserTmpl),
)

type sessionSummaryData struct {
	Activity  string
	StartedAt *time.Time
	DurationS *int
	Sensitive domain.SessionSensitiveData
}

// SessionSummary builds a Request that asks the LLM to produce a markdown
// summary of a single workout session. Task-level parameters (temperature,
// thinking mode) are applied by the router, not here.
//
// Must be called inside a WorkoutSession.UseSensitiveData callback so that
// sensitive fields in sd are zeroed immediately after the call returns.
func SessionSummary(ws *domain.WorkoutSession, sd domain.SessionSensitiveData) (llm.Request, error) {
	data := sessionSummaryData{
		Activity:  orDefault(ws.Activity, "unspecified"),
		StartedAt: ws.StartedAt,
		DurationS: ws.DurationS,
		Sensitive: sd,
	}
	var buf bytes.Buffer
	if err := sessionSummaryUserTemplate.Execute(&buf, data); err != nil {
		return llm.Request{}, fmt.Errorf("render session summary user prompt: %w", err)
	}
	return llm.Request{
		Messages: []llm.Message{
			{Role: "system", Content: fitnessCoachSystem},
			{Role: "user", Content: buf.String()},
		},
	}, nil
}

func orDefault(s *string, def string) string {
	if s == nil {
		return def
	}
	return *s
}
