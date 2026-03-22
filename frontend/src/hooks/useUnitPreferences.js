// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useAuth } from "../Auth.jsx";

// System → default unit mappings (mirrors backend domain/units.go).
const FITNESS_WEIGHT_UNIT = { metric: "kg", imperial: "lb" };

// Grams per unit (mirrors backend domain/units.go massToGrams).
const MASS_TO_GRAMS = { kg: 1000, lb: 453.592 };

/**
 * Convert a fitness weight value between kg and lb.
 * Returns the converted value rounded to at most 2 decimal places.
 *
 * @param {number} value
 * @param {string} fromUnit — "kg" or "lb"
 * @param {string} toUnit   — "kg" or "lb"
 * @returns {number}
 */
export function convertFitnessWeight(value, fromUnit, toUnit) {
	if (fromUnit === toUnit) return value;
	const grams = value * (MASS_TO_GRAMS[fromUnit] ?? 1);
	const converted = grams / (MASS_TO_GRAMS[toUnit] ?? 1);
	return parseFloat(converted.toFixed(2));
}
const COOKING_MASS_UNIT = { metric: "g", imperial: "oz", us_customary: "oz" };
const COOKING_VOLUME_UNIT = {
	metric: "ml",
	imperial: "fl_oz",
	us_customary: "cup",
};

/**
 * Returns the resolved default units based on:
 *   1. adHocOverride (ephemeral page-scoped override)
 *   2. User profile preference (from AuthContext)
 *   3. Hardcoded fallback: metric
 *
 * @param {{ fitnessUnitSystem?: string, cookingUnitSystem?: string }} adHocOverride
 * @returns {{ fitnessWeightUnit: string, cookingMassUnit: string, cookingVolumeUnit: string }}
 */
export function useUnitPreferences(adHocOverride = {}) {
	const { user } = useAuth();

	const fitnessUnitSystem =
		adHocOverride.fitnessUnitSystem ?? user?.fitness_unit_system ?? "metric";
	const cookingUnitSystem =
		adHocOverride.cookingUnitSystem ?? user?.cooking_unit_system ?? "metric";

	return {
		fitnessWeightUnit: FITNESS_WEIGHT_UNIT[fitnessUnitSystem] ?? "kg",
		cookingMassUnit: COOKING_MASS_UNIT[cookingUnitSystem] ?? "g",
		cookingVolumeUnit: COOKING_VOLUME_UNIT[cookingUnitSystem] ?? "ml",
	};
}
