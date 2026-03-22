// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// PreparationStore manages preparation and preparation_ingredient persistence.
type PreparationStore struct{}

func NewPreparationStore() *PreparationStore {
	return &PreparationStore{}
}

// List returns preparations for the org, without ingredients.
func (s *PreparationStore) List(ctx context.Context, q Querier, orgID domain.OrgID) ([]domain.PreparationLite, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, description, yield_amount, yield_unit, is_public, org_id, created_at
		FROM cove.preparations
		WHERE org_id = $1 OR is_public = true
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list preparations: %w", err)
	}
	defer rows.Close()

	list := []domain.PreparationLite{}
	for rows.Next() {
		var p domain.PreparationLite
		if err := rows.Scan(&p.ID, &p.Name, &p.Description, &p.YieldAmount, &p.YieldUnit, &p.IsPublic, &p.OrgID, &p.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan preparation: %w", err)
		}
		list = append(list, p)
	}
	return list, rows.Err()
}

// Get fetches a preparation with its ingredients.
func (s *PreparationStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.PreparationID) (*domain.Preparation, error) {
	var p domain.Preparation
	var stepsJSON []byte
	err := q.QueryRowContext(ctx, `
		SELECT id, name, description, yield_amount, yield_unit, steps, is_public,
		       org_id, created_by, created_at, updated_by, updated_at
		FROM cove.preparations
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, id, orgID).Scan(
		&p.ID, &p.Name, &p.Description, &p.YieldAmount, &p.YieldUnit, &stepsJSON, &p.IsPublic,
		&p.OrgID, &p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get preparation: %w", err)
	}
	p.Steps = []domain.PreparationStep{}
	if err := json.Unmarshal(stepsJSON, &p.Steps); err != nil {
		return nil, fmt.Errorf("unmarshal steps: %w", err)
	}
	ingredients, err := s.listIngredients(ctx, q, orgID, id)
	if err != nil {
		return nil, err
	}
	p.Ingredients = ingredients
	return &p, nil
}

// Create inserts a new preparation. org_id and created_by are set by the bookkeeping trigger.
func (s *PreparationStore) Create(ctx context.Context, q Querier, p domain.PreparationParams) (*domain.Preparation, error) {
	stepsJSON, err := json.Marshal(p.Steps)
	if err != nil {
		return nil, fmt.Errorf("marshal steps: %w", err)
	}
	var id domain.PreparationID
	err = q.QueryRowContext(ctx, `
		INSERT INTO cove.preparations (name, description, yield_amount, yield_unit, steps, is_public)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`, p.Name, p.Description, p.YieldAmount, p.YieldUnit, stepsJSON, p.IsPublic).Scan(&id)
	if isUniqueViolation(err) {
		return nil, ErrDuplicate
	}
	if err != nil {
		return nil, fmt.Errorf("create preparation: %w", err)
	}
	identity, _ := domain.IdentityFromContext(ctx)
	return s.Get(ctx, q, identity.OrgID, id)
}

// Update modifies a preparation, explicitly filtering by org_id for defense-in-depth.
func (s *PreparationStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.PreparationID, p domain.PreparationParams) (*domain.Preparation, error) {
	stepsJSON, err := json.Marshal(p.Steps)
	if err != nil {
		return nil, fmt.Errorf("marshal steps: %w", err)
	}
	res, err := q.ExecContext(ctx, `
		UPDATE cove.preparations
		SET name = $1, description = $2, yield_amount = $3, yield_unit = $4, steps = $5, is_public = $6
		WHERE id = $7 AND org_id = $8
	`, p.Name, p.Description, p.YieldAmount, p.YieldUnit, stepsJSON, p.IsPublic, id, orgID)
	if err != nil {
		return nil, fmt.Errorf("update preparation: %w", err)
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

// Delete removes a preparation. Ingredients cascade via ON DELETE CASCADE.
func (s *PreparationStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, id domain.PreparationID) error {
	res, err := q.ExecContext(ctx, `
		DELETE FROM cove.preparations WHERE id = $1 AND org_id = $2
	`, id, orgID)
	if err != nil {
		return fmt.Errorf("delete preparation: %w", err)
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

// listIngredients fetches all ingredients for a preparation.
func (s *PreparationStore) listIngredients(ctx context.Context, q Querier, orgID domain.OrgID, preparationID domain.PreparationID) ([]domain.PreparationIngredient, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, preparation_id, ingredient_id, name, amount, unit, prep
		FROM cove.preparation_ingredients
		WHERE preparation_id = $1 AND org_id = $2
		ORDER BY id
	`, preparationID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list preparation ingredients: %w", err)
	}
	defer rows.Close()

	list := []domain.PreparationIngredient{}
	for rows.Next() {
		var i domain.PreparationIngredient
		if err := rows.Scan(&i.ID, &i.PreparationID, &i.IngredientID, &i.Name, &i.Amount, &i.Unit, &i.Prep); err != nil {
			return nil, fmt.Errorf("scan preparation ingredient: %w", err)
		}
		list = append(list, i)
	}
	return list, rows.Err()
}

// AddIngredient inserts an ingredient into a preparation. Verifies preparation ownership via subquery.
func (s *PreparationStore) AddIngredient(ctx context.Context, q Querier, orgID domain.OrgID, preparationID domain.PreparationID, p domain.PreparationIngredientParams) (*domain.PreparationIngredient, error) {
	var id domain.PreparationIngredientID
	err := q.QueryRowContext(ctx, `
		INSERT INTO cove.preparation_ingredients (preparation_id, ingredient_id, name, amount, unit, prep)
		SELECT $1, $2, $3, $4, $5, $6
		WHERE EXISTS (SELECT 1 FROM cove.preparations WHERE id = $1 AND org_id = $7)
		RETURNING id
	`, preparationID, p.IngredientID, p.Name, p.Amount, p.Unit, p.Prep, orgID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("add preparation ingredient: %w", err)
	}
	return &domain.PreparationIngredient{
		ID:            id,
		PreparationID: preparationID,
		IngredientID:  p.IngredientID,
		Name:          p.Name,
		Amount:        p.Amount,
		Unit:          p.Unit,
		Prep:          p.Prep,
	}, nil
}

// UpdateIngredient modifies a preparation ingredient, filtering by preparation_id and org_id for defense-in-depth.
func (s *PreparationStore) UpdateIngredient(ctx context.Context, q Querier, orgID domain.OrgID, preparationID domain.PreparationID, id domain.PreparationIngredientID, p domain.PreparationIngredientParams) (*domain.PreparationIngredient, error) {
	var updated domain.PreparationIngredient
	err := q.QueryRowContext(ctx, `
		UPDATE cove.preparation_ingredients
		SET name = $1, amount = $2, unit = $3, prep = $4
		WHERE id = $5 AND preparation_id = $6 AND org_id = $7
		RETURNING id, preparation_id, ingredient_id, name, amount, unit, prep
	`, p.Name, p.Amount, p.Unit, p.Prep, id, preparationID, orgID).Scan(
		&updated.ID, &updated.PreparationID, &updated.IngredientID,
		&updated.Name, &updated.Amount, &updated.Unit, &updated.Prep,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("update preparation ingredient: %w", err)
	}
	return &updated, nil
}

// DeleteIngredient removes a preparation ingredient, filtering by preparation_id and org_id for defense-in-depth.
func (s *PreparationStore) DeleteIngredient(ctx context.Context, q Querier, orgID domain.OrgID, preparationID domain.PreparationID, id domain.PreparationIngredientID) error {
	res, err := q.ExecContext(ctx, `
		DELETE FROM cove.preparation_ingredients WHERE id = $1 AND preparation_id = $2 AND org_id = $3
	`, id, preparationID, orgID)
	if err != nil {
		return fmt.Errorf("delete preparation ingredient: %w", err)
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
