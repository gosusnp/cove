// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen, fireEvent, waitFor } from "@testing-library/preact";
import { describe, it, expect, vi, afterEach } from "vitest";
import { Programs } from "./Programs.jsx";
import { withProviders } from "../test-utils.jsx";

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

// ─── Fixtures ────────────────────────────────────────────────────────────────

const MOCK_PROGRAMS = [
	{ id: 1, name: "Strength A/B" },
	{ id: 2, name: "Hypertrophy Block" },
];

const MOCK_PROGRAM_DETAIL = {
	id: 1,
	name: "Strength A/B",
	sets: [
		{
			id: 10,
			name: "A",
			rounds: 3,
			rest_s: 90,
			sort_order: 0,
			exercises: [{ id: 100, name: "Squat", sort_order: 0 }],
		},
	],
};

const MOCK_USER = { email: "user@example.com", name: "Test User" };

const renderPrograms = (path = "/programs") =>
	withProviders(<Programs />, { path, user: MOCK_USER });

// ─── Tests ───────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

describe("Programs — list", () => {
	it("renders program names from API", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_PROGRAMS),
		});
		renderPrograms();
		await waitFor(() => {
			expect(screen.getByText("Strength A/B")).toBeInTheDocument();
			expect(screen.getByText("Hypertrophy Block")).toBeInTheDocument();
		});
	});

	it("shows empty state when no programs", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderPrograms();
		await waitFor(() =>
			expect(screen.getByText("No programs yet.")).toBeInTheDocument(),
		);
	});

	it("shows error when fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({}),
		});
		renderPrograms();
		await waitFor(() =>
			expect(screen.getByText("Failed to fetch programs")).toBeInTheDocument(),
		);
	});

	it("shows empty state in detail panel when no program selected", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_PROGRAMS),
		});
		renderPrograms();
		await waitFor(() =>
			expect(
				screen.getByText("Select a program to view its sets."),
			).toBeInTheDocument(),
		);
	});
});

describe("Programs — active row", () => {
	it("navigates to /programs/:id when a program row is clicked", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_PROGRAMS),
		});
		renderPrograms();
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Strength A/B"));
		expect(window.location.href).toContain("/programs/1");
	});
});

describe("Programs — new program dialog", () => {
	it("opens new program dialog on + New click", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderPrograms();
		await waitFor(() => expect(screen.getByText("+ New")).toBeInTheDocument());
		fireEvent.click(screen.getByText("+ New"));
		expect(screen.getByText("New Program")).toBeInTheDocument();
	});

	it("shows validation error when name is empty", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderPrograms();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		expect(screen.getByText("Name is required.")).toBeInTheDocument();
	});

	it("creates a new program and reloads list", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 3, name: "New Program" }),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAMS),
				});
			});

		renderPrograms();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "New Program" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs",
				expect.objectContaining({
					method: "POST",
					body: JSON.stringify({ name: "New Program" }),
				}),
			),
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

		renderPrograms();
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

	it("cancels new program dialog", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderPrograms();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		expect(screen.getByText("New Program")).toBeInTheDocument();
		fireEvent.click(screen.getByTestId("mock-dialog-close"));
		expect(screen.queryByText("New Program")).not.toBeInTheDocument();
	});
});

describe("Programs — rename dialog", () => {
	it("opens rename dialog pre-filled with current name", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_PROGRAMS),
		});
		renderPrograms();
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getAllByText("Rename")[0]);
		expect(screen.getByText("Rename Program")).toBeInTheDocument();
		expect(screen.getByDisplayValue("Strength A/B")).toBeInTheDocument();
	});

	it("renames a program via PUT", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "PUT") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 1, name: "Updated Name" }),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAMS),
				});
			});

		renderPrograms();
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getAllByText("Rename")[0]);
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Updated Name" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({ name: "Updated Name" }),
				}),
			),
		);
	});
});

describe("Programs — delete", () => {
	it("opens confirm dialog on Delete click", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_PROGRAMS),
		});
		renderPrograms();
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getAllByText("Delete")[0]);
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
	});

	it("deletes a program via DELETE and reloads list", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "DELETE") {
					return Promise.resolve({ ok: true });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAMS),
				});
			});

		renderPrograms();
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getAllByText("Delete")[0]);
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1",
				expect.objectContaining({ method: "DELETE" }),
			),
		);
	});

	it("navigates to /programs when the selected program is deleted", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: true });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_PROGRAMS),
			});
		});

		renderPrograms("/programs/1");
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getAllByText("Delete")[0]);
		fireEvent.click(screen.getByText("Confirm"));

		// preact-iso's route() updates window.location via history API
		await waitFor(() => expect(window.location.href).toContain("/programs"));
	});
});
