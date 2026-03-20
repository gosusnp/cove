// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import "time"

// -----------------------------------------------------------------------------
// Ingredient
// -----------------------------------------------------------------------------

type IngredientID IntID[struct{ ingredient struct{} }]

// IngredientParams holds the mutable fields for creating or updating an ingredient.
type IngredientParams struct {
	Name            string
	FdcID           *int
	CaloriesPer100g float64
	ProteinPer100g  float64
	FatPer100g      float64
	CarbsPer100g    float64
	DensityGPerMl   *float64
	IsPublic        bool
}

// Ingredient is a food ingredient with nutritional data per 100g.
type Ingredient struct {
	ID              IngredientID `json:"id"`
	Name            string       `json:"name"`
	FdcID           *int         `json:"fdc_id,omitempty"`
	CaloriesPer100g float64      `json:"calories_per_100g"`
	ProteinPer100g  float64      `json:"protein_per_100g"`
	FatPer100g      float64      `json:"fat_per_100g"`
	CarbsPer100g    float64      `json:"carbs_per_100g"`
	DensityGPerMl   *float64     `json:"density_g_per_ml,omitempty"`
	OrgID           OrgID        `json:"org_id"`
	IsPublic        bool         `json:"is_public"`
	CreatedBy       UserID       `json:"created_by"`
	CreatedAt       time.Time    `json:"created_at"`
	UpdatedBy       *UserID      `json:"updated_by,omitempty"`
	UpdatedAt       time.Time    `json:"updated_at"`
}
