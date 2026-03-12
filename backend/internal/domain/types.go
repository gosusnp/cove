// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"time"
)

// -----------------------------------------------------------------------------
// Workout Session
// -----------------------------------------------------------------------------

type WorkoutSessionID IntID[struct{ workoutSession struct{} }]

// WorkoutSession represents a single training session.
// program_structure being non-nil implicitly indicates a structured session.
type WorkoutSession struct {
	ID               WorkoutSessionID `json:"id"`
	OrgID            OrgID            `json:"org_id"`
	UserID           UserID           `json:"user_id"`
	ProgramID        *ProgramID       `json:"program_id,omitempty"`
	ProgramName      *string          `json:"program_name,omitempty"`
	ProgramStructure *string          `json:"program_structure,omitempty"`
	Activity         *string          `json:"activity,omitempty"`
	DurationS        *int             `json:"duration_s,omitempty"`
	PerceivedEffort  *int             `json:"perceived_effort,omitempty"`
	SessionNotes     *string          `json:"session_notes,omitempty"`
	StartedAt        *time.Time       `json:"started_at,omitempty"`
	CompletedAt      *time.Time       `json:"completed_at,omitempty"`
	CreatedBy        UserID           `json:"created_by"`
	CreatedAt        time.Time        `json:"created_at"`
	UpdatedBy        *UserID          `json:"updated_by,omitempty"`
	UpdatedAt        time.Time        `json:"updated_at"`
}

// -----------------------------------------------------------------------------
// Program
// -----------------------------------------------------------------------------

type ProgramID IntID[struct{ program struct{} }]

// ProgramLite is a trimmed version of a program.
type ProgramLite struct {
	ID       ProgramID `json:"id"`
	Name     string    `json:"name"`
	OrgID    OrgID     `json:"org_id"`
	IsPublic bool      `json:"is_public"`
}

// Program is the complete program hierarchy.
type Program struct {
	ID          ProgramID    `json:"id"`
	Name        string       `json:"name"`
	Description *string      `json:"description,omitempty"`
	OrgID       OrgID        `json:"org_id"`
	IsPublic    bool         `json:"is_public"`
	CreatedBy   UserID       `json:"created_by"`
	CreatedAt   time.Time    `json:"created_at"`
	UpdatedBy   *UserID      `json:"updated_by,omitempty"`
	UpdatedAt   time.Time    `json:"updated_at"`
	Sets        []ProgramSet `json:"sets"`
}

type ProgramSet struct {
	ID                  int64             `json:"id"`
	Name                *string           `json:"name,omitempty"`
	Rounds              int               `json:"rounds"`
	IntraSetRestSeconds *int              `json:"rest_s,omitempty"`
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
