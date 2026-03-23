// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
)

func TestProgramVersions_ListAndRollback(t *testing.T) {
	app := NewTestApp(t)

	// 1. Setup user and org
	uID, orgID := app.SeedUserWithOrg("user@example.com", "user-sub")

	// 2. Create program (Version 0 state)
	pLite := app.SeedProgramForUser(context.Background(), "Original Name", uID, orgID)

	// Create a set and exercise to have some structure
	ex := app.SeedExerciseForUser(context.Background(), "Pushup", nil, uID, orgID)
	ps := app.SeedProgramSet(pLite.ID, 3)
	app.SeedProgramExercise(pLite.ID, ps.ID, ex.ID)

	// 3. Update program (Trigger Version 1: "Original Name")
	// Using service directly to ensure trigger fires
	id := &domain.Identity{UserID: uID, OrgID: orgID}
	ctx := domain.NewContext(context.Background(), id)
	newName := "Updated Name"
	_, err := app.Programs.Update(ctx, pLite.ID, newName, nil, nil, false)
	if err != nil {
		t.Fatalf("update program: %v", err)
	}

	// 4. List versions
	r := app.AuthRequest("GET", "/api/programs/"+strconv.FormatInt(int64(pLite.ID), 10)+"/versions", nil, uID)
	w := app.Do(r)
	if w.Code != http.StatusOK {
		t.Fatalf("list versions status: %d body: %s", w.Code, w.Body.String())
	}

	var versions []domain.ProgramVersionMeta
	if err := json.NewDecoder(w.Body).Decode(&versions); err != nil {
		t.Fatalf("decode versions: %v", err)
	}

	if len(versions) != 3 {
		t.Fatalf("expected 3 versions (CreateSet, CreateExercise, UpdateName), got %d", len(versions))
	}

	// 5. Rollback to version
	// Restore "Original Name"
	rollbackPath := "/api/programs/" + strconv.FormatInt(int64(pLite.ID), 10) + "/versions/" + strconv.FormatInt(int64(versions[0].ID), 10) + "/rollback"

	rRollback := app.AuthRequest("POST", rollbackPath, nil, uID)
	wRollback := app.Do(rRollback)
	if wRollback.Code != http.StatusNoContent {
		t.Fatalf("rollback status: %d body: %s", wRollback.Code, wRollback.Body.String())
	}

	// 6. Verify program state is reverted
	pCurrent, err := app.Programs.Get(ctx, pLite.ID)
	if err != nil {
		t.Fatalf("get program: %v", err)
	}
	if pCurrent.Name != "Original Name" {
		t.Errorf("expected program name reverted to 'Original Name', got '%s'", pCurrent.Name)
	}

	// 7. Check that rollback created a NEW version (capturing "Updated Name")
	rList2 := app.AuthRequest("GET", "/api/programs/"+strconv.FormatInt(int64(pLite.ID), 10)+"/versions", nil, uID)
	wList2 := app.Do(rList2)
	var versions2 []domain.ProgramVersionMeta
	if err := json.NewDecoder(wList2.Body).Decode(&versions2); err != nil {
		t.Fatalf("decode versions after rollback: %v", err)
	}

	if len(versions2) != 4 {
		t.Errorf("expected 4 versions after rollback, got %d", len(versions2))
	}
}

func TestProgramVersions_ValidationAndIsolation(t *testing.T) {
	app := NewTestApp(t)

	// 1. Setup Org A and User A
	uA, oA := app.SeedUserWithOrg("userA@example.com", "sub-A")
	pA := app.SeedProgramForUser(context.Background(), "Program A", uA, oA)

	// Create a version for Program A (via update)
	ctxA := domain.NewContext(context.Background(), &domain.Identity{UserID: uA, OrgID: oA})
	if _, err := app.Programs.Update(ctxA, pA.ID, "Program A Updated", nil, nil, false); err != nil {
		t.Fatalf("update program A: %v", err)
	}

	// Get the version ID
	rVersions := app.AuthRequest("GET", "/api/programs/"+strconv.FormatInt(int64(pA.ID), 10)+"/versions", nil, uA)
	wVersions := app.Do(rVersions)
	var versions []domain.ProgramVersionMeta
	if err := json.NewDecoder(wVersions.Body).Decode(&versions); err != nil {
		t.Fatalf("decode versions: %v", err)
	}
	vA := versions[0].ID

	// 2. Setup Org B and User B
	uB, oB := app.SeedUserWithOrg("userB@example.com", "sub-B")
	pB := app.SeedProgramForUser(context.Background(), "Program B", uB, oB)

	// ── Cross-org isolation ──

	// User B tries to list versions of Program A -> 404 (Program not found due to RLS)
	rList := app.AuthRequest("GET", "/api/programs/"+strconv.FormatInt(int64(pA.ID), 10)+"/versions", nil, uB)
	if w := app.Do(rList); w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for cross-org list versions, got %d", w.Code)
	}

	// User B tries to get a specific version of Program A -> 404 (Version not found due to RLS)
	rGet := app.AuthRequest("GET", "/api/programs/"+strconv.FormatInt(int64(pA.ID), 10)+"/versions/"+strconv.FormatInt(int64(vA), 10), nil, uB)
	if w := app.Do(rGet); w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for cross-org get version, got %d", w.Code)
	}

	// User B tries to rollback Program B using User A's version ID -> 404
	rRollback := app.AuthRequest("POST", "/api/programs/"+strconv.FormatInt(int64(pB.ID), 10)+"/versions/"+strconv.FormatInt(int64(vA), 10)+"/rollback", nil, uB)
	if w := app.Do(rRollback); w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for cross-org rollback, got %d", w.Code)
	}

	// ── Validation ──

	// GET with non-existent version ID -> 404
	rNonExistent := app.AuthRequest("GET", "/api/programs/"+strconv.FormatInt(int64(pA.ID), 10)+"/versions/999999", nil, uA)
	if w := app.Do(rNonExistent); w.Code != http.StatusNotFound {
		t.Errorf("expected 404 for non-existent version, got %d", w.Code)
	}

	// Create a version for Program B (within Org B)
	ctxB := domain.NewContext(context.Background(), &domain.Identity{UserID: uB, OrgID: oB})
	if _, err := app.Programs.Update(ctxB, pB.ID, "Program B Updated", nil, nil, false); err != nil {
		t.Fatalf("update program B: %v", err)
	}
	rVersionsB := app.AuthRequest("GET", "/api/programs/"+strconv.FormatInt(int64(pB.ID), 10)+"/versions", nil, uB)
	var versionsB []domain.ProgramVersionMeta
	if err := json.NewDecoder(app.Do(rVersionsB).Body).Decode(&versionsB); err != nil {
		t.Fatalf("decode versions B: %v", err)
	}
	vB := versionsB[0].ID

	// Create Program C in Org B
	pC := app.SeedProgramForUser(context.Background(), "Program C", uB, oB)

	// User B tries to rollback Program C using Program B's version ID -> 400
	rMismatch := app.AuthRequest("POST", "/api/programs/"+strconv.FormatInt(int64(pC.ID), 10)+"/versions/"+strconv.FormatInt(int64(vB), 10)+"/rollback", nil, uB)
	wMismatch := app.Do(rMismatch)
	if wMismatch.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for cross-program rollback, got %d body: %s", wMismatch.Code, wMismatch.Body.String())
	}
	if !strings.Contains(wMismatch.Body.String(), "version does not belong to this program") {
		t.Errorf("expected mismatch error message, got: %s", wMismatch.Body.String())
	}
}
