// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
)

type IngredientStore struct{}

func NewIngredientStore() *IngredientStore {
	return &IngredientStore{}
}

func (s *IngredientStore) List(ctx context.Context, q Querier, orgID domain.OrgID) ([]domain.Ingredient, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, fdc_id, calories_per_100g, protein_per_100g, fat_per_100g, carbs_per_100g,
		       density_g_per_ml, org_id, is_public, created_by, created_at, updated_by, updated_at
		FROM cove.ingredients
		WHERE org_id = $1 OR is_public = true
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list ingredients: %w", err)
	}
	defer rows.Close()

	ingredients := []domain.Ingredient{}
	for rows.Next() {
		var i domain.Ingredient
		if err := rows.Scan(
			&i.ID, &i.Name, &i.FdcID, &i.CaloriesPer100g, &i.ProteinPer100g, &i.FatPer100g, &i.CarbsPer100g,
			&i.DensityGPerMl, &i.OrgID, &i.IsPublic, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan ingredient: %w", err)
		}
		ingredients = append(ingredients, i)
	}
	return ingredients, rows.Err()
}

func (s *IngredientStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.IngredientID) (*domain.Ingredient, error) {
	var i domain.Ingredient
	err := q.QueryRowContext(ctx, `
		SELECT id, name, fdc_id, calories_per_100g, protein_per_100g, fat_per_100g, carbs_per_100g,
		       density_g_per_ml, org_id, is_public, created_by, created_at, updated_by, updated_at
		FROM cove.ingredients
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, id, orgID).Scan(
		&i.ID, &i.Name, &i.FdcID, &i.CaloriesPer100g, &i.ProteinPer100g, &i.FatPer100g, &i.CarbsPer100g,
		&i.DensityGPerMl, &i.OrgID, &i.IsPublic, &i.CreatedBy, &i.CreatedAt, &i.UpdatedBy, &i.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get ingredient: %w", err)
	}
	return &i, nil
}

func (s *IngredientStore) Create(ctx context.Context, q Querier, p domain.IngredientParams) (*domain.Ingredient, error) {
	var id domain.IngredientID
	err := q.QueryRowContext(ctx, `
		INSERT INTO cove.ingredients
		  (name, fdc_id, calories_per_100g, protein_per_100g, fat_per_100g, carbs_per_100g, density_g_per_ml, is_public)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id
	`, p.Name, p.FdcID, p.CaloriesPer100g, p.ProteinPer100g, p.FatPer100g, p.CarbsPer100g, p.DensityGPerMl, p.IsPublic,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create ingredient: %w", err)
	}

	// Identity is guaranteed by service.withScopedTx caller.
	idInfo, _ := domain.IdentityFromContext(ctx)
	return s.Get(ctx, q, idInfo.OrgID, id)
}

func (s *IngredientStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.IngredientID, p domain.IngredientParams) (*domain.Ingredient, error) {
	res, err := q.ExecContext(ctx, `
		UPDATE cove.ingredients
		SET name = $1, fdc_id = $2, calories_per_100g = $3, protein_per_100g = $4,
		    fat_per_100g = $5, carbs_per_100g = $6, density_g_per_ml = $7, is_public = $8
		WHERE id = $9 AND org_id = $10
	`, p.Name, p.FdcID, p.CaloriesPer100g, p.ProteinPer100g, p.FatPer100g, p.CarbsPer100g, p.DensityGPerMl, p.IsPublic, id, orgID)
	if err != nil {
		return nil, fmt.Errorf("update ingredient: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(ctx, q, orgID, id)
}

func (s *IngredientStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, id domain.IngredientID) error {
	res, err := q.ExecContext(ctx, `DELETE FROM cove.ingredients WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return fmt.Errorf("delete ingredient: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
