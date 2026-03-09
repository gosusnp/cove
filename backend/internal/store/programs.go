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

// jsonProgramExercise is the internal representation for JSONB serialization of program exercises.
type jsonProgramExercise struct {
	ID                    int64             `json:"id"`
	ExerciseID            domain.ExerciseID `json:"exercise_id"`
	Name                  string            `json:"name"`
	Laterality            *string           `json:"laterality,omitempty"`
	TargetReps            *int              `json:"reps,omitempty"`
	TargetDurationSeconds *int              `json:"duration_s,omitempty"`
	TargetWeightKg        *float64          `json:"weight_kg,omitempty"`
	SortOrder             *int              `json:"sort_order,omitempty"`
}

// jsonProgramSet is the internal representation for JSONB serialization of program sets.
type jsonProgramSet struct {
	ID                  int64                 `json:"id"`
	Name                *string               `json:"name,omitempty"`
	Rounds              int                   `json:"rounds"`
	IntraSetRestSeconds *int                  `json:"rest_s,omitempty"`
	SortOrder           *int                  `json:"sort_order,omitempty"`
	Exercises           []jsonProgramExercise `json:"exercises"`
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

// GetDetail returns the full program hierarchy: sets with their exercises, all read from the denormalized programs.sets JSONB column.
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

	for _, r := range rawSets {
		sd := domain.ProgramSet{
			ID:                  r.ID,
			Name:                r.Name,
			Rounds:              r.Rounds,
			IntraSetRestSeconds: r.IntraSetRestSeconds,
			SortOrder:           r.SortOrder,
			Exercises:           []domain.ProgramExercise{},
		}
		for _, ex := range r.Exercises {
			sd.Exercises = append(sd.Exercises, domain.ProgramExercise{
				ID:                    ex.ID,
				ExerciseID:            ex.ExerciseID,
				Name:                  ex.Name,
				Laterality:            ex.Laterality,
				TargetReps:            ex.TargetReps,
				TargetDurationSeconds: ex.TargetDurationSeconds,
				TargetWeightKg:        ex.TargetWeightKg,
				SortOrder:             ex.SortOrder,
			})
		}
		p.Sets = append(p.Sets, sd)
	}

	return &p, nil
}

// SyncProgramJSON rebuilds the programs.sets JSONB column from the program_sets and program_exercises tables.
func (s *ProgramStore) SyncProgramJSON(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID) error {
	res, err := q.ExecContext(ctx, `
		UPDATE programs SET sets = (
			SELECT COALESCE(jsonb_agg(
				jsonb_build_object(
					'id', ps.id,
					'name', ps.name,
					'rounds', ps.rounds,
					'rest_s', ps.intra_set_rest_seconds,
					'sort_order', ps.sort_order,
					'exercises', (
						SELECT COALESCE(jsonb_agg(
							jsonb_build_object(
								'id', pe.id,
								'exercise_id', pe.exercise_id,
								'name', e.name,
								'laterality', pe.laterality,
								'reps', pe.target_reps,
								'duration_s', pe.target_duration_seconds,
								'weight_kg', pe.target_weight_kg,
								'sort_order', pe.sort_order
							) ORDER BY pe.sort_order NULLS LAST, pe.id
						), '[]'::jsonb)
						FROM program_exercises pe
						JOIN exercises e ON e.id = pe.exercise_id
						WHERE pe.program_set_id = ps.id
					)
				) ORDER BY ps.sort_order NULLS LAST, ps.id
			), '[]'::jsonb)
			FROM program_sets ps WHERE ps.program_id = $1
		)
		WHERE id = $1 AND org_id = $2
	`, programID, orgID)
	if err != nil {
		return fmt.Errorf("sync program json: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("sync program json rows affected: %w", err)
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
	if err := s.SyncProgramJSON(ctx, q, orgID, programID); err != nil {
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
	if err := s.SyncProgramJSON(ctx, q, orgID, programID); err != nil {
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
	return s.SyncProgramJSON(ctx, q, orgID, programID)
}

// CreateExercise inserts a new row into program_exercises and syncs the programs.sets JSONB column.
func (s *ProgramStore) CreateExercise(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*ProgramExercise, error) {
	var id int64
	err := q.QueryRowContext(ctx,
		`INSERT INTO program_exercises (program_set_id, exercise_id, laterality, target_reps, target_duration_seconds, target_weight_kg, sort_order)
		 SELECT $1, $2, $3, $4, $5, $6, $7
		 WHERE EXISTS (
		 	SELECT 1 FROM program_sets ps
		 	JOIN programs p ON p.id = ps.program_id
		 	WHERE ps.id = $1 AND ps.program_id = $8 AND p.org_id = $9
		 )
		 RETURNING id`,
		setID, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder, programID, orgID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("create program exercise: %w", err)
	}
	if err := s.SyncProgramJSON(ctx, q, orgID, programID); err != nil {
		return nil, err
	}
	return &ProgramExercise{
		ID:                    id,
		ProgramSetID:          setID,
		ExerciseID:            exerciseID,
		Laterality:            laterality,
		TargetReps:            targetReps,
		TargetDurationSeconds: targetDurationSeconds,
		TargetWeightKg:        targetWeightKg,
		SortOrder:             sortOrder,
	}, nil
}

// UpdateExercise updates an existing program_exercises row and syncs the programs.sets JSONB column.
func (s *ProgramStore) UpdateExercise(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID, id int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*ProgramExercise, error) {
	res, err := q.ExecContext(ctx,
		`UPDATE program_exercises pe
		 SET exercise_id = $1, laterality = $2, target_reps = $3, target_duration_seconds = $4, target_weight_kg = $5, sort_order = $6
		 FROM program_sets ps
		 JOIN programs p ON p.id = ps.program_id
		 WHERE pe.id = $7 AND pe.program_set_id = $8
		   AND pe.program_set_id = ps.id
		   AND ps.program_id = $9
		   AND p.org_id = $10`,
		exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder, id, setID, programID, orgID,
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
	if err := s.SyncProgramJSON(ctx, q, orgID, programID); err != nil {
		return nil, err
	}
	return &ProgramExercise{
		ID:                    id,
		ProgramSetID:          setID,
		ExerciseID:            exerciseID,
		Laterality:            laterality,
		TargetReps:            targetReps,
		TargetDurationSeconds: targetDurationSeconds,
		TargetWeightKg:        targetWeightKg,
		SortOrder:             sortOrder,
	}, nil
}

// DeleteExercise removes a program_exercises row and syncs the programs.sets JSONB column.
func (s *ProgramStore) DeleteExercise(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID, id int64) error {
	res, err := q.ExecContext(ctx,
		`DELETE FROM program_exercises pe
		 USING program_sets ps, programs p
		 WHERE pe.id = $1 AND pe.program_set_id = $2
		   AND pe.program_set_id = ps.id
		   AND ps.program_id = $3
		   AND ps.program_id = p.id
		   AND p.org_id = $4`,
		id, setID, programID, orgID,
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
	return s.SyncProgramJSON(ctx, q, orgID, programID)
}

// GetExercise reads a single exercise from the denormalized JSONB column.
func (s *ProgramStore) GetExercise(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID, id int64) (*ProgramExercise, error) {
	exercises, err := s.ListExercises(ctx, q, orgID, programID, setID)
	if err != nil {
		return nil, err
	}
	for i := range exercises {
		if exercises[i].ID == id {
			return &exercises[i], nil
		}
	}
	return nil, ErrNotFound
}

// ListExercises reads all exercises for a set from the denormalized JSONB column.
func (s *ProgramStore) ListExercises(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID int64) ([]ProgramExercise, error) {
	var setsJSON []byte
	err := q.QueryRowContext(ctx, `
		SELECT COALESCE(sets, '[]'::jsonb) FROM programs
		WHERE id = $1 AND (org_id = $2 OR is_public = true)
	`, programID, orgID).Scan(&setsJSON)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("list exercises: %w", err)
	}

	var rawSets []jsonProgramSet
	if err := json.Unmarshal(setsJSON, &rawSets); err != nil {
		return nil, fmt.Errorf("unmarshal sets: %w", err)
	}

	for _, rs := range rawSets {
		if rs.ID != setID {
			continue
		}
		exercises := make([]ProgramExercise, len(rs.Exercises))
		for i, ex := range rs.Exercises {
			exercises[i] = ProgramExercise{
				ID:                    ex.ID,
				ProgramSetID:          setID,
				ExerciseID:            ex.ExerciseID,
				Laterality:            ex.Laterality,
				TargetReps:            ex.TargetReps,
				TargetDurationSeconds: ex.TargetDurationSeconds,
				TargetWeightKg:        ex.TargetWeightKg,
				SortOrder:             ex.SortOrder,
			}
		}
		return exercises, nil
	}
	return nil, ErrNotFound
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
