// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package markdown_test

import (
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
)

func sptr(s string) *crypto.SensitiveString {
	ss := crypto.NewSensitiveString(s)
	return &ss
}

func TestTrainingProfile_Empty(t *testing.T) {
	data := domain.TrainingProfileSensitiveData{}
	got := markdown.TrainingProfile(data)
	if !strings.Contains(got, "currently empty") {
		t.Errorf("expected empty profile message, got: %q", got)
	}
}

func TestTrainingProfile_Full(t *testing.T) {
	data := domain.TrainingProfileSensitiveData{
		Motivation:  sptr("I want to be strong."),
		Constraints: sptr("No heavy squats."),
		Disciplines: []domain.TrainingProfileDiscipline{
			{
				Name:          sptr("Weightlifting"),
				YearsPractice: ptr(5.0),
				Level:         sptr("intermediate"),
				Notes:         sptr("Focus on snatch."),
			},
		},
	}

	got := markdown.TrainingProfile(data)

	if !strings.Contains(got, "## Motivation") || !strings.Contains(got, "I want to be strong.") {
		t.Errorf("missing motivation, got: %q", got)
	}
	if !strings.Contains(got, "## Constraints") || !strings.Contains(got, "No heavy squats.") {
		t.Errorf("missing constraints, got: %q", got)
	}
	if !strings.Contains(got, "## Disciplines") || !strings.Contains(got, "Weightlifting") {
		t.Errorf("missing disciplines, got: %q", got)
	}
	if !strings.Contains(got, "5 years") || !strings.Contains(got, "intermediate") {
		t.Errorf("missing discipline details, got: %q", got)
	}
	if !strings.Contains(got, "Focus on snatch.") {
		t.Errorf("missing discipline notes, got: %q", got)
	}
}
