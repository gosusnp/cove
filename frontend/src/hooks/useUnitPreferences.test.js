// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { describe, expect, it } from "vitest";
import {
	convertFitnessWeight,
	convertUnit,
	quantizeForDisplay,
} from "./useUnitPreferences.js";

describe("quantizeForDisplay", () => {
	it("rounds to the nearest step for each unit", () => {
		expect(quantizeForDisplay(220.46, "lb")).toBe(220.5);
		expect(quantizeForDisplay(99.793, "kg")).toBe(99.79);
		expect(quantizeForDisplay(283.495, "g")).toBe(283.5);
		expect(quantizeForDisplay(29.6, "ml")).toBe(30);
		expect(quantizeForDisplay(0.4, "tsp")).toBe(0.375); // nearest 1/8
		expect(quantizeForDisplay(0.9, "tbsp")).toBe(1.0); // nearest 1/2
		expect(quantizeForDisplay(0.6, "cup")).toBe(0.625); // nearest 1/8
		expect(quantizeForDisplay(1.23, "fl_oz")).toBe(1.2);
		expect(quantizeForDisplay(1.234, "l")).toBe(1.23);
		expect(quantizeForDisplay(1.23, "oz")).toBe(1.2);
	});

	it("passes through values for units with no step defined", () => {
		expect(quantizeForDisplay(1.23456, "unit")).toBe(1.23456);
	});
});

describe("convertUnit", () => {
	it("returns value unchanged when units are the same", () => {
		expect(convertUnit(100, "g", "g")).toBe(100);
		expect(convertUnit(1, "cup", "cup")).toBe(1);
	});

	it("converts between mass units", () => {
		expect(convertUnit(1000, "g", "kg")).toBe(1); // 1000g → 1kg
		expect(convertUnit(1, "kg", "g")).toBe(1000); // 1kg → 1000g
		expect(convertUnit(453.592, "g", "lb")).toBe(1); // 453.592g → 1lb
	});

	it("converts between volume units", () => {
		expect(convertUnit(1, "cup", "tbsp")).toBe(16); // 1 cup → 16 tbsp
		expect(convertUnit(3, "tsp", "tbsp")).toBe(1); // 3 tsp → 1 tbsp
		expect(convertUnit(1000, "ml", "l")).toBe(1); // 1000ml → 1l
	});

	it("returns value unchanged for incompatible units without density", () => {
		expect(convertUnit(100, "g", "ml")).toBe(100); // mass ↔ volume, no density
		expect(convertUnit(1, "cup", "kg")).toBe(1);
		expect(convertUnit(5, "unit", "g")).toBe(5); // count
	});

	it("converts mass to volume using density", () => {
		// Water: 1 g/ml. 236.588 g → 1 cup (quantized: nearest 0.125)
		expect(convertUnit(236.588, "g", "cup", 1.0)).toBe(1.0);
		// 500 g of water → 500 ml
		expect(convertUnit(500, "g", "ml", 1.0)).toBe(500);
	});

	it("converts volume to mass using density", () => {
		// 1 cup of water (1 g/ml) → 236.588 g → quantized to 236.6
		expect(convertUnit(1, "cup", "g", 1.0)).toBe(236.6);
		// 500 ml of water → 500 g
		expect(convertUnit(500, "ml", "g", 1.0)).toBe(500);
	});

	it("quantizes the result", () => {
		expect(convertUnit(100, "kg", "lb")).toBe(220.5); // nearest 0.5 lb
		expect(convertUnit(1, "tbsp", "tsp")).toBe(3); // exact
		expect(convertUnit(0.5, "cup", "tbsp")).toBe(8); // exact
	});
});

describe("convertFitnessWeight", () => {
	it("returns value unchanged when units are the same", () => {
		expect(convertFitnessWeight(80, "kg", "kg")).toBe(80);
		expect(convertFitnessWeight(185, "lb", "lb")).toBe(185);
	});

	it("converts kg to lb and quantizes to nearest 0.5", () => {
		expect(convertFitnessWeight(100, "kg", "lb")).toBe(220.5);
		expect(convertFitnessWeight(80, "kg", "lb")).toBe(176.5);
	});

	it("converts lb to kg and quantizes to nearest 0.01", () => {
		expect(convertFitnessWeight(225, "lb", "kg")).toBe(102.06);
		expect(convertFitnessWeight(185, "lb", "kg")).toBe(83.91);
	});
});
