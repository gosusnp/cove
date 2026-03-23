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

// ListVersions returns metadata for available snapshots of a program (no snapshot payload).
func (s *ProgramStore) ListVersions(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID) ([]domain.ProgramVersionMeta, error) {
	rows, err := q.QueryContext(ctx, `
		SELECT id, created_by, created_at
		FROM cove.program_versions
		WHERE program_id = $1 AND org_id = $2
		ORDER BY created_at DESC
	`, programID, orgID)
	if err != nil {
		return nil, fmt.Errorf("list program versions: %w", err)
	}
	defer rows.Close()

	versions := []domain.ProgramVersionMeta{}
	for rows.Next() {
		var v domain.ProgramVersionMeta
		v.ProgramID = programID
		v.OrgID = orgID
		if err := rows.Scan(&v.ID, &v.CreatedBy, &v.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan program version: %w", err)
		}
		versions = append(versions, v)
	}
	return versions, rows.Err()
}

// GetVersion retrieves a full program snapshot by version ID.
func (s *ProgramStore) GetVersion(ctx context.Context, q Querier, orgID domain.OrgID, versionID domain.ProgramVersionID) (*domain.ProgramVersion, error) {
	var v domain.ProgramVersion
	var snapshotJSON []byte

	err := q.QueryRowContext(ctx, `
		SELECT id, program_id, created_by, created_at, snapshot
		FROM cove.program_versions
		WHERE id = $1 AND org_id = $2
	`, versionID, orgID).Scan(
		&v.ID, &v.ProgramID, &v.CreatedBy, &v.CreatedAt, &snapshotJSON,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("get program version: %w", err)
	}
	v.OrgID = orgID

	// Unmarshal snapshot into a struct matching the DB JSON structure (from jsonb_build_object in trigger)
	var raw struct {
		Name        string           `json:"name"`
		Description *string          `json:"description"`
		Activity    *string          `json:"activity"`
		IsPublic    bool             `json:"is_public"`
		Sets        []JSONProgramSet `json:"sets"`
	}
	if err := json.Unmarshal(snapshotJSON, &raw); err != nil {
		return nil, fmt.Errorf("unmarshal snapshot: %w", err)
	}

	v.Snapshot = domain.ProgramSnapshot{
		Name:        raw.Name,
		Description: raw.Description,
		Activity:    raw.Activity,
		IsPublic:    raw.IsPublic,
		Sets:        make([]domain.ProgramSet, len(raw.Sets)),
	}

	for i, r := range raw.Sets {
		v.Snapshot.Sets[i] = domain.ProgramSet{
			ID:                  r.ID,
			Name:                r.Name,
			Rounds:              r.Rounds,
			IntraSetRestSeconds: r.IntraSetRestSeconds,
			Exercises:           make([]domain.ProgramExercise, len(r.Exercises)),
		}
		for j, ex := range r.Exercises {
			v.Snapshot.Sets[i].Exercises[j] = domain.ProgramExercise{
				ID:                    ex.ID,
				ExerciseID:            ex.ExerciseID,
				Name:                  ex.NameSnapshot, // Map name_snapshot back to Name
				Laterality:            ex.Laterality,
				TargetReps:            ex.TargetReps,
				TargetDurationSeconds: ex.TargetDurationSeconds,
				TargetWeight:          ex.TargetWeight,
				WeightUnit:            ex.WeightUnit,
			}
		}
	}

	return &v, nil
}

// Restore applies a snapshot to the programs table.
// This triggers a new version entry for the state BEFORE the restore.
func (s *ProgramStore) Restore(ctx context.Context, q Querier, orgID domain.OrgID, programID domain.ProgramID, snapshot domain.ProgramSnapshot) error {
	// Convert domain.ProgramSet back to JSONProgramSet for storage
	sets := make([]JSONProgramSet, len(snapshot.Sets))
	for i, s := range snapshot.Sets {
		sets[i] = JSONProgramSet{
			ID:                  s.ID,
			Name:                s.Name,
			Rounds:              s.Rounds,
			IntraSetRestSeconds: s.IntraSetRestSeconds,
			Exercises:           make([]JSONProgramExercise, len(s.Exercises)),
		}
		for j, ex := range s.Exercises {
			sets[i].Exercises[j] = JSONProgramExercise{
				ID:                    ex.ID,
				ExerciseID:            ex.ExerciseID,
				NameSnapshot:          ex.Name, // Map Name to NameSnapshot
				Laterality:            ex.Laterality,
				TargetReps:            ex.TargetReps,
				TargetDurationSeconds: ex.TargetDurationSeconds,
				TargetWeight:          ex.TargetWeight,
				WeightUnit:            ex.WeightUnit,
			}
		}
	}

	setsJSON, err := json.Marshal(sets)
	if err != nil {
		return fmt.Errorf("marshal sets for restore: %w", err)
	}

	res, err := q.ExecContext(ctx, `
		UPDATE cove.programs
		SET name = $1, description = $2, activity = $3, is_public = $4, sets = $5
		WHERE id = $6 AND org_id = $7
	`, snapshot.Name, snapshot.Description, snapshot.Activity, snapshot.IsPublic, setsJSON, programID, orgID)

	if err != nil {
		return fmt.Errorf("restore program: %w", err)
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
