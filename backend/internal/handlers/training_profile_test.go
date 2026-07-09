// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/crypto"
	"github.com/gosusnp/cove/backend/internal/domain"
)

type trainingProfileResp struct {
	UserID      domain.UserID                      `json:"user_id"`
	Motivation  *string                            `json:"motivation,omitempty"`
	Constraints *string                            `json:"constraints,omitempty"`
	Disciplines []trainingProfileDisciplineRequest `json:"disciplines,omitempty"`
}

func TestTrainingProfileHandler(t *testing.T) {
	t.Run("GET - not found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		r := app.AuthRequest(http.MethodGet, "/api/users/me/training-profile", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("PUT - create new profile", func(t *testing.T) {
		app := NewTestApp(t)
		u1, _ := app.SeedUserWithOrg("u1@test.com", "sub1")
		motivation := "To get strong"
		disciplineName := "Bouldering"
		body := mustJSON(t, map[string]any{
			"motivation": motivation,
			"disciplines": []any{
				map[string]any{
					"name":           disciplineName,
					"years_practice": 5.5,
					"level":          "V7",
				},
			},
		})
		r := app.AuthRequest(http.MethodPut, "/api/users/me/training-profile", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got trainingProfileResp
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.UserID != u1 {
			t.Errorf("got user_id %v, want %v", got.UserID, u1)
		}
		if got.Motivation == nil || *got.Motivation != motivation {
			t.Errorf("got motivation %v, want %q", got.Motivation, motivation)
		}
		if len(got.Disciplines) != 1 || got.Disciplines[0].Name == nil || *got.Disciplines[0].Name != disciplineName {
			t.Errorf("got disciplines %+v", got.Disciplines)
		}
	})

	t.Run("PATCH - partial update", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		motivation := "Original motivation"
		constraints := "Original constraints"
		app.SeedTrainingProfile(context.Background(), u1, o1, domain.TrainingProfileSensitiveData{
			Motivation:  crypto.NewSensitiveStringFromPtr(&motivation),
			Constraints: crypto.NewSensitiveStringFromPtr(&constraints),
		})

		newMotivation := "Updated motivation"
		body := mustJSON(t, map[string]any{"motivation": newMotivation})
		r := app.AuthRequest(http.MethodPatch, "/api/users/me/training-profile", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got trainingProfileResp
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Motivation == nil || *got.Motivation != newMotivation {
			t.Errorf("got motivation %v, want %q", got.Motivation, newMotivation)
		}
		if got.Constraints == nil || *got.Constraints != constraints {
			t.Errorf("got constraints %v, want %q — should be preserved", got.Constraints, constraints)
		}
	})

	t.Run("PUT - update existing profile", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		motivation1 := "Motivation 1"
		app.SeedTrainingProfile(context.Background(), u1, o1, domain.TrainingProfileSensitiveData{
			Motivation: crypto.NewSensitiveStringFromPtr(&motivation1),
		})

		motivation2 := "Motivation 2"
		body := mustJSON(t, map[string]any{"motivation": motivation2})
		r := app.AuthRequest(http.MethodPut, "/api/users/me/training-profile", body, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got trainingProfileResp
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Motivation == nil || *got.Motivation != motivation2 {
			t.Errorf("got motivation %v, want %q", got.Motivation, motivation2)
		}
	})

	t.Run("GET - found", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		motivation := "Test motivation"
		app.SeedTrainingProfile(context.Background(), u1, o1, domain.TrainingProfileSensitiveData{
			Motivation: crypto.NewSensitiveStringFromPtr(&motivation),
		})

		r := app.AuthRequest(http.MethodGet, "/api/users/me/training-profile", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusOK {
			t.Errorf("got status %d, want %d", w.Code, http.StatusOK)
		}
		var got trainingProfileResp
		if err := json.NewDecoder(w.Body).Decode(&got); err != nil {
			t.Fatalf("decode: %v", err)
		}
		if got.Motivation == nil || *got.Motivation != motivation {
			t.Errorf("got motivation %v, want %q", got.Motivation, motivation)
		}
	})

	t.Run("isolation - user B cannot see user A profile", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		u2, _ := app.SeedUserWithOrg("u2@test.com", "sub2")

		motivation := "Private motivation"
		app.SeedTrainingProfile(context.Background(), u1, o1, domain.TrainingProfileSensitiveData{
			Motivation: crypto.NewSensitiveStringFromPtr(&motivation),
		})

		// u2 tries to get their own profile (which doesn't exist)
		r := app.AuthRequest(http.MethodGet, "/api/users/me/training-profile", nil, u2)
		w := app.Do(r)

		if w.Code != http.StatusNotFound {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNotFound)
		}
	})

	t.Run("DELETE", func(t *testing.T) {
		app := NewTestApp(t)
		u1, o1 := app.SeedUserWithOrg("u1@test.com", "sub1")
		app.SeedTrainingProfile(context.Background(), u1, o1, domain.TrainingProfileSensitiveData{})

		r := app.AuthRequest(http.MethodDelete, "/api/users/me/training-profile", nil, u1)
		w := app.Do(r)

		if w.Code != http.StatusNoContent {
			t.Errorf("got status %d, want %d", w.Code, http.StatusNoContent)
		}

		// Verify gone
		r2 := app.AuthRequest(http.MethodGet, "/api/users/me/training-profile", nil, u1)
		w2 := app.Do(r2)
		if w2.Code != http.StatusNotFound {
			t.Errorf("after delete: got status %d, want %d", w2.Code, http.StatusNotFound)
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		app := NewTestApp(t)
		r := httptest.NewRequest(http.MethodGet, "/api/users/me/training-profile", nil)
		w := app.Do(r)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("got status %d, want %d", w.Code, http.StatusUnauthorized)
		}
	})
}
