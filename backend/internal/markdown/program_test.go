// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package markdown_test

import (
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
)

func ptr[T any](v T) *T { return &v }

func TestProgram_NoSets(t *testing.T) {
	p := &domain.Program{Name: "Rest Day"}
	got := markdown.Program(p)
	if !strings.HasPrefix(got, "# Rest Day\n") {
		t.Errorf("expected heading, got: %q", got)
	}
}

func TestProgram_WithDescription(t *testing.T) {
	p := &domain.Program{
		Name:        "Push Day",
		Description: ptr("Upper body push focus."),
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "Upper body push focus.") {
		t.Errorf("expected description in output, got: %q", got)
	}
}

func TestProgram_SetHeading_Named(t *testing.T) {
	p := &domain.Program{
		Name: "Full Body",
		Sets: []domain.ProgramSet{
			{ID: 1, Name: ptr("A"), Rounds: 3, IntraSetRestSeconds: ptr(90)},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "## A (3 rounds · 90s rest)") {
		t.Errorf("expected named set heading, got: %q", got)
	}
}

func TestProgram_SetHeading_Unnamed(t *testing.T) {
	p := &domain.Program{
		Name: "Full Body",
		Sets: []domain.ProgramSet{
			{ID: 2, Rounds: 4},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "## Set 2 (4 rounds)") {
		t.Errorf("expected unnamed set heading with ID, got: %q", got)
	}
}

func TestProgram_SetHeading_NoRest(t *testing.T) {
	p := &domain.Program{
		Name: "Full Body",
		Sets: []domain.ProgramSet{
			{ID: 1, Name: ptr("B"), Rounds: 2},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "## B (2 rounds)") {
		t.Errorf("expected no rest in heading when unset, got: %q", got)
	}
}

func TestProgram_ExerciseLine_AllFields(t *testing.T) {
	p := &domain.Program{
		Name: "Strength",
		Sets: []domain.ProgramSet{
			{
				ID:     1,
				Rounds: 3,
				Exercises: []domain.ProgramExercise{
					{
						Name:                  "Bench Press",
						Laterality:            ptr("bilateral"),
						TargetReps:            ptr(8),
						TargetDurationSeconds: nil,
						TargetWeight:          ptr(80.0),
					},
				},
			},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "- Bench Press — bilateral · 8 reps · +80kg") {
		t.Errorf("unexpected exercise line, got: %q", got)
	}
}

func TestProgram_ExerciseLine_ExplicitPoundUnit(t *testing.T) {
	u := domain.UnitPound
	p := &domain.Program{
		Name: "Strength",
		Sets: []domain.ProgramSet{
			{
				ID:     1,
				Rounds: 3,
				Exercises: []domain.ProgramExercise{
					{
						Name:         "Bench Press",
						TargetReps:   ptr(5),
						TargetWeight: ptr(62.5),
						WeightUnit:   &u,
					},
				},
			},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "+62.5lb") {
		t.Errorf("expected +62.5lb in output, got: %q", got)
	}
}

func TestProgram_ExerciseLine_DurationOnly(t *testing.T) {
	p := &domain.Program{
		Name: "Cardio",
		Sets: []domain.ProgramSet{
			{
				ID:     1,
				Rounds: 1,
				Exercises: []domain.ProgramExercise{
					{Name: "Plank", TargetDurationSeconds: ptr(60)},
				},
			},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "- Plank — 60s") {
		t.Errorf("unexpected duration-only exercise line, got: %q", got)
	}
	if strings.Contains(got, "bodyweight") {
		t.Errorf("expected no bodyweight label, got: %q", got)
	}
}

func TestProgram_ExerciseLine_AssistedWeight(t *testing.T) {
	p := &domain.Program{
		Name: "Strength",
		Sets: []domain.ProgramSet{
			{
				ID:     1,
				Rounds: 3,
				Exercises: []domain.ProgramExercise{
					{Name: "Pull-up", TargetReps: ptr(10), TargetWeight: ptr(-10.0)},
				},
			},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "-10kg (assisted)") {
		t.Errorf("expected assisted weight label, got: %q", got)
	}
}

func TestProgram_ExerciseLine_NoDetails(t *testing.T) {
	p := &domain.Program{
		Name: "Conditioning",
		Sets: []domain.ProgramSet{
			{
				ID:     1,
				Rounds: 1,
				Exercises: []domain.ProgramExercise{
					{Name: "Wrist Roller"},
				},
			},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "- Wrist Roller\n") {
		t.Errorf("expected no em dash when no details, got: %q", got)
	}
}

func TestProgram_ExerciseLine_Bodyweight(t *testing.T) {
	p := &domain.Program{
		Name: "Calisthenics",
		Sets: []domain.ProgramSet{
			{
				ID:     1,
				Rounds: 3,
				Exercises: []domain.ProgramExercise{
					{Name: "Push-up", TargetReps: ptr(20)},
				},
			},
		},
	}
	got := markdown.Program(p)
	if strings.Contains(got, "bodyweight") {
		t.Errorf("expected no bodyweight label for nil weight, got: %q", got)
	}
	if !strings.Contains(got, "- Push-up — 20 reps") {
		t.Errorf("expected exercise without weight label, got: %q", got)
	}
}

func TestProgram_MultipleSetsAndExercises(t *testing.T) {
	p := &domain.Program{
		Name: "Push/Pull",
		Sets: []domain.ProgramSet{
			{
				ID: 1, Name: ptr("Push"), Rounds: 3,
				Exercises: []domain.ProgramExercise{
					{Name: "Bench Press", TargetReps: ptr(8), TargetWeight: ptr(80.0)},
					{Name: "Overhead Press", TargetReps: ptr(10), TargetWeight: ptr(50.0)},
				},
			},
			{
				ID: 2, Name: ptr("Pull"), Rounds: 3,
				Exercises: []domain.ProgramExercise{
					{Name: "Barbell Row", TargetReps: ptr(8), TargetWeight: ptr(70.0)},
				},
			},
		},
	}
	got := markdown.Program(p)
	if !strings.Contains(got, "## Push") {
		t.Errorf("expected Push set heading, got: %q", got)
	}
	if !strings.Contains(got, "## Pull") {
		t.Errorf("expected Pull set heading, got: %q", got)
	}
	if !strings.Contains(got, "Bench Press") || !strings.Contains(got, "Barbell Row") {
		t.Errorf("expected both exercises, got: %q", got)
	}
}
