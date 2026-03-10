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
	RawMux http.Handler // Raw mux with OAuth (no /api prefix)

	// Services
	Exercises *service.ExerciseService
	Programs  *service.ProgramService
	Users     *service.UserService

	// Stores (for direct seeding/verification)
	UserStore *store.UserStore
	OrgStore  *store.OrgStore

	// Internal
	programOwners map[domain.ProgramID]*domain.Identity
	systemToken   string
}

// NewTestApp creates a fully wired application with a fresh, migrated database.
func NewTestApp(t *testing.T) *TestApp {
	t.Helper()

	database := testutil.NewDB(t)

	// Stores
	exStore := store.NewExerciseStore()
	uStore := store.NewUserStore()
	oStore := store.NewOrgStore()

	// Services
	exSvc := service.NewExerciseService(database, exStore)
	pSvc := service.NewProgramService(database)
	uSvc := service.NewUserService(database, uStore, oStore)

	// Create system user for raw mux auth
	sysUser, _, err := uSvc.GetOrCreate(context.Background(), domain.Email("system@test.com"), domain.GoogleSub("system-sub"))
	if err != nil {
		t.Fatalf("create system user: %v", err)
	}
	sysToken, _, err := uSvc.CreateSession(context.Background(), sysUser.ID, "127.0.0.1", "test-agent", "Test OS")
	if err != nil {
		t.Fatalf("create system session: %v", err)
	}

	// Handlers & Mux
	apiMux := http.NewServeMux()
	NewExerciseHandler(exSvc).RegisterRoutes(apiMux)
	NewProgramHandler(pSvc).RegisterRoutes(apiMux)
	NewProgramSetHandler(pSvc).RegisterRoutes(apiMux)
	NewProgramExerciseHandler(pSvc).RegisterRoutes(apiMux)
	NewUserHandler(uSvc).RegisterRoutes(apiMux)

	// Apply OAuth middleware with /api prefix as in server.go
	handler := http.StripPrefix("/api", middleware.OAuth(uSvc, apiMux))
	// Raw mux has OAuth but no /api prefix; DoRaw auto-injects system token
	rawHandler := middleware.OAuth(uSvc, apiMux)

	app := &TestApp{
		T:             t,
		DB:            database,
		Mux:           handler,
		RawMux:        rawHandler,
		Exercises:     exSvc,
		Programs:      pSvc,
		Users:         uSvc,
		UserStore:     uStore,
		OrgStore:      oStore,
		programOwners: make(map[domain.ProgramID]*domain.Identity),
		systemToken:   sysToken,
	}
	return app
}

// Do executes an HTTP request against the app's mux and returns the recorder.
func (a *TestApp) Do(r *http.Request) *httptest.ResponseRecorder {
	a.T.Helper()
	w := httptest.NewRecorder()
	a.Mux.ServeHTTP(w, r)
	return w
}

// DoRaw executes an HTTP request against the app's raw mux (no /api prefix) and returns the recorder.
// It automatically injects the system Bearer token if no Authorization header is set.
func (a *TestApp) DoRaw(r *http.Request) *httptest.ResponseRecorder {
	a.T.Helper()
	if r.Header.Get("Authorization") == "" {
		r.Header.Set("Authorization", "Bearer "+a.systemToken)
	}
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
	token, _, err := a.Users.CreateSession(context.Background(), userID, "127.0.0.1", "test-agent", "Test OS")
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

// SeedUserWithOrg creates a user and returns their ID and OrgID.
func (a *TestApp) SeedUserWithOrg(email, sub string) (domain.UserID, domain.OrgID) {
	a.T.Helper()
	u, _, err := a.Users.GetOrCreate(context.Background(), domain.Email(email), domain.GoogleSub(sub))
	if err != nil {
		a.T.Fatalf("seed user: %v", err)
	}
	// Get org ID from org_members
	var orgID domain.OrgID
	err = a.DB.QueryRow(`SELECT org_id FROM org_members WHERE user_id = $1`, u.ID).Scan(&orgID)
	if err != nil {
		a.T.Fatalf("get seeded org: %v", err)
	}
	return u.ID, orgID
}

// SeedExercise creates a public exercise.
func (a *TestApp) SeedExercise(name string, progression *string) *domain.Exercise {
	a.T.Helper()
	// All exercises now require an org and user.
	// We'll create/use a default "system" user for this seeder.
	u, o := a.SeedUserWithOrg("system@test.com", "system-sub")
	id := &domain.Identity{UserID: u, OrgID: o}
	ctx := domain.NewContext(context.Background(), id)

	ex, err := a.Exercises.Create(ctx, name, progression, nil, true)
	if err != nil {
		a.T.Fatalf("seed exercise: %v", err)
	}
	return ex
}

// SeedExerciseForUser creates an exercise owned by the user's org.
func (a *TestApp) SeedExerciseForUser(ctx context.Context, name string, progression *string, userID domain.UserID, orgID domain.OrgID) *domain.Exercise {
	a.T.Helper()
	id := &domain.Identity{UserID: userID, OrgID: orgID}
	authCtx := domain.NewContext(ctx, id)

	ex, err := a.Exercises.Create(authCtx, name, progression, nil, false)
	if err != nil {
		a.T.Fatalf("seed exercise for user: %v", err)
	}
	return ex
}

// SeedProgram creates a public program and records the owner identity.
func (a *TestApp) SeedProgram(name string) *domain.ProgramLite {
	a.T.Helper()
	u, o := a.SeedUserWithOrg("system@test.com", "system-sub")
	id := &domain.Identity{UserID: u, OrgID: o}
	ctx := domain.NewContext(context.Background(), id)

	p, err := a.Programs.Create(ctx, name, nil, true)
	if err != nil {
		a.T.Fatalf("seed program: %v", err)
	}
	a.programOwners[p.ID] = id
	return p
}

// SeedProgramForUser creates a program owned by the user's org and records the owner identity.
func (a *TestApp) SeedProgramForUser(ctx context.Context, name string, userID domain.UserID, orgID domain.OrgID) *domain.ProgramLite {
	a.T.Helper()
	id := &domain.Identity{UserID: userID, OrgID: orgID}
	authCtx := domain.NewContext(ctx, id)

	p, err := a.Programs.Create(authCtx, name, nil, false)
	if err != nil {
		a.T.Fatalf("seed program for user: %v", err)
	}
	a.programOwners[p.ID] = id
	return p
}

// SeedProgramSet creates a set for a program using the program owner's identity.
func (a *TestApp) SeedProgramSet(programID domain.ProgramID, rounds int) *store.ProgramSet {
	a.T.Helper()
	identity := a.programOwners[programID]
	if identity == nil {
		a.T.Fatalf("no owner recorded for program %v; call SeedProgram or SeedProgramForUser first", programID)
	}
	ctx := domain.NewContext(context.Background(), identity)
	ps, err := a.Programs.CreateSet(ctx, programID, nil, rounds, nil)
	if err != nil {
		a.T.Fatalf("seed program set: %v", err)
	}
	return ps
}

// SeedProgramExercise adds an exercise to a set.
func (a *TestApp) SeedProgramExercise(programID domain.ProgramID, setID int64, exerciseID domain.ExerciseID) *store.ProgramExercise {
	a.T.Helper()

	identity := a.programOwners[programID]
	if identity == nil {
		a.T.Fatalf("no owner recorded for program %v", programID)
	}
	ctx := domain.NewContext(context.Background(), identity)

	pe, err := a.Programs.CreateExercise(ctx, programID, setID, exerciseID, nil, nil, nil, nil)
	if err != nil {
		a.T.Fatalf("seed program exercise: %v", err)
	}
	return pe
}
