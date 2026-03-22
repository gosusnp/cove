// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { useAuth } from "../Auth.jsx";

// System → default unit mappings (mirrors backend domain/units.go).
const FITNESS_WEIGHT_UNIT = { metric: "kg", imperial: "lb" };
const COOKING_MASS_UNIT = { metric: "g", imperial: "oz", us_customary: "oz" };
const COOKING_VOLUME_UNIT = {
	metric: "ml",
	imperial: "fl_oz",
	us_customary: "cup",
};

// Grams per unit (mirrors backend domain/units.go massToGrams).
const MASS_TO_GRAMS = { g: 1, kg: 1000, oz: 28.3495, lb: 453.592 };

// Millilitres per unit (mirrors backend domain/units.go volumeToMl).
const VOLUME_TO_ML = {
	ml: 1,
	l: 1000,
	tsp: 4.92892,
	tbsp: 14.7868,
	fl_oz: 29.5735,
	cup: 236.588,
};

// Smallest meaningful display increment per unit.
// Units absent from this map are not quantized (pass-through).
//
// Distinct from the backend's FitnessWeightSteps / CookingVolumeSteps, which
// govern input step size during authoring. This map governs display precision
// after a unit conversion (e.g. toggling kg → lb on an exercise).
export const DISPLAY_STEPS = {
	kg: 0.01,
	lb: 0.5,
	g: 0.1,
	oz: 0.1,
	ml: 1,
	l: 0.01,
	fl_oz: 0.1,
	tsp: 0.125,
	tbsp: 0.5,
	cup: 0.125,
};

// Returns the number of decimal places encoded in a step value.
function stepDecimals(step) {
	const s = step.toString();
	const i = s.indexOf(".");
	return i === -1 ? 0 : s.length - i - 1;
}

/**
 * Round value to the nearest meaningful display increment for unit.
 * If unit has no entry in DISPLAY_STEPS, returns value unchanged.
 *
 * @param {number} value
 * @param {string} unit
 * @returns {number}
 */
export function quantizeForDisplay(value, unit) {
	const step = DISPLAY_STEPS[unit];
	if (!step) return value;
	return parseFloat(
		(Math.round(value / step) * step).toFixed(stepDecimals(step)),
	);
}

/**
 * Convert a fitness weight value between kg and lb, quantized for display.
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
	return quantizeForDisplay(converted, toUnit);
}

/**
 * Convert a value between any two compatible units (mass↔mass or volume↔volume),
 * quantized for display. Returns value unchanged when units are incompatible
 * (e.g. mass↔volume, or count units like "unit").
 *
 * @param {number} value
 * @param {string} fromUnit
 * @param {string} toUnit
 * @returns {number}
 */
export function convertUnit(value, fromUnit, toUnit) {
	if (fromUnit === toUnit) return value;

	const fromGrams = MASS_TO_GRAMS[fromUnit];
	const toGrams = MASS_TO_GRAMS[toUnit];
	if (fromGrams != null && toGrams != null) {
		return quantizeForDisplay((value * fromGrams) / toGrams, toUnit);
	}

	const fromMl = VOLUME_TO_ML[fromUnit];
	const toMl = VOLUME_TO_ML[toUnit];
	if (fromMl != null && toMl != null) {
		return quantizeForDisplay((value * fromMl) / toMl, toUnit);
	}

	// Incompatible units (mass↔volume, count, free-text) — return unchanged.
	return value;
}

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
