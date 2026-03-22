// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package fdc

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// newTestClient creates a Client pointed at the given test server URL.
func newTestClient(apiKey string, server *httptest.Server) *Client {
	return &Client{apiKey: apiKey, baseURL: server.URL, http: server.Client()}
}

func TestClient_Search_success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("api_key") != "testkey" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		resp := fdcSearchResponse{
			Foods: []fdcFood{
				{
					FDCID:       2,
					Description: "Chicken Leg",
					DataType:    "SR Legacy",
					FoodNutrients: []fdcFoodNutrient{
						{NutrientID: nutrientEnergy, Value: 184},
					},
				},
				{
					FDCID:       1,
					Description: "Chicken Breast",
					DataType:    "Foundation",
					FoodNutrients: []fdcFoodNutrient{
						{NutrientID: nutrientEnergy, Value: 120},
						{NutrientID: nutrientProtein, Value: 23},
						{NutrientID: nutrientFat, Value: 2.6},
						{NutrientID: nutrientCarbs, Value: 0},
					},
				},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient("testkey", srv)
	results, err := c.Search(context.Background(), "chicken")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("got %d results, want 2", len(results))
	}

	// Foundation must come first.
	if results[0].DataType != "Foundation" {
		t.Errorf("results[0].DataType = %q, want Foundation", results[0].DataType)
	}
	if results[1].DataType != "SR Legacy" {
		t.Errorf("results[1].DataType = %q, want SR Legacy", results[1].DataType)
	}

	// Nutrients extracted correctly.
	f := results[0]
	if f.CaloriesPer100g != 120 {
		t.Errorf("CaloriesPer100g = %v, want 120", f.CaloriesPer100g)
	}
	if f.ProteinPer100g != 23 {
		t.Errorf("ProteinPer100g = %v, want 23", f.ProteinPer100g)
	}
	if f.FatPer100g != 2.6 {
		t.Errorf("FatPer100g = %v, want 2.6", f.FatPer100g)
	}
	if f.CarbsPer100g != 0 {
		t.Errorf("CarbsPer100g = %v, want 0", f.CarbsPer100g)
	}
}

func TestClient_Search_nonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	c := newTestClient("key", srv)
	_, err := c.Search(context.Background(), "broccoli")
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

func TestClient_Search_malformedJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{not valid json`))
	}))
	defer srv.Close()

	c := newTestClient("key", srv)
	_, err := c.Search(context.Background(), "egg")
	if err == nil {
		t.Fatal("expected error for malformed JSON, got nil")
	}
}

func TestClient_Search_missingNutrientReturnsZero(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := fdcSearchResponse{
			Foods: []fdcFood{
				{FDCID: 1, Description: "Water", DataType: "Foundation", FoodNutrients: nil},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	c := newTestClient("key", srv)
	results, err := c.Search(context.Background(), "water")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results[0].CaloriesPer100g != 0 {
		t.Errorf("expected 0 calories for food with no nutrients, got %v", results[0].CaloriesPer100g)
	}
}

func TestClient_GetFood_density(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detail := fdcFoodDetail{
			FDCID:       42,
			Description: "Whole milk",
			FoodPortions: []fdcFoodPortion{
				// 1 cup (236.588 ml) weighs 244 g → density ≈ 1.0313 g/ml
				{Amount: 1, GramWeight: 244, MeasureUnit: fdcMeasureUnit{Name: "cup"}},
				// 1 tablespoon (14.7868 ml) weighs ~15.25 g → density ≈ 1.0313 g/ml
				{Amount: 1, GramWeight: 15.25, MeasureUnit: fdcMeasureUnit{Name: "tablespoon"}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(detail)
	}))
	defer srv.Close()

	c := newTestClient("key", srv)
	result, err := c.GetFood(context.Background(), 42)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FDCID != 42 {
		t.Errorf("FDCID = %d, want 42", result.FDCID)
	}
	if result.DensityGPerMl == nil {
		t.Fatal("expected DensityGPerMl to be non-nil")
	}
	// Average of ~1.0313 and ~1.0314 ≈ 1.031
	if *result.DensityGPerMl < 1.0 || *result.DensityGPerMl > 1.1 {
		t.Errorf("DensityGPerMl = %v, want ~1.031", *result.DensityGPerMl)
	}
}

func TestClient_GetFood_noDensityWhenNoVolumePortions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		detail := fdcFoodDetail{
			FDCID:       7,
			Description: "Beef, ground",
			FoodPortions: []fdcFoodPortion{
				// "serving" is not a volume unit
				{Amount: 1, GramWeight: 85, MeasureUnit: fdcMeasureUnit{Name: "serving"}},
			},
		}
		_ = json.NewEncoder(w).Encode(detail)
	}))
	defer srv.Close()

	c := newTestClient("key", srv)
	result, err := c.GetFood(context.Background(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DensityGPerMl != nil {
		t.Errorf("expected nil DensityGPerMl for non-volume portions, got %v", *result.DensityGPerMl)
	}
}

func TestClient_GetFood_nonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := newTestClient("key", srv)
	_, err := c.GetFood(context.Background(), 999)
	if err == nil {
		t.Fatal("expected error for non-200 response, got nil")
	}
}

func TestComputeDensity_averagesAcrossPortions(t *testing.T) {
	portions := []fdcFoodPortion{
		{Amount: 1, GramWeight: 236.588, MeasureUnit: fdcMeasureUnit{Name: "cup"}},        // 1 g/ml
		{Amount: 1, GramWeight: 14.7868, MeasureUnit: fdcMeasureUnit{Name: "tablespoon"}}, // 1 g/ml
	}
	d := computeDensity(portions)
	if d == nil {
		t.Fatal("expected non-nil density")
	}
	if *d < 0.999 || *d > 1.001 {
		t.Errorf("density = %v, want ~1.0", *d)
	}
}

func TestComputeDensity_nilForNoPortions(t *testing.T) {
	if computeDensity(nil) != nil {
		t.Error("expected nil for empty portions")
	}
}

func TestComputeDensity_skipsZeroAmountOrWeight(t *testing.T) {
	portions := []fdcFoodPortion{
		{Amount: 0, GramWeight: 100, MeasureUnit: fdcMeasureUnit{Name: "cup"}},
		{Amount: 1, GramWeight: 0, MeasureUnit: fdcMeasureUnit{Name: "cup"}},
	}
	if computeDensity(portions) != nil {
		t.Error("expected nil when all portions have zero amount or weight")
	}
}
