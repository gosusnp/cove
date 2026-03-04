// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

// -----------------------------------------------------------------------------
// Program
// -----------------------------------------------------------------------------

type ProgramID IntID[struct{ program struct{} }]

// Program is the full program hierarchy.
type Program struct {
	ID   ProgramID    `json:"id"`
	Name string       `json:"name"`
	Sets []ProgramSet `json:"sets"`
}

// ProgramLite is a trimmed version of a program, usually used in lists.
type ProgramLite struct {
	ID   ProgramID `json:"id"`
	Name string    `json:"name"`
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
}

// ExerciseLite is a trimmed version of an exercise.
type ExerciseLite struct {
	ID   ExerciseID `json:"id"`
	Name string     `json:"name"`
}
