// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
)

type ProgramStore struct {
	db *sql.DB
}

func NewProgramStore(db *sql.DB) *ProgramStore {
	return &ProgramStore{db: db}
}

func (s *ProgramStore) List() ([]domain.ProgramLite, error) {
	rows, err := s.db.Query(`SELECT id, name FROM programs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	defer rows.Close()

	programs := []domain.ProgramLite{}
	for rows.Next() {
		var p domain.ProgramLite
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, fmt.Errorf("scan program: %w", err)
		}
		programs = append(programs, p)
	}
	return programs, rows.Err()
}

func (s *ProgramStore) Get(id domain.ProgramID) (*domain.ProgramLite, error) {
	var p domain.ProgramLite
	err := s.db.QueryRow(`SELECT id, name FROM programs WHERE id = $1`, id).
		Scan(&p.ID, &p.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program: %w", err)
	}
	return &p, nil
}

func (s *ProgramStore) Create(name string) (*domain.ProgramLite, error) {
	var id domain.ProgramID
	err := s.db.QueryRow(
		`INSERT INTO programs (name) VALUES ($1) RETURNING id`, name,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create program: %w", err)
	}
	return s.Get(id)
}

func (s *ProgramStore) Update(id domain.ProgramID, name string) (*domain.ProgramLite, error) {
	res, err := s.db.Exec(`UPDATE programs SET name = $1 WHERE id = $2`, name, id)
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
	return s.Get(id)
}

// GetDetail returns the full program hierarchy: sets with their exercises.
// Uses 3 queries: program, sets, and all exercises for those sets.
func (s *ProgramStore) GetDetail(id domain.ProgramID) (*domain.Program, error) {
	p, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	detail := &domain.Program{ID: p.ID, Name: p.Name, Sets: []domain.ProgramSet{}}

	setRows, err := s.db.Query(
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
		setIndex[sd.ID] = len(detail.Sets)
		detail.Sets = append(detail.Sets, sd)
	}
	if err := setRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate program sets: %w", err)
	}

	if len(detail.Sets) == 0 {
		return detail, nil
	}

	exRows, err := s.db.Query(
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
		detail.Sets[idx].Exercises = append(detail.Sets[idx].Exercises, ped)
	}
	if err := exRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate program exercises: %w", err)
	}

	return detail, nil
}

func (s *ProgramStore) Delete(id domain.ProgramID) error {
	res, err := s.db.Exec(`DELETE FROM programs WHERE id = $1`, id)
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
