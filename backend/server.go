// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package main

import (
	"net/http"

	"github.com/gosusnp/cove/backend/internal/fdc"
	"github.com/gosusnp/cove/backend/internal/handlers"
	"github.com/gosusnp/cove/backend/internal/middleware"
	"github.com/gosusnp/cove/backend/internal/service"
)

// Services bundles all application service dependencies for NewAPIHandler.
type Services struct {
	Users            *service.UserService
	Exercises        *service.ExerciseService
	Programs         *service.ProgramService
	WorkoutSessions  *service.WorkoutSessionService
	TrainingProfiles *service.TrainingProfileService
	Ingredients      *service.IngredientService
	Recipes          *service.RecipeService
	Preparations     *service.PreparationService
	FDCClient        *fdc.Client
}

// NewAPIHandler assembles the API sub-mux and returns a handler that serves
// all /api/... routes, with OAuth and Cache-Control: no-store middleware applied.
func NewAPIHandler(svcs Services, secureCookies bool) http.Handler {
	apiMux := http.NewServeMux()
	handlers.NewActivityHandler().RegisterRoutes(apiMux)
	handlers.NewExerciseHandler(svcs.Exercises).RegisterRoutes(apiMux)
	handlers.NewProgramHandler(svcs.Programs).RegisterRoutes(apiMux)
	handlers.NewProgramSetHandler(svcs.Programs).RegisterRoutes(apiMux)
	handlers.NewProgramExerciseHandler(svcs.Programs).RegisterRoutes(apiMux)
	handlers.NewUserHandler(svcs.Users, secureCookies).RegisterRoutes(apiMux)
	handlers.NewWorkoutSessionHandler(svcs.WorkoutSessions).RegisterRoutes(apiMux)
	handlers.NewTrainingProfileHandler(svcs.TrainingProfiles).RegisterRoutes(apiMux)
	handlers.NewIngredientHandler(svcs.Ingredients).RegisterRoutes(apiMux)
	handlers.NewFDCHandler(svcs.FDCClient).RegisterRoutes(apiMux)
	handlers.NewRecipeHandler(svcs.Recipes).RegisterRoutes(apiMux)
	handlers.NewPreparationHandler(svcs.Preparations).RegisterRoutes(apiMux)

	adminMux := http.NewServeMux()
	handlers.NewServiceAccountHandler(svcs.Users).RegisterRoutes(adminMux)
	apiMux.Handle("/admin/", middleware.RequireAdmin(adminMux))

	return http.StripPrefix("/api", middleware.NoStore(middleware.OAuth(svcs.Users, apiMux)))
}
