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
		&ws.ProgramID, &ws.ProgramName, &ws.ProgramStructure,
		&ws.Activity, &ws.DurationS, &ws.PerceivedEffort, &ws.SessionNotes,
		&ws.StartedAt, &ws.CompletedAt,
		&ws.CreatedBy, &ws.CreatedAt, &ws.UpdatedBy, &ws.UpdatedAt,
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
	program_id, program_name, program_structure,
	activity, duration_s, perceived_effort, session_notes,
	started_at, completed_at,
	created_by, created_at, updated_by, updated_at`

func (s *WorkoutSessionStore) get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	row := q.QueryRowContext(ctx, `
		SELECT`+workoutSessionColumns+`
		FROM workout_sessions
		WHERE id = $1 AND org_id = $2
	`, id, orgID)
	ws, err := scanWorkoutSession(row)
	if err != nil {
		return nil, fmt.Errorf("get workout session: %w", err)
	}
	return ws, nil
}

func (s *WorkoutSessionStore) List(ctx context.Context, q Querier, orgID domain.OrgID, userID domain.UserID) ([]domain.WorkoutSession, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT`+workoutSessionColumns+`
		FROM workout_sessions
		WHERE org_id = $1 AND user_id = $2
		ORDER BY COALESCE(started_at, created_at) DESC
	`, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("list workout sessions: %w", err)
	}
	defer rows.Close()

	sessions := []domain.WorkoutSession{}
	for rows.Next() {
		ws, err := scanWorkoutSession(rows)
		if err != nil {
			return nil, fmt.Errorf("scan workout session: %w", err)
		}
		sessions = append(sessions, *ws)
	}
	return sessions, rows.Err()
}

func (s *WorkoutSessionStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID) (*domain.WorkoutSession, error) {
	return s.get(ctx, q, orgID, id)
}

// WorkoutSessionParams holds the mutable fields for creating or updating a workout session.
type WorkoutSessionParams struct {
	ProgramID        *domain.ProgramID
	ProgramName      *string
	ProgramStructure *string
	Activity         *string
	DurationS        *int
	PerceivedEffort  *int
	SessionNotes     *string
	StartedAt        *time.Time
	CompletedAt      *time.Time
}

func (s *WorkoutSessionStore) Create(ctx context.Context, q Querier, p WorkoutSessionParams) (*domain.WorkoutSession, error) {
	// org_id and created_by are set by the bookkeeping trigger via ScopedQuerier.
	// user_id is the owner of the session and must be set explicitly.
	idInfo, _ := domain.IdentityFromContext(ctx)
	var id domain.WorkoutSessionID
	err := q.QueryRowContext(ctx, `
		INSERT INTO workout_sessions (
			user_id,
			program_id, program_name, program_structure,
			activity, duration_s, perceived_effort, session_notes,
			started_at, completed_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id
	`,
		idInfo.UserID,
		p.ProgramID, p.ProgramName, p.ProgramStructure,
		p.Activity, p.DurationS, p.PerceivedEffort, p.SessionNotes,
		p.StartedAt, p.CompletedAt,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create workout session: %w", err)
	}

	return s.get(ctx, q, idInfo.OrgID, id)
}

func (s *WorkoutSessionStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID, p WorkoutSessionParams) (*domain.WorkoutSession, error) {
	res, err := q.ExecContext(ctx, `
		UPDATE workout_sessions SET
			program_id = $1, program_name = $2, program_structure = $3,
			activity = $4, duration_s = $5, perceived_effort = $6, session_notes = $7,
			started_at = $8, completed_at = $9
		WHERE id = $10 AND org_id = $11
	`,
		p.ProgramID, p.ProgramName, p.ProgramStructure,
		p.Activity, p.DurationS, p.PerceivedEffort, p.SessionNotes,
		p.StartedAt, p.CompletedAt,
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
	res, err := q.ExecContext(ctx, `DELETE FROM workout_sessions WHERE id = $1 AND org_id = $2`, id, orgID)
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
