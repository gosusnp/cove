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

// listIngredients fetches all ingredients for a preparation, including ingredient density.
func (s *PreparationStore) listIngredients(ctx context.Context, q Querier, orgID domain.OrgID, preparationID domain.PreparationID) ([]domain.PreparationIngredient, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT pi.id, pi.preparation_id, pi.ingredient_id, pi.preparation_ref_id,
		       pi.name, pi.amount, pi.unit, pi.prep,
		       ing.density_g_per_ml
		FROM cove.preparation_ingredients pi
		LEFT JOIN cove.ingredients ing ON ing.id = pi.ingredient_id
		WHERE pi.preparation_id = $1 AND pi.org_id = $2
		ORDER BY pi.id
	`, preparationID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list preparation ingredients: %w", err)
	}
	defer rows.Close()

	list := []domain.PreparationIngredient{}
	for rows.Next() {
		var i domain.PreparationIngredient
		if err := rows.Scan(&i.ID, &i.PreparationID, &i.IngredientID, &i.PreparationRefID,
			&i.Name, &i.Amount, &i.Unit, &i.Prep, &i.DensityGPerMl); err != nil {
			return nil, fmt.Errorf("scan preparation ingredient: %w", err)
		}
		list = append(list, i)
	}
	return list, rows.Err()
}

// getIngredient fetches a single preparation ingredient by id, including density from the linked ingredient.
func (s *PreparationStore) getIngredient(ctx context.Context, q Querier, orgID domain.OrgID, id domain.PreparationIngredientID) (*domain.PreparationIngredient, error) {
	var i domain.PreparationIngredient
	err := q.QueryRowContext(ctx, `
		SELECT pi.id, pi.preparation_id, pi.ingredient_id, pi.preparation_ref_id,
		       pi.name, pi.amount, pi.unit, pi.prep,
		       ing.density_g_per_ml
		FROM cove.preparation_ingredients pi
		LEFT JOIN cove.ingredients ing ON ing.id = pi.ingredient_id
		WHERE pi.id = $1 AND pi.org_id = $2
	`, id, orgID).Scan(&i.ID, &i.PreparationID, &i.IngredientID, &i.PreparationRefID,
		&i.Name, &i.Amount, &i.Unit, &i.Prep, &i.DensityGPerMl)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get preparation ingredient: %w", err)
	}
	return &i, nil
}

// AddIngredient inserts an ingredient (or sub-preparation ref) into a preparation.
// Verifies preparation ownership via subquery. Exactly one of p.IngredientID or p.PreparationRefID must be non-nil.
func (s *PreparationStore) AddIngredient(ctx context.Context, q Querier, orgID domain.OrgID, preparationID domain.PreparationID, p domain.PreparationIngredientParams) (*domain.PreparationIngredient, error) {
	var id domain.PreparationIngredientID
	err := q.QueryRowContext(ctx, `
		INSERT INTO cove.preparation_ingredients (preparation_id, ingredient_id, preparation_ref_id, name, amount, unit, prep)
		SELECT $1, $2, $3, $4, $5, $6, $7
		WHERE EXISTS (SELECT 1 FROM cove.preparations WHERE id = $1 AND org_id = $8)
		RETURNING id
	`, preparationID, p.IngredientID, p.PreparationRefID, p.Name, p.Amount, p.Unit, p.Prep, orgID).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("add preparation ingredient: %w", err)
	}
	return s.getIngredient(ctx, q, orgID, id)
}

// HasCircularRef reports whether adding an edge fromPrepID → toPrepID would create a cycle.
// It traverses the existing sub-preparation graph starting from toPrepID and checks if
// fromPrepID is reachable.
func (s *PreparationStore) HasCircularRef(ctx context.Context, q Querier, fromPrepID, toPrepID domain.PreparationID) (bool, error) {
	if fromPrepID == toPrepID {
		return true, nil
	}
	var exists bool
	err := q.QueryRowContext(ctx, `
		WITH RECURSIVE deps(id) AS (
			SELECT preparation_ref_id AS id
			FROM cove.preparation_ingredients
			WHERE preparation_id = $2 AND preparation_ref_id IS NOT NULL
			UNION
			SELECT pi.preparation_ref_id
			FROM cove.preparation_ingredients pi
			JOIN deps d ON pi.preparation_id = d.id
			WHERE pi.preparation_ref_id IS NOT NULL
		)
		SELECT EXISTS(SELECT 1 FROM deps WHERE id = $1)
	`, fromPrepID, toPrepID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("has circular ref: %w", err)
	}
	return exists, nil
}

// UpdateIngredient modifies a preparation ingredient, filtering by preparation_id and org_id for defense-in-depth.
func (s *PreparationStore) UpdateIngredient(ctx context.Context, q Querier, orgID domain.OrgID, preparationID domain.PreparationID, id domain.PreparationIngredientID, p domain.PreparationIngredientParams) (*domain.PreparationIngredient, error) {
	res, err := q.ExecContext(ctx, `
		UPDATE cove.preparation_ingredients
		SET name = $1, amount = $2, unit = $3, prep = $4,
		    ingredient_id = $5, preparation_ref_id = $6
		WHERE id = $7 AND preparation_id = $8 AND org_id = $9
	`, p.Name, p.Amount, p.Unit, p.Prep, p.IngredientID, p.PreparationRefID, id, preparationID, orgID)
	if err != nil {
		return nil, fmt.Errorf("update preparation ingredient: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.getIngredient(ctx, q, orgID, id)
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
