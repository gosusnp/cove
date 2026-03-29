// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"github.com/gosusnp/cove/backend/internal/store"
)

var ErrNotFound = store.ErrNotFound
var ErrDuplicate = store.ErrDuplicate
var ErrConflict = store.ErrConflict

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }

// ExternalServiceError signals that a third-party service was unreachable or
// returned an unexpected error. Handlers should map this to 422.
type ExternalServiceError struct{ Msg string }

func (e *ExternalServiceError) Error() string { return e.Msg }
