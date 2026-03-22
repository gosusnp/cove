// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain

import (
	"fmt"
	"math"
)

// Unit represents a unit of measurement.
type Unit string

// Mass units.
const (
	UnitGram     Unit = "g"
	UnitKilogram Unit = "kg"
	UnitOunce    Unit = "oz"
	UnitPound    Unit = "lb"
)

// Volume units.
const (
	UnitMilliliter Unit = "ml"
	UnitLiter      Unit = "l"
	UnitTeaspoon   Unit = "tsp"
	UnitTablespoon Unit = "tbsp"
	UnitFluidOunce Unit = "fl_oz"
	UnitCup        Unit = "cup"
)

// Count unit.
const (
	UnitEach Unit = "unit"
)

// UnitCategory classifies a unit by what it measures.
type UnitCategory string

const (
	UnitCategoryMass   UnitCategory = "mass"
	UnitCategoryVolume UnitCategory = "volume"
	UnitCategoryCount  UnitCategory = "count"
)

// Category returns the measurement category of the unit.
func (u Unit) Category() UnitCategory {
	switch u {
	case UnitGram, UnitKilogram, UnitOunce, UnitPound:
		return UnitCategoryMass
	case UnitMilliliter, UnitLiter, UnitTeaspoon, UnitTablespoon, UnitFluidOunce, UnitCup:
		return UnitCategoryVolume
	default:
		return UnitCategoryCount
	}
}

// Valid reports whether the unit is a known value.
func (u Unit) Valid() bool {
	switch u {
	case UnitGram, UnitKilogram, UnitOunce, UnitPound,
		UnitMilliliter, UnitLiter, UnitTeaspoon, UnitTablespoon, UnitFluidOunce, UnitCup,
		UnitEach:
		return true
	}
	return false
}

// massToGrams maps each mass unit to its value in grams.
var massToGrams = map[Unit]float64{
	UnitGram:     1.0,
	UnitKilogram: 1000.0,
	UnitOunce:    28.3495,
	UnitPound:    453.592,
}

// volumeToMl maps each volume unit to its value in milliliters.
var volumeToMl = map[Unit]float64{
	UnitMilliliter: 1.0,
	UnitLiter:      1000.0,
	UnitTeaspoon:   4.92892,
	UnitTablespoon: 14.7868,
	UnitFluidOunce: 29.5735,
	UnitCup:        236.588,
}

// ConvertMass converts an amount from one mass unit to another.
func ConvertMass(amount float64, from, to Unit) (float64, error) {
	fromFactor, ok := massToGrams[from]
	if !ok {
		return 0, fmt.Errorf("unit %q is not a mass unit", from)
	}
	toFactor, ok := massToGrams[to]
	if !ok {
		return 0, fmt.Errorf("unit %q is not a mass unit", to)
	}
	return amount * fromFactor / toFactor, nil
}

// ConvertVolume converts an amount from one volume unit to another.
func ConvertVolume(amount float64, from, to Unit) (float64, error) {
	fromFactor, ok := volumeToMl[from]
	if !ok {
		return 0, fmt.Errorf("unit %q is not a volume unit", from)
	}
	toFactor, ok := volumeToMl[to]
	if !ok {
		return 0, fmt.Errorf("unit %q is not a volume unit", to)
	}
	return amount * fromFactor / toFactor, nil
}

// ConvertMassToVolume converts a mass amount to a volume unit using ingredient density.
// densityGPerMl is the ingredient's density in grams per milliliter (from FDC data).
func ConvertMassToVolume(amount float64, from Unit, densityGPerMl float64, to Unit) (float64, error) {
	if densityGPerMl <= 0 {
		return 0, fmt.Errorf("invalid density: %g", densityGPerMl)
	}
	grams, err := ConvertMass(amount, from, UnitGram)
	if err != nil {
		return 0, err
	}
	ml := grams / densityGPerMl
	return ConvertVolume(ml, UnitMilliliter, to)
}

// UnitSystem represents a measurement system preference.
type UnitSystem string

const (
	UnitSystemMetric      UnitSystem = "metric"
	UnitSystemImperial    UnitSystem = "imperial"
	UnitSystemUSCustomary UnitSystem = "us_customary"
)

// Valid reports whether the system is a known value.
func (s UnitSystem) Valid() bool {
	switch s {
	case UnitSystemMetric, UnitSystemImperial, UnitSystemUSCustomary:
		return true
	}
	return false
}

// DefaultFitnessWeightUnit returns the default weight unit for the fitness domain.
func (s UnitSystem) DefaultFitnessWeightUnit() Unit {
	switch s {
	case UnitSystemImperial:
		return UnitPound
	default:
		return UnitKilogram
	}
}

// DefaultCookingMassUnit returns the default mass unit for the cooking domain.
func (s UnitSystem) DefaultCookingMassUnit() Unit {
	switch s {
	case UnitSystemMetric:
		return UnitGram
	default:
		return UnitOunce
	}
}

// DefaultCookingVolumeUnit returns the default volume unit for the cooking domain.
func (s UnitSystem) DefaultCookingVolumeUnit() Unit {
	switch s {
	case UnitSystemMetric:
		return UnitMilliliter
	case UnitSystemImperial:
		return UnitFluidOunce
	default:
		return UnitCup
	}
}

// QuantizeSteps maps a Unit to its smallest practical increment.
// Units absent from the map are not quantized (pass-through).
type QuantizeSteps map[Unit]float64

// FitnessWeightSteps defines standard plate increments for program building.
// These represent realistic program targets, not exact plate-loading calculations.
var FitnessWeightSteps = QuantizeSteps{
	UnitPound:    2.5,
	UnitKilogram: 1.25,
	UnitOunce:    0.25,
	UnitGram:     1.0,
}

// CookingVolumeSteps defines practical measuring tool increments for cooking.
// Precise units (ml, fl_oz, l) are intentionally absent — no rounding needed.
var CookingVolumeSteps = QuantizeSteps{
	UnitCup:        1.0 / 8,
	UnitTablespoon: 0.5,
	UnitTeaspoon:   0.25,
}

// Quantize rounds amount to the nearest step for the given unit.
// If the unit has no entry in steps, the original amount is returned unchanged.
func Quantize(amount float64, unit Unit, steps QuantizeSteps) float64 {
	step, ok := steps[unit]
	if !ok || step == 0 {
		return amount
	}
	return math.Round(amount/step) * step
}
