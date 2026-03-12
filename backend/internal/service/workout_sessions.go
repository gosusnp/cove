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

type WorkoutSessionService struct {
	db    *sql.DB
	store *store.WorkoutSessionStore
}

func NewWorkoutSessionService(db *sql.DB, s *store.WorkoutSessionStore) *WorkoutSessionService {
	return &WorkoutSessionService{db: db, store: s}
}

func (s *WorkoutSessionService) List(ctx context.Context) ([]domain.WorkoutSession, error) {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var list []domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, id.OrgID, id.UserID)
		return err
	})
	return list, err
}

func (s *WorkoutSessionService) Get(ctx context.Context, id domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var ws *domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ws, err = s.store.Get(ctx, q, identity.OrgID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ws, err
}

func (s *WorkoutSessionService) Create(ctx context.Context, p store.WorkoutSessionParams) (*domain.WorkoutSession, error) {
	if _, ok := domain.IdentityFromContext(ctx); !ok {
		return nil, ErrUnauthorized
	}

	var ws *domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ws, err = s.store.Create(ctx, q, p)
		return err
	})
	return ws, err
}

func (s *WorkoutSessionService) Update(ctx context.Context, id domain.WorkoutSessionID, p store.WorkoutSessionParams) (*domain.WorkoutSession, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var ws *domain.WorkoutSession
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ws, err = s.store.Update(ctx, q, identity.OrgID, id, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ws, err
}

func (s *WorkoutSessionService) Delete(ctx context.Context, id domain.WorkoutSessionID) error {
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
