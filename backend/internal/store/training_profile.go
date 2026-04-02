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

type TrainingProfileStore struct{}

func NewTrainingProfileStore() *TrainingProfileStore {
	return &TrainingProfileStore{}
}

func scanTrainingProfile(sc scanner) (*domain.UserTrainingProfile, error) {
	var tp domain.UserTrainingProfile
	err := sc.Scan(
		&tp.UserID, &tp.OrgID,
		&tp.CreatedBy, &tp.CreatedAt, &tp.UpdatedBy, &tp.UpdatedAt,
		tp.SensitiveDataScanner(),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &tp, nil
}

const trainingProfileColumns = `
	user_id, org_id,
	created_by, created_at, updated_by, updated_at,
	training_profile`

func (s *TrainingProfileStore) Get(ctx context.Context, q Querier, userID domain.UserID) (*domain.UserTrainingProfile, error) {
	row := q.QueryRowContext(ctx, `
		SELECT `+trainingProfileColumns+`
		FROM cove.user_training_profiles
		WHERE user_id = $1
	`, userID)
	tp, err := scanTrainingProfile(row)
	if err != nil {
		return nil, fmt.Errorf("get training profile: %w", err)
	}
	return tp, nil
}

func (s *TrainingProfileStore) Upsert(ctx context.Context, q Querier, userID domain.UserID, orgID domain.OrgID, sensitiveData []byte) (*domain.UserTrainingProfile, error) {
	// org_id and created_by are set by the bookkeeping trigger via ScopedQuerier,
	// but we include org_id explicitly for defense-in-depth.
	_, err := q.ExecContext(ctx, `
		INSERT INTO cove.user_training_profiles (
			user_id, org_id, training_profile
		) VALUES ($1, $2, $3)
		ON CONFLICT (user_id) DO UPDATE SET
			training_profile = EXCLUDED.training_profile,
			org_id = EXCLUDED.org_id
	`, userID, orgID, sensitiveData)
	if err != nil {
		return nil, fmt.Errorf("upsert training profile: %w", err)
	}
	return s.Get(ctx, q, userID)
}

func (s *TrainingProfileStore) Delete(ctx context.Context, q Querier, userID domain.UserID) error {
	res, err := q.ExecContext(ctx, `DELETE FROM cove.user_training_profiles WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("delete training profile: %w", err)
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
