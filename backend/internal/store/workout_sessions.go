// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
)

type WorkoutSessionStore struct{}

func NewWorkoutSessionStore() *WorkoutSessionStore {
	return &WorkoutSessionStore{}
}

type scanner interface {
	Scan(dest ...any) error
}

func scanWorkoutSession(sc scanner) (*domain.WorkoutSession, error) {
	var ws domain.WorkoutSession
	err := sc.Scan(
		&ws.ID, &ws.OrgID, &ws.UserID,
		&ws.ProgramID,
		&ws.Activity, &ws.DurationS,
		&ws.StartedAt, &ws.CompletedAt,
		&ws.SummaryGeneratedAt,
		&ws.CreatedBy, &ws.CreatedAt, &ws.UpdatedBy, &ws.UpdatedAt,
		ws.SensitiveDataScanner(),
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ws, nil
}

const workoutSessionColumns = `
	id, org_id, user_id,
	program_id,
	activity, duration_s,
	started_at, completed_at,
	summary_generated_at,
	created_by, created_at, updated_by, updated_at,
	sensitive_data`

func (s *WorkoutSessionStore) get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	row := q.QueryRowContext(ctx, `
		SELECT`+workoutSessionColumns+`
		FROM cove.workout_sessions
		WHERE id = $1 AND org_id = $2
	`, id, orgID)
	ws, err := scanWorkoutSession(row)
	if err != nil {
		return nil, fmt.Errorf("get workout session: %w", err)
	}
	return ws, nil
}

func (s *WorkoutSessionStore) List(ctx context.Context, q Querier, orgID domain.OrgID, userID domain.UserID) ([]*domain.WorkoutSession, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT`+workoutSessionColumns+`
		FROM cove.workout_sessions
		WHERE org_id = $1 AND user_id = $2
		ORDER BY COALESCE(started_at, created_at) DESC
	`, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("list workout sessions: %w", err)
	}
	defer rows.Close()

	sessions := []*domain.WorkoutSession{}
	for rows.Next() {
		ws, err := scanWorkoutSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout session: %w", err)
		}
		sessions = append(sessions, ws)
	}
	return sessions, rows.Err()
}

func (s *WorkoutSessionStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	return s.get(ctx, q, orgID, id)
}

// WorkoutSessionParams holds the mutable fields for creating or updating a workout session.
type WorkoutSessionParams struct {
	ProgramID     *domain.ProgramID
	Activity      *string
	DurationS     *int
	StartedAt     *time.Time
	CompletedAt   *time.Time
	SensitiveData domain.SessionSensitiveData
}

func (s *WorkoutSessionStore) Create(ctx context.Context, q Querier, p WorkoutSessionParams, sensitiveData []byte) (*domain.WorkoutSession, error) {
	// org_id and created_by are set by the bookkeeping trigger via ScopedQuerier.
	// user_id is the owner of the session and must be set explicitly.
	idInfo, _ := domain.IdentityFromContext(ctx)
	var id domain.WorkoutSessionID
	err := q.QueryRowContext(ctx, `
		INSERT INTO cove.workout_sessions (
			user_id,
			program_id,
			activity, duration_s,
			started_at, completed_at,
			sensitive_data
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`,
		idInfo.UserID,
		p.ProgramID,
		p.Activity, p.DurationS,
		p.StartedAt, p.CompletedAt,
		sensitiveData,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create workout session: %w", err)
	}

	return s.get(ctx, q, idInfo.OrgID, id)
}

func (s *WorkoutSessionStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID, p WorkoutSessionParams, sensitiveData []byte, setSummaryNow bool) (*domain.WorkoutSession, error) {
	res, err := q.ExecContext(ctx, `
		UPDATE cove.workout_sessions SET
			program_id = $1,
			activity = $2, duration_s = $3,
			started_at = $4, completed_at = $5,
			sensitive_data = $6,
			summary_generated_at = CASE WHEN $7 THEN NOW() ELSE summary_generated_at END
		WHERE id = $8 AND org_id = $9
	`,
		p.ProgramID,
		p.Activity, p.DurationS,
		p.StartedAt, p.CompletedAt,
		sensitiveData,
		setSummaryNow,
		id, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("update workout session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.get(ctx, q, orgID, id)
}

func (s *WorkoutSessionStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID) error {
	res, err := q.ExecContext(ctx, `DELETE FROM cove.workout_sessions WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return fmt.Errorf("delete workout session: %w", err)
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
