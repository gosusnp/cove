// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import "github.com/gosusnp/cove/api/store"

var ErrNotFound = store.ErrNotFound
var ErrDuplicate = store.ErrDuplicate

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }
