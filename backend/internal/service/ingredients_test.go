// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
	"github.com/gosusnp/cove/backend/internal/fdc"
	"github.com/gosusnp/cove/backend/internal/store"
	"github.com/gosusnp/cove/backend/internal/testutil"
)

func newTestIngredientService(t *testing.T, fdcClient *fdc.Client) (*IngredientService, context.Context) {
	t.Helper()
	db := testutil.NewDB(t)
	svc := NewIngredientService(db, store.NewIngredientStore(), fdcClient)

	uSvc := NewUserService(db, store.NewUserStore(), store.NewOrgStore())
	user, _, _ := uSvc.GetOrCreate(context.Background(), "ing@example.com", "sub-ing")

	var orgID domain.OrgID
	_ = db.QueryRow(`SELECT org_id FROM cove.org_members WHERE user_id = $1`, user.ID).Scan(&orgID)

	ctx := domain.NewContext(context.Background(), &domain.Identity{
		UserID: user.ID,
		OrgID:  orgID,
	})
	return svc, ctx
}

func TestIngredientService_Create_fdcDensity(t *testing.T) {
	t.Run("populates density from FDC when fdc_id is set", func(t *testing.T) {
		// Serve a food detail response with one volume portion: 1 cup = 240g → density ~1.014 g/ml.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			resp := map[string]any{
				"fdcId":       42,
				"description": "Whole milk",
				"foodPortions": []map[string]any{
					{
						"amount":      1,
						"gramWeight":  240,
						"measureUnit": map[string]any{"name": "cup"},
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(resp)
		}))
		defer srv.Close()

		fdcClient := fdc.NewClientWithBaseURL("testkey", srv.URL)
		svc, ctx := newTestIngredientService(t, fdcClient)

		fdcID := 42
		ing, err := svc.Create(ctx, domain.IngredientParams{
			Name:            "whole milk",
			FdcID:           &fdcID,
			CaloriesPer100g: 61,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ing.DensityGPerMl == nil {
			t.Fatal("expected DensityGPerMl to be populated from FDC, got nil")
		}
		// 240g / 236.588ml ≈ 1.014
		if *ing.DensityGPerMl < 1.0 || *ing.DensityGPerMl > 1.1 {
			t.Errorf("DensityGPerMl = %v, want ~1.014", *ing.DensityGPerMl)
		}
	})

	t.Run("skips FDC call when density already set", func(t *testing.T) {
		called := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
			w.WriteHeader(http.StatusInternalServerError)
		}))
		defer srv.Close()

		fdcClient := fdc.NewClientWithBaseURL("testkey", srv.URL)
		svc, ctx := newTestIngredientService(t, fdcClient)

		fdcID := 1
		density := 0.9
		ing, err := svc.Create(ctx, domain.IngredientParams{
			Name:          "olive oil",
			FdcID:         &fdcID,
			DensityGPerMl: &density,
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if called {
			t.Error("expected FDC GetFood to be skipped when density is already set")
		}
		if ing.DensityGPerMl == nil || *ing.DensityGPerMl != 0.9 {
			t.Errorf("DensityGPerMl = %v, want 0.9", ing.DensityGPerMl)
		}
	})

	t.Run("FDC failure is silently ignored", func(t *testing.T) {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusServiceUnavailable)
		}))
		defer srv.Close()

		fdcClient := fdc.NewClientWithBaseURL("testkey", srv.URL)
		svc, ctx := newTestIngredientService(t, fdcClient)

		fdcID := 99
		ing, err := svc.Create(ctx, domain.IngredientParams{
			Name:  "mystery food",
			FdcID: &fdcID,
		})
		if err != nil {
			t.Fatalf("FDC failure should not fail the Create call, got: %v", err)
		}
		if ing.DensityGPerMl != nil {
			t.Errorf("expected nil DensityGPerMl on FDC failure, got %v", *ing.DensityGPerMl)
		}
	})

	t.Run("no fdc_id skips FDC call", func(t *testing.T) {
		called := false
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			called = true
		}))
		defer srv.Close()

		fdcClient := fdc.NewClientWithBaseURL("testkey", srv.URL)
		svc, ctx := newTestIngredientService(t, fdcClient)

		_, err := svc.Create(ctx, domain.IngredientParams{Name: "salt"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if called {
			t.Error("expected FDC GetFood not to be called when fdc_id is nil")
		}
	})
}
