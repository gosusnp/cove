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

type ProgramStore struct{}

func NewProgramStore() *ProgramStore {
	return &ProgramStore{}
}

func (s *ProgramStore) List(ctx context.Context, q Querier, orgID domain.OrgID) ([]domain.ProgramLite, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, org_id, is_public FROM programs 
		WHERE org_id = $1 OR is_public = true
		ORDER BY name
	`, orgID)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	defer rows.Close()

	programs := []domain.ProgramLite{}
	for rows.Next() {
		var p domain.ProgramLite
		if err := rows.Scan(&p.ID, &p.Name, &p.OrgID, &p.IsPublic); err != nil {
			return nil, fmt.Errorf("scan program: %w", err)
		}
		programs = append(programs, p)
	}
	return programs, rows.Err()
}

func (s *ProgramStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID) (*domain.ProgramLite, error) {
	var p domain.ProgramLite
	err := q.QueryRowContext(ctx, `
		SELECT id, name, org_id, is_public FROM programs 
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, id, orgID).
		Scan(&p.ID, &p.Name, &p.OrgID, &p.IsPublic)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program: %w", err)
	}
	return &p, nil
}

func (s *ProgramStore) Create(ctx context.Context, q Querier, name string, description *string, isPublic bool) (*domain.ProgramLite, error) {
	var id domain.ProgramID
	err := q.QueryRowContext(ctx,
		`INSERT INTO programs (name, description, is_public) VALUES ($1, $2, $3) RETURNING id`,
		name, description, isPublic,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create program: %w", err)
	}

	idInfo, _ := domain.IdentityFromContext(ctx)
	return s.Get(ctx, q, idInfo.OrgID, id)
}

func (s *ProgramStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID, name string, description *string, isPublic bool) (*domain.ProgramLite, error) {
	res, err := q.ExecContext(ctx,
		`UPDATE programs SET name = $1, description = $2, is_public = $3 
		 WHERE id = $4 AND org_id = $5`,
		name, description, isPublic, id, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("update program: %w", err)
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

// GetDetail returns the full program hierarchy: sets with their exercises.
// Uses 3 queries: program, sets, and all exercises for those sets.
func (s *ProgramStore) GetDetail(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID) (*domain.Program, error) {
	var p domain.Program
	err := q.QueryRowContext(ctx, `
		SELECT id, name, description, org_id, is_public, created_by, created_at, updated_by, updated_at
		FROM programs 
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, id, orgID).Scan(
		&p.ID, &p.Name, &p.Description, &p.OrgID, &p.IsPublic,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program detail: %w", err)
	}

	p.Sets = []domain.ProgramSet{}

	setRows, err := q.QueryContext(ctx,
		`SELECT id, name, rounds, intra_set_rest_seconds, sort_order FROM program_sets WHERE program_id = $1 ORDER BY sort_order, id`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get program sets: %w", err)
	}
	defer setRows.Close()

	setIndex := map[int64]int{}
	for setRows.Next() {
		var sd domain.ProgramSet
		sd.Exercises = []domain.ProgramExercise{}
		if err := setRows.Scan(&sd.ID, &sd.Name, &sd.Rounds, &sd.IntraSetRestSeconds, &sd.SortOrder); err != nil {
			return nil, fmt.Errorf("scan program set: %w", err)
		}
		setIndex[sd.ID] = len(p.Sets)
		p.Sets = append(p.Sets, sd)
	}
	if err := setRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate program sets: %w", err)
	}

	if len(p.Sets) == 0 {
		return &p, nil
	}

	exRows, err := q.QueryContext(ctx,
		`SELECT pe.id, pe.program_set_id, pe.exercise_id, e.name, pe.laterality, pe.target_reps, pe.target_duration_seconds, pe.target_weight_kg, pe.sort_order
		 FROM program_exercises pe
		 JOIN exercises e ON e.id = pe.exercise_id
		 JOIN program_sets ps ON ps.id = pe.program_set_id
		 WHERE ps.program_id = $1
		 ORDER BY pe.program_set_id, pe.sort_order, pe.id`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get program exercises: %w", err)
	}
	defer exRows.Close()

	for exRows.Next() {
		var ped domain.ProgramExercise
		var setID int64
		if err := exRows.Scan(&ped.ID, &setID, &ped.ExerciseID, &ped.Name, &ped.Laterality, &ped.TargetReps, &ped.TargetDurationSeconds, &ped.TargetWeightKg, &ped.SortOrder); err != nil {
			return nil, fmt.Errorf("scan program exercise: %w", err)
		}
		idx := setIndex[setID]
		p.Sets[idx].Exercises = append(p.Sets[idx].Exercises, ped)
	}
	if err := exRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate program exercises: %w", err)
	}

	return &p, nil
}

func (s *ProgramStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID) error {
	res, err := q.ExecContext(ctx, `DELETE FROM programs WHERE id = $1 AND org_id = $2`, id, orgID)
	if err != nil {
		return fmt.Errorf("delete program: %w", err)
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
