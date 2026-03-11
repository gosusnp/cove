// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

type ExerciseService struct {
	db    *sql.DB
	store *store.ExerciseStore
}

func NewExerciseService(db *sql.DB, s *store.ExerciseStore) *ExerciseService {
	return &ExerciseService{db: db, store: s}
}

func normalizeName(s string) string {
	return strings.Join(strings.Fields(s), " ")
}

func (s *ExerciseService) List(ctx context.Context) ([]domain.Exercise, error) {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var list []domain.Exercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, id.OrgID)
		return err
	})
	return list, err
}

// ExerciseResolution is the result of a bulk ID lookup.
// Found contains exercises that were visible to the caller.
// Missing contains IDs that were not found or excluded by visibility rules.
type ExerciseResolution struct {
	Found   []domain.Exercise   `json:"found"`
	Missing []domain.ExerciseID `json:"missing"`
}

func (s *ExerciseService) GetByIDs(ctx context.Context, ids []domain.ExerciseID) (ExerciseResolution, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ExerciseResolution{}, ErrUnauthorized
	}

	var found []domain.Exercise
	var missing []domain.ExerciseID
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		found, missing, err = s.store.GetByIDs(ctx, q, identity.OrgID, ids)
		return err
	})
	if err != nil {
		return ExerciseResolution{}, err
	}
	return ExerciseResolution{Found: found, Missing: missing}, nil
}

func (s *ExerciseService) Get(ctx context.Context, id domain.ExerciseID) (*domain.Exercise, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var ex *domain.Exercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ex, err = s.store.Get(ctx, q, identity.OrgID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ex, err
}

func (s *ExerciseService) Create(ctx context.Context, name string, progression *string, description *string, isPublic bool) (*domain.Exercise, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	var ex *domain.Exercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ex, err = s.store.Create(ctx, q, name, progression, description, isPublic)
		return err
	})
	return ex, err
}

func (s *ExerciseService) Update(ctx context.Context, id domain.ExerciseID, name string, progression *string, description *string, isPublic bool) (*domain.Exercise, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	var ex *domain.Exercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ex, err = s.store.Update(ctx, q, identity.OrgID, id, name, progression, description, isPublic)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ex, err
}

func (s *ExerciseService) Delete(ctx context.Context, id domain.ExerciseID) error {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		return s.store.Delete(ctx, q, identity.OrgID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
