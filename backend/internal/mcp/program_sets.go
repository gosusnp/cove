// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// updateProgramSetSchema returns an explicit JSON schema for update_program_set.
// A custom schema is required because the params struct uses domain.Optional[T] fields,
// which jsonschema-go renders as {Value, Set} objects rather than plain primitives.
func updateProgramSetSchema() *jsonschema.Schema {
	return &jsonschema.Schema{
		Type:                 "object",
		AdditionalProperties: &jsonschema.Schema{Not: &jsonschema.Schema{}},
		Required:             []string{"program_id", "id"},
		Properties: map[string]*jsonschema.Schema{
			"program_id": {Type: "integer", Description: "ID of the program that owns the set."},
			"id":         {Type: "integer", Description: "ID of the set to update."},
			"updated_at": {Type: "string", Format: "date-time", Description: "ISO 8601 version token for optimistic locking. Omit for last-write-wins."},
			"name":       {Type: "string", Description: "Label for the set, e.g. 'A', 'B', 'Superset 1'. Omit to leave unchanged."},
			"rounds":     {Type: "integer", Description: "Number of times to complete the full set. Omit to leave unchanged."},
			"rest_s":     {Type: "integer", Description: "Rest in seconds between rounds. Omit to leave unchanged."},
		},
	}
}

type updateProgramSetParams struct {
	ProgramID           int64      `json:"program_id"`
	ID                  int64      `json:"id"`
	UpdatedAt           *time.Time `json:"updated_at"`
	Name                *string    `json:"name"`
	Rounds              *int       `json:"rounds"`
	IntraSetRestSeconds *int       `json:"rest_s"`
}

func registerProgramSetTools(server *mcp.Server, svc *service.ProgramService) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "update_program_set",
		Description: "Partially update a program set. Only provided fields are changed; omitted fields retain their current values. updated_at is the version token for optimistic locking; omit for last-write-wins.",
		InputSchema: updateProgramSetSchema(),
	}, func(ctx context.Context, _ *mcp.CallToolRequest, params updateProgramSetParams) (*mcp.CallToolResult, struct{}, error) {
		patch := service.ProgramSetPatch{UpdatedAt: params.UpdatedAt}
		if params.Name != nil {
			patch.Name = domain.Optional[*string]{Value: params.Name, Set: true}
		}
		if params.Rounds != nil {
			patch.Rounds = domain.Optional[int]{Value: *params.Rounds, Set: true}
		}
		if params.IntraSetRestSeconds != nil {
			patch.IntraSetRestSeconds = domain.Optional[*int]{Value: params.IntraSetRestSeconds, Set: true}
		}
		ps, err := svc.PatchSet(ctx, domain.ProgramID(params.ProgramID), params.ID, patch)
		if err != nil {
			return nil, struct{}{}, err
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: formatProgramSet(ps)}}}, struct{}{}, nil
	})

}

// formatProgramSet formats a store.ProgramSet as a confirmation one-liner.
func formatProgramSet(ps *store.ProgramSet) string {
	label := fmt.Sprintf("Set %d", ps.ID)
	if ps.Name != nil && *ps.Name != "" {
		label = *ps.Name
	}
	rounds := "1 round"
	if ps.Rounds != 1 {
		rounds = fmt.Sprintf("%d rounds", ps.Rounds)
	}
	detail := rounds
	if ps.IntraSetRestSeconds != nil && *ps.IntraSetRestSeconds > 0 {
		detail += fmt.Sprintf(" · %ds rest", *ps.IntraSetRestSeconds)
	}
	return fmt.Sprintf("Set updated: **%s** (id %d, program %d) — %s\n", label, ps.ID, ps.ProgramID, detail)
}

// formatProgramExercise formats a store.ProgramExercise as a confirmation one-liner.
func formatProgramExercise(pe *store.ProgramExercise) string {
	var details []string
	if pe.Laterality != nil {
		details = append(details, *pe.Laterality)
	}
	if pe.TargetReps != nil {
		details = append(details, fmt.Sprintf("%d reps", *pe.TargetReps))
	}
	if pe.TargetDurationSeconds != nil {
		details = append(details, fmt.Sprintf("%ds", *pe.TargetDurationSeconds))
	}
	if w := formatWeight(pe.TargetWeight, pe.WeightUnit); w != "" {
		details = append(details, w)
	}

	line := fmt.Sprintf("Exercise updated: id %d (set %d, exercise id %d)", pe.ID, pe.ProgramSetID, pe.ExerciseID)
	if len(details) > 0 {
		line += " — " + strings.Join(details, " · ")
	}
	return line + "\n"
}

// formatWeight returns a human-readable weight string, or empty for bodyweight.
func formatWeight(weight *float64, unit *domain.Unit) string {
	if weight == nil || *weight == 0 {
		return ""
	}
	u := domain.UnitKilogram
	if unit != nil {
		u = *unit
	}
	if *weight > 0 {
		return fmt.Sprintf("+%g%s", *weight, u)
	}
	return fmt.Sprintf("%g%s (assisted)", *weight, u)
}
