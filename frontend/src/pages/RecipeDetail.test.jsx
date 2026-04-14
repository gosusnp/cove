// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { RecipeDetail } from "./RecipeDetail.jsx";

// ─── Mocks ──────────────────────────────────────────────────────────────────

vi.mock("../components/ui/Accordion.jsx", () => ({
	Accordion: ({ children }) => <div>{children}</div>,
	AccordionItem: ({ children }) => <div>{children}</div>,
	AccordionTrigger: ({ children }) => <div>{children}</div>,
	AccordionContent: ({ children }) => <div>{children}</div>,
}));

vi.mock("../components/ui/ConfirmDialog.jsx", () => ({
	ConfirmDialog: ({ openSignal, title, onConfirm }) =>
		openSignal.value ? (
			<div data-testid="mock-confirm-dialog">
				<p>{title}</p>
				<button
					type="button"
					onClick={async () => {
						await onConfirm();
						openSignal.value = false;
					}}
				>
					Confirm
				</button>
				<button
					type="button"
					onClick={() => {
						openSignal.value = false;
					}}
				>
					Cancel
				</button>
			</div>
		) : null,
}));

vi.mock("../components/ui/Combobox.jsx", () => ({
	Combobox: ({ label, value, onChange, options, placeholder, freeform }) => (
		<div>
			{label && <label htmlFor="mock-combobox">{label}</label>}
			<input
				id="mock-combobox"
				data-testid={freeform ? "mock-combobox-freeform" : "mock-combobox"}
				value={value ?? ""}
				placeholder={placeholder}
				onInput={(e) => onChange(e.target.value)}
				list="mock-combobox-options"
			/>
			<datalist id="mock-combobox-options">
				{options.map((o) => (
					<option key={o.value} value={o.value}>
						{o.label}
					</option>
				))}
			</datalist>
		</div>
	),
}));

vi.mock("../components/ui/EditableMarkdown.jsx", () => ({
	EditableMarkdown: ({ value, placeholder, onSave }) => (
		<div data-testid="mock-editable-markdown">
			<span>{value || placeholder}</span>
			<button type="button" onClick={() => onSave("Updated description")}>
				Save description
			</button>
		</div>
	),
}));

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const MOCK_USER = { email: "user@example.com", name: "Test User" };

const MOCK_INGREDIENTS = [
	{ id: 42, name: "Ground beef" },
	{ id: 43, name: "Tomato paste" },
];

const MOCK_FDC_RESULTS = [
	{
		fdc_id: 101,
		name: "Chicken Breast",
		data_type: "Foundation",
		calories_per_100g: 120,
		protein_per_100g: 23,
		fat_per_100g: 2.6,
		carbs_per_100g: 0,
	},
	{
		fdc_id: 102,
		name: "Chicken Leg",
		data_type: "SR Legacy",
		calories_per_100g: 184,
		protein_per_100g: 18,
		fat_per_100g: 12,
		carbs_per_100g: 0,
	},
];

const MOCK_PREP = {
	id: 10,
	name: "Bolognese Sauce",
	description: "A rich meat sauce",
	yield_amount: 0,
	yield_unit: "",
	steps: [],
	ingredients: [],
	is_public: false,
};

const MOCK_RECIPE_WITH_PREP = {
	id: 1,
	name: "Pasta Bolognese",
	servings: 4,
	description: null,
	yield_amount: null,
	yield_unit: null,
	is_public: false,
	preparations: [
		{
			id: 1,
			recipe_id: 1,
			preparation_id: 10,
			position: 1,
			amount: 1,
			unit: "serving",
		},
	],
};

const MOCK_RECIPE_EMPTY = {
	id: 2,
	name: "Empty Recipe",
	servings: 2,
	description: null,
	yield_amount: null,
	yield_unit: null,
	is_public: false,
	preparations: [],
};

const renderDetail = (recipeId = 1, onRecipeUpdated = vi.fn()) =>
	withProviders(
		<RecipeDetail recipeId={recipeId} onRecipeUpdated={onRecipeUpdated} />,
		{ user: MOCK_USER },
	);

function mockFetchForRecipe(recipe = MOCK_RECIPE_EMPTY) {
	vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		if (url === `/api/recipes/${recipe.id}` && !opts?.method) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(recipe),
			});
		}
		if (url.startsWith("/api/preparations/") && !opts?.method) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_PREP),
			});
		}
		// /api/preparations list (for AddComponentForm)
		if (url === "/api/preparations" && !opts?.method) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve([MOCK_PREP]),
			});
		}
		return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
	});
}

function mockFetchForIngredients({
	fdcResults = null,
	createIngredientResponse = { id: 99, name: "Chicken Breast" },
} = {}) {
	vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		if (url === `/api/recipes/1` && !opts?.method) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_RECIPE_WITH_PREP),
			});
		}
		if (url === `/api/preparations/10` && !opts?.method) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_PREP),
			});
		}
		if (url === "/api/preparations" && !opts?.method) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve([MOCK_PREP]),
			});
		}
		if (url === "/api/ingredients" && !opts?.method) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_INGREDIENTS),
			});
		}
		if (url.startsWith("/api/fdc/search") && !opts?.method) {
			const foods = fdcResults ?? MOCK_FDC_RESULTS;
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve({ foods }),
			});
		}
		if (url === "/api/ingredients" && opts?.method === "POST") {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(createIngredientResponse),
			});
		}
		if (
			url.startsWith("/api/preparations/") &&
			url.endsWith("/ingredients") &&
			opts?.method === "POST"
		) {
			return Promise.resolve({
				ok: true,
				json: () =>
					Promise.resolve({
						id: 1,
						ingredient_id: 42,
						name: "Ground beef",
						amount: 2,
						unit: "cups",
						prep: null,
					}),
			});
		}
		return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
	});
}

async function openAddIngredientForm() {
	renderDetail(1);
	await waitFor(() =>
		expect(screen.getByText("Bolognese Sauce")).toBeInTheDocument(),
	);
	fireEvent.click(screen.getByRole("button", { name: "Add ingredient\u2026" }));
	await waitFor(() =>
		expect(screen.getByTestId("mock-combobox-freeform")).toBeInTheDocument(),
	);
}

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

describe("RecipeDetail — loading & errors", () => {
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
			expect(screen.getByText("Failed to load recipe")).toBeInTheDocument(),
		);
	});
});

describe("RecipeDetail — content", () => {
	it("shows recipe name after fetch", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);
	});

	it("shows servings count", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("2 servings")).toBeInTheDocument(),
		);
	});

	it("shows singular serving when servings is 1", async () => {
		mockFetchForRecipe({ ...MOCK_RECIPE_EMPTY, servings: 1 });
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("1 serving")).toBeInTheDocument(),
		);
	});

	it("shows empty components state when recipe has no preparations", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(
				screen.getByText("No preparations yet. Add one to get started."),
			).toBeInTheDocument(),
		);
	});

	it("renders preparation sections when recipe has linked preparations", async () => {
		mockFetchForRecipe(MOCK_RECIPE_WITH_PREP);
		renderDetail(1);
		await waitFor(() =>
			expect(screen.getByText("Bolognese Sauce")).toBeInTheDocument(),
		);
	});

	it("shows amount in recipe for each component", async () => {
		mockFetchForRecipe(MOCK_RECIPE_WITH_PREP);
		renderDetail(1);
		await waitFor(() =>
			expect(screen.getByText("1 serving")).toBeInTheDocument(),
		);
	});
});

describe("RecipeDetail — edit recipe name", () => {
	it("clicking recipe name activates edit mode", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Empty Recipe"));
		expect(screen.getByDisplayValue("Empty Recipe")).toBeInTheDocument();
	});

	it("saves updated recipe name on Enter", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (url === "/api/recipes/2" && opts?.method === "PUT") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({ ...MOCK_RECIPE_EMPTY, name: "New Name" }),
					});
				}
				if (url === "/api/recipes/2") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_RECIPE_EMPTY),
					});
				}
				return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
			});

		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Empty Recipe"));
		const input = screen.getByDisplayValue("Empty Recipe");
		fireEvent.input(input, { target: { value: "New Name" } });
		fireEvent.keyDown(input, { key: "Enter" });

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/recipes/2",
				expect.objectContaining({ method: "PUT" }),
			),
		);
	});

	it("cancels name edit on Escape", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Empty Recipe"));
		const input = screen.getByDisplayValue("Empty Recipe");
		fireEvent.input(input, { target: { value: "Changed" } });
		fireEvent.keyDown(input, { key: "Escape" });

		// Input should be gone, original name restored
		expect(screen.queryByDisplayValue("Changed")).not.toBeInTheDocument();
		expect(screen.getByText("Empty Recipe")).toBeInTheDocument();
	});
});

describe("RecipeDetail — add component form", () => {
	it("shows AddComponentForm when Add button is clicked", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Add"));
		expect(screen.getByText("Add Component")).toBeInTheDocument();
	});

	it("hides the Add button while form is open", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		// Before clicking: no form title
		expect(screen.queryByText("Add Component")).not.toBeInTheDocument();
		fireEvent.click(screen.getByText("Add"));
		// After clicking: form is shown and the trigger button is hidden
		expect(screen.getByText("Add Component")).toBeInTheDocument();
		// The icon+text "Add" button (the trigger) is gone; only the form's submit remains
		const addButtons = screen.queryAllByText("Add");
		// All remaining "Add" text nodes belong to the form submit button, not the trigger
		expect(addButtons.every((btn) => btn.type === "submit")).toBe(true);
	});

	it("cancels add component form", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Add"));
		expect(screen.getByText("Add Component")).toBeInTheDocument();
		fireEvent.click(screen.getByText("Cancel"));
		expect(screen.queryByText("Add Component")).not.toBeInTheDocument();
	});

	it("shows error when submitting without a value", async () => {
		mockFetchForRecipe(MOCK_RECIPE_EMPTY);
		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Add"));
		await waitFor(() =>
			expect(screen.getByTestId("mock-combobox-freeform")).toBeInTheDocument(),
		);
		fireEvent.submit(
			screen.getByTestId("mock-combobox-freeform").closest("form"),
		);
		await waitFor(() =>
			expect(
				screen.getByText("Select or name a preparation."),
			).toBeInTheDocument(),
		);
	});

	it("links existing preparation when an existing option is selected", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (url === "/api/recipes/2" && !opts?.method) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_RECIPE_EMPTY),
					});
				}
				if (url === "/api/preparations" && !opts?.method) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve([MOCK_PREP]),
					});
				}
				if (url === "/api/recipes/2/preparations" && opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({
								id: 1,
								recipe_id: 2,
								preparation_id: 10,
								position: 1,
								amount: 1,
								unit: "serving",
							}),
					});
				}
				if (url === `/api/preparations/${MOCK_PREP.id}` && !opts?.method) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_PREP),
					});
				}
				return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
			});

		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Add"));
		await waitFor(() =>
			expect(screen.getByTestId("mock-combobox-freeform")).toBeInTheDocument(),
		);

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "existing:10" },
		});
		fireEvent.submit(
			screen.getByTestId("mock-combobox-freeform").closest("form"),
		);

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/recipes/2/preparations",
				expect.objectContaining({ method: "POST" }),
			),
		);
		// Should NOT have called POST /api/preparations (create new)
		expect(fetchSpy).not.toHaveBeenCalledWith(
			"/api/preparations",
			expect.objectContaining({ method: "POST" }),
		);
	});

	it("creates a new preparation when a non-numeric name is typed", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (url === "/api/recipes/2" && !opts?.method) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_RECIPE_EMPTY),
					});
				}
				if (url === "/api/preparations" && opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ ...MOCK_PREP, id: 99 }),
					});
				}
				if (url === "/api/preparations" && !opts?.method) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve([]),
					});
				}
				if (url === "/api/recipes/2/preparations" && opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({
								id: 5,
								recipe_id: 2,
								preparation_id: 99,
								position: 1,
								amount: 1,
								unit: "serving",
							}),
					});
				}
				if (url === "/api/preparations/99" && !opts?.method) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ ...MOCK_PREP, id: 99 }),
					});
				}
				return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
			});

		renderDetail(2);
		await waitFor(() =>
			expect(screen.getByText("Empty Recipe")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Add"));
		await waitFor(() =>
			expect(screen.getByTestId("mock-combobox-freeform")).toBeInTheDocument(),
		);

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "Brand New Sauce" },
		});
		fireEvent.submit(
			screen.getByTestId("mock-combobox-freeform").closest("form"),
		);

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/preparations",
				expect.objectContaining({
					method: "POST",
					body: expect.stringContaining("Brand New Sauce"),
				}),
			),
		);
		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/recipes/2/preparations",
				expect.objectContaining({ method: "POST" }),
			),
		);
	});
});

const MOCK_PREP_WITH_INGREDIENT = {
	...MOCK_PREP,
	ingredients: [
		{
			id: 5,
			preparation_id: 10,
			ingredient_id: null,
			name: "Ground beef",
			amount: 500,
			unit: "g",
			prep: null,
		},
	],
};

describe("RecipeDetail — remove component", () => {
	it("shows remove confirm dialog when trash button is clicked", async () => {
		mockFetchForRecipe(MOCK_RECIPE_WITH_PREP);
		renderDetail(1);
		await waitFor(() =>
			expect(screen.getByText("Bolognese Sauce")).toBeInTheDocument(),
		);

		// The Trash2 button is the only button with a lucide-trash-2 svg inside
		const trashButton = screen
			.getAllByRole("button")
			.find((btn) => btn.querySelector(".lucide-trash-2") !== null);
		expect(trashButton).toBeDefined();
		fireEvent.click(trashButton);
		await waitFor(() =>
			expect(screen.getByText("Remove component")).toBeInTheDocument(),
		);
	});

	it("removes component from list after confirming", async () => {
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (url === "/api/recipes/1" && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_RECIPE_WITH_PREP),
				});
			}
			if (url === `/api/preparations/${MOCK_PREP.id}` && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PREP),
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
		});

		renderDetail(1);
		await waitFor(() =>
			expect(screen.getByText("Bolognese Sauce")).toBeInTheDocument(),
		);

		const trashButton = screen
			.getAllByRole("button")
			.find((btn) => btn.querySelector(".lucide-trash-2") !== null);
		expect(trashButton).toBeDefined();
		fireEvent.click(trashButton);

		await waitFor(() =>
			expect(screen.getByText("Remove component")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(screen.queryByText("Bolognese Sauce")).not.toBeInTheDocument(),
		);
	});

	it("shows error and keeps component when DELETE fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (url === "/api/recipes/1" && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_RECIPE_WITH_PREP),
				});
			}
			if (url === `/api/preparations/${MOCK_PREP.id}` && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PREP),
				});
			}
			if (opts?.method === "DELETE") {
				return Promise.resolve({
					ok: false,
					json: () => Promise.resolve({ error: "not found" }),
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
		});

		renderDetail(1);
		await waitFor(() =>
			expect(screen.getByText("Bolognese Sauce")).toBeInTheDocument(),
		);

		const trashButton = screen
			.getAllByRole("button")
			.find((btn) => btn.querySelector(".lucide-trash-2") !== null);
		fireEvent.click(trashButton);
		await waitFor(() =>
			expect(screen.getByText("Remove component")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(screen.getByText("not found")).toBeInTheDocument(),
		);
		// Component must still be in the list
		expect(screen.getByText("Bolognese Sauce")).toBeInTheDocument();
	});
});

describe("RecipeDetail — delete ingredient", () => {
	it("shows error and keeps ingredient when DELETE fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (url === "/api/recipes/1" && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_RECIPE_WITH_PREP),
				});
			}
			if (url === `/api/preparations/${MOCK_PREP.id}` && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PREP_WITH_INGREDIENT),
				});
			}
			if (opts?.method === "DELETE") {
				return Promise.resolve({
					ok: false,
					json: () => Promise.resolve({ error: "delete failed" }),
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
		});

		renderDetail(1);
		await waitFor(() =>
			expect(screen.getByText("Ground beef")).toBeInTheDocument(),
		);

		// The X button next to the ingredient — find the button within the ingredient row
		// There are multiple X buttons; we need the one adjacent to "Ground beef"
		const ingredientText = screen.getByText(/Ground beef/);
		const ingredientRow = ingredientText.closest("div");
		const xButtons = ingredientRow.querySelectorAll("button");
		// Last button in the row is the delete (X) button
		fireEvent.click(xButtons[xButtons.length - 1]);

		await waitFor(() =>
			expect(screen.getByText("delete failed")).toBeInTheDocument(),
		);
		// Ingredient must still be in the list
		expect(screen.getByText("Ground beef")).toBeInTheDocument();
	});
});

describe("RecipeDetail — AddIngredientForm", () => {
	it("submitting without selecting an ingredient shows validation error", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		// The Add button is not shown until an ingredient is selected, so
		// trigger submit via form directly to simulate the edge case.
		const form = screen.getByTestId("mock-combobox-freeform").closest("form");
		fireEvent.submit(form);

		await waitFor(() =>
			expect(
				screen.getByText("Select an ingredient or preparation first"),
			).toBeInTheDocument(),
		);
	});

	it("shows combobox (freeform) when Add ingredient is clicked", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();
		expect(screen.getByTestId("mock-combobox-freeform")).toBeInTheDocument();
	});

	it("selecting an existing ingredient shows prep fields without Name display", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		const combobox = screen.getByTestId("mock-combobox-freeform");
		fireEvent.input(combobox, { target: { value: "ing:42" } });

		// Amount and unit fields should appear
		await waitFor(() =>
			expect(screen.getByLabelText("Amount")).toBeInTheDocument(),
		);
		expect(screen.queryByLabelText("Name (display)")).not.toBeInTheDocument();
		// Prep field should appear
		expect(screen.getByLabelText("Prep")).toBeInTheDocument();
	});

	it("typing a non-existing name enters FDC mode with auto-search", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		const combobox = screen.getByTestId("mock-combobox-freeform");
		fireEvent.input(combobox, { target: { value: "chicken breast" } });

		// Should enter FDC mode — shows "Search FDC" label and results
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		expect(screen.getByText("Chicken Leg")).toBeInTheDocument();
	});

	it("Back button in FDC mode returns to select mode", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "chicken" },
		});
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByRole("button", { name: "← Back" }));

		await waitFor(() =>
			expect(screen.getByTestId("mock-combobox-freeform")).toBeInTheDocument(),
		);
		expect(screen.queryByText("Chicken Breast")).not.toBeInTheDocument();
	});

	it("selecting an FDC result enters confirm mode and shows macros", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "chicken" },
		});
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByText("Chicken Breast"));

		// Confirm mode: shows macros and Name (display) field
		await waitFor(() =>
			expect(screen.getByLabelText("Name (display)")).toBeInTheDocument(),
		);
		// Calorie info should be visible
		expect(screen.getByText(/120/)).toBeInTheDocument();
	});

	it("Change button in confirm mode returns to FDC mode", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "chicken" },
		});
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Chicken Breast"));
		await waitFor(() =>
			expect(screen.getByLabelText("Name (display)")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByRole("button", { name: "← Change" }));

		await waitFor(() =>
			expect(screen.getByLabelText("Search FDC")).toBeInTheDocument(),
		);
		expect(screen.queryByLabelText("Name (display)")).not.toBeInTheDocument();
		expect(
			screen.queryByRole("button", { name: "← Change" }),
		).not.toBeInTheDocument();
	});

	it("Cancel button hides the form", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));

		await waitFor(() =>
			expect(
				screen.queryByTestId("mock-combobox-freeform"),
			).not.toBeInTheDocument(),
		);
	});

	it("saving with an existing ingredient calls prep/ingredients endpoint", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "ing:42" },
		});
		await waitFor(() =>
			expect(screen.getByLabelText("Amount")).toBeInTheDocument(),
		);

		fireEvent.input(screen.getByLabelText("Amount"), {
			target: { value: "2" },
		});
		fireEvent.input(screen.getByTestId("mock-combobox"), {
			target: { value: "cup" },
		});

		fireEvent.click(screen.getByTestId("add-ingredient-submit"));

		await waitFor(() => {
			const calls = global.fetch.mock.calls;
			const prepCall = calls.find(
				([url, opts]) =>
					url === "/api/preparations/10/ingredients" && opts?.method === "POST",
			);
			expect(prepCall).toBeDefined();
			const body = JSON.parse(prepCall[1].body);
			expect(body.ingredient_id).toBe(42);
			expect(body.amount).toBe(2);
			expect(body.unit).toBe("cup");
		});
	});

	it("saving with a sub-preparation sends preparation_ref_id in POST body", async () => {
		// Use a preparation with id=20 (different from prepId=10 so it is not filtered out)
		const BIGA_PREP = {
			id: 20,
			name: "Biga",
			yield_amount: 500,
			yield_unit: "g",
		};
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (url === `/api/recipes/1` && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_RECIPE_WITH_PREP),
				});
			}
			if (url === `/api/preparations/10` && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PREP),
				});
			}
			if (url === "/api/preparations" && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve([MOCK_PREP, BIGA_PREP]),
				});
			}
			if (url === "/api/ingredients" && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_INGREDIENTS),
				});
			}
			if (
				url.startsWith("/api/preparations/") &&
				url.endsWith("/ingredients") &&
				opts?.method === "POST"
			) {
				return Promise.resolve({
					ok: true,
					json: () =>
						Promise.resolve({
							id: 5,
							preparation_ref_id: 20,
							name: "Biga",
							amount: 200,
							unit: "g",
						}),
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
		});

		await openAddIngredientForm();

		// Select the sub-preparation (prep:20)
		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "prep:20" },
		});
		await waitFor(() =>
			expect(screen.getByLabelText("Amount")).toBeInTheDocument(),
		);

		fireEvent.input(screen.getByLabelText("Amount"), {
			target: { value: "200" },
		});

		fireEvent.click(screen.getByTestId("add-ingredient-submit"));

		await waitFor(() => {
			const calls = global.fetch.mock.calls;
			const prepCall = calls.find(
				([url, opts]) =>
					url === "/api/preparations/10/ingredients" && opts?.method === "POST",
			);
			expect(prepCall).toBeDefined();
			const body = JSON.parse(prepCall[1].body);
			expect(body.preparation_ref_id).toBe(20);
			expect(body.ingredient_id).toBeUndefined();
			expect(body.amount).toBe(200);
		});
	});

	it("saving in confirm mode creates ingredient first then adds to prep", async () => {
		mockFetchForIngredients({
			createIngredientResponse: { id: 99, name: "Chicken Breast" },
		});
		await openAddIngredientForm();

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "chicken" },
		});
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Chicken Breast"));
		await waitFor(() =>
			expect(screen.getByLabelText("Name (display)")).toBeInTheDocument(),
		);

		fireEvent.input(screen.getByLabelText("Amount"), {
			target: { value: "300" },
		});

		fireEvent.click(screen.getByTestId("add-ingredient-submit"));

		await waitFor(() => {
			const calls = global.fetch.mock.calls;
			// POST /api/ingredients first
			const createCall = calls.find(
				([url, opts]) => url === "/api/ingredients" && opts?.method === "POST",
			);
			expect(createCall).toBeDefined();
			const createBody = JSON.parse(createCall[1].body);
			expect(createBody.fdc_id).toBe(101);
			expect(createBody.is_public).toBe(true);

			// Then POST to prep/ingredients with the new id
			const prepCall = calls.find(
				([url, opts]) =>
					url === "/api/preparations/10/ingredients" && opts?.method === "POST",
			);
			expect(prepCall).toBeDefined();
			const prepBody = JSON.parse(prepCall[1].body);
			expect(prepBody.ingredient_id).toBe(99);
			expect(prepBody.amount).toBe(300);
		});
	});

	it("unit defaults to the cooking preference unit (g for metric)", async () => {
		mockFetchForIngredients();
		await openAddIngredientForm();

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "ing:42" },
		});
		await waitFor(() =>
			expect(screen.getByLabelText("Amount")).toBeInTheDocument(),
		);

		// Do not change unit — leave as default
		fireEvent.click(screen.getByTestId("add-ingredient-submit"));

		await waitFor(() => {
			const calls = global.fetch.mock.calls;
			const prepCall = calls.find(
				([url, opts]) =>
					url === "/api/preparations/10/ingredients" && opts?.method === "POST",
			);
			expect(prepCall).toBeDefined();
			const body = JSON.parse(prepCall[1].body);
			// Default cooking preference (no user pref → metric fallback) resolves to "g"
			expect(body.unit).toBe("g");
		});
	});

	it("shows error message when save fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (url === `/api/recipes/1` && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_RECIPE_WITH_PREP),
				});
			}
			if (url === `/api/preparations/10` && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PREP),
				});
			}
			if (url === "/api/preparations" && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve([MOCK_PREP]),
				});
			}
			if (url === "/api/ingredients" && !opts?.method) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_INGREDIENTS),
				});
			}
			if (
				url.startsWith("/api/preparations/") &&
				url.endsWith("/ingredients") &&
				opts?.method === "POST"
			) {
				return Promise.resolve({
					ok: false,
					json: () => Promise.resolve({ error: "ingredient already added" }),
				});
			}
			return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
		});

		renderDetail(1);
		await waitFor(() =>
			expect(screen.getByText("Bolognese Sauce")).toBeInTheDocument(),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Add ingredient\u2026" }),
		);
		await waitFor(() =>
			expect(screen.getByTestId("mock-combobox-freeform")).toBeInTheDocument(),
		);

		fireEvent.input(screen.getByTestId("mock-combobox-freeform"), {
			target: { value: "ing:42" },
		});
		await waitFor(() =>
			expect(screen.getByLabelText("Amount")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getByTestId("add-ingredient-submit"));

		await waitFor(() =>
			expect(screen.getByText("ingredient already added")).toBeInTheDocument(),
		);
	});
});
