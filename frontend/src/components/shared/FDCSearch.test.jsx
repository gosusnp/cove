// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../../test-utils.jsx";
import { FDCSearch } from "./FDCSearch.jsx";

const MOCK_FOODS = [
	{ fdc_id: 171477, name: "Chicken Breast", data_type: "Foundation" },
	{ fdc_id: 171478, name: "Ground Beef", data_type: "SR Legacy" },
];

function mockSearchOk(foods = MOCK_FOODS) {
	vi.spyOn(global, "fetch").mockResolvedValue({
		ok: true,
		json: () => Promise.resolve({ foods }),
	});
}

function renderSearch(props = {}) {
	return withProviders(<FDCSearch onSelect={vi.fn()} {...props} />, {});
}

afterEach(() => {
	vi.restoreAllMocks();
});

describe("FDCSearch — rendering", () => {
	it("renders search field and Search button", () => {
		renderSearch();
		expect(screen.getByLabelText("Search FDC")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Search" })).toBeInTheDocument();
	});

	it("shows Cancel button when onCancel provided", () => {
		renderSearch({ onCancel: vi.fn() });
		expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
	});

	it("omits Cancel button when onCancel not provided", () => {
		renderSearch();
		expect(
			screen.queryByRole("button", { name: "Cancel" }),
		).not.toBeInTheDocument();
	});
});

describe("FDCSearch — search", () => {
	it("clicking Search fetches results", async () => {
		mockSearchOk();
		renderSearch();

		fireEvent.input(screen.getByLabelText("Search FDC"), {
			target: { value: "chicken" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Search" }));

		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
	});

	it("pressing Enter in the search field triggers search", async () => {
		mockSearchOk();
		renderSearch();

		fireEvent.input(screen.getByLabelText("Search FDC"), {
			target: { value: "chicken" },
		});
		fireEvent.keyDown(screen.getByLabelText("Search FDC"), { key: "Enter" });

		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
	});

	it("shows Foundation badge for Foundation foods", async () => {
		mockSearchOk();
		renderSearch();
		fireEvent.input(screen.getByLabelText("Search FDC"), {
			target: { value: "chicken" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Search" }));
		await waitFor(() =>
			expect(screen.getByText("Foundation")).toBeInTheDocument(),
		);
		expect(screen.getByText("SR Legacy")).toBeInTheDocument();
	});

	it("shows 'No results found.' when results are empty", async () => {
		mockSearchOk([]);
		renderSearch();
		fireEvent.input(screen.getByLabelText("Search FDC"), {
			target: { value: "xyz" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Search" }));
		await waitFor(() =>
			expect(screen.getByText("No results found.")).toBeInTheDocument(),
		);
	});

	it("shows error message when fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({ ok: false });
		renderSearch();
		fireEvent.input(screen.getByLabelText("Search FDC"), {
			target: { value: "chicken" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Search" }));
		await waitFor(() =>
			expect(screen.getByText("Search failed")).toBeInTheDocument(),
		);
	});

	it("auto-searches when initialQuery is provided", async () => {
		mockSearchOk();
		renderSearch({ initialQuery: "oats" });
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
	});
});

describe("FDCSearch — selection and cancel", () => {
	it("clicking a result calls onSelect with the food", async () => {
		mockSearchOk();
		const onSelect = vi.fn();
		renderSearch({ onSelect });
		fireEvent.input(screen.getByLabelText("Search FDC"), {
			target: { value: "chicken" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Search" }));
		await waitFor(() =>
			expect(screen.getByText("Chicken Breast")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Chicken Breast"));
		expect(onSelect).toHaveBeenCalledWith(MOCK_FOODS[0]);
	});

	it("clicking Cancel calls onCancel", () => {
		const onCancel = vi.fn();
		renderSearch({ onCancel });
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(onCancel).toHaveBeenCalled();
	});
});
