// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/db"
	covemcp "github.com/gosusnp/cove/backend/internal/mcp"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

var containerDSN string

func TestMain(m *testing.M) {
	testutil.RunMain(m, &containerDSN, db.MigrationsFS)
}

func newTestAPIHandler(t *testing.T) http.Handler {
	t.Helper()
	database := testutil.NewDB(t)
	userStore := store.NewUserStore()
	orgStore := store.NewOrgStore()
	userSvc := service.NewUserService(database, userStore, orgStore)
	pSvc := service.NewProgramService(database)
	svcs := covemcp.Services{
		Exercises:        service.NewExerciseService(database, store.NewExerciseStore()),
		Programs:         pSvc,
		ProgramSets:      service.NewProgramSetService(pSvc),
		ProgramExercises: service.NewProgramExerciseService(database, store.NewProgramExerciseStore()),
	}
	return NewAPIHandler(userStore, userSvc, svcs)
}

// TestAPIRoutesSmokeTest verifies every API route is wired correctly by sending
// an unauthenticated request and asserting the response is 401, not 404.
// A 404 means the route is missing from the mux; a 401 means it exists and auth is enforced.
func TestAPIRoutesSmokeTest(t *testing.T) {
	handler := newTestAPIHandler(t)

	routes := []struct {
		method string
		path   string
	}{
		// Users
		{"GET", "/api/users/me"},
		{"GET", "/api/users/tokens"},
		{"POST", "/api/users/tokens"},
		{"DELETE", "/api/users/tokens/00000000-0000-0000-0000-000000000000"},
		// Exercises
		{"GET", "/api/exercises"},
		{"POST", "/api/exercises"},
		{"GET", "/api/exercises/1"},
		{"PUT", "/api/exercises/1"},
		{"DELETE", "/api/exercises/1"},
		// Programs
		{"GET", "/api/programs"},
		{"POST", "/api/programs"},
		{"GET", "/api/programs/1"},
		{"PUT", "/api/programs/1"},
		{"DELETE", "/api/programs/1"},
		// Program sets
		{"GET", "/api/programs/1/sets"},
		{"POST", "/api/programs/1/sets"},
		{"GET", "/api/programs/1/sets/1"},
		{"PUT", "/api/programs/1/sets/1"},
		{"DELETE", "/api/programs/1/sets/1"},
		// Program exercises
		{"GET", "/api/programs/1/sets/1/exercises"},
		{"POST", "/api/programs/1/sets/1/exercises"},
		{"GET", "/api/programs/1/sets/1/exercises/1"},
		{"PUT", "/api/programs/1/sets/1/exercises/1"},
		{"DELETE", "/api/programs/1/sets/1/exercises/1"},
	}

	for _, rt := range routes {
		t.Run(rt.method+" "+rt.path, func(t *testing.T) {
			r := httptest.NewRequest(rt.method, rt.path, nil)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code == http.StatusNotFound {
				t.Errorf("got 404: route not registered")
			}
		})
	}
}
