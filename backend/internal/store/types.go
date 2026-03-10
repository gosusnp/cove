// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import "github.com/gosusnp/cove/backend/internal/domain"

type ProgramSet struct {
	ID                  int64            `json:"id"`
	ProgramID           domain.ProgramID `json:"program_id"`
	Name                *string          `json:"name,omitempty"`
	Rounds              int              `json:"rounds"`
	IntraSetRestSeconds *int             `json:"rest_s,omitempty"`
}

type ProgramExercise struct {
	ID                    int64             `json:"id"`
	ProgramSetID          int64             `json:"program_set_id"`
	ExerciseID            domain.ExerciseID `json:"exercise_id"`
	Laterality            *string           `json:"laterality,omitempty"`
	TargetReps            *int              `json:"reps,omitempty"`
	TargetDurationSeconds *int              `json:"duration_s,omitempty"`
	TargetWeightKg        *float64          `json:"weight_kg,omitempty"`
}
