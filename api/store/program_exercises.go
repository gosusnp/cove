// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"database/sql"
	"errors"
	"fmt"
)

type ProgramExerciseStore struct {
	db *sql.DB
}

func NewProgramExerciseStore(db *sql.DB) *ProgramExerciseStore {
	return &ProgramExerciseStore{db: db}
}

const programExerciseColumns = `id, program_set_id, exercise_id, laterality, target_reps, target_duration_seconds, target_weight_kg, sort_order`

func scanProgramExercise(row interface{ Scan(...any) error }) (*ProgramExercise, error) {
	var pe ProgramExercise
	if err := row.Scan(&pe.ID, &pe.ProgramSetID, &pe.ExerciseID, &pe.Laterality, &pe.TargetReps, &pe.TargetDurationSeconds, &pe.TargetWeightKg, &pe.SortOrder); err != nil {
		return nil, err
	}
	return &pe, nil
}

func (s *ProgramExerciseStore) List(programSetID int64) ([]ProgramExercise, error) {
	rows, err := s.db.Query(
		`SELECT `+programExerciseColumns+` FROM program_exercises WHERE program_set_id = ? ORDER BY sort_order, id`,
		programSetID,
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

func (s *ProgramExerciseStore) Get(programSetID, id int64) (*ProgramExercise, error) {
	row := s.db.QueryRow(
		`SELECT `+programExerciseColumns+` FROM program_exercises WHERE id = ? AND program_set_id = ?`,
		id, programSetID,
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

func (s *ProgramExerciseStore) Create(programSetID, exerciseID int64, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*ProgramExercise, error) {
	res, err := s.db.Exec(
		`INSERT INTO program_exercises (program_set_id, exercise_id, laterality, target_reps, target_duration_seconds, target_weight_kg, sort_order) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		programSetID, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder,
	)
	if err != nil {
		return nil, fmt.Errorf("create program exercise: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return s.Get(programSetID, id)
}

func (s *ProgramExerciseStore) Update(programSetID, id, exerciseID int64, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*ProgramExercise, error) {
	res, err := s.db.Exec(
		`UPDATE program_exercises SET exercise_id = ?, laterality = ?, target_reps = ?, target_duration_seconds = ?, target_weight_kg = ?, sort_order = ? WHERE id = ? AND program_set_id = ?`,
		exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder, id, programSetID,
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
	return s.Get(programSetID, id)
}

func (s *ProgramExerciseStore) Delete(programSetID, id int64) error {
	res, err := s.db.Exec(
		`DELETE FROM program_exercises WHERE id = ? AND program_set_id = ?`,
		id, programSetID,
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
