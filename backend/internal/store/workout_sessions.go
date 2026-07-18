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
		&ws.HealthConnectID, &ws.Source, &ws.SourceActivity,
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
	health_connect_id, source, source_activity,
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

func (s *WorkoutSessionStore) List(ctx context.Context, q Querier, orgID domain.OrgID, userID domain.UserID, from, to, updatedSince *time.Time) ([]*domain.WorkoutSession, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT`+workoutSessionColumns+`
		FROM cove.workout_sessions
		WHERE org_id = $1 AND user_id = $2
		  AND ($3::timestamptz IS NULL OR started_at >= $3)
		  AND ($4::timestamptz IS NULL OR started_at < $4)
		  AND ($5::timestamptz IS NULL OR updated_at >= $5)
		ORDER BY COALESCE(started_at, created_at) DESC
	`, orgID, userID, from, to, updatedSince)
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

func (s *WorkoutSessionStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID, p WorkoutSessionParams, sensitiveData []byte, setSummaryNow bool, updatedAt *time.Time) (*domain.WorkoutSession, error) {
	res, err := q.ExecContext(ctx, `
		UPDATE cove.workout_sessions SET
			program_id = $1,
			activity = $2, duration_s = $3,
			started_at = $4, completed_at = $5,
			sensitive_data = $6,
			summary_generated_at = CASE WHEN $7 THEN NOW() ELSE summary_generated_at END
		WHERE id = $8 AND org_id = $9 AND ($10::timestamptz IS NULL OR updated_at = $10)
	`,
		p.ProgramID,
		p.Activity, p.DurationS,
		p.StartedAt, p.CompletedAt,
		sensitiveData,
		setSummaryNow,
		id, orgID, updatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("update workout session: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		if _, err := s.get(ctx, q, orgID, id); errors.Is(err, ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, ErrConflict
	}
	return s.get(ctx, q, orgID, id)
}

// SetHealthConnectID links a session to a Health Connect record by storing its
// UUID. Uses IS DISTINCT FROM so the bookkeeping trigger does not fire when the
// value is already set to the same UUID (preventing a spurious updated_at bump).
func (s *WorkoutSessionStore) SetHealthConnectID(ctx context.Context, q Querier, orgID domain.OrgID, id domain.WorkoutSessionID, hcID string) error {
	res, err := q.ExecContext(ctx, `
		UPDATE cove.workout_sessions
		SET health_connect_id = $3
		WHERE id = $1 AND org_id = $2 AND health_connect_id IS DISTINCT FROM $3
	`, id, orgID, hcID)
	if err != nil {
		return fmt.Errorf("set health connect id: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		// Either the value was already the same (noop) or the session doesn't
		// exist. Distinguish by attempting a get.
		if _, err := s.get(ctx, q, orgID, id); errors.Is(err, ErrNotFound) {
			return ErrNotFound
		}
		// Noop — HC ID already set to the same value.
	}
	return nil
}

// HealthConnectSessionParams holds the fields provided by the Health Connect
// import path. Source and health_connect_id are managed by the store.
type HealthConnectSessionParams struct {
	SourceActivity *string
	Activity       *string
	DurationS      *int
	StartedAt      *time.Time
	CompletedAt    *time.Time
}

// UpsertHealthConnect creates or updates a session keyed on its Health Connect
// record UUID. Returns the session and true if it was newly created, false if
// updated. Source is fixed to "health_connect" on insert and never changed on
// update. On conflict, only non-NULL fields in p are applied; omitted fields
// retain their current stored values (patch-upsert semantics).
func (s *WorkoutSessionStore) UpsertHealthConnect(ctx context.Context, q Querier, userID domain.UserID, orgID domain.OrgID, hcID string, p HealthConnectSessionParams) (*domain.WorkoutSession, bool, error) {
	var id domain.WorkoutSessionID
	var created bool
	err := q.QueryRowContext(ctx, `
		INSERT INTO cove.workout_sessions (
			user_id,
			health_connect_id, source, source_activity,
			activity, duration_s,
			started_at, completed_at
		) VALUES ($1, $2, 'health_connect', $3, $4, $5, $6, $7)
		ON CONFLICT (health_connect_id) DO UPDATE SET
			source_activity = COALESCE(EXCLUDED.source_activity, workout_sessions.source_activity),
			activity        = COALESCE(EXCLUDED.activity,         workout_sessions.activity),
			duration_s      = COALESCE(EXCLUDED.duration_s,       workout_sessions.duration_s),
			started_at      = COALESCE(EXCLUDED.started_at,       workout_sessions.started_at),
			completed_at    = COALESCE(EXCLUDED.completed_at,     workout_sessions.completed_at)
		RETURNING id, (xmax = 0)
	`,
		userID,
		hcID, p.SourceActivity,
		p.Activity, p.DurationS,
		p.StartedAt, p.CompletedAt,
	).Scan(&id, &created)
	if err != nil {
		return nil, false, fmt.Errorf("upsert health connect session: %w", err)
	}

	ws, err := s.get(ctx, q, orgID, id)
	if err != nil {
		return nil, false, err
	}
	return ws, created, nil
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
