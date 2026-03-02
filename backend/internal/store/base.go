// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
)

// Querier is implemented by both *sql.DB and *sql.Tx, allowing stores to
// execute queries within or outside a transaction transparently.
type Querier interface {
	Query(query string, args ...any) (*sql.Rows, error)
	Exec(query string, args ...any) (sql.Result, error)
	QueryRow(query string, args ...any) *sql.Row

	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// baseStore holds the query executor and is embedded by all stores.
type baseStore struct {
	db Querier
}

// withTx returns a copy of the baseStore that executes queries within tx.
func (b baseStore) withTx(tx *sql.Tx) baseStore {
	return baseStore{db: tx}
}
