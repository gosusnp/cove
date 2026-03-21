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

// RecipeStore manages recipe and recipe_preparation persistence.
type RecipeStore struct{}

func NewRecipeStore() *RecipeStore {
	return &RecipeStore{}
}

// List returns recipes for the org, without preparations.
func (s *RecipeStore) List(ctx context.Context, q Querier, orgID domain.OrgID) ([]domain.RecipeLite, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, description, servings, is_public, org_id, created_at
		FROM cove.recipes
		WHERE org_id = $1 OR is_public = true
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list recipes: %w", err)
	}
	defer rows.Close()

	list := []domain.RecipeLite{}
	for rows.Next() {
		var r domain.RecipeLite
		if err := rows.Scan(&r.ID, &r.Name, &r.Description, &r.Servings, &r.IsPublic, &r.OrgID, &r.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan recipe: %w", err)
		}
		list = append(list, r)
	}
	return list, rows.Err()
}

// Get fetches a recipe with its preparations.
func (s *RecipeStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.RecipeID) (*domain.Recipe, error) {
	var r domain.Recipe
	err := q.QueryRowContext(ctx, `
		SELECT id, name, description, yield_amount, yield_unit, servings, is_public,
		       org_id, created_by, created_at, updated_by, updated_at
		FROM cove.recipes
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, id, orgID).Scan(
		&r.ID, &r.Name, &r.Description, &r.YieldAmount, &r.YieldUnit, &r.Servings, &r.IsPublic,
		&r.OrgID, &r.CreatedBy, &r.CreatedAt, &r.UpdatedBy, &r.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get recipe: %w", err)
	}
	preparations, err := s.listPreparations(ctx, q, orgID, id)
	if err != nil {
		return nil, err
	}
	r.Preparations = preparations
	return &r, nil
}

// Create inserts a new recipe. org_id and created_by are set by the bookkeeping trigger.
func (s *RecipeStore) Create(ctx context.Context, q Querier, p domain.RecipeParams) (*domain.Recipe, error) {
	var id domain.RecipeID
	err := q.QueryRowContext(ctx, `
		INSERT INTO cove.recipes (name, description, yield_amount, yield_unit, servings, is_public)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, p.Name, p.Description, p.YieldAmount, p.YieldUnit, p.Servings, p.IsPublic).Scan(&id)
	if isUniqueViolation(err) {
		return nil, ErrDuplicate
	}
	if err != nil {
		return nil, fmt.Errorf("create recipe: %w", err)
	}
	identity, _ := domain.IdentityFromContext(ctx)
	return s.Get(ctx, q, identity.OrgID, id)
}

// Update modifies a recipe, explicitly filtering by org_id for defense-in-depth.
func (s *RecipeStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.RecipeID, p domain.RecipeParams) (*domain.Recipe, error) {
	res, err := q.ExecContext(ctx, `
		UPDATE cove.recipes
		SET name = $1, description = $2, yield_amount = $3, yield_unit = $4, servings = $5, is_public = $6
		WHERE id = $7 AND org_id = $8
	`, p.Name, p.Description, p.YieldAmount, p.YieldUnit, p.Servings, p.IsPublic, id, orgID)
	if err != nil {
		return nil, fmt.Errorf("update recipe: %w", err)
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

// Delete removes a recipe. recipe_preparations cascade via ON DELETE CASCADE.
func (s *RecipeStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, id domain.RecipeID) error {
	res, err := q.ExecContext(ctx, `
		DELETE FROM cove.recipes WHERE id = $1 AND org_id = $2
	`, id, orgID)
	if err != nil {
		return fmt.Errorf("delete recipe: %w", err)
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

// listPreparations fetches all preparations linked to a recipe, ordered by position.
func (s *RecipeStore) listPreparations(ctx context.Context, q Querier, orgID domain.OrgID, recipeID domain.RecipeID) ([]domain.RecipePreparation, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, recipe_id, preparation_id, position, amount, unit
		FROM cove.recipe_preparations
		WHERE recipe_id = $1 AND org_id = $2
		ORDER BY position
	`, recipeID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list recipe preparations: %w", err)
	}
	defer rows.Close()

	list := []domain.RecipePreparation{}
	for rows.Next() {
		var rp domain.RecipePreparation
		if err := rows.Scan(&rp.ID, &rp.RecipeID, &rp.PreparationID, &rp.Position, &rp.Amount, &rp.Unit); err != nil {
			return nil, fmt.Errorf("scan recipe preparation: %w", err)
		}
		list = append(list, rp)
	}
	return list, rows.Err()
}

// AddPreparation links a preparation to a recipe. Verifies recipe ownership via subquery.
func (s *RecipeStore) AddPreparation(ctx context.Context, q Querier, orgID domain.OrgID, recipeID domain.RecipeID, p domain.RecipePreparationParams) (*domain.RecipePreparation, error) {
	var id domain.RecipePreparationID
	err := q.QueryRowContext(ctx, `
		INSERT INTO cove.recipe_preparations (recipe_id, preparation_id, position, amount, unit)
		SELECT $1, $2, $3, $4, $5
		WHERE EXISTS (SELECT 1 FROM cove.recipes WHERE id = $1 AND org_id = $6)
		RETURNING id
	`, recipeID, p.PreparationID, p.Position, p.Amount, p.Unit, orgID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("add recipe preparation: %w", err)
	}
	return &domain.RecipePreparation{
		ID:            id,
		RecipeID:      recipeID,
		PreparationID: p.PreparationID,
		Position:      p.Position,
		Amount:        p.Amount,
		Unit:          p.Unit,
	}, nil
}

// UpdatePreparation modifies a recipe preparation link, filtering by org_id for defense-in-depth.
func (s *RecipeStore) UpdatePreparation(ctx context.Context, q Querier, orgID domain.OrgID, id domain.RecipePreparationID, p domain.RecipePreparationParams) (*domain.RecipePreparation, error) {
	var rp domain.RecipePreparation
	err := q.QueryRowContext(ctx, `
		UPDATE cove.recipe_preparations
		SET position = $1, amount = $2, unit = $3
		WHERE id = $4 AND org_id = $5
		RETURNING id, recipe_id, preparation_id, position, amount, unit
	`, p.Position, p.Amount, p.Unit, id, orgID).Scan(
		&rp.ID, &rp.RecipeID, &rp.PreparationID, &rp.Position, &rp.Amount, &rp.Unit,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update recipe preparation: %w", err)
	}
	return &rp, nil
}

// DeletePreparation removes a recipe preparation link, filtering by org_id for defense-in-depth.
func (s *RecipeStore) DeletePreparation(ctx context.Context, q Querier, orgID domain.OrgID, id domain.RecipePreparationID) error {
	res, err := q.ExecContext(ctx, `
		DELETE FROM cove.recipe_preparations WHERE id = $1 AND org_id = $2
	`, id, orgID)
	if err != nil {
		return fmt.Errorf("delete recipe preparation: %w", err)
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
