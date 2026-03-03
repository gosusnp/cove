// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"

	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/db"
	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

// TestApp encapsulates the entire application stack for integration testing.
type TestApp struct {
	T      *testing.T
	DB     *sql.DB
	Mux    http.Handler // Fully wired /api handler with OAuth
	RawMux http.Handler // Raw mux without /api prefix or OAuth

	// Services
	Exercises        *service.ExerciseService
	Programs         *service.ProgramService
	ProgramSets      *service.ProgramSetService
	ProgramExercises *service.ProgramExerciseService
	Users            *service.UserService

	// Stores (for direct seeding/verification)
	UserStore *store.UserStore
	OrgStore  *store.OrgStore
}

// NewTestApp creates a fully wired application with a fresh, migrated database.
func NewTestApp(t *testing.T) *TestApp {
	t.Helper()

	database := testutil.NewDB(t, containerDSN, db.MigrationsFS)

	// Stores
	exStore := store.NewExerciseStore(database)
	_ = store.NewProgramStore(database)
	psStore := store.NewProgramSetStore(database)
	peStore := store.NewProgramExerciseStore(database)
	uStore := store.NewUserStore()
	oStore := store.NewOrgStore()

	// Services
	exSvc := service.NewExerciseService(exStore)
	pSvc := service.NewProgramService(database)
	psSvc := service.NewProgramSetService(psStore)
	peSvc := service.NewProgramExerciseService(peStore)
	uSvc := service.NewUserService(database, uStore, oStore)

	// Handlers & Mux
	apiMux := http.NewServeMux()
	NewExerciseHandler(exSvc).RegisterRoutes(apiMux)
	NewProgramHandler(pSvc).RegisterRoutes(apiMux)
	NewProgramSetHandler(psSvc).RegisterRoutes(apiMux)
	NewProgramExerciseHandler(peSvc).RegisterRoutes(apiMux)
	NewUserHandler(uSvc).RegisterRoutes(apiMux)

	// Apply OAuth middleware with /api prefix as in server.go
	handler := http.StripPrefix("/api", middleware.OAuth(uSvc, apiMux))

	return &TestApp{
		T:                t,
		DB:               database,
		Mux:              handler,
		RawMux:           apiMux,
		Exercises:        exSvc,
		Programs:         pSvc,
		ProgramSets:      psSvc,
		ProgramExercises: peSvc,
		Users:            uSvc,
		UserStore:        uStore,
		OrgStore:         oStore,
	}
}

// Do executes an HTTP request against the app's mux and returns the recorder.
func (a *TestApp) Do(r *http.Request) *httptest.ResponseRecorder {
	a.T.Helper()
	w := httptest.NewRecorder()
	a.Mux.ServeHTTP(w, r)
	return w
}

// DoRaw executes an HTTP request against the app's raw mux (no OAuth, no /api prefix) and returns the recorder.
func (a *TestApp) DoRaw(r *http.Request) *httptest.ResponseRecorder {
	a.T.Helper()
	w := httptest.NewRecorder()
	a.RawMux.ServeHTTP(w, r)
	return w
}

// AuthRequest creates a request with a valid Bearer token for the given user.
func (a *TestApp) AuthRequest(method, path string, body any, userID domain.UserID) *http.Request {
	a.T.Helper()
	var rbody io.Reader
	if body != nil {
		if r, ok := body.(io.Reader); ok {
			rbody = r
		} else {
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(body); err != nil {
				a.T.Fatalf("encode body: %v", err)
			}
			rbody = &buf
		}
	}

	r := httptest.NewRequest(method, path, rbody)
	token, err := a.Users.CreateSession(context.Background(), userID, "127.0.0.1", "test-agent", "Test OS")
	if err != nil {
		a.T.Fatalf("create session: %v", err)
	}
	r.Header.Set("Authorization", "Bearer "+token)
	return r
}

// SeedUser creates a user and return their ID.
func (a *TestApp) SeedUser(email, sub string) domain.UserID {
	a.T.Helper()
	u, _, err := a.Users.GetOrCreate(context.Background(), domain.Email(email), domain.GoogleSub(sub))
	if err != nil {
		a.T.Fatalf("seed user: %v", err)
	}
	return u.ID
}

// SeedExercise creates an exercise and returns it.
func (a *TestApp) SeedExercise(name string, progression *string) *store.ExerciseDetail {
	a.T.Helper()
	ex, err := a.Exercises.Create(name, progression)
	if err != nil {
		a.T.Fatalf("seed exercise: %v", err)
	}
	return ex
}

// SeedProgram creates a program and returns it.
func (a *TestApp) SeedProgram(name string) *store.Program {
	a.T.Helper()
	p, err := a.Programs.Create(name)
	if err != nil {
		a.T.Fatalf("seed program: %v", err)
	}
	return p
}

// SeedProgramSet creates a set for a program.
func (a *TestApp) SeedProgramSet(programID int64, rounds int) *store.ProgramSet {
	a.T.Helper()
	ps, err := a.ProgramSets.Create(programID, nil, rounds, nil, nil)
	if err != nil {
		a.T.Fatalf("seed program set: %v", err)
	}
	return ps
}

// SeedProgramExercise adds an exercise to a set.
func (a *TestApp) SeedProgramExercise(setID, exerciseID int64) *store.ProgramExercise {
	a.T.Helper()
	pe, err := a.ProgramExercises.Create(setID, exerciseID, nil, nil, nil, nil, nil)
	if err != nil {
		a.T.Fatalf("seed program exercise: %v", err)
	}
	return pe
}
