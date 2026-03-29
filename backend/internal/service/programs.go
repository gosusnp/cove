// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
	"github.com/gosusnp/cove/backend/internal/store"
)

// validateWeightUnit returns a ValidationError when weight is set but the unit
// is missing or is not a mass unit.
func validateWeightUnit(weight *float64, unit *domain.Unit) error {
	if weight == nil {
		return nil
	}
	if unit == nil {
		return &ValidationError{Msg: "weight_unit is required when weight is set"}
	}
	if !unit.Valid() || unit.Category() != domain.UnitCategoryMass {
		return &ValidationError{Msg: "weight_unit must be a mass unit (g, kg, oz, lb)"}
	}
	return nil
}

// normalizeActivity trims whitespace from the activity value and returns nil
// for empty strings, ensuring the database never stores an empty string.
func normalizeActivity(s *string) *string {
	if s == nil {
		return nil
	}
	trimmed := strings.TrimSpace(*s)
	if trimmed == "" {
		return nil
	}
	return &trimmed
}

// ProgramEnriched wraps a full program with a pre-rendered Markdown structure.
type ProgramEnriched struct {
	*domain.Program
	Structure string `json:"structure"`
}

type ProgramExerciseInput struct {
	ExerciseID            domain.ExerciseID `json:"exercise_id"`
	Laterality            *string           `json:"laterality,omitempty"`
	TargetReps            *int              `json:"reps,omitempty"`
	TargetDurationSeconds *int              `json:"duration_s,omitempty"`
	TargetWeight          *float64          `json:"weight,omitempty"`
	WeightUnit            *domain.Unit      `json:"weight_unit,omitempty"`
}

type ProgramSetInput struct {
	Name                *string                `json:"name,omitempty"`
	Rounds              int                    `json:"rounds"`
	IntraSetRestSeconds *int                   `json:"rest_s,omitempty"`
	Exercises           []ProgramExerciseInput `json:"exercises"`
}

// ProgramService holds a direct *sql.DB reference in addition to its stores
// because all write operations require a transaction with a FOR UPDATE lock on
// the programs row. Other services only need their own store.
type ProgramService struct {
	db      *sql.DB
	store   *store.ProgramStore
	exStore *store.ExerciseStore
}

func NewProgramService(db *sql.DB, exStore *store.ExerciseStore) *ProgramService {
	return &ProgramService{db: db, store: store.NewProgramStore(), exStore: exStore}
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

func (s *ProgramService) Get(ctx context.Context, id domain.ProgramID) (*ProgramEnriched, error) {
	idInfo, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var p *domain.Program
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		p, err = s.store.Get(ctx, q, idInfo.OrgID, id)
		if err != nil {
			return err
		}

		// Collect unique exercise IDs for bulk name resolution.
		seen := map[domain.ExerciseID]bool{}
		var exerciseIDs []domain.ExerciseID
		for _, set := range p.Sets {
			for _, ex := range set.Exercises {
				if !seen[ex.ExerciseID] {
					seen[ex.ExerciseID] = true
					exerciseIDs = append(exerciseIDs, ex.ExerciseID)
				}
			}
		}

		if len(exerciseIDs) > 0 {
			// Bulk fetch live names; missing IDs fall back to name_snapshot already in Name.
			found, _, err := s.exStore.GetByIDs(ctx, q, idInfo.OrgID, exerciseIDs)
			if err != nil {
				return fmt.Errorf("resolve exercise names: %w", err)
			}
			liveNames := make(map[domain.ExerciseID]string, len(found))
			for _, e := range found {
				liveNames[e.ID] = e.Name
			}
			for i, set := range p.Sets {
				for j, ex := range set.Exercises {
					if name, ok := liveNames[ex.ExerciseID]; ok {
						p.Sets[i].Exercises[j].Name = name
					}
					// else: Name retains the name_snapshot set by the store.
				}
			}
		}

		return nil
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &ProgramEnriched{Program: p, Structure: markdown.Program(p)}, nil
}

func (s *ProgramService) Create(ctx context.Context, name string, description *string, activity *string, isPublic bool) (*domain.ProgramLite, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	activity = normalizeActivity(activity)

	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var p *domain.ProgramLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		p, err = s.store.Create(ctx, q, id.OrgID, name, description, activity, isPublic)
		return err
	})
	return p, err
}

func (s *ProgramService) Update(ctx context.Context, id domain.ProgramID, updatedAt *time.Time, name string, description *string, activity *string, isPublic bool) (*domain.ProgramLite, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	activity = normalizeActivity(activity)

	idInfo, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return nil, ErrUnauthorized
	}

	var p *domain.ProgramLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		var err error
		p, err = s.store.Update(ctx, q, idInfo.OrgID, id, name, description, activity, isPublic, updatedAt)
		return err
	})
	if errors.Is(err, store.ErrConflict) {
		return nil, ErrConflict
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

func (s *ProgramService) CreateSet(ctx context.Context, programID domain.ProgramID, name *string, rounds int, intraSetRestSeconds *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	var ps *store.ProgramSet
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		ps, err = s.store.CreateSet(ctx, q, idInfo.OrgID, programID, name, rounds, intraSetRestSeconds)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return ps, err
}

func (s *ProgramService) UpdateSet(ctx context.Context, programID domain.ProgramID, id int64, updatedAt *time.Time, name *string, rounds int, intraSetRestSeconds *int) (*store.ProgramSet, error) {
	if rounds < 1 {
		rounds = 1
	}
	var ps *store.ProgramSet
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		var err error
		ps, err = s.store.UpdateSet(ctx, q, idInfo.OrgID, programID, id, name, rounds, intraSetRestSeconds, updatedAt)
		return err
	})
	if errors.Is(err, store.ErrConflict) {
		return nil, ErrConflict
	}
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

func (s *ProgramService) CreateExercise(ctx context.Context, programID domain.ProgramID, setID int64, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeight *float64, weightUnit *domain.Unit) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	if err := validateWeightUnit(targetWeight, weightUnit); err != nil {
		return nil, err
	}

	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)

		// Resolve exercise name, enforcing org filter and visibility.
		found, missing, err := s.exStore.GetByIDs(ctx, q, idInfo.OrgID, []domain.ExerciseID{exerciseID})
		if err != nil {
			return fmt.Errorf("resolve exercise name: %w", err)
		}
		if len(missing) > 0 {
			return ErrNotFound
		}
		nameSnapshot := found[0].Name

		pe, err = s.store.CreateExercise(ctx, q, idInfo.OrgID, programID, setID, exerciseID, nameSnapshot, laterality, targetReps, targetDurationSeconds, targetWeight, weightUnit)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return pe, err
}

func (s *ProgramService) UpdateExercise(ctx context.Context, programID domain.ProgramID, setID, id int64, updatedAt *time.Time, exerciseID domain.ExerciseID, laterality *string, targetReps, targetDurationSeconds *int, targetWeight *float64, weightUnit *domain.Unit) (*store.ProgramExercise, error) {
	if exerciseID == 0 {
		return nil, &ValidationError{Msg: "exercise_id is required"}
	}
	if err := validateWeightUnit(targetWeight, weightUnit); err != nil {
		return nil, err
	}

	var pe *store.ProgramExercise
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)

		// Validate exercise_id visibility before mutating.
		_, missing, err := s.exStore.GetByIDs(ctx, q, idInfo.OrgID, []domain.ExerciseID{exerciseID})
		if err != nil {
			return fmt.Errorf("validate exercise: %w", err)
		}
		if len(missing) > 0 {
			return ErrNotFound
		}

		pe, err = s.store.UpdateExercise(ctx, q, idInfo.OrgID, programID, setID, id, exerciseID, laterality, targetReps, targetDurationSeconds, targetWeight, weightUnit, updatedAt)
		return err
	})
	if errors.Is(err, store.ErrConflict) {
		return nil, ErrConflict
	}
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

// StructureEntry describes one set and its ordered exercises for ReorderStructure.
type StructureEntry struct {
	SetID       int64   `json:"set_id"`
	ExerciseIDs []int64 `json:"exercise_ids"`
}

// ReorderStructure atomically reorders sets and exercises within a program.
func (s *ProgramService) ReorderStructure(ctx context.Context, programID domain.ProgramID, structure []StructureEntry) error {
	entries := make([]store.ProgramStructureEntry, len(structure))
	for i, e := range structure {
		entries[i] = store.ProgramStructureEntry{SetID: e.SetID, ExerciseIDs: e.ExerciseIDs}
	}

	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)
		return s.store.ReorderStructure(ctx, q, idInfo.OrgID, programID, entries)
	})
	var bse *store.BadStructureError
	if errors.As(err, &bse) {
		return &ValidationError{Msg: bse.Msg}
	}
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

func (s *ProgramService) CreateFull(ctx context.Context, name string, description *string, activity *string, isPublic bool, sets []ProgramSetInput) (*domain.ProgramLite, error) {
	name = normalizeName(name)
	if name == "" {
		return nil, &ValidationError{Msg: "name is required"}
	}
	activity = normalizeActivity(activity)

	// Collect unique exercise IDs preserving insertion order for deterministic error messages.
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

	var programLite *domain.ProgramLite
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		idInfo, _ := domain.IdentityFromContext(ctx)

		exerciseNames := map[domain.ExerciseID]string{}
		if len(ids) > 0 {
			found, missing, err := s.exStore.GetByIDs(ctx, q, idInfo.OrgID, ids)
			if err != nil {
				return fmt.Errorf("check exercises: %w", err)
			}
			for _, e := range found {
				exerciseNames[e.ID] = e.Name
			}
			for _, missingID := range missing {
				return &ValidationError{Msg: fmt.Sprintf("exercise_id %d not found or access denied", missingID)}
			}
		}

		p, err := s.store.Create(ctx, q, idInfo.OrgID, name, description, activity, isPublic)
		if err != nil {
			return err
		}

		// Build the sets JSONB structure directly without going through the dropped tables.
		jsonSets := make([]store.JSONProgramSet, 0, len(sets))
		var nextSetID int64 = 1
		for _, set := range sets {
			if set.Rounds < 1 {
				set.Rounds = 1
			}
			setID := nextSetID
			nextSetID++

			jsonExercises := make([]store.JSONProgramExercise, 0, len(set.Exercises))
			var nextExID int64 = 1
			for _, ex := range set.Exercises {
				jsonExercises = append(jsonExercises, store.JSONProgramExercise{
					ID:                    nextExID,
					ExerciseID:            ex.ExerciseID,
					NameSnapshot:          exerciseNames[ex.ExerciseID],
					Laterality:            ex.Laterality,
					TargetReps:            ex.TargetReps,
					TargetDurationSeconds: ex.TargetDurationSeconds,
					TargetWeight:          ex.TargetWeight,
					WeightUnit:            ex.WeightUnit,
				})
				nextExID++
			}
			jsonSets = append(jsonSets, store.JSONProgramSet{
				ID:                  setID,
				Name:                set.Name,
				Rounds:              set.Rounds,
				IntraSetRestSeconds: set.IntraSetRestSeconds,
				Exercises:           jsonExercises,
			})
		}

		if err := s.store.WriteSetsForNewProgram(ctx, q, p.ID, jsonSets); err != nil {
			return err
		}

		programLite = p
		return nil
	})

	return programLite, err
}
