// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

// Package fdc provides a client for the USDA Food Data Central API.
package fdc

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"time"
)

const (
	defaultBaseURL = "https://api.nal.usda.gov/fdc/v1"

	// Nutrient IDs used in FDC responses.
	nutrientEnergy  = 1008 // kcal
	nutrientProtein = 1003
	nutrientFat     = 1004 // Total lipid
	nutrientCarbs   = 1005 // Carbohydrate, by difference
)

// Client calls the USDA Food Data Central API.
type Client struct {
	apiKey  string
	baseURL string
	http    *http.Client
}

// NewClient returns a Client using the given FDC API key.
func NewClient(apiKey string) *Client {
	return &Client{apiKey: apiKey, baseURL: defaultBaseURL, http: &http.Client{Timeout: 10 * time.Second}}
}

// FoodResult is one food item returned by a FDC search.
type FoodResult struct {
	FDCID           int      `json:"fdc_id"`
	Name            string   `json:"name"`
	DataType        string   `json:"data_type"`
	CaloriesPer100g float64  `json:"calories_per_100g"`
	ProteinPer100g  float64  `json:"protein_per_100g"`
	FatPer100g      float64  `json:"fat_per_100g"`
	CarbsPer100g    float64  `json:"carbs_per_100g"`
	DensityGPerMl   *float64 `json:"density_g_per_ml,omitempty"`
}

// fdcSearchResponse is the top-level shape of the FDC search endpoint.
type fdcSearchResponse struct {
	Foods []fdcFood `json:"foods"`
}

type fdcFood struct {
	FDCID         int               `json:"fdcId"`
	Description   string            `json:"description"`
	DataType      string            `json:"dataType"`
	FoodNutrients []fdcFoodNutrient `json:"foodNutrients"`
}

type fdcFoodNutrient struct {
	NutrientID int     `json:"nutrientId"`
	Value      float64 `json:"value"`
}

// Search queries FDC for Foundation and SR Legacy foods matching query.
// Foundation results are listed before SR Legacy results.
func (c *Client) Search(ctx context.Context, query string) ([]FoodResult, error) {
	params := url.Values{}
	params.Set("query", query)
	params.Add("dataType", "Foundation")
	params.Add("dataType", "SR Legacy")
	params.Set("pageSize", "50")
	params.Set("api_key", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/foods/search?"+params.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("fdc search request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fdc search: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fdc search: unexpected status %d", resp.StatusCode)
	}

	var body fdcSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("fdc search decode: %w", err)
	}

	// Sort Foundation before SR Legacy.
	sort.SliceStable(body.Foods, func(i, j int) bool {
		return body.Foods[i].DataType == "Foundation" && body.Foods[j].DataType != "Foundation"
	})

	results := make([]FoodResult, 0, len(body.Foods))
	for _, f := range body.Foods {
		results = append(results, FoodResult{
			FDCID:           f.FDCID,
			Name:            f.Description,
			DataType:        f.DataType,
			CaloriesPer100g: f.nutrient(nutrientEnergy),
			ProteinPer100g:  f.nutrient(nutrientProtein),
			FatPer100g:      f.nutrient(nutrientFat),
			CarbsPer100g:    f.nutrient(nutrientCarbs),
		})
	}
	return results, nil
}

// nutrient returns the value for the given nutrient ID, or 0 if not found.
func (f *fdcFood) nutrient(id int) float64 {
	for _, n := range f.FoodNutrients {
		if n.NutrientID == id {
			return n.Value
		}
	}
	return 0
}
