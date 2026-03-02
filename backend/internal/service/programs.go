// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gosusnp/cove/backend/internal/store"
)

type ProgramExerciseInput struct {
	ExerciseID            int64    `json:"exercise_id"`
	Laterality            *string  `json:"laterality,omitempty"`
	TargetReps            *int     `json:"reps,omitempty"`
	TargetDurationSeconds *int     `json:"duration_s,omitempty"`
	TargetWeightKg        *float64 `json:"weight_kg,omitempty"`
	SortOrder             *int     `json:"sort_order,omitempty"`
}

type ProgramSetInput struct {
	Name                *string                `json:"name,omitempty"`
	Rounds              int                    `json:"rounds"`
	IntraSetRestSeconds *int                   `json:"rest_s,omitempty"`
	SortOrder           *int                   `json:"sort_order,omitempty"`
	Exercises           []ProgramExerciseInput `json:"exercises"`
}

// ProgramService holds a direct *sql.DB reference in addition to the store
// because CreateFull requires a transaction that spans programs, program_sets,
// and program_exercises. Other services only need their own store.
type ProgramService struct {
	db    *sql.DB
	store *store.ProgramStore
}

func NewProgramService(db *sql.DB) *ProgramService {
	return &ProgramService{db: db, store: store.NewProgramStore(db)}
}

func (s *ProgramService) List() ([]store.Program, error) {
	return s.store.List()
}

func (s *ProgramService) GetDetail(id int64) (*store.ProgramDetail, error) {
	p, err := s.store.GetDetail(id)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *ProgramService) Create(name string) (*store.Program, error) {
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	return s.store.Create(name)
}

func (s *ProgramService) Update(id int64, name string) (*store.Program, error) {
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	p, err := s.store.Update(id, name)
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *ProgramService) Delete(id int64) error {
	if err := s.store.Delete(id); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			return ErrNotFound
		}
		return err
	}
	return nil
}

func (s *ProgramService) CreateFull(name string, sets []ProgramSetInput) (*store.Program, error) {
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}

	if err := s.validateExerciseIDs(sets); err != nil {
		return nil, err
	}

	tx, err := s.db.Begin()
	if err != nil {
		return nil, fmt.Errorf("begin transaction: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck

	var programID int64
	if err := tx.QueryRow(`INSERT INTO programs (name) VALUES ($1) RETURNING id`, name).Scan(&programID); err != nil {
		return nil, fmt.Errorf("create program: %w", err)
	}

	for _, set := range sets {
		if set.Rounds < 1 {
			set.Rounds = 1
		}
		var setID int64
		if err := tx.QueryRow(
			`INSERT INTO program_sets (program_id, name, rounds, intra_set_rest_seconds, sort_order) VALUES ($1, $2, $3, $4, $5) RETURNING id`,
			programID, set.Name, set.Rounds, set.IntraSetRestSeconds, set.SortOrder,
		).Scan(&setID); err != nil {
			return nil, fmt.Errorf("create program set: %w", err)
		}

		for _, ex := range set.Exercises {
			_, err := tx.Exec(
				`INSERT INTO program_exercises (program_set_id, exercise_id, laterality, target_reps, target_duration_seconds, target_weight_kg, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
				setID, ex.ExerciseID, ex.Laterality, ex.TargetReps, ex.TargetDurationSeconds, ex.TargetWeightKg, ex.SortOrder,
			)
			if err != nil {
				return nil, fmt.Errorf("create program exercise: %w", err)
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	return &store.Program{ID: programID, Name: name}, nil
}

func (s *ProgramService) validateExerciseIDs(sets []ProgramSetInput) error {
	// Collect unique IDs preserving insertion order for deterministic error messages.
	seen := map[int64]bool{}
	var ids []int64
	for _, set := range sets {
		for _, ex := range set.Exercises {
			if !seen[ex.ExerciseID] {
				seen[ex.ExerciseID] = true
				ids = append(ids, ex.ExerciseID)
			}
		}
	}
	if len(ids) == 0 {
		return nil
	}

	// Single query to fetch all existing IDs.
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	rows, err := s.db.Query(`SELECT id FROM exercises WHERE id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return fmt.Errorf("check exercises: %w", err)
	}
	defer rows.Close()

	found := map[int64]bool{}
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("scan exercise id: %w", err)
		}
		found[id] = true
	}
	if err := rows.Err(); err != nil {
		return fmt.Errorf("iterate exercises: %w", err)
	}

	for _, id := range ids {
		if !found[id] {
			return &ValidationError{Msg: fmt.Sprintf("exercise_id %d not found", id)}
		}
	}
	return nil
}
