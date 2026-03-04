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

type ExerciseStore struct{}

func NewExerciseStore() *ExerciseStore {
	return &ExerciseStore{}
}

func (s *ExerciseStore) List(ctx context.Context, q Querier, orgID domain.OrgID) ([]domain.Exercise, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, progression, description, org_id, is_public, created_by, created_at, updated_by, updated_at
		FROM exercises 
		WHERE org_id = $1 OR is_public = true
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}
	defer rows.Close()

	exercises := []domain.Exercise{}
	for rows.Next() {
		var e domain.Exercise
		if err := rows.Scan(
			&e.ID, &e.Name, &e.Progression, &e.Description, &e.OrgID, &e.IsPublic,
			&e.CreatedBy, &e.CreatedAt, &e.UpdatedBy, &e.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan exercise: %w", err)
		}
		exercises = append(exercises, e)
	}
	return exercises, rows.Err()
}

func (s *ExerciseStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ExerciseID) (*domain.Exercise, error) {
	var e domain.Exercise
	err := q.QueryRowContext(ctx, `
		SELECT id, name, progression, description, org_id, is_public, created_by, created_at, updated_by, updated_at
		FROM exercises 
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, id, orgID).Scan(
		&e.ID, &e.Name, &e.Progression, &e.Description, &e.OrgID, &e.IsPublic,
		&e.CreatedBy, &e.CreatedAt, &e.UpdatedBy, &e.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get exercise: %w", err)
	}
	return &e, nil
}

func (s *ExerciseStore) Create(ctx context.Context, q Querier, name string, progression *string, description *string, isPublic bool) (*domain.Exercise, error) {
	var id domain.ExerciseID
	// Note: org_id and created_by are handled by the trigger via ScopedQuerier session variables.
	err := q.QueryRowContext(ctx,
		`INSERT INTO exercises (name, progression, description, is_public) VALUES ($1, $2, $3, $4) RETURNING id`,
		name, progression, description, isPublic,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create exercise: %w", err)
	}

	// Identity is guaranteed by service.withScopedTx caller.
	idInfo, _ := domain.IdentityFromContext(ctx)
	return s.Get(ctx, q, idInfo.OrgID, id)
}

func (s *ExerciseStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ExerciseID, name string, progression *string, description *string, isPublic bool) (*domain.Exercise, error) {
	res, err := q.ExecContext(ctx,
		`UPDATE exercises SET name = $1, progression = $2, description = $3, is_public = $4 
		 WHERE id = $5 AND org_id = $6`,
		name, progression, description, isPublic, id, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("update exercise: %w", err)
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

func (s *ExerciseStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ExerciseID) error {
	res, err := q.ExecContext(ctx, `DELETE FROM exercises WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return fmt.Errorf("delete exercise: %w", err)
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
