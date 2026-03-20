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

type IngredientService struct {
	db    *sql.DB
	store *store.IngredientStore
}

func NewIngredientService(db *sql.DB, s *store.IngredientStore) *IngredientService {
	return &IngredientService{db: db, store: s}
}

func (s *IngredientService) List(ctx context.Context) ([]domain.Ingredient, error) {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var list []domain.Ingredient
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, id.OrgID)
		return err
	})
	return list, err
}

func (s *IngredientService) Get(ctx context.Context, id domain.IngredientID) (*domain.Ingredient, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var ing *domain.Ingredient
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ing, err = s.store.Get(ctx, q, identity.OrgID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ing, err
}

func (s *IngredientService) Create(ctx context.Context, p domain.IngredientParams) (*domain.Ingredient, error) {
	p.Name = normalizeName(p.Name)
	if p.Name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}

	var ing *domain.Ingredient
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ing, err = s.store.Create(ctx, q, p)
		return err
	})
	return ing, err
}

func (s *IngredientService) Update(ctx context.Context, id domain.IngredientID, p domain.IngredientParams) (*domain.Ingredient, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	p.Name = normalizeName(p.Name)
	if p.Name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}

	var ing *domain.Ingredient
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ing, err = s.store.Update(ctx, q, identity.OrgID, id, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ing, err
}

func (s *IngredientService) Delete(ctx context.Context, id domain.IngredientID) error {
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
