// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

type Exercise struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type ExerciseDetail struct {
	ID          int64   `json:"id"`
	Name        string  `json:"name"`
	Progression *string `json:"progression,omitempty"`
}

type Program struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

// ProgramDetail is the full program hierarchy returned by GET /programs/{id}.
type ProgramDetail struct {
	ID   int64              `json:"id"`
	Name string             `json:"name"`
	Sets []ProgramSetDetail `json:"sets"`
}

type ProgramSetDetail struct {
	ID                  int64                   `json:"id"`
	Name                *string                 `json:"name,omitempty"`
	Rounds              int                     `json:"rounds"`
	IntraSetRestSeconds *int                    `json:"rest_s,omitempty"`
	SortOrder           *int                    `json:"sort_order,omitempty"`
	Exercises           []ProgramExerciseDetail `json:"exercises"`
}

type ProgramExerciseDetail struct {
	ID                    int64    `json:"id"`
	ExerciseID            int64    `json:"exercise_id"`
	Name                  string   `json:"name"`
	Laterality            *string  `json:"laterality,omitempty"`
	TargetReps            *int     `json:"reps,omitempty"`
	TargetDurationSeconds *int     `json:"duration_s,omitempty"`
	TargetWeightKg        *float64 `json:"weight_kg,omitempty"`
	SortOrder             *int     `json:"sort_order,omitempty"`
}

type ProgramSet struct {
	ID                  int64   `json:"id"`
	ProgramID           int64   `json:"program_id"`
	Name                *string `json:"name,omitempty"`
	Rounds              int     `json:"rounds"`
	IntraSetRestSeconds *int    `json:"rest_s,omitempty"`
	SortOrder           *int    `json:"sort_order,omitempty"`
}

type ProgramExercise struct {
	ID                    int64    `json:"id"`
	ProgramSetID          int64    `json:"program_set_id"`
	ExerciseID            int64    `json:"exercise_id"`
	Laterality            *string  `json:"laterality,omitempty"`
	TargetReps            *int     `json:"reps,omitempty"`
	TargetDurationSeconds *int     `json:"duration_s,omitempty"`
	TargetWeightKg        *float64 `json:"weight_kg,omitempty"`
	SortOrder             *int     `json:"sort_order,omitempty"`
}
