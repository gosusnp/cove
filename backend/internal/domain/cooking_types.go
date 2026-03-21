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

// -----------------------------------------------------------------------------
// Preparation
// -----------------------------------------------------------------------------

type PreparationID IntID[struct{ preparation struct{} }]
type PreparationIngredientID IntID[struct{ preparationIngredient struct{} }]

// PreparationStep is one step in a preparation's instructions.
type PreparationStep struct {
	Description string `json:"description"`
}

// PreparationIngredient is an ingredient used in a preparation.
type PreparationIngredient struct {
	ID            PreparationIngredientID `json:"id"`
	PreparationID PreparationID           `json:"preparation_id"`
	IngredientID  IngredientID            `json:"ingredient_id"`
	Name          string                  `json:"name"`
	Amount        float64                 `json:"amount"`
	Unit          string                  `json:"unit"`
	Prep          *string                 `json:"prep,omitempty"`
}

// PreparationParams holds the mutable fields for creating or updating a preparation.
type PreparationParams struct {
	Name        string
	Description *string
	YieldAmount float64
	YieldUnit   string
	Steps       []PreparationStep
	IsPublic    bool
}

// PreparationIngredientParams holds the mutable fields for creating or updating a preparation ingredient.
type PreparationIngredientParams struct {
	IngredientID IngredientID
	Name         string
	Amount       float64
	Unit         string
	Prep         *string
}

// PreparationLite is a trimmed projection used for list endpoints.
type PreparationLite struct {
	ID          PreparationID `json:"id"`
	Name        string        `json:"name"`
	Description *string       `json:"description,omitempty"`
	YieldAmount float64       `json:"yield_amount"`
	YieldUnit   string        `json:"yield_unit"`
	IsPublic    bool          `json:"is_public"`
	OrgID       OrgID         `json:"org_id"`
	CreatedAt   time.Time     `json:"created_at"`
}

// Preparation is the full entity including steps and ingredients.
type Preparation struct {
	ID          PreparationID           `json:"id"`
	Name        string                  `json:"name"`
	Description *string                 `json:"description,omitempty"`
	YieldAmount float64                 `json:"yield_amount"`
	YieldUnit   string                  `json:"yield_unit"`
	Steps       []PreparationStep       `json:"steps"`
	IsPublic    bool                    `json:"is_public"`
	Ingredients []PreparationIngredient `json:"ingredients"`
	OrgID       OrgID                   `json:"org_id"`
	CreatedBy   UserID                  `json:"created_by"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedBy   *UserID                 `json:"updated_by,omitempty"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// -----------------------------------------------------------------------------
// Recipe
// -----------------------------------------------------------------------------

type RecipeID IntID[struct{ recipe struct{} }]
type RecipePreparationID IntID[struct{ recipePreparation struct{} }]

// RecipePreparation links a preparation to a recipe with quantity and position.
type RecipePreparation struct {
	ID            RecipePreparationID `json:"id"`
	RecipeID      RecipeID            `json:"recipe_id"`
	PreparationID PreparationID       `json:"preparation_id"`
	Position      int                 `json:"position"`
	Amount        float64             `json:"amount"`
	Unit          string              `json:"unit"`
}

// RecipePreparationParams holds the mutable fields for adding or updating a recipe preparation.
type RecipePreparationParams struct {
	PreparationID PreparationID
	Position      int
	Amount        float64
	Unit          string
}

// RecipeParams holds the mutable fields for creating or updating a recipe.
type RecipeParams struct {
	Name        string
	Description *string
	YieldAmount *float64
	YieldUnit   *string
	Servings    int
	IsPublic    bool
}

// RecipeLite is a trimmed projection used for list endpoints.
type RecipeLite struct {
	ID          RecipeID  `json:"id"`
	Name        string    `json:"name"`
	Description *string   `json:"description,omitempty"`
	Servings    int       `json:"servings"`
	IsPublic    bool      `json:"is_public"`
	OrgID       OrgID     `json:"org_id"`
	CreatedAt   time.Time `json:"created_at"`
}

// Recipe is the full entity including linked preparations.
type Recipe struct {
	ID           RecipeID            `json:"id"`
	Name         string              `json:"name"`
	Description  *string             `json:"description,omitempty"`
	YieldAmount  *float64            `json:"yield_amount,omitempty"`
	YieldUnit    *string             `json:"yield_unit,omitempty"`
	Servings     int                 `json:"servings"`
	IsPublic     bool                `json:"is_public"`
	Preparations []RecipePreparation `json:"preparations"`
	OrgID        OrgID               `json:"org_id"`
	CreatedBy    UserID              `json:"created_by"`
	CreatedAt    time.Time           `json:"created_at"`
	UpdatedBy    *UserID             `json:"updated_by,omitempty"`
	UpdatedAt    time.Time           `json:"updated_at"`
}
