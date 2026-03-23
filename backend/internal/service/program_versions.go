// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

// ListVersions returns metadata for available snapshots of a program (no snapshot payload).
func (s *ProgramService) ListVersions(ctx context.Context, programID domain.ProgramID) ([]domain.ProgramVersionMeta, error) {
	var versions []domain.ProgramVersionMeta
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		id, _ := domain.IdentityFromContext(ctx)

		// Verify program existence and access first to return 404 if denied
		if _, err := s.store.GetLite(ctx, q, id.OrgID, programID); err != nil {
			return err
		}

		var err error
		versions, err = s.store.ListVersions(ctx, q, id.OrgID, programID)
		return err
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return versions, nil
}

// GetVersion retrieves a single historical version of the program.
func (s *ProgramService) GetVersion(ctx context.Context, programID domain.ProgramID, versionID domain.ProgramVersionID) (*domain.ProgramVersion, error) {
	var version *domain.ProgramVersion
	err := withScopedTx(ctx, s.db, func(q store.Querier) error {
		id, _ := domain.IdentityFromContext(ctx)

		v, err := s.store.GetVersion(ctx, q, id.OrgID, versionID)
		if err != nil {
			return err
		}

		if v.ProgramID != programID {
			return &ValidationError{Msg: "version does not belong to this program"}
		}
		version = v
		return nil
	})
	if errors.Is(err, store.ErrNotFound) {
		return nil, ErrNotFound
	}
	return version, err
}

// Rollback restores a historical version of the program.
func (s *ProgramService) Rollback(ctx context.Context, programID domain.ProgramID, versionID domain.ProgramVersionID) error {
	return withScopedTx(ctx, s.db, func(q store.Querier) error {
		id, _ := domain.IdentityFromContext(ctx)

		// 1. Fetch the version (verifies org access)
		v, err := s.store.GetVersion(ctx, q, id.OrgID, versionID)
		if err != nil {
			return err
		}

		// 2. Verify program ownership
		if v.ProgramID != programID {
			return &ValidationError{Msg: "version does not belong to this program"}
		}

		// 3. Apply rollback
		if err := s.store.Restore(ctx, q, id.OrgID, programID, v.Snapshot); err != nil {
			return fmt.Errorf("restore: %w", err)
		}
		return nil
	})
}
