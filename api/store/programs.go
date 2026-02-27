// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"database/sql"
	"errors"
	"fmt"
)

type ProgramStore struct {
	db *sql.DB
}

func NewProgramStore(db *sql.DB) *ProgramStore {
	return &ProgramStore{db: db}
}

func (s *ProgramStore) List() ([]Program, error) {
	rows, err := s.db.Query(`SELECT id, name FROM programs ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("list programs: %w", err)
	}
	defer rows.Close()

	programs := []Program{}
	for rows.Next() {
		var p Program
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, fmt.Errorf("scan program: %w", err)
		}
		programs = append(programs, p)
	}
	return programs, rows.Err()
}

func (s *ProgramStore) Get(id int64) (*Program, error) {
	var p Program
	err := s.db.QueryRow(`SELECT id, name FROM programs WHERE id = ?`, id).
		Scan(&p.ID, &p.Name)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program: %w", err)
	}
	return &p, nil
}

func (s *ProgramStore) Create(name string) (*Program, error) {
	res, err := s.db.Exec(`INSERT INTO programs (name) VALUES (?)`, name)
	if err != nil {
		return nil, fmt.Errorf("create program: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("last insert id: %w", err)
	}
	return s.Get(id)
}

func (s *ProgramStore) Update(id int64, name string) (*Program, error) {
	res, err := s.db.Exec(`UPDATE programs SET name = ? WHERE id = ?`, name, id)
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
func (s *ProgramStore) GetDetail(id int64) (*ProgramDetail, error) {
	p, err := s.Get(id)
	if err != nil {
		return nil, err
	}

	detail := &ProgramDetail{ID: p.ID, Name: p.Name, Sets: []ProgramSetDetail{}}

	setRows, err := s.db.Query(
		`SELECT id, name, rounds, intra_set_rest_seconds, sort_order FROM program_sets WHERE program_id = ? ORDER BY sort_order, id`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get program sets: %w", err)
	}
	defer setRows.Close()

	setIndex := map[int64]int{}
	for setRows.Next() {
		var sd ProgramSetDetail
		sd.Exercises = []ProgramExerciseDetail{}
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
		 WHERE ps.program_id = ?
		 ORDER BY pe.program_set_id, pe.sort_order, pe.id`,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("get program exercises: %w", err)
	}
	defer exRows.Close()

	for exRows.Next() {
		var ped ProgramExerciseDetail
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

func (s *ProgramStore) Delete(id int64) error {
	res, err := s.db.Exec(`DELETE FROM programs WHERE id = ?`, id)
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
