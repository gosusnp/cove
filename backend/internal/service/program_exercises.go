// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

type ProgramExerciseService struct {
	db    *sql.DB
	store *store.ProgramExerciseStore
}

func NewProgramExerciseService(db *sql.DB, s *store.ProgramExerciseStore) *ProgramExerciseService {
	return &ProgramExerciseService{db: db, store: s}
}

func (s *ProgramExerciseService) List(ctx context.Context, setID int64) ([]store.ProgramExercise, error) {
	var list []store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		list, err = s.store.List(ctx, q, idInfo.OrgID, setID)
		return err
	})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (s *ProgramExerciseService) Get(ctx context.Context, setID, id int64) (*store.ProgramExercise, error) {
	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		pe, err = s.store.Get(ctx, q, idInfo.OrgID, setID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return pe, err
}

func (s *ProgramExerciseService) Create(ctx context.Context, setID int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		pe, err = s.store.Create(ctx, q, idInfo.OrgID, setID, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return pe, err
}

func (s *ProgramExerciseService) Update(ctx context.Context, setID, id int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		pe, err = s.store.Update(ctx, q, idInfo.OrgID, setID, id, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return pe, err
}

func (s *ProgramExerciseService) Delete(ctx context.Context, setID, id int64) error {
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		return s.store.Delete(ctx, q, idInfo.OrgID, setID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
