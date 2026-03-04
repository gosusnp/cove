// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/store"
)

var ErrUnauthorized = errors.New("unauthorized")

// withScopedTx executes the given function within a transaction, automatically
// applying RLS session variables if an identity is present in the context.
// If no identity is present, it returns ErrUnauthorized.
func withScopedTx(ctx context.Context, db *sql.DB, fn func(store.Querier) error) error {
	id, ok := domain.IdentityFromContext(ctx)
	if !ok {
		return ErrUnauthorized
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	q := store.NewScopedQuerier(tx, id.OrgID.String(), id.UserID.String())

	if err := fn(q); err != nil {
		return err
	}

	return tx.Commit()
}
