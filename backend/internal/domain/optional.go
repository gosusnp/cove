// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import "encoding/json"

// Optional represents a value that may or may not have been explicitly provided
// by the caller. It is used in patch operations to distinguish "field not
// present in the request" (Set=false) from "field explicitly set, even to
// nil/zero" (Set=true).
//
// UnmarshalJSON is only invoked when the JSON key is present, so Set stays
// false for absent fields regardless of the value type.
type Optional[T any] struct {
	Value T
	Set   bool
}

// UnmarshalJSON implements json.Unmarshaler.
func (o *Optional[T]) UnmarshalJSON(b []byte) error {
	o.Set = true
	return json.Unmarshal(b, &o.Value)
}
