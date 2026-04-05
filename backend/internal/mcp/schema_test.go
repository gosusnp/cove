// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
)

// checkNoOptionalLeakage verifies that no property in the schema has a type of
// "object" with "Value"/"Set" sub-properties — the signature of a leaked
// domain.Optional[T] that jsonschema-go emits when an explicit InputSchema is
// not provided for patch tools.
func checkNoOptionalLeakage(t *testing.T, name string, schema *jsonschema.Schema) {
	t.Helper()
	for prop, s := range schema.Properties {
		if s.Type == "object" {
			if _, hasValue := s.Properties["Value"]; hasValue {
				t.Errorf("%s: property %q leaks domain.Optional[T] as {Value, Set} object", name, prop)
			}
		}
	}
}

// checkRequired verifies that all expected fields appear in the schema's Required list.
func checkRequired(t *testing.T, name string, schema *jsonschema.Schema, fields ...string) {
	t.Helper()
	required := make(map[string]bool, len(schema.Required))
	for _, r := range schema.Required {
		required[r] = true
	}
	for _, f := range fields {
		if !required[f] {
			t.Errorf("%s: expected %q in required fields, got %v", name, f, schema.Required)
		}
	}
}

// checkPropertyType verifies a named property has the expected primitive type.
func checkPropertyType(t *testing.T, name, prop, wantType string, schema *jsonschema.Schema) {
	t.Helper()
	s, ok := schema.Properties[prop]
	if !ok {
		t.Errorf("%s: property %q not found", name, prop)
		return
	}
	if s.Type != wantType {
		t.Errorf("%s: property %q has type %q, want %q", name, prop, s.Type, wantType)
	}
}

func TestUpdateProgramSchema(t *testing.T) {
	s := updateProgramSchema()
	checkNoOptionalLeakage(t, "update_program", s)
	checkRequired(t, "update_program", s, "id")
	checkPropertyType(t, "update_program", "id", "integer", s)
	checkPropertyType(t, "update_program", "name", "string", s)
	checkPropertyType(t, "update_program", "is_public", "boolean", s)
}

func TestUpdateProgramSetSchema(t *testing.T) {
	s := updateProgramSetSchema()
	checkNoOptionalLeakage(t, "update_program_set", s)
	checkRequired(t, "update_program_set", s, "program_id", "id")
	checkPropertyType(t, "update_program_set", "rounds", "integer", s)
	checkPropertyType(t, "update_program_set", "name", "string", s)
}

func TestUpdateProgramExerciseSchema(t *testing.T) {
	s := updateProgramExerciseSchema()
	checkNoOptionalLeakage(t, "update_program_exercise", s)
	checkRequired(t, "update_program_exercise", s, "program_id", "set_id", "id")
	checkPropertyType(t, "update_program_exercise", "reps", "integer", s)
	checkPropertyType(t, "update_program_exercise", "weight", "number", s)
	checkPropertyType(t, "update_program_exercise", "weight_unit", "string", s)
}

func TestCreateProgramFullSchema(t *testing.T) {
	s := createProgramFullSchema()
	checkNoOptionalLeakage(t, "create_program_full", s)
	checkRequired(t, "create_program_full", s, "name", "sets")
	checkPropertyType(t, "create_program_full", "name", "string", s)
	checkPropertyType(t, "create_program_full", "is_public", "boolean", s)

	sets, ok := s.Properties["sets"]
	if !ok {
		t.Fatal("create_program_full: property \"sets\" not found")
	}
	if sets.Type != "array" {
		t.Errorf("create_program_full: \"sets\" has type %q, want \"array\"", sets.Type)
	}
	if sets.Items == nil {
		t.Error("create_program_full: \"sets\" has no items schema")
	}
}

func TestTrainingProfileUpdateSchema(t *testing.T) {
	s := trainingProfileUpdateSchema()
	checkNoOptionalLeakage(t, "update_training_profile", s)
	// All fields optional — no required list expected
	if len(s.Required) != 0 {
		t.Errorf("update_training_profile: expected no required fields, got %v", s.Required)
	}
	checkPropertyType(t, "update_training_profile", "motivation", "string", s)
	checkPropertyType(t, "update_training_profile", "constraints", "string", s)

	disciplines, ok := s.Properties["disciplines"]
	if !ok {
		t.Fatal("update_training_profile: property \"disciplines\" not found")
	}
	if disciplines.Type != "array" {
		t.Errorf("update_training_profile: \"disciplines\" has type %q, want \"array\"", disciplines.Type)
	}
}
