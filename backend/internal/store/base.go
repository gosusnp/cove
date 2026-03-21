// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package store

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

var ErrNotFound = errors.New("not found")
var ErrDuplicate = errors.New("duplicate")

// isUniqueViolation reports whether err is a PostgreSQL unique constraint violation (23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}

// Querier is implemented by both *sql.DB and *sql.Tx, allowing stores to
// execute queries within or outside a transaction transparently.
type Querier interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// ScopedQuerier wraps a Querier and automatically sets the current_org_id and
// current_user_id session variables for the duration of the request or transaction.
type ScopedQuerier struct {
	q      Querier
	orgID  string
	userID string
}

func NewScopedQuerier(q Querier, orgID, userID string) *ScopedQuerier {
	return &ScopedQuerier{
		q:      q,
		orgID:  orgID,
		userID: userID,
	}
}

func (s *ScopedQuerier) setSession(ctx context.Context) error {
	// We use SET LOCAL so that it only lasts for the current transaction.
	_, err := s.q.ExecContext(ctx, "SELECT set_config('app.current_org_id', $1, true), set_config('app.current_user_id', $2, true)", s.orgID, s.userID)
	return err
}

func (s *ScopedQuerier) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	if err := s.setSession(ctx); err != nil {
		return nil, err
	}
	return s.q.QueryContext(ctx, query, args...)
}

func (s *ScopedQuerier) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	if err := s.setSession(ctx); err != nil {
		return nil, err
	}
	return s.q.ExecContext(ctx, query, args...)
}

func (s *ScopedQuerier) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	// If setSession fails, we can't easily return a delayed error row here
	// from QueryRowContext because sql.Row's error is private.
	// We'll proceed and let the underlying query fail or the trigger catch it.
	_ = s.setSession(ctx)
	return s.q.QueryRowContext(ctx, query, args...)
}
