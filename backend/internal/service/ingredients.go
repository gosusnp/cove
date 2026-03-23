// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/fdc"
	"github.com/gosusnp/cove/backend/internal/store"
)

type IngredientService struct {
	db    *sql.DB
	store *store.IngredientStore
	fdc   *fdc.Client // nullable; density computation is best-effort
}

func NewIngredientService(db *sql.DB, s *store.IngredientStore, fdcClient *fdc.Client) *IngredientService {
	return &IngredientService{db: db, store: s, fdc: fdcClient}
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

	// Best-effort: fetch density from FDC when an FDC ID is provided and density is unknown.
	if p.FdcID != nil && s.fdc != nil && p.DensityGPerMl == nil {
		if food, err := s.fdc.GetFood(ctx, *p.FdcID); err == nil {
			p.DensityGPerMl = food.DensityGPerMl
		}
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

func (s *IngredientService) RefreshFromFDC(ctx context.Context, id domain.IngredientID) (*domain.Ingredient, error) {
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
	if err != nil {
		return nil, err
	}
	if ing.FdcID == nil {
		return nil, &ValidationError{Msg: "ingredient has no FDC ID"}
	}
	if s.fdc == nil {
		return nil, fmt.Errorf("FDC client not configured")
	}

	food, err := s.fdc.GetFood(ctx, *ing.FdcID)
	if err != nil {
		return nil, &ExternalServiceError{Msg: "FDC is currently unavailable, please try again"}
	}

	p := domain.IngredientParams{
		Name:            ing.Name,
		FdcID:           ing.FdcID,
		CaloriesPer100g: food.CaloriesPer100g,
		ProteinPer100g:  food.ProteinPer100g,
		FatPer100g:      food.FatPer100g,
		CarbsPer100g:    food.CarbsPer100g,
		DensityGPerMl:   food.DensityGPerMl,
		IsPublic:        ing.IsPublic,
	}

	err = withScopedTx(ctx, s.db, func(q store.Querier) error {
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
