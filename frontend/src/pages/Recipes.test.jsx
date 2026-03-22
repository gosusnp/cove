// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { Recipes } from "./Recipes.jsx";

// ─── Mocks ──────────────────────────────────────────────────────────────────

let dialogSignal = null;
vi.mock("../components/ui/Dialog.jsx", () => ({
	Dialog: ({ children, openSignal }) => {
		dialogSignal = openSignal;
		return openSignal.value ? (
			<div data-testid="mock-dialog">{children}</div>
		) : null;
	},
	DialogContent: ({ children }) => <div>{children}</div>,
	DialogTitle: ({ children }) => <h2>{children}</h2>,
	DialogClose: ({ children }) => (
		<button
			type="button"
			data-testid="mock-dialog-close"
			onClick={() => {
				if (dialogSignal) dialogSignal.value = false;
			}}
		>
			{children}
		</button>
	),
}));

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

vi.mock("./RecipeDetail.jsx", () => ({
	RecipeDetail: ({ recipeId }) => (
		<div data-testid="mock-recipe-detail">Recipe {recipeId}</div>
	),
}));

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const MOCK_RECIPES = [
	{ id: 1, name: "Pasta Bolognese", servings: 4 },
	{ id: 2, name: "Chicken Stir Fry", servings: 2 },
];

const MOCK_USER = { email: "user@example.com", name: "Test User" };

const renderRecipes = (path = "/cook/recipes") =>
	withProviders(<Recipes />, { path, user: MOCK_USER });

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

describe("Recipes — list", () => {
	it("renders recipe names from API", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_RECIPES),
		});
		renderRecipes();
		await waitFor(() => {
			expect(screen.getByText("Pasta Bolognese")).toBeInTheDocument();
			expect(screen.getByText("Chicken Stir Fry")).toBeInTheDocument();
		});
	});

	it("shows servings sublabel", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_RECIPES),
		});
		renderRecipes();
		await waitFor(() => {
			expect(screen.getByText("4 servings")).toBeInTheDocument();
			expect(screen.getByText("2 servings")).toBeInTheDocument();
		});
	});

	it("shows singular 'serving' when servings is 1", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([{ id: 3, name: "Oatmeal", servings: 1 }]),
		});
		renderRecipes();
		await waitFor(() =>
			expect(screen.getByText("1 serving")).toBeInTheDocument(),
		);
	});

	it("shows empty state when no recipes", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderRecipes();
		await waitFor(() =>
			expect(screen.getByText("No recipes yet.")).toBeInTheDocument(),
		);
	});

	it("shows error when fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({}),
		});
		renderRecipes();
		await waitFor(() =>
			expect(screen.getByText("Failed to fetch recipes")).toBeInTheDocument(),
		);
	});

	it("shows empty state in detail panel when no recipe selected", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_RECIPES),
		});
		renderRecipes("/cook/recipes");
		await waitFor(() =>
			expect(
				screen.getByText("Select a recipe to view its details."),
			).toBeInTheDocument(),
		);
	});
});

describe("Recipes — active row", () => {
	it("navigates to /cook/recipes/:id when a row is clicked", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_RECIPES),
		});
		renderRecipes();
		await waitFor(() =>
			expect(screen.getByText("Pasta Bolognese")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Pasta Bolognese"));
		expect(window.location.href).toContain("/cook/recipes/1");
	});
});

describe("Recipes — new recipe dialog", () => {
	it("opens new recipe dialog on + New click", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderRecipes();
		await waitFor(() => expect(screen.getByText("+ New")).toBeInTheDocument());
		fireEvent.click(screen.getByText("+ New"));
		expect(screen.getByText("New Recipe")).toBeInTheDocument();
	});

	it("shows validation error when name is empty", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderRecipes();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		expect(screen.getByText("Name is required.")).toBeInTheDocument();
	});

	it("shows validation error when servings is not a positive number", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderRecipes();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "My Recipe" },
		});
		fireEvent.input(screen.getByLabelText("Servings"), {
			target: { value: "0" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		expect(
			screen.getByText("Servings must be a positive number."),
		).toBeInTheDocument();
	});

	it("creates a new recipe and navigates to it", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({ id: 3, name: "My Recipe", servings: 2 }),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_RECIPES),
				});
			});

		renderRecipes();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "My Recipe" },
		});
		fireEvent.input(screen.getByLabelText("Servings"), {
			target: { value: "2" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/recipes",
				expect.objectContaining({
					method: "POST",
					body: JSON.stringify({ name: "My Recipe", servings: 2 }),
				}),
			),
		);
		await waitFor(() =>
			expect(window.location.href).toContain("/cook/recipes/3"),
		);
	});

	it("shows API error when create fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "POST") {
				return Promise.resolve({
					ok: false,
					json: () => Promise.resolve({ error: "Name already taken" }),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve([]),
			});
		});

		renderRecipes();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Dupe" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));

		await waitFor(() =>
			expect(screen.getByText("Name already taken")).toBeInTheDocument(),
		);
	});

	it("cancels new recipe dialog", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderRecipes();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		expect(screen.getByText("New Recipe")).toBeInTheDocument();
		fireEvent.click(screen.getByTestId("mock-dialog-close"));
		expect(screen.queryByText("New Recipe")).not.toBeInTheDocument();
	});
});
