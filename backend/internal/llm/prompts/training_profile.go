// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package prompts

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"text/template"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
)

//go:embed build_training_profile.tmpl
var buildTrainingProfileTmpl string

var buildTrainingProfileTemplate = template.Must(
	template.New("build_training_profile").
		Funcs(promptFuncs).
		Parse(buildTrainingProfileTmpl),
)

type buildTrainingProfileData struct {
	ProfileMarkdown string
}

// BuildTrainingProfile renders the onboarding prompt for gathering user training profile info.
// If tp is provided, its sensitive data is used to populate the "Current Profile" section.
func BuildTrainingProfile(ctx context.Context, tp *domain.UserTrainingProfile) (string, error) {
	var data buildTrainingProfileData
	if tp != nil {
		err := tp.UseSensitiveData(ctx, func(sd domain.TrainingProfileSensitiveData) error {
			data.ProfileMarkdown = markdown.TrainingProfile(sd)
			return nil
		})
		if err != nil {
			return "", fmt.Errorf("render current profile for prompt: %w", err)
		}
	}

	var buf bytes.Buffer
	if err := buildTrainingProfileTemplate.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("render build training profile prompt: %w", err)
	}
	return buf.String(), nil
}
