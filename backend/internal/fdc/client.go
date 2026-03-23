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
	"strings"
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

// NewClientWithBaseURL returns a Client with a custom base URL. Intended for tests.
func NewClientWithBaseURL(apiKey, baseURL string) *Client {
	return &Client{apiKey: apiKey, baseURL: baseURL, http: &http.Client{Timeout: 10 * time.Second}}
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

// fdcFoodDetail is the shape returned by GET /food/{fdcId}.
type fdcFoodDetail struct {
	FDCID         int                     `json:"fdcId"`
	Description   string                  `json:"description"`
	FoodNutrients []fdcDetailFoodNutrient `json:"foodNutrients"`
	FoodPortions  []fdcFoodPortion        `json:"foodPortions"`
}

// fdcDetailFoodNutrient is the nutrient entry shape in the food detail endpoint.
// The nutrient ID is nested under a "nutrient" object, unlike the search endpoint.
type fdcDetailFoodNutrient struct {
	Nutrient struct {
		ID int `json:"id"`
	} `json:"nutrient"`
	Amount float64 `json:"amount"`
}

// fdcFoodPortion is one serving-size entry within a food detail response.
type fdcFoodPortion struct {
	Amount      float64        `json:"amount"`
	GramWeight  float64        `json:"gramWeight"`
	MeasureUnit fdcMeasureUnit `json:"measureUnit"`
}

type fdcMeasureUnit struct {
	Name string `json:"name"`
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

// GetFood fetches a single food by FDC ID and returns its density if computable from portions.
func (c *Client) GetFood(ctx context.Context, fdcID int) (*FoodResult, error) {
	params := url.Values{}
	params.Set("api_key", c.apiKey)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("%s/food/%d?%s", c.baseURL, fdcID, params.Encode()), nil)
	if err != nil {
		return nil, fmt.Errorf("fdc get food request: %w", err)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fdc get food: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fdc get food: unexpected status %d", resp.StatusCode)
	}

	var body fdcFoodDetail
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("fdc get food decode: %w", err)
	}

	return &FoodResult{
		FDCID:           body.FDCID,
		Name:            body.Description,
		CaloriesPer100g: body.detailNutrient(nutrientEnergy),
		ProteinPer100g:  body.detailNutrient(nutrientProtein),
		FatPer100g:      body.detailNutrient(nutrientFat),
		CarbsPer100g:    body.detailNutrient(nutrientCarbs),
		DensityGPerMl:   computeDensity(body.FoodPortions),
	}, nil
}

// detailNutrient returns the amount for the given nutrient ID from a food detail response, or 0 if not found.
func (f *fdcFoodDetail) detailNutrient(id int) float64 {
	for _, n := range f.FoodNutrients {
		if n.Nutrient.ID == id {
			return n.Amount
		}
	}
	return 0
}

// volumeUnitMl maps known FDC measure unit names (lowercase) to ml equivalents.
var volumeUnitMl = map[string]float64{
	"cup":         236.588,
	"tablespoon":  14.7868,
	"teaspoon":    4.92892,
	"fluid ounce": 29.5735,
	"milliliter":  1.0,
	"liter":       1000.0,
}

// computeDensity derives g/ml density by averaging across all volume-based portions.
// Returns nil when no usable volume portions are present.
func computeDensity(portions []fdcFoodPortion) *float64 {
	var sum float64
	var count int
	for _, p := range portions {
		if p.Amount <= 0 || p.GramWeight <= 0 {
			continue
		}
		ml, ok := volumeUnitMl[strings.ToLower(strings.TrimSpace(p.MeasureUnit.Name))]
		if !ok {
			continue
		}
		sum += p.GramWeight / (p.Amount * ml)
		count++
	}
	if count == 0 {
		return nil
	}
	d := sum / float64(count)
	return &d
}
