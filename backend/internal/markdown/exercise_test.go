// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package markdown_test

import (
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
)

func TestExercise_MinimalPrivate(t *testing.T) {
	e := &domain.Exercise{ID: domain.ExerciseID(1), Name: "Squat", IsPublic: false}
	got := markdown.Exercise(e)
	if !strings.Contains(got, "Squat") {
		t.Errorf("expected name in output, got: %q", got)
	}
	if !strings.Contains(got, "private") {
		t.Errorf("expected private visibility, got: %q", got)
	}
	if !strings.Contains(got, "1") {
		t.Errorf("expected id in output, got: %q", got)
	}
}

func TestExercise_PublicWithProgressionAndDescription(t *testing.T) {
	e := &domain.Exercise{
		ID:          domain.ExerciseID(42),
		Name:        "Bench Press",
		IsPublic:    true,
		Progression: ptr("Add 2.5kg per week"),
		Description: ptr("Standard barbell bench press"),
	}
	got := markdown.Exercise(e)
	if !strings.Contains(got, "public") {
		t.Errorf("expected public visibility, got: %q", got)
	}
	if !strings.Contains(got, "Add 2.5kg per week") {
		t.Errorf("expected progression in output, got: %q", got)
	}
	if !strings.Contains(got, "Standard barbell bench press") {
		t.Errorf("expected description in output, got: %q", got)
	}
}

func TestExercise_OmitsEmptyProgression(t *testing.T) {
	empty := ""
	e := &domain.Exercise{ID: domain.ExerciseID(1), Name: "Plank", Progression: &empty}
	got := markdown.Exercise(e)
	if strings.Contains(got, "Progression:") {
		t.Errorf("expected no progression line for empty string, got: %q", got)
	}
}

func TestExerciseList_Empty(t *testing.T) {
	got := markdown.ExerciseList(nil)
	if !strings.Contains(got, "No exercises found") {
		t.Errorf("expected empty message, got: %q", got)
	}
}

func TestExerciseList_Multiple(t *testing.T) {
	es := []domain.Exercise{
		{ID: domain.ExerciseID(1), Name: "Squat", IsPublic: false},
		{ID: domain.ExerciseID(2), Name: "Deadlift", IsPublic: true, Progression: ptr("5x5 linear")},
	}
	got := markdown.ExerciseList(es)
	if !strings.Contains(got, "Squat") || !strings.Contains(got, "Deadlift") {
		t.Errorf("expected both exercises in output, got: %q", got)
	}
	if !strings.Contains(got, "private") {
		t.Errorf("expected private visibility for Squat, got: %q", got)
	}
	if !strings.Contains(got, "5x5 linear") {
		t.Errorf("expected progression for Deadlift, got: %q", got)
	}
}
