// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package crypto

import (
	"encoding/json"
	"fmt"
	"io"
)

// SensitiveString is a string value backed by []byte to allow explicit zeroing
// of the underlying memory. Go string interning means *string fields cannot be
// reliably zeroed; SensitiveString avoids that risk for sensitive plaintext.
//
// It is JSON-transparent: it marshals and unmarshals as a plain JSON string,
// so it is compatible with EncryptedField's JSON round-trip without any
// changes to the encryption layer.
//
// Always use *SensitiveString with omitempty in structs so that absent fields
// are omitted from JSON rather than serialized as "".
type SensitiveString struct {
	b []byte
}

// NewSensitiveString creates a SensitiveString from a plain string.
func NewSensitiveString(s string) SensitiveString {
	b := make([]byte, len(s))
	copy(b, s)
	return SensitiveString{b: b}
}

// String returns a copy of the value as a Go string. The returned string is
// short-lived by convention — use it only for building response values and
// let it be collected promptly.
func (s SensitiveString) String() string {
	return string(s.b)
}

// Zero overwrites the backing bytes with zeros. Call this when the value is
// no longer needed to reduce the window during which plaintext is in memory.
func (s *SensitiveString) Zero() {
	for i := range s.b {
		s.b[i] = 0
	}
	s.b = s.b[:0]
}

// Format implements fmt.Formatter. All verbs emit "[REDACTED]" to prevent
// sensitive values from appearing in logs or error messages.
func (s SensitiveString) Format(f fmt.State, _ rune) {
	_, _ = io.WriteString(f, "SensitiveString[REDACTED]")
}

// GoString implements fmt.GoStringer so that %#v also emits "SensitiveString[REDACTED]".
func (s SensitiveString) GoString() string {
	return "SensitiveString[REDACTED]"
}

// MarshalJSON serializes the value as a JSON string.
func (s SensitiveString) MarshalJSON() ([]byte, error) {
	return json.Marshal(string(s.b))
}

// UnmarshalJSON deserializes a JSON string into the backing []byte.
func (s *SensitiveString) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return fmt.Errorf("SensitiveString: %w", err)
	}
	s.b = make([]byte, len(str))
	copy(s.b, str)
	return nil
}
