// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID        uuid.UUID
	Email     string
	GoogleSub string
	CreatedAt time.Time
}

type Org struct {
	ID uuid.UUID
}

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

type PAT struct {
	ID         uuid.UUID  `json:"id"`
	Name       string     `json:"name"`
	CreatedAt  time.Time  `json:"created_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
}

type Session struct {
	ID              uuid.UUID  `json:"id"`
	CreatedAt       time.Time  `json:"created_at"`
	LastUsedAt      *time.Time `json:"last_used_at"`
	InitialIPMasked *string    `json:"initial_ip_masked,omitempty"`
	InitialBrowser  *string    `json:"initial_browser,omitempty"`
	InitialOS       *string    `json:"initial_os,omitempty"`
	LastIPMasked    *string    `json:"last_ip_masked,omitempty"`
	LastBrowser     *string    `json:"last_browser,omitempty"`
	LastOS          *string    `json:"last_os,omitempty"`
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
