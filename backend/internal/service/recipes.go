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

// RecipeService manages recipe business logic.
type RecipeService struct {
	db    *sql.DB
	store *store.RecipeStore
}

func NewRecipeService(db *sql.DB, s *store.RecipeStore) *RecipeService {
	return &RecipeService{db: db, store: s}
}

// List returns all recipes for the authenticated org.
func (s *RecipeService) List(ctx context.Context) ([]domain.RecipeLite, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	var list []domain.RecipeLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, identity.OrgID)
		return err
	})
	return list, err
}

// Get retrieves a recipe with its preparations.
func (s *RecipeService) Get(ctx context.Context, id domain.RecipeID) (*domain.Recipe, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	var r *domain.Recipe
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		r, err = s.store.Get(ctx, q, identity.OrgID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return r, err
}

// Create inserts a new recipe with validation.
func (s *RecipeService) Create(ctx context.Context, p domain.RecipeParams) (*domain.Recipe, error) {
	if _, ok := domain.IdentityFromContext(ctx); !ok {
		return nil, ErrUnauthorized
	}
	p.Name = normalizeName(p.Name)
	if p.Name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	if p.Servings <= 0 {
		return nil, &ValidationError{Msg: "servings must be greater than zero"}
	}
	var r *domain.Recipe
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		r, err = s.store.Create(ctx, q, p)
		return err
	})
	return r, err
}

// Update modifies a recipe with validation.
func (s *RecipeService) Update(ctx context.Context, id domain.RecipeID, p domain.RecipeParams) (*domain.Recipe, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	p.Name = normalizeName(p.Name)
	if p.Name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	if p.Servings <= 0 {
		return nil, &ValidationError{Msg: "servings must be greater than zero"}
	}
	var r *domain.Recipe
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		r, err = s.store.Update(ctx, q, identity.OrgID, id, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return r, err
}

// Delete removes a recipe and its preparation links.
func (s *RecipeService) Delete(ctx context.Context, id domain.RecipeID) error {
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

// AddPreparation links a preparation to a recipe.
func (s *RecipeService) AddPreparation(ctx context.Context, recipeID domain.RecipeID, p domain.RecipePreparationParams) (*domain.RecipePreparation, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	if p.Position <= 0 {
		return nil, &ValidationError{Msg: "position must be greater than zero"}
	}
	if p.Amount <= 0 {
		return nil, &ValidationError{Msg: "amount must be greater than zero"}
	}
	if strings.TrimSpace(p.Unit) == "" {
		return nil, &ValidationError{Msg: "unit is required"}
	}
	var rp *domain.RecipePreparation
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		rp, err = s.store.AddPreparation(ctx, q, identity.OrgID, recipeID, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return rp, err
}

// UpdatePreparation modifies a recipe preparation link.
func (s *RecipeService) UpdatePreparation(ctx context.Context, id domain.RecipePreparationID, p domain.RecipePreparationParams) (*domain.RecipePreparation, error) {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}
	if p.Position <= 0 {
		return nil, &ValidationError{Msg: "position must be greater than zero"}
	}
	if p.Amount <= 0 {
		return nil, &ValidationError{Msg: "amount must be greater than zero"}
	}
	if strings.TrimSpace(p.Unit) == "" {
		return nil, &ValidationError{Msg: "unit is required"}
	}
	var rp *domain.RecipePreparation
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		rp, err = s.store.UpdatePreparation(ctx, q, identity.OrgID, id, p)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return rp, err
}

// DeletePreparation removes a preparation link from a recipe.
func (s *RecipeService) DeletePreparation(ctx context.Context, id domain.RecipePreparationID) error {
	identity, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		return s.store.DeletePreparation(ctx, q, identity.OrgID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}
