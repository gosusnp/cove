// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { IngredientDetail } from "./IngredientDetail.jsx";

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const MOCK_USER = { email: "user@example.com", name: "Test User" };

const MOCK_INGREDIENT = {
	id: 1,
	name: "Chicken Breast",
	fdc_id: 171477,
	calories_per_100g: 165,
	protein_per_100g: 31,
	fat_per_100g: 3.6,
	carbs_per_100g: 0,
	density_g_per_ml: null,
	org_id: "00000000-0000-0000-0000-000000000001",
	is_public: true,
};

const renderDetail = (ingredientId = 1) =>
	withProviders(<IngredientDetail ingredientId={ingredientId} />, {
		user: MOCK_USER,
	});

function mockDefaultFetch(ingredient = MOCK_INGREDIENT) {
	vi.spyOn(global, "fetch").mockResolvedValue({
		ok: true,
		json: () => Promise.resolve(ingredient),
	});
}

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

describe("IngredientDetail — loading & errors", () => {
	it("shows Loading… while fetching", () => {
		vi.spyOn(global, "fetch").mockReturnValue(new Promise(() => {}));
		renderDetail();
		expect(screen.getByText("Loading…")).toBeInTheDocument();
	});

	it("shows error when fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({}),
		});
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Failed to load ingredient")).toBeInTheDocument(),
		);
	});
});

describe("IngredientDetail — content", () => {
	it("shows ingredient name after fetch", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
	});

	it("shows all macro fields", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		expect(screen.getByText("165 kcal")).toBeInTheDocument();
		expect(screen.getByText("31 g")).toBeInTheDocument();
		expect(screen.getByText("3.6 g")).toBeInTheDocument();
		expect(screen.getByText("0 g")).toBeInTheDocument();
	});

	it("shows FDC ID when present", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("FDC ID: 171477")).toBeInTheDocument(),
		);
	});

	it("omits FDC ID section when absent", async () => {
		mockDefaultFetch({ ...MOCK_INGREDIENT, fdc_id: null });
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		expect(screen.queryByText(/FDC ID/)).not.toBeInTheDocument();
	});

	it("shows density when present", async () => {
		mockDefaultFetch({ ...MOCK_INGREDIENT, density_g_per_ml: 1.05 });
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("1.05 g/ml")).toBeInTheDocument(),
		);
	});

	it("omits density row when absent", async () => {
		mockDefaultFetch({ ...MOCK_INGREDIENT, density_g_per_ml: null });
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		expect(screen.queryByText(/g\/ml/)).not.toBeInTheDocument();
	});
});
