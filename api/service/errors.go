package service

import "github.com/gosusnp/cove/api/store"

var ErrNotFound = store.ErrNotFound

type ValidationError struct{ Msg string }

func (e *ValidationError) Error() string { return e.Msg }
