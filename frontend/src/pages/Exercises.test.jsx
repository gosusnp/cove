// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { Exercises } from "./Exercises.jsx";

// ─── Mocks ───────────────────────────────────────────────────────────────────

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

vi.mock("../components/ui/Switch.jsx", () => ({
	Switch: ({ checkedSignal, id }) => (
		<input
			type="checkbox"
			id={id}
			checked={checkedSignal?.value ?? false}
			onChange={(e) => {
				if (checkedSignal) checkedSignal.value = e.target.checked;
			}}
		/>
	),
}));

vi.mock("./ExerciseDetail.jsx", () => ({
	ExerciseDetail: ({ exerciseId }) => (
		<div data-testid="mock-exercise-detail">Exercise {exerciseId}</div>
	),
}));

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const MOCK_EXERCISES = [
	{
		id: 1,
		name: "Diamond Push-up",
		progression: "Push-up",
		description: "Triceps and chest",
		is_public: true,
	},
	{
		id: 2,
		name: "Archer Push-up",
		progression: "Push-up",
		description: "Unilateral focus",
		is_public: false,
	},
];

const MOCK_USER = { email: "jane@example.com", name: "Jane Smith" };

const renderExercises = (path = "/exercises") =>
	withProviders(<Exercises />, { path, user: MOCK_USER });

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

describe("Exercises — list", () => {
	it("renders exercise names from API", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_EXERCISES),
		});
		renderExercises();
		await waitFor(() => {
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument();
			expect(screen.getByText("Archer Push-up")).toBeInTheDocument();
		});
	});

	it("shows empty state when no exercises", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderExercises();
		await waitFor(() =>
			expect(screen.getByText("No exercises yet.")).toBeInTheDocument(),
		);
	});

	it("shows error when fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({}),
		});
		renderExercises();
		await waitFor(() =>
			expect(screen.getByText("Failed to fetch exercises")).toBeInTheDocument(),
		);
	});

	it("shows empty state in detail panel when no exercise selected", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_EXERCISES),
		});
		renderExercises();
		await waitFor(() =>
			expect(
				screen.getByText("Select an exercise to view its details."),
			).toBeInTheDocument(),
		);
	});
});

describe("Exercises — active row", () => {
	it("navigates to /exercises/:id when a row is clicked", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_EXERCISES),
		});
		renderExercises();
		await waitFor(() =>
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByText("Diamond Push-up"));
		expect(window.location.href).toContain("/exercises/1");
	});
});

describe("Exercises — new exercise dialog", () => {
	it("opens new exercise dialog on + New click", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderExercises();
		await waitFor(() => expect(screen.getByText("+ New")).toBeInTheDocument());
		fireEvent.click(screen.getByText("+ New"));
		expect(screen.getByText("New Exercise")).toBeInTheDocument();
	});

	it("shows validation error when name is empty", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderExercises();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));
		expect(screen.getByText("Name is required.")).toBeInTheDocument();
	});

	it("creates a new exercise via POST", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 3, name: "Muscle-up" }),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve([]),
				});
			});

		renderExercises();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Muscle-up" },
		});
		fireEvent.input(screen.getByLabelText("Progression"), {
			target: { value: "Pull-up" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises",
				expect.objectContaining({
					method: "POST",
					body: JSON.stringify({
						name: "Muscle-up",
						progression: "Pull-up",
						description: null,
						is_public: false,
					}),
				}),
			),
		);
	});

	it("shows API error when create fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "POST") {
				return Promise.resolve({
					ok: false,
					json: () => Promise.resolve({ error: "Name must be unique" }),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve([]),
			});
		});

		renderExercises();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Diamond Push-up" },
		});
		fireEvent.submit(screen.getByLabelText("Name").closest("form"));

		await waitFor(() =>
			expect(screen.getByText("Name must be unique")).toBeInTheDocument(),
		);
	});

	it("cancels the new exercise dialog", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderExercises();
		await waitFor(() => screen.getByText("+ New"));
		fireEvent.click(screen.getByText("+ New"));
		expect(screen.getByText("New Exercise")).toBeInTheDocument();
		fireEvent.click(screen.getByTestId("mock-dialog-close"));
		expect(screen.queryByText("New Exercise")).not.toBeInTheDocument();
	});
});
