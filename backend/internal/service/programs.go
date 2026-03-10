// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

type ProgramExerciseInput struct {
	ExerciseID            domain.ExerciseID `json:"exercise_id"`
	Laterality            *string           `json:"laterality,omitempty"`
	TargetReps            *int              `json:"reps,omitempty"`
	TargetDurationSeconds *int              `json:"duration_s,omitempty"`
	TargetWeightKg        *float64          `json:"weight_kg,omitempty"`
	SortOrder             *int              `json:"sort_order,omitempty"`
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
	return &ProgramService{db: db, store: store.NewProgramStore()}
}

func (s *ProgramService) List(ctx context.Context) ([]domain.ProgramLite, error) {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var list []domain.ProgramLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		list, err = s.store.List(ctx, q, id.OrgID)
		return err
	})
	return list, err
}

func (s *ProgramService) GetDetail(ctx context.Context, id domain.ProgramID) (*domain.Program, error) {
	idInfo, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var p *domain.Program
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		p, err = s.store.GetDetail(ctx, q, idInfo.OrgID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *ProgramService) Create(ctx context.Context, name string, description *string, isPublic bool) (*domain.ProgramLite, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}

	var p *domain.ProgramLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		p, err = s.store.Create(ctx, q, name, description, isPublic)
		return err
	})
	return p, err
}

func (s *ProgramService) Update(ctx context.Context, id domain.ProgramID, name string, description *string, isPublic bool) (*domain.ProgramLite, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}

	idInfo, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var p *domain.ProgramLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		p, err = s.store.Update(ctx, q, idInfo.OrgID, id, name, description, isPublic)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return p, err
}

func (s *ProgramService) ListSets(ctx context.Context, programID domain.ProgramID) ([]store.ProgramSet, error) {
	var sets []store.ProgramSet
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		sets, err = s.store.ListSets(ctx, q, idInfo.OrgID, programID)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return sets, err
}

func (s *ProgramService) GetSet(ctx context.Context, programID domain.ProgramID, id int64) (*store.ProgramSet, error) {
	var ps *store.ProgramSet
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		ps, err = s.store.GetSet(ctx, q, idInfo.OrgID, programID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ps, err
}

func (s *ProgramService) CreateSet(ctx context.Context, programID domain.ProgramID, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	var ps *store.ProgramSet
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		ps, err = s.store.CreateSet(ctx, q, idInfo.OrgID, programID, name, rounds, intraSetRestSeconds, sortOrder)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ps, err
}

func (s *ProgramService) UpdateSet(ctx context.Context, programID domain.ProgramID, id int64, name *string, rounds int, intraSetRestSeconds, sortOrder *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	var ps *store.ProgramSet
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		ps, err = s.store.UpdateSet(ctx, q, idInfo.OrgID, programID, id, name, rounds, intraSetRestSeconds, sortOrder)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ps, err
}

func (s *ProgramService) DeleteSet(ctx context.Context, programID domain.ProgramID, id int64) error {
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		return s.store.DeleteSet(ctx, q, idInfo.OrgID, programID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *ProgramService) ListExercises(ctx context.Context, programID domain.ProgramID, setID int64) ([]store.ProgramExercise, error) {
	var list []store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		list, err = s.store.ListExercises(ctx, q, idInfo.OrgID, programID, setID)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return list, err
}

func (s *ProgramService) GetExercise(ctx context.Context, programID domain.ProgramID, setID, id int64) (*store.ProgramExercise, error) {
	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		pe, err = s.store.GetExercise(ctx, q, idInfo.OrgID, programID, setID, id)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return pe, err
}

func (s *ProgramService) CreateExercise(ctx context.Context, programID domain.ProgramID, setID int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		pe, err = s.store.CreateExercise(ctx, q, idInfo.OrgID, programID, setID, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return pe, err
}

func (s *ProgramService) UpdateExercise(ctx context.Context, programID domain.ProgramID, setID, id int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeightKg *float64, sortOrder *int) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		pe, err = s.store.UpdateExercise(ctx, q, idInfo.OrgID, programID, setID, id, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeightKg, sortOrder)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return pe, err
}

func (s *ProgramService) DeleteExercise(ctx context.Context, programID domain.ProgramID, setID, id int64) error {
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		return s.store.DeleteExercise(ctx, q, idInfo.OrgID, programID, setID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *ProgramService) Delete(ctx context.Context, id domain.ProgramID) error {
	idInfo, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		return s.store.Delete(ctx, q, idInfo.OrgID, id)
	})
	if errors.Is(err, store.ErrNotFound) {
		return ErrNotFound
	}
	return err
}

func (s *ProgramService) CreateFull(ctx context.Context, name string, description *string, isPublic bool, sets []ProgramSetInput) (*domain.ProgramLite, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}

	if err := s.validateExerciseIDs(ctx, sets); err != nil {
		return nil, err
	}

	var programLite *domain.ProgramLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var programID domain.ProgramID
		if err := q.QueryRowContext(ctx, `INSERT INTO programs (name, description, is_public) VALUES ($1, $2, $3) RETURNING id`, name, description, isPublic).Scan(&programID); err != nil {
			return fmt.Errorf("create program: %w", err)
		}

		var nextSetID int64 = 1
		for _, set := range sets {
			if set.Rounds < 1 {
				set.Rounds = 1
			}
			setID := nextSetID
			nextSetID++
			if _, err := q.ExecContext(ctx,
				`INSERT INTO program_sets (id, program_id, name, rounds, intra_set_rest_seconds, sort_order) VALUES ($1, $2, $3, $4, $5, $6)`,
				setID, programID, set.Name, set.Rounds, set.IntraSetRestSeconds, set.SortOrder,
			); err != nil {
				return fmt.Errorf("create program set: %w", err)
			}

			var nextExID int64 = 1
			for _, ex := range set.Exercises {
				exID := nextExID
				nextExID++
				if _, err := q.ExecContext(ctx,
					`INSERT INTO program_exercises (id, program_set_id, exercise_id, laterality, target_reps, target_duration_seconds, target_weight_kg, sort_order) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
					exID, setID, ex.ExerciseID, ex.Laterality, ex.TargetReps, ex.TargetDurationSeconds, ex.TargetWeightKg, ex.SortOrder,
				); err != nil {
					return fmt.Errorf("create program exercise: %w", err)
				}
			}
		}

		idInfo, _ := domain.IdentityFromContext(ctx)
		if len(sets) > 0 {
			if err := s.store.SyncProgramJSON(ctx, q, idInfo.OrgID, programID); err != nil {
				return err
			}
		}
		var err error
		programLite, err = s.store.Get(ctx, q, idInfo.OrgID, programID)
		return err
	})

	return programLite, err
}

func (s *ProgramService) validateExerciseIDs(ctx context.Context, sets []ProgramSetInput) error {
	// Collect unique IDs preserving insertion order for deterministic error messages.
	seen := map[domain.ExerciseID]bool{}
	var ids []domain.ExerciseID
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

	idInfo, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	// Single query to fetch all existing IDs.
	// We must also check that the user has access to these exercises.
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids)+1)
	args[0] = idInfo.OrgID
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args[i+1] = id
	}
	rows, err := s.db.QueryContext(ctx, `SELECT id FROM exercises WHERE (org_id = $1 OR is_public = true) AND id IN (`+strings.Join(placeholders, ",")+`)`, args...)
	if err != nil {
		return fmt.Errorf("check exercises: %w", err)
	}
	defer rows.Close()

	found := map[domain.ExerciseID]bool{}
	for rows.Next() {
		var id domain.ExerciseID
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
			return &ValidationError{Msg: fmt.Sprintf("exercise_id %d not found or access denied", id)}
		}
	}
	return nil
}
