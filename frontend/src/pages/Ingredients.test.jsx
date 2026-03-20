// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { Ingredients } from "./Ingredients.jsx";

// ─── Mocks ───────────────────────────────────────────────────────────────────

vi.mock("../components/ui/ListDetail.jsx", () => ({
	ListDetail: ({ list, detail, emptyState, hasDetail }) => (
		<div data-testid="mock-list-detail">
			<div data-testid="list-panel">{list}</div>
			<div data-testid="detail-panel">
				{hasDetail ? detail : <p>{emptyState}</p>}
			</div>
		</div>
	),
}));

vi.mock("./IngredientDetail.jsx", () => ({
	IngredientDetail: ({ ingredientId }) => (
		<div data-testid="mock-ingredient-detail">Ingredient {ingredientId}</div>
	),
}));

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const MOCK_INGREDIENTS = [
	{
		id: 1,
		name: "Chicken Breast",
		calories_per_100g: 165,
		protein_per_100g: 31,
		fat_per_100g: 3.6,
		carbs_per_100g: 0,
	},
	{
		id: 2,
		name: "Oats",
		calories_per_100g: 389,
		protein_per_100g: 17,
		fat_per_100g: 7,
		carbs_per_100g: 66,
	},
];

const MOCK_USER = { email: "jane@example.com", name: "Jane Smith" };

const renderIngredients = (path = "/cook/ingredients") =>
	withProviders(<Ingredients />, { path, user: MOCK_USER });

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

describe("Ingredients — list", () => {
	it("shows Loading… while fetching", () => {
		vi.spyOn(global, "fetch").mockReturnValue(new Promise(() => {}));
		renderIngredients();
		expect(screen.getByText("Loading…")).toBeInTheDocument();
	});

	it("renders ingredient names after fetch", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_INGREDIENTS),
		});
		renderIngredients();
		await waitFor(() => {
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument();
			expect(screen.getByText("Oats")).toBeInTheDocument();
		});
	});

	it("shows calorie sublabel", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_INGREDIENTS),
		});
		renderIngredients();
		await waitFor(() =>
			expect(screen.getByText("165 kcal / 100 g")).toBeInTheDocument(),
		);
	});

	it("shows empty state when list is empty", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderIngredients();
		await waitFor(() =>
			expect(screen.getByText("No ingredients yet.")).toBeInTheDocument(),
		);
	});

	it("shows error when fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({}),
		});
		renderIngredients();
		await waitFor(() =>
			expect(
				screen.getByText("Failed to fetch ingredients"),
			).toBeInTheDocument(),
		);
	});

	it("shows empty state in detail panel when no ingredient selected", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_INGREDIENTS),
		});
		renderIngredients("/cook/ingredients");
		await waitFor(() =>
			expect(
				screen.getByText("Select an ingredient to view its details."),
			).toBeInTheDocument(),
		);
	});
});

describe("Ingredients — active row", () => {
	it("navigates to /cook/ingredients/:id when a row is clicked", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_INGREDIENTS),
		});
		renderIngredients();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Chicken Breast"));
		expect(window.location.href).toContain("/cook/ingredients/1");
	});
});
