// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package mcp

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/markdown"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestSessionService(t *testing.T) (*service.WorkoutSessionService, context.Context) {
	t.Helper()
	database := testutil.NewDB(t)

	uSvc := service.NewUserService(database, store.NewUserStore(), store.NewOrgStore())
	user, _, err := uSvc.GetOrCreate(context.Background(), "test@example.com", "sub123")
	if err != nil {
		t.Fatalf("create user: %v", err)
	}

	var orgID domain.OrgID
	if err := database.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user.ID).Scan(&orgID); err != nil {
		t.Fatalf("get org: %v", err)
	}

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})

	svc := service.NewWorkoutSessionService(database, store.NewWorkoutSessionStore(), crypto.NewTestEncryptor())
	return svc, ctx
}

func TestBuildSessionParams_DateParsing(t *testing.T) {
	valid := "2026-06-01T09:00:00Z"
	invalid := "2026-06-01" // date-only, not RFC3339

	t.Run("valid started_at is parsed", func(t *testing.T) {
		p, err := buildSessionParams(createSessionParams{StartedAt: &valid})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.StartedAt == nil {
			t.Fatal("StartedAt is nil, want parsed time")
		}
		if p.StartedAt.UTC().Format(time.RFC3339) != valid {
			t.Errorf("StartedAt = %v, want %v", p.StartedAt, valid)
		}
	})

	t.Run("invalid started_at returns error", func(t *testing.T) {
		_, err := buildSessionParams(createSessionParams{StartedAt: &invalid})
		if err == nil {
			t.Fatal("expected error for date-only started_at, got nil")
		}
		if !strings.Contains(err.Error(), "RFC3339") {
			t.Errorf("error message should mention RFC3339, got: %v", err)
		}
	})

	t.Run("valid completed_at is parsed", func(t *testing.T) {
		p, err := buildSessionParams(createSessionParams{CompletedAt: &valid})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if p.CompletedAt == nil {
			t.Fatal("CompletedAt is nil, want parsed time")
		}
	})

	t.Run("invalid completed_at returns error", func(t *testing.T) {
		_, err := buildSessionParams(createSessionParams{CompletedAt: &invalid})
		if err == nil {
			t.Fatal("expected error for date-only completed_at, got nil")
		}
		if !strings.Contains(err.Error(), "RFC3339") {
			t.Errorf("error message should mention RFC3339, got: %v", err)
		}
	})
}

func TestCreateSessionTool_SensitiveDataRoundtrip(t *testing.T) {
	sessions, ctx := newTestSessionService(t)

	notes := "felt strong today"
	rpe := 8
	progName := "PPL"
	structure := "5x5 squat @ 100kg"
	activity := "strength"

	p := store.WorkoutSessionParams{
		Activity: &activity,
		SensitiveData: domain.SessionSensitiveData{
			PerceivedEffort:  &rpe,
			SessionNotes:     crypto.NewSensitiveStringFromPtr(&notes),
			ProgramName:      crypto.NewSensitiveStringFromPtr(&progName),
			ProgramStructure: crypto.NewSensitiveStringFromPtr(&structure),
		},
	}

	ws, err := sessions.Create(ctx, p)
	if err != nil {
		t.Fatalf("create session: %v", err)
	}

	var text string
	if err := ws.UseSensitiveData(ctx, func(sd domain.SessionSensitiveData) error {
		text = fmt.Sprintf("**ID:** %d\n\n%s", ws.ID, markdown.SessionEntry(ws, sd))
		return nil
	}); err != nil {
		t.Fatalf("UseSensitiveData: %v", err)
	}

	if !strings.Contains(text, fmt.Sprintf("**ID:** %d", ws.ID)) {
		t.Errorf("output missing ID prefix:\n%s", text)
	}
	if !strings.Contains(text, "RPE 8") {
		t.Errorf("output missing RPE:\n%s", text)
	}
	if !strings.Contains(text, notes) {
		t.Errorf("output missing session notes:\n%s", text)
	}
	if !strings.Contains(text, progName) {
		t.Errorf("output missing program name:\n%s", text)
	}
}
