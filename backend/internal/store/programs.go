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

// JSONProgramExercise is the JSONB representation of a program exercise stored in programs.sets.
type JSONProgramExercise struct {
	ID                    int64             `json:"id"`
	ExerciseID            domain.ExerciseID `json:"exercise_id"`
	NameSnapshot          string            `json:"name_snapshot"`
	Laterality            *string           `json:"laterality,omitempty"`
	TargetReps            *int              `json:"reps,omitempty"`
	TargetDurationSeconds *int              `json:"duration_s,omitempty"`
	TargetWeightKg        *float64          `json:"weight_kg,omitempty"`
}

// JSONProgramSet is the JSONB representation of a program set stored in programs.sets.
type JSONProgramSet struct {
	ID                  int64                 `json:"id"`
	Name                *string               `json:"name,omitempty"`
	Rounds              int                   `json:"rounds"`
	IntraSetRestSeconds *int                  `json:"rest_s,omitempty"`
	Exercises           []JSONProgramExercise `json:"exercises"`
}

// jsonProgramExercise is an alias for the exported type, kept for internal use.
type jsonProgramExercise = JSONProgramExercise

// jsonProgramSet is an alias for the exported type, kept for internal use.
type jsonProgramSet = JSONProgramSet

type ProgramStore struct{}

func NewProgramStore() *ProgramStore {
	return &ProgramStore{}
}

func (s *ProgramStore) List(ctx context.Context, q Querier, orgID domain.OrgID) ([]domain.ProgramLite, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, name, org_id, is_public FROM cove.programs
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

func (s *ProgramStore) GetLite(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID) (*domain.ProgramLite, error) {
	var p domain.ProgramLite
	err := q.QueryRowContext(ctx, `
		SELECT id, name, org_id, is_public FROM cove.programs
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

func (s *ProgramStore) Create(ctx context.Context, q Querier, orgID domain.OrgID, name string, description *string, isPublic bool) (*domain.ProgramLite, error) {
	var id domain.ProgramID
	err := q.QueryRowContext(ctx,
		`INSERT INTO cove.programs (name, description, is_public) VALUES ($1, $2, $3) RETURNING id`,
		name, description, isPublic,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("create program: %w", err)
	}

	return s.GetLite(ctx, q, orgID, id)
}

func (s *ProgramStore) Update(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID, name string, description *string, isPublic bool) (*domain.ProgramLite, error) {
	res, err := q.ExecContext(ctx,
		`UPDATE cove.programs SET name = $1, description = $2, is_public = $3
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
	return s.GetLite(ctx, q, orgID, id)
}

// Get returns the full program hierarchy: sets with their exercises, all read from the denormalized programs.sets JSONB column.
func (s *ProgramStore) Get(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID) (*domain.Program, error) {
	var p domain.Program
	var setsJSON []byte
	err := q.QueryRowContext(ctx, `
		SELECT id, name, description, org_id, is_public, created_by, created_at, updated_by, updated_at, sets
		FROM cove.programs
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
			Exercises:           []domain.ProgramExercise{},
		}
		for _, ex := range r.Exercises {
			sd.Exercises = append(sd.Exercises, domain.ProgramExercise{
				ID:                    ex.ID,
				ExerciseID:            ex.ExerciseID,
				Name:                  ex.NameSnapshot,
				Laterality:            ex.Laterality,
				TargetReps:            ex.TargetReps,
				TargetDurationSeconds: ex.TargetDurationSeconds,
				TargetWeightKg:        ex.TargetWeightKg,
			})
		}
		p.Sets = append(p.Sets, sd)
	}

	return &p, nil
}

// ListSets returns the set metadata for a program from the denormalized JSONB column.
func (s *ProgramStore) ListSets(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID) ([]ProgramSet, error) {
	var setsJSON []byte
	err := q.QueryRowContext(ctx, `
		SELECT COALESCE(sets, '[]'::jsonb) FROM cove.programs
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

// lockProgram acquires a FOR UPDATE row lock on the programs row and verifies ownership.
// It returns ErrNotFound if the row does not exist or belongs to a different org.
// All write operations that mutate the programs.sets JSONB column must call this first
// to serialize concurrent mutations to the same program.
func (s *ProgramStore) lockProgram(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID) error {
	var id domain.ProgramID
	err := q.QueryRowContext(ctx,
		`SELECT id FROM cove.programs WHERE id = $1 AND org_id = $2 FOR UPDATE`,
		programID, orgID,
	).Scan(&id)
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	if err != nil {
		return fmt.Errorf("lock program: %w", err)
	}
	return nil
}

// nextSetID reads the programs.sets JSONB column and returns max(set.id) + 1.
// Must be called inside a transaction after lockProgram.
func (s *ProgramStore) nextSetID(ctx context.Context, q Querier, programID domain.ProgramID) (int64, error) {
	var setsJSON []byte
	if err := q.QueryRowContext(ctx,
		`SELECT COALESCE(sets, '[]'::jsonb) FROM cove.programs WHERE id = $1`,
		programID,
	).Scan(&setsJSON); err != nil {
		return 0, fmt.Errorf("read sets for id: %w", err)
	}
	var raw []jsonProgramSet
	if err := json.Unmarshal(setsJSON, &raw); err != nil {
		return 0, fmt.Errorf("unmarshal sets for id: %w", err)
	}
	var maxID int64
	for _, r := range raw {
		if r.ID > maxID {
			maxID = r.ID
		}
	}
	return maxID + 1, nil
}

// nextExerciseID reads the programs.sets JSONB column and returns max(exercise.id) + 1
// across all exercises in all sets of the program.
// Must be called inside a transaction after lockProgram.
func (s *ProgramStore) nextExerciseID(ctx context.Context, q Querier, programID domain.ProgramID) (int64, error) {
	var setsJSON []byte
	if err := q.QueryRowContext(ctx,
		`SELECT COALESCE(sets, '[]'::jsonb) FROM cove.programs WHERE id = $1`,
		programID,
	).Scan(&setsJSON); err != nil {
		return 0, fmt.Errorf("read sets for exercise id: %w", err)
	}
	var raw []jsonProgramSet
	if err := json.Unmarshal(setsJSON, &raw); err != nil {
		return 0, fmt.Errorf("unmarshal sets for exercise id: %w", err)
	}
	var maxID int64
	for _, r := range raw {
		for _, ex := range r.Exercises {
			if ex.ID > maxID {
				maxID = ex.ID
			}
		}
	}
	return maxID + 1, nil
}

// readSets reads and unmarshals the programs.sets JSONB column.
// Must be called inside a transaction after lockProgram.
func (s *ProgramStore) readSets(ctx context.Context, q Querier, programID domain.ProgramID) ([]jsonProgramSet, error) {
	var setsJSON []byte
	if err := q.QueryRowContext(ctx,
		`SELECT COALESCE(sets, '[]'::jsonb) FROM cove.programs WHERE id = $1`,
		programID,
	).Scan(&setsJSON); err != nil {
		return nil, fmt.Errorf("read sets: %w", err)
	}
	var raw []jsonProgramSet
	if err := json.Unmarshal(setsJSON, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal sets: %w", err)
	}
	return raw, nil
}

// writeSets marshals sets and updates the programs.sets JSONB column.
func (s *ProgramStore) writeSets(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, sets []jsonProgramSet) error {
	data, err := json.Marshal(sets)
	if err != nil {
		return fmt.Errorf("marshal sets: %w", err)
	}
	res, err := q.ExecContext(ctx,
		`UPDATE cove.programs SET sets = $1 WHERE id = $2 AND org_id = $3`,
		data, programID, orgID,
	)
	if err != nil {
		return fmt.Errorf("write sets: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("write sets rows affected: %w", err)
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// CreateSet appends a new set to the programs.sets JSONB column.
func (s *ProgramStore) CreateSet(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, name *string, rounds int, intraSetRestSeconds *int) (*ProgramSet, error) {
	if err := s.lockProgram(ctx, q, orgID, programID); err != nil {
		return nil, err
	}
	id, err := s.nextSetID(ctx, q, programID)
	if err != nil {
		return nil, err
	}
	sets, err := s.readSets(ctx, q, programID)
	if err != nil {
		return nil, err
	}
	sets = append(sets, jsonProgramSet{
		ID:                  id,
		Name:                name,
		Rounds:              rounds,
		IntraSetRestSeconds: intraSetRestSeconds,
		Exercises:           []jsonProgramExercise{},
	})
	if err := s.writeSets(ctx, q, orgID, programID, sets); err != nil {
		return nil, err
	}
	return &ProgramSet{
		ID:                  id,
		ProgramID:           programID,
		Name:                name,
		Rounds:              rounds,
		IntraSetRestSeconds: intraSetRestSeconds,
	}, nil
}

// UpdateSet updates an existing set in the programs.sets JSONB column.
func (s *ProgramStore) UpdateSet(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, id int64, name *string, rounds int, intraSetRestSeconds *int) (*ProgramSet, error) {
	if err := s.lockProgram(ctx, q, orgID, programID); err != nil {
		return nil, err
	}
	sets, err := s.readSets(ctx, q, programID)
	if err != nil {
		return nil, err
	}
	found := false
	for i, r := range sets {
		if r.ID == id {
			sets[i].Name = name
			sets[i].Rounds = rounds
			sets[i].IntraSetRestSeconds = intraSetRestSeconds
			found = true
			break
		}
	}
	if !found {
		return nil, ErrNotFound
	}
	if err := s.writeSets(ctx, q, orgID, programID, sets); err != nil {
		return nil, err
	}
	return &ProgramSet{
		ID:                  id,
		ProgramID:           programID,
		Name:                name,
		Rounds:              rounds,
		IntraSetRestSeconds: intraSetRestSeconds,
	}, nil
}

// DeleteSet removes a set from the programs.sets JSONB column.
func (s *ProgramStore) DeleteSet(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, id int64) error {
	if err := s.lockProgram(ctx, q, orgID, programID); err != nil {
		return err
	}
	sets, err := s.readSets(ctx, q, programID)
	if err != nil {
		return err
	}
	filtered := sets[:0]
	found := false
	for _, r := range sets {
		if r.ID == id {
			found = true
			continue
		}
		filtered = append(filtered, r)
	}
	if !found {
		return ErrNotFound
	}
	return s.writeSets(ctx, q, orgID, programID, filtered)
}

// CreateExercise appends a new exercise to the target set in the programs.sets JSONB column.
// nameSnapshot is written once and never updated; callers are responsible for resolving it
// via ExerciseService before calling this method.
func (s *ProgramStore) CreateExercise(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID int64, exerciseID domain.ExerciseID, nameSnapshot string, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64) (*ProgramExercise, error) {
	if err := s.lockProgram(ctx, q, orgID, programID); err != nil {
		return nil, err
	}

	sets, err := s.readSets(ctx, q, programID)
	if err != nil {
		return nil, err
	}

	// Verify the target set belongs to this program.
	setFound := false
	for _, r := range sets {
		if r.ID == setID {
			setFound = true
			break
		}
	}
	if !setFound {
		return nil, ErrNotFound
	}

	id, err := s.nextExerciseID(ctx, q, programID)
	if err != nil {
		return nil, err
	}

	newEx := jsonProgramExercise{
		ID:                    id,
		ExerciseID:            exerciseID,
		NameSnapshot:          nameSnapshot,
		Laterality:            laterality,
		TargetReps:            targetReps,
		TargetDurationSeconds: targetDurationSeconds,
		TargetWeightKg:        targetWeightKg,
	}
	for i, r := range sets {
		if r.ID == setID {
			sets[i].Exercises = append(sets[i].Exercises, newEx)
			break
		}
	}
	if err := s.writeSets(ctx, q, orgID, programID, sets); err != nil {
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
	}, nil
}

// UpdateExercise updates an existing exercise in the programs.sets JSONB column.
// name_snapshot is never mutated on update; the existing value is preserved.
func (s *ProgramStore) UpdateExercise(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID, id int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64) (*ProgramExercise, error) {
	if err := s.lockProgram(ctx, q, orgID, programID); err != nil {
		return nil, err
	}

	sets, err := s.readSets(ctx, q, programID)
	if err != nil {
		return nil, err
	}

	found := false
	for i, r := range sets {
		if r.ID != setID {
			continue
		}
		for j, ex := range r.Exercises {
			if ex.ID == id {
				sets[i].Exercises[j] = jsonProgramExercise{
					ID:                    id,
					ExerciseID:            exerciseID,
					NameSnapshot:          ex.NameSnapshot,
					Laterality:            laterality,
					TargetReps:            targetReps,
					TargetDurationSeconds: targetDurationSeconds,
					TargetWeightKg:        targetWeightKg,
				}
				found = true
				break
			}
		}
		break
	}
	if !found {
		return nil, ErrNotFound
	}
	if err := s.writeSets(ctx, q, orgID, programID, sets); err != nil {
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
	}, nil
}

// DeleteExercise removes an exercise from the target set in the programs.sets JSONB column.
func (s *ProgramStore) DeleteExercise(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, setID, id int64) error {
	if err := s.lockProgram(ctx, q, orgID, programID); err != nil {
		return err
	}
	sets, err := s.readSets(ctx, q, programID)
	if err != nil {
		return err
	}
	found := false
	for i, r := range sets {
		if r.ID != setID {
			continue
		}
		filtered := r.Exercises[:0]
		for _, ex := range r.Exercises {
			if ex.ID == id {
				found = true
				continue
			}
			filtered = append(filtered, ex)
		}
		sets[i].Exercises = filtered
		break
	}
	if !found {
		return ErrNotFound
	}
	return s.writeSets(ctx, q, orgID, programID, sets)
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
		SELECT COALESCE(sets, '[]'::jsonb) FROM cove.programs
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
			}
		}
		return exercises, nil
	}
	return nil, ErrNotFound
}

// WriteSetsForNewProgram updates programs.sets for a freshly-created program row.
// It is used by CreateFull where the program was just inserted in the same transaction
// and there is no need for a lockProgram call.
func (s *ProgramStore) WriteSetsForNewProgram(ctx context.Context, q Querier, programID domain.ProgramID, sets []JSONProgramSet) error {
	if len(sets) == 0 {
		return nil
	}
	data, err := json.Marshal(sets)
	if err != nil {
		return fmt.Errorf("marshal sets for new program: %w", err)
	}
	_, err = q.ExecContext(ctx,
		`UPDATE cove.programs SET sets = $1 WHERE id = $2`,
		data, programID,
	)
	if err != nil {
		return fmt.Errorf("write sets for new program: %w", err)
	}
	return nil
}

// ProgramStructureEntry describes one set and its ordered exercise IDs for ReorderStructure.
type ProgramStructureEntry struct {
	SetID       int64
	ExerciseIDs []int64
}

// BadStructureError is returned by ReorderStructure when the request is inconsistent
// with the program's current set/exercise inventory.
type BadStructureError struct {
	Msg string
}

func (e *BadStructureError) Error() string { return e.Msg }

// ReorderStructure atomically reorders sets and exercises within a program.
// structure must contain exactly one entry per existing set. The union of all exercise IDs
// across entries must match exactly the full set of exercise IDs currently in the program;
// cross-set moves are allowed, but additions and deletions are not.
func (s *ProgramStore) ReorderStructure(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, structure []ProgramStructureEntry) error {
	if err := s.lockProgram(ctx, q, orgID, programID); err != nil {
		return err
	}
	sets, err := s.readSets(ctx, q, programID)
	if err != nil {
		return err
	}

	if len(structure) != len(sets) {
		return &BadStructureError{Msg: fmt.Sprintf("expected %d sets, got %d", len(sets), len(structure))}
	}

	setIndex := make(map[int64]jsonProgramSet, len(sets))
	for _, r := range sets {
		setIndex[r.ID] = r
	}

	allExercises := make(map[int64]jsonProgramExercise)
	for _, r := range sets {
		for _, ex := range r.Exercises {
			allExercises[ex.ID] = ex
		}
	}

	seenSets := make(map[int64]bool, len(structure))
	seenExercises := make(map[int64]bool, len(allExercises))
	for _, entry := range structure {
		if _, ok := setIndex[entry.SetID]; !ok {
			return &BadStructureError{Msg: fmt.Sprintf("set %d not found", entry.SetID)}
		}
		if seenSets[entry.SetID] {
			return &BadStructureError{Msg: fmt.Sprintf("duplicate set_id %d", entry.SetID)}
		}
		seenSets[entry.SetID] = true
		for _, exID := range entry.ExerciseIDs {
			if _, ok := allExercises[exID]; !ok {
				return &BadStructureError{Msg: fmt.Sprintf("exercise %d not found in program", exID)}
			}
			if seenExercises[exID] {
				return &BadStructureError{Msg: fmt.Sprintf("duplicate exercise_id %d", exID)}
			}
			seenExercises[exID] = true
		}
	}
	if len(seenExercises) != len(allExercises) {
		return &BadStructureError{Msg: "all exercises must be included exactly once"}
	}

	newSets := make([]jsonProgramSet, 0, len(structure))
	for _, entry := range structure {
		cur := setIndex[entry.SetID]
		newExercises := make([]jsonProgramExercise, 0, len(entry.ExerciseIDs))
		for _, exID := range entry.ExerciseIDs {
			newExercises = append(newExercises, allExercises[exID])
		}
		cur.Exercises = newExercises
		newSets = append(newSets, cur)
	}

	return s.writeSets(ctx, q, orgID, programID, newSets)
}

func (s *ProgramStore) Delete(ctx context.Context, q Querier, orgID domain.OrgID, id domain.ProgramID) error {
	res, err := q.ExecContext(ctx, `DELETE FROM cove.programs WHERE id = $1 AND org_id = $2`, id, orgID)
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
