// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import "errors"

var ErrNotFound = errors.New("not found")
var ErrDuplicate = errors.New("duplicate")

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }
