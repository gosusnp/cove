// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

package domain_test

import (
	"testing"

	"github.com/gosusnp/cove/backend/internal/domain"
)

func TestUnitSystem_Valid(t *testing.T) {
	cases := []struct {
		s    domain.UnitSystem
		want bool
	}{
		{domain.UnitSystemMetric, true},
		{domain.UnitSystemImperial, true},
		{domain.UnitSystemUSCustomary, true},
		{"furlong", false},
		{"", false},
	}
	for _, c := range cases {
		if got := c.s.Valid(); got != c.want {
			t.Errorf("UnitSystem(%q).Valid() = %v, want %v", c.s, got, c.want)
		}
	}
}

func TestUnitSystem_DefaultFitnessWeightUnit(t *testing.T) {
	if domain.UnitSystemMetric.DefaultFitnessWeightUnit() != domain.UnitKilogram {
		t.Error("metric fitness should be kg")
	}
	if domain.UnitSystemImperial.DefaultFitnessWeightUnit() != domain.UnitPound {
		t.Error("imperial fitness should be lb")
	}
	// us_customary falls back to kg (not applicable to fitness, but must not panic)
	if domain.UnitSystemUSCustomary.DefaultFitnessWeightUnit() != domain.UnitKilogram {
		t.Error("us_customary fitness fallback should be kg")
	}
}

func TestUnitSystem_DefaultCookingUnits(t *testing.T) {
	cases := []struct {
		s      domain.UnitSystem
		mass   domain.Unit
		volume domain.Unit
	}{
		{domain.UnitSystemMetric, domain.UnitGram, domain.UnitMilliliter},
		{domain.UnitSystemImperial, domain.UnitOunce, domain.UnitFluidOunce},
		{domain.UnitSystemUSCustomary, domain.UnitOunce, domain.UnitCup},
	}
	for _, c := range cases {
		if got := c.s.DefaultCookingMassUnit(); got != c.mass {
			t.Errorf("%q DefaultCookingMassUnit = %q, want %q", c.s, got, c.mass)
		}
		if got := c.s.DefaultCookingVolumeUnit(); got != c.volume {
			t.Errorf("%q DefaultCookingVolumeUnit = %q, want %q", c.s, got, c.volume)
		}
	}
}
