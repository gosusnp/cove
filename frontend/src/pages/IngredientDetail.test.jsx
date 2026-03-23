// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { IngredientDetail } from "./IngredientDetail.jsx";

// ─── Mocks ────────────────────────────────────────────────────────────────────

vi.mock("../components/ui/ConfirmDialog.jsx", () => ({
	ConfirmDialog: ({ openSignal, title, onConfirm }) =>
		openSignal.value ? (
			<div data-testid="mock-confirm-dialog">
				<p>{title}</p>
				<button
					type="button"
					onClick={async () => {
						try {
							await onConfirm();
							openSignal.value = false;
						} catch (_err) {
							// error displayed by the component under test
						}
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

vi.mock("../components/shared/FDCSearch.jsx", () => ({
	FDCSearch: ({ onSelect, onCancel }) => (
		<div data-testid="mock-fdc-search">
			<button
				type="button"
				onClick={() =>
					onSelect({
						fdc_id: 99999,
						name: "Oats",
						calories_per_100g: 389,
						protein_per_100g: 17,
						fat_per_100g: 7,
						carbs_per_100g: 66,
						density_g_per_ml: null,
					})
				}
			>
				Select Oats
			</button>
			{onCancel && (
				<button type="button" onClick={onCancel}>
					CancelSearch
				</button>
			)}
		</div>
	),
}));

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

	it("shows FDC ID value when present", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => expect(screen.getByText("171477")).toBeInTheDocument());
	});

	it("shows — for FDC ID when absent", async () => {
		mockDefaultFetch({ ...MOCK_INGREDIENT, fdc_id: null });
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		expect(screen.queryByText("171477")).not.toBeInTheDocument();
	});

	it("Sync button is disabled when fdc_id is absent", async () => {
		mockDefaultFetch({ ...MOCK_INGREDIENT, fdc_id: null });
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		expect(screen.getByLabelText("Sync nutrition from FDC")).toBeDisabled();
	});

	it("Sync button is enabled when fdc_id is present", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		expect(screen.getByLabelText("Sync nutrition from FDC")).not.toBeDisabled();
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

describe("IngredientDetail — rename", () => {
	it("clicking pencil shows input pre-filled with current name", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Edit ingredient name" }),
			).toBeInTheDocument(),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		expect(screen.getByDisplayValue("Chicken Breast")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
	});

	it("cancel restores the original name display", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(screen.getByText("Chicken Breast")).toBeInTheDocument();
		expect(
			screen.queryByDisplayValue("Chicken Breast"),
		).not.toBeInTheDocument();
	});

	it("shows validation error when name is cleared", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.input(screen.getByDisplayValue("Chicken Breast"), {
			target: { value: "" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));
		expect(screen.getByText("Name is required.")).toBeInTheDocument();
	});

	it("saves new name via PUT and updates display", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "PUT") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({ ...MOCK_INGREDIENT, name: "Turkey Breast" }),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_INGREDIENT),
				});
			});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.input(screen.getByDisplayValue("Chicken Breast"), {
			target: { value: "Turkey Breast" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/ingredients/1",
				expect.objectContaining({ method: "PUT" }),
			),
		);
		await waitFor(() =>
			expect(screen.getByText("Turkey Breast")).toBeInTheDocument(),
		);
	});

	it("shows API error on save failure", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "PUT") {
				return Promise.resolve({
					ok: false,
					json: () => Promise.resolve({ error: "Name already taken" }),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_INGREDIENT),
			});
		});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Edit ingredient name" }),
		);
		fireEvent.input(screen.getByDisplayValue("Chicken Breast"), {
			target: { value: "Taken Name" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(screen.getByText("Name already taken")).toBeInTheDocument(),
		);
	});
});

describe("IngredientDetail — FDC sync", () => {
	it("clicking Sync calls POST /fdc-sync and updates macros", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({ ...MOCK_INGREDIENT, calories_per_100g: 200 }),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_INGREDIENT),
				});
			});

		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByLabelText("Sync nutrition from FDC"),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByLabelText("Sync nutrition from FDC"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/ingredients/1/fdc-sync",
				expect.objectContaining({ method: "POST" }),
			),
		);
		await waitFor(() =>
			expect(screen.getByText("200 kcal")).toBeInTheDocument(),
		);
	});

	it("shows error when sync fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "POST") {
				return Promise.resolve({
					ok: false,
					json: () =>
						Promise.resolve({
							error: "FDC is currently unavailable, please try again",
						}),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_INGREDIENT),
			});
		});

		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByLabelText("Sync nutrition from FDC"),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByLabelText("Sync nutrition from FDC"));

		await waitFor(() =>
			expect(
				screen.getByText("FDC is currently unavailable, please try again"),
			).toBeInTheDocument(),
		);
	});
});

describe("IngredientDetail — change FDC ID", () => {
	it("clicking Change shows FDCSearch", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Change" }),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Change" }));
		expect(screen.getByTestId("mock-fdc-search")).toBeInTheDocument();
	});

	it("selecting a food from FDCSearch calls PUT with new FDC data", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "PUT") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({
								...MOCK_INGREDIENT,
								fdc_id: 99999,
								calories_per_100g: 389,
							}),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_INGREDIENT),
				});
			});

		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Change" }),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Change" }));
		fireEvent.click(screen.getByRole("button", { name: "Select Oats" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/ingredients/1",
				expect.objectContaining({
					method: "PUT",
					body: expect.stringContaining('"fdc_id":99999'),
				}),
			),
		);
		await waitFor(() =>
			expect(screen.queryByTestId("mock-fdc-search")).not.toBeInTheDocument(),
		);
	});

	it("cancel in FDCSearch hides the search panel", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Change" }),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Change" }));
		expect(screen.getByTestId("mock-fdc-search")).toBeInTheDocument();
		fireEvent.click(screen.getByRole("button", { name: "CancelSearch" }));
		expect(screen.queryByTestId("mock-fdc-search")).not.toBeInTheDocument();
	});
});

describe("IngredientDetail — delete", () => {
	it("clicking Delete opens confirm dialog", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Delete ingredient" }),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete ingredient" }));
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
		expect(screen.getByText("Delete Ingredient")).toBeInTheDocument();
	});

	it("confirming delete calls DELETE and invokes onDeleted", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "DELETE") {
					return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_INGREDIENT),
				});
			});

		const onDeleted = vi.fn();
		withProviders(<IngredientDetail ingredientId={1} onDeleted={onDeleted} />, {
			user: MOCK_USER,
		});

		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Delete ingredient" }),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete ingredient" }));
		fireEvent.click(screen.getByRole("button", { name: "Confirm" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/ingredients/1",
				expect.objectContaining({ method: "DELETE" }),
			),
		);
		await waitFor(() => expect(onDeleted).toHaveBeenCalled());
	});

	it("shows error when delete fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: false, json: () => Promise.resolve({}) });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_INGREDIENT),
			});
		});

		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Delete ingredient" }),
			).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete ingredient" }));
		fireEvent.click(screen.getByRole("button", { name: "Confirm" }));

		await waitFor(() =>
			expect(
				screen.getByText("Failed to delete ingredient."),
			).toBeInTheDocument(),
		);
	});
});
