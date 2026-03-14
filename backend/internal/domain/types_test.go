// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"fmt"
	"testing"

	"github.com/gosusnp/cove/backend/internal/crypto"
)

func sensitivePtr(s string) *crypto.SensitiveString {
	ss := crypto.NewSensitiveString(s)
	return &ss
}

func TestSessionSensitiveData_FormatRedacted(t *testing.T) {
	effort := 7
	sd := SessionSensitiveData{
		PerceivedEffort: &effort,
		SessionNotes:    sensitivePtr("my notes"),
	}
	for _, tc := range []struct {
		verb string
		got  string
	}{
		{"%v", fmt.Sprintf("%v", sd)},
		{"%s", fmt.Sprintf("%s", sd)},
		{"%q", fmt.Sprintf("%q", sd)},
		{"%+v", fmt.Sprintf("%+v", sd)},
		{"%#v", fmt.Sprintf("%#v", sd)},
	} {
		if tc.got != "SessionSensitiveData[REDACTED]" {
			t.Errorf("verb %s = %q, want SessionSensitiveData[REDACTED]", tc.verb, tc.got)
		}
	}
}

func TestSessionSensitiveData_IsEmpty(t *testing.T) {
	effort := 7

	tests := []struct {
		name string
		sd   SessionSensitiveData
		want bool
	}{
		{"all nil", SessionSensitiveData{}, true},
		{"perceived effort set", SessionSensitiveData{PerceivedEffort: &effort}, false},
		{"session notes set", SessionSensitiveData{SessionNotes: sensitivePtr("felt great")}, false},
		{"program name set", SessionSensitiveData{ProgramName: sensitivePtr("Strength")}, false},
		{"program structure set", SessionSensitiveData{ProgramStructure: sensitivePtr("3x5")}, false},
		{"all set", SessionSensitiveData{
			PerceivedEffort:  &effort,
			SessionNotes:     sensitivePtr("felt great"),
			ProgramName:      sensitivePtr("Strength"),
			ProgramStructure: sensitivePtr("3x5"),
		}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.sd.IsEmpty(); got != tt.want {
				t.Errorf("IsEmpty() = %v, want %v", got, tt.want)
			}
		})
	}
}
