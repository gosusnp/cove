// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
)

// jsonProgramSet is the internal representation for JSONB serialization of program sets.
type jsonProgramSet struct {
	ID                  int64   `json:"id"`
	Name                *string `json:"name,omitempty"`
	Rounds              int     `json:"rounds"`
	IntraSetRestSeconds *int    `json:"rest_s,omitempty"`
	SortOrder           *int    `json:"sort_order,omitempty"`
}

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
// Sets are read from the denormalized programs.sets JSONB column; exercises are joined from program_exercises.
func (s *ProgramStore) GetDetail(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID) (*domain.Program, error) {
	var p domain.Program
	var setsJSON []byte
	err := q.QueryRowContext(ctx, `
		SELECT id, name, description, org_id, is_public, created_by, created_at, updated_by, updated_at, sets
		FROM programs
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, id, orgID).Scan(
		&p.ID, &p.Name, &p.Description, &p.OrgID, &p.IsPublic,
		&p.CreatedBy, &p.CreatedAt, &p.UpdatedBy, &p.UpdatedAt, &setsJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program detail: %w", err)
	}

	p.Sets = []domain.ProgramSet{}

	var rawSets []jsonProgramSet
	if err := json.Unmarshal(setsJSON, &rawSets); err != nil {
		return nil, fmt.Errorf("unmarshal sets: %w", err)
	}

	setIndex := map[int64]int{}
	for _, r := range rawSets {
		sd := domain.ProgramSet{
			ID:                  r.ID,
			Name:                r.Name,
			Rounds:              r.Rounds,
			IntraSetRestSeconds: r.IntraSetRestSeconds,
			SortOrder:           r.SortOrder,
			Exercises:           []domain.ProgramExercise{},
		}
		setIndex[r.ID] = len(p.Sets)
		p.Sets = append(p.Sets, sd)
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

// SyncSetsJSON rebuilds the programs.sets JSONB column from the program_sets table.
func (s *ProgramStore) SyncSetsJSON(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID) error {
	res, err := q.ExecContext(ctx, `
		UPDATE programs SET sets = (
			SELECT COALESCE(jsonb_agg(
				jsonb_build_object(
					'id', ps.id,
					'name', ps.name,
					'rounds', ps.rounds,
					'rest_s', ps.intra_set_rest_seconds,
					'sort_order', ps.sort_order
				) ORDER BY ps.sort_order NULLS LAST, ps.id
			), '[]'::jsonb)
			FROM program_sets ps WHERE ps.program_id = $1
		)
		WHERE id = $1 AND org_id = $2
	`, programID, orgID)
	if err != nil {
		return fmt.Errorf("sync sets json: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sync sets json rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// ListSets returns the set metadata for a program from the denormalized JSONB column.
func (s *ProgramStore) ListSets(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID) ([]ProgramSet, error) {
	var setsJSON []byte
	err := q.QueryRowContext(ctx, `
		SELECT COALESCE(sets, '[]'::jsonb) FROM programs
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, programID, orgID).Scan(&setsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("list sets: %w", err)
	}

	var raw []jsonProgramSet
	if err := json.Unmarshal(setsJSON, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal sets: %w", err)
	}

	sets := make([]ProgramSet, len(raw))
	for i, r := range raw {
		sets[i] = ProgramSet{
			ID:                  r.ID,
			ProgramID:           programID,
			Name:                r.Name,
			Rounds:              r.Rounds,
			IntraSetRestSeconds: r.IntraSetRestSeconds,
			SortOrder:           r.SortOrder,
		}
	}
	return sets, nil
}

// GetSet returns a single set by id, reading from the denormalized JSONB column.
func (s *ProgramStore) GetSet(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, id int64) (*ProgramSet, error) {
	sets, err := s.ListSets(ctx, q, orgID, programID)
	if err != nil {
		return nil, err
	}
	for i := range sets {
		if sets[i].ID == id {
			return &sets[i], nil
		}
	}
	return nil, ErrNotFound
}

// CreateSet inserts a new row into program_sets and syncs the programs.sets JSONB column.
func (s *ProgramStore) CreateSet(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*ProgramSet, error) {
	var id int64
	err := q.QueryRowContext(ctx,
		`INSERT INTO program_sets (program_id, name, rounds, intra_set_rest_seconds, sort_order)
		 SELECT $1, $3, $4, $5, $6 FROM programs WHERE id = $1 AND org_id = $2
		 RETURNING id`,
		programID, orgID, name, rounds, intraSetRestSeconds, sortOrder,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("create program set: %w", err)
	}
	if err := s.SyncSetsJSON(ctx, q, orgID, programID); err != nil {
		return nil, err
	}
	return &ProgramSet{
		ID:                  id,
		ProgramID:           programID,
		Name:                name,
		Rounds:              rounds,
		IntraSetRestSeconds: intraSetRestSeconds,
		SortOrder:           sortOrder,
	}, nil
}

// UpdateSet updates an existing program set row and syncs the programs.sets JSONB column.
func (s *ProgramStore) UpdateSet(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, id int64, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*ProgramSet, error) {
	res, err := q.ExecContext(ctx,
		`UPDATE program_sets SET name = $1, rounds = $2, intra_set_rest_seconds = $3, sort_order = $4
		 WHERE id = $5 AND program_id = $6
		   AND EXISTS (SELECT 1 FROM programs WHERE id = $6 AND org_id = $7)`,
		name, rounds, intraSetRestSeconds, sortOrder, id, programID, orgID,
	)
	if err != nil {
		return nil, fmt.Errorf("update program set: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return nil, ErrNotFound
	}
	if err := s.SyncSetsJSON(ctx, q, orgID, programID); err != nil {
		return nil, err
	}
	return &ProgramSet{
		ID:                  id,
		ProgramID:           programID,
		Name:                name,
		Rounds:              rounds,
		IntraSetRestSeconds: intraSetRestSeconds,
		SortOrder:           sortOrder,
	}, nil
}

// DeleteSet removes a program set row and syncs the programs.sets JSONB column.
func (s *ProgramStore) DeleteSet(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, id int64) error {
	res, err := q.ExecContext(ctx,
		`DELETE FROM program_sets WHERE id = $1 AND program_id = $2
		  AND EXISTS (SELECT 1 FROM programs WHERE id = $2 AND org_id = $3)`,
		id, programID, orgID,
	)
	if err != nil {
		return fmt.Errorf("delete program set: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return s.SyncSetsJSON(ctx, q, orgID, programID)
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
