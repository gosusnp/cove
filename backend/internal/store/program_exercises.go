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

type ProgramExerciseStore struct{}

func NewProgramExerciseStore() *ProgramExerciseStore {
	return &ProgramExerciseStore{}
}

const programExerciseColumns = `pe.id, pe.program_set_id, pe.exercise_id, pe.laterality, pe.target_reps, pe.target_duration_seconds, pe.target_weight_kg, pe.sort_order`

func scanProgramExercise(row interface{ Scan(...any) error }) (*ProgramExercise, error) {
	var pe ProgramExercise
	if err := row.Scan(&pe.ID, &pe.ProgramSetID, &pe.ExerciseID, &pe.Laterality, &pe.TargetReps, &pe.TargetDurationSeconds, &pe.TargetWeightKg, &pe.SortOrder); err != nil {
		return nil, err
	}
	return &pe, nil
}

func (s *ProgramExerciseStore) List(ctx context.Context, q Querier, orgID domain.OrgID, programSetID int64) ([]ProgramExercise, error) {
	rows, err := q.QueryContext(ctx,
		`SELECT `+programExerciseColumns+`
		 FROM program_exercises pe
		 JOIN program_sets ps ON ps.id = pe.program_set_id
		 JOIN programs p ON p.id = ps.program_id
		 WHERE pe.program_set_id = $1 AND (p.org_id = $2 OR p.is_public = true)
		 ORDER BY pe.sort_order, pe.id`,
		programSetID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("list program exercises: %w", err)
	}
	defer rows.Close()

	exercises := []ProgramExercise{}
	for rows.Next() {
		pe, err := scanProgramExercise(rows)
		if err != nil {
			return nil, fmt.Errorf("scan program exercise: %w", err)
		}
		exercises = append(exercises, *pe)
	}
	return exercises, rows.Err()
}

func (s *ProgramExerciseStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, programSetID, id int64) (*ProgramExercise, error) {
	row := q.QueryRowContext(ctx,
		`SELECT `+programExerciseColumns+`
		 FROM program_exercises pe
		 JOIN program_sets ps ON ps.id = pe.program_set_id
		 JOIN programs p ON p.id = ps.program_id
		 WHERE pe.id = $1 AND pe.program_set_id = $2 AND (p.org_id = $3 OR p.is_public = true)`,
		id, programSetID, orgID,
	)
	pe, err := scanProgramExercise(row)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program exercise: %w", err)
	}
	return pe, nil
}

func (s *ProgramExerciseStore) Create(ctx context.Context, q Querier, orgID domain.OrgID, programSetID int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*ProgramExercise, error) {
	var id int64
	// We use a subquery to ensure the program set belongs to the org
	err := q.QueryRowContext(ctx,
		`INSERT INTO program_exercises (program_set_id, exercise_id, laterality, target_reps, target_duration_seconds, target_weight_kg, sort_order)
		 SELECT $1, $2, $3, $4, $5, $6, $7
		 WHERE EXISTS (
		 	SELECT 1 FROM program_sets ps
		 	JOIN programs p ON p.id = ps.program_id
		 	WHERE ps.id = $1 AND p.org_id = $8
		 )
		 RETURNING id`,
		programSetID, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder, orgID,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		// Could mean program set not found or not owned by org
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("create program exercise: %w", err)
	}
	return s.Get(ctx, q, orgID, programSetID, id)
}

func (s *ProgramExerciseStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, programSetID, id int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*ProgramExercise, error) {
	res, err := q.ExecContext(ctx,
		`UPDATE program_exercises pe
		 SET exercise_id = $1, laterality = $2, target_reps = $3, target_duration_seconds = $4, target_weight_kg = $5, sort_order = $6
		 FROM program_sets ps
		 JOIN programs p ON p.id = ps.program_id
		 WHERE pe.id = $7 AND pe.program_set_id = $8
		   AND pe.program_set_id = ps.id
		   AND p.org_id = $9`,
		exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder, id, programSetID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("update program exercise: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	return s.Get(ctx, q, orgID, programSetID, id)
}

func (s *ProgramExerciseStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, programSetID, id int64) error {
	res, err := q.ExecContext(ctx,
		`DELETE FROM program_exercises pe
		 USING program_sets ps, programs p
		 WHERE pe.id = $1 AND pe.program_set_id = $2
		   AND pe.program_set_id = ps.id
		   AND ps.program_id = p.id
		   AND p.org_id = $3`,
		id, programSetID, orgID,
	)
	if err != nil {
		return fmt.Errorf("delete program exercise: %w", err)
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
