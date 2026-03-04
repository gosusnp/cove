// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"time"
)

// -----------------------------------------------------------------------------
// Program
// -----------------------------------------------------------------------------

type ProgramID IntID[struct{ program struct{} }]

// ProgramLite is a trimmed version of a program.
type ProgramLite struct {
	ID   ProgramID `json:"id"`
	Name string    `json:"name"`
}

// Program is the complete program hierarchy.
type Program struct {
	ID   ProgramID    `json:"id"`
	Name string       `json:"name"`
	Sets []ProgramSet `json:"sets"`
}

type ProgramSet struct {
	ID                  int64             `json:"id"`
	Name                *string           `json:"name,omitempty"`
	Rounds              int               `json:"rounds"`
	IntraSetRestSeconds *int              `json:"rest_s,omitempty"`
	SortOrder           *int              `json:"sort_order,omitempty"`
	Exercises           []ProgramExercise `json:"exercises"`
}

type ProgramExercise struct {
	ID                    int64      `json:"id"`
	ExerciseID            ExerciseID `json:"exercise_id"`
	Name                  string     `json:"name"`
	Laterality            *string    `json:"laterality,omitempty"`
	TargetReps            *int       `json:"reps,omitempty"`
	TargetDurationSeconds *int       `json:"duration_s,omitempty"`
	TargetWeightKg        *float64   `json:"weight_kg,omitempty"`
	SortOrder             *int       `json:"sort_order,omitempty"`
}

// -----------------------------------------------------------------------------
// Exercise
// -----------------------------------------------------------------------------

type ExerciseID IntID[struct{ exercise struct{} }]

// Exercise is the complete exercise definition.
type Exercise struct {
	ID          ExerciseID `json:"id"`
	Name        string     `json:"name"`
	Progression *string    `json:"progression,omitempty"`
	Description *string    `json:"description,omitempty"`
	OrgID       OrgID      `json:"org_id"`
	IsPublic    bool       `json:"is_public"`
	CreatedBy   UserID     `json:"created_by"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedBy   *UserID    `json:"updated_by,omitempty"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ExerciseLite is a trimmed version of an exercise.
type ExerciseLite struct {
	ID   ExerciseID `json:"id"`
	Name string     `json:"name"`
}
