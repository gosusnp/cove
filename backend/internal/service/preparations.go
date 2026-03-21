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

// PreparationService manages preparation business logic.
type PreparationService struct {
	db    *sql.DB
	store *store.PreparationStore
}

func NewPreparationService(db *sql.DB, s *store.PreparationStore) *PreparationService {
	return &PreparationService{db: db, store: s}
}

// List returns all preparations for the authenticated org.
func (s *PreparationService) List(ctx context.Context) ([]domain.PreparationLite, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	var list []domain.PreparationLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, identity.OrgID)
		return err
	})
	return list, err
}

// Get retrieves a preparation with its ingredients.
func (s *PreparationService) Get(ctx context.Context, id domain.PreparationID) (*domain.Preparation, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	var p *domain.Preparation
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		p, err = s.store.Get(ctx, q, identity.OrgID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

// Create inserts a new preparation with validation.
func (s *PreparationService) Create(ctx context.Context, p domain.PreparationParams) (*domain.Preparation, error) {
	if _, ok := domain.IdentityFromContext(ctx); !ok {
		return nil, ErrUnauthorized
	}
	p.Name = normalizeName(p.Name)
	if p.Name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	if p.YieldAmount <= 0 {
		return nil, &ValidationError{Msg: "yield_amount must be greater than zero"}
	}
	if strings.TrimSpace(p.YieldUnit) == "" {
		return nil, &ValidationError{Msg: "yield_unit is required"}
	}
	if p.Steps == nil {
		p.Steps = []domain.PreparationStep{}
	}
	var prep *domain.Preparation
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		prep, err = s.store.Create(ctx, q, p)
		return err
	})
	return prep, err
}

// Update modifies a preparation with validation.
func (s *PreparationService) Update(ctx context.Context, id domain.PreparationID, p domain.PreparationParams) (*domain.Preparation, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	p.Name = normalizeName(p.Name)
	if p.Name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	if p.YieldAmount <= 0 {
		return nil, &ValidationError{Msg: "yield_amount must be greater than zero"}
	}
	if strings.TrimSpace(p.YieldUnit) == "" {
		return nil, &ValidationError{Msg: "yield_unit is required"}
	}
	if p.Steps == nil {
		p.Steps = []domain.PreparationStep{}
	}
	var prep *domain.Preparation
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		prep, err = s.store.Update(ctx, q, identity.OrgID, id, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return prep, err
}

// Delete removes a preparation and its ingredients.
func (s *PreparationService) Delete(ctx context.Context, id domain.PreparationID) error {
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

// AddIngredient adds an ingredient to a preparation.
func (s *PreparationService) AddIngredient(ctx context.Context, preparationID domain.PreparationID, p domain.PreparationIngredientParams) (*domain.PreparationIngredient, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(p.Name) == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	if p.Amount <= 0 {
		return nil, &ValidationError{Msg: "amount must be greater than zero"}
	}
	if strings.TrimSpace(p.Unit) == "" {
		return nil, &ValidationError{Msg: "unit is required"}
	}
	var ingredient *domain.PreparationIngredient
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ingredient, err = s.store.AddIngredient(ctx, q, identity.OrgID, preparationID, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ingredient, err
}

// UpdateIngredient modifies a preparation ingredient.
func (s *PreparationService) UpdateIngredient(ctx context.Context, id domain.PreparationIngredientID, p domain.PreparationIngredientParams) (*domain.PreparationIngredient, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	if strings.TrimSpace(p.Name) == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	if p.Amount <= 0 {
		return nil, &ValidationError{Msg: "amount must be greater than zero"}
	}
	if strings.TrimSpace(p.Unit) == "" {
		return nil, &ValidationError{Msg: "unit is required"}
	}
	var ingredient *domain.PreparationIngredient
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		ingredient, err = s.store.UpdateIngredient(ctx, q, identity.OrgID, id, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ingredient, err
}

// DeleteIngredient removes an ingredient from a preparation.
func (s *PreparationService) DeleteIngredient(ctx context.Context, id domain.PreparationIngredientID) error {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		return s.store.DeleteIngredient(ctx, q, identity.OrgID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
