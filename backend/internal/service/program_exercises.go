// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

type ProgramExerciseService struct {
	programs *ProgramService
}

func NewProgramExerciseService(programs *ProgramService) *ProgramExerciseService {
	return &ProgramExerciseService{programs: programs}
}

func (s *ProgramExerciseService) List(ctx context.Context, programID domain.ProgramID, setID int64) ([]store.ProgramExercise, error) {
	return s.programs.ListExercises(ctx, programID, setID)
}

func (s *ProgramExerciseService) Get(ctx context.Context, programID domain.ProgramID, setID, id int64) (*store.ProgramExercise, error) {
	return s.programs.GetExercise(ctx, programID, setID, id)
}

func (s *ProgramExerciseService) Create(ctx context.Context, programID domain.ProgramID, setID int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	return s.programs.CreateExercise(ctx, programID, setID, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
}

func (s *ProgramExerciseService) Update(ctx context.Context, programID domain.ProgramID, setID, id int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	return s.programs.UpdateExercise(ctx, programID, setID, id, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
}

func (s *ProgramExerciseService) Delete(ctx context.Context, programID domain.ProgramID, setID, id int64) error {
	return s.programs.DeleteExercise(ctx, programID, setID, id)
}
