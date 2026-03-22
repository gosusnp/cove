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
