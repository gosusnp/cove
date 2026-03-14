// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package crypto

import (
	"encoding/json"
	"testing"
)

func TestSensitiveString_RoundTrip(t *testing.T) {
	original := NewSensitiveString("hello, world")

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}
	if string(data) != `"hello, world"` {
		t.Errorf("Marshal = %s, want %q", data, "hello, world")
	}

	var got SensitiveString
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if got.String() != "hello, world" {
		t.Errorf("String() = %q, want %q", got.String(), "hello, world")
	}
}

func TestSensitiveString_Zero(t *testing.T) {
	s := NewSensitiveString("secret")
	s.Zero()
	if s.String() != "" {
		t.Errorf("after Zero, String() = %q, want empty", s.String())
	}
}

func TestSensitiveString_EmptyRoundTrip(t *testing.T) {
	var s SensitiveString
	data, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("Marshal empty: %v", err)
	}
	if string(data) != `""` {
		t.Errorf("Marshal empty = %s, want %q", data, "")
	}

	var got SensitiveString
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal empty: %v", err)
	}
	if got.String() != "" {
		t.Errorf("String() = %q, want empty", got.String())
	}
}

func TestSensitiveString_IndependentCopy(t *testing.T) {
	// NewSensitiveString must copy — mutating the original string source
	// must not affect the SensitiveString backing bytes.
	src := "original"
	s := NewSensitiveString(src)
	// We can't mutate a Go string literal, but we can verify the backing
	// slice is independent by zeroing s and confirming the value is gone
	// without affecting a second copy.
	s2 := NewSensitiveString(s.String())
	s.Zero()
	if s2.String() != "original" {
		t.Errorf("s2 affected by s.Zero(); got %q", s2.String())
	}
}
