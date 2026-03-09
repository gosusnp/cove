// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"net/http"

	covemcp "github.com/gosusnp/cove/backend/internal/mcp"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
	"github.com/gosusnp/cove/backend/internal/store"

	"github.com/gosusnp/cove/backend/internal/handlers"
)

// NewAPIHandler assembles the API sub-mux and returns a handler that serves
// all /api/... routes, with OAuth middleware applied.
func NewAPIHandler(userStore *store.UserStore, userSvc *service.UserService, svcs covemcp.Services) http.Handler {
	apiMux := http.NewServeMux()
	handlers.NewExerciseHandler(svcs.Exercises).RegisterRoutes(apiMux)
	handlers.NewProgramHandler(svcs.Programs).RegisterRoutes(apiMux)
	handlers.NewProgramSetHandler(svcs.Programs).RegisterRoutes(apiMux)
	handlers.NewProgramExerciseHandler(svcs.Programs).RegisterRoutes(apiMux)
	handlers.NewUserHandler(userSvc).RegisterRoutes(apiMux)
	return http.StripPrefix("/api", middleware.OAuth(userSvc, apiMux))
}
