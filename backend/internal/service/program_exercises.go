// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"errors"

	"github.com/gosusnp/cove/backend/internal/store"
)

type ProgramExerciseService struct {
	store *store.ProgramExerciseStore
}

func NewProgramExerciseService(s *store.ProgramExerciseStore) *ProgramExerciseService {
	return &ProgramExerciseService{store: s}
}

func (s *ProgramExerciseService) List(setID int64) ([]store.ProgramExercise, error) {
	return s.store.List(setID)
}

func (s *ProgramExerciseService) Get(setID, id int64) (*store.ProgramExercise, error) {
	e, err := s.store.Get(setID, id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return e, err
}

func (s *ProgramExerciseService) Create(setID, exerciseID int64, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	return s.store.Create(setID, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
}

func (s *ProgramExerciseService) Update(setID, id, exerciseID int64, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	e, err := s.store.Update(setID, id, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return e, err
}

func (s *ProgramExerciseService) Delete(setID, id int64) error {
	if err := s.store.Delete(setID, id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}
