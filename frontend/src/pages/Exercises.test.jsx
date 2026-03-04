// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { screen, fireEvent, waitFor, within } from "@testing-library/preact";
import { describe, it, expect, vi, afterEach, beforeEach } from "vitest";
import { Exercises } from "./Exercises.jsx";
import { withProviders } from "../test-utils.jsx";

let dialogSignal = null;
vi.mock("../components/ui/Dialog.jsx", () => ({
	Dialog: ({ children, openSignal }) => {
		dialogSignal = openSignal;
		return openSignal.value ? (
			<div data-testid="mock-dialog">{children}</div>
		) : null;
	},
	DialogContent: ({ children }) => (
		<div data-testid="mock-dialog-content">{children}</div>
	),
	DialogTitle: ({ children }) => <h2>{children}</h2>,
	DialogDescription: ({ children }) => <p>{children}</p>,
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

// Mock Switch since it uses Radix and might need jsdom mocks if not already present
vi.mock("../components/ui/Switch.jsx", () => ({
	Switch: ({ checkedSignal, id }) => (
		<input
			type="checkbox"
			id={id}
			checked={checkedSignal.value}
			onChange={(e) => {
				checkedSignal.value = e.target.checked;
			}}
		/>
	),
}));

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

const renderExercises = (opts = {}) =>
	withProviders(<Exercises />, {
		path: "/exercises",
		user: MOCK_USER,
		...opts,
	});

describe("Exercises", () => {
	beforeEach(() => {
		vi.stubGlobal(
			"confirm",
			vi.fn(() => true),
		);
		vi.stubGlobal("alert", vi.fn());
	});

	afterEach(() => {
		vi.restoreAllMocks();
		vi.unstubAllGlobals();
	});

	it("renders the page heading", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_EXERCISES),
		});
		renderExercises();
		expect(
			screen.getByRole("heading", { name: "Exercises" }),
		).toBeInTheDocument();
		await waitFor(() =>
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument(),
		);
	});

	it("renders exercises returned by the API", async () => {
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
			expect(screen.getByText("No exercises yet")).toBeInTheDocument(),
		);
	});

	it("opens create dialog on button click", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderExercises();
		await waitFor(() => expect(screen.getByText("Create")).not.toBeDisabled());
		fireEvent.click(screen.getByText("Create"));
		expect(screen.getByText("New Exercise")).toBeInTheDocument();
	});

	it("creates a new exercise", async () => {
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
		await waitFor(() => expect(screen.getByText("Create")).toBeInTheDocument());
		fireEvent.click(screen.getByText("Create"));

		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Muscle-up" },
		});
		fireEvent.input(screen.getByLabelText("Progression"), {
			target: { value: "Pull-up" },
		});

		fireEvent.submit(screen.getByText("Save Exercise").closest("form"));

		await waitFor(() => {
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
			);
		});
	});

	it("shows a validation error when name is missing", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderExercises();
		await waitFor(() => expect(screen.getByText("Create")).toBeInTheDocument());
		fireEvent.click(screen.getByText("Create"));

		fireEvent.submit(screen.getByText("Save Exercise").closest("form"));

		expect(screen.getByText("Name is required.")).toBeInTheDocument();
	});

	it("cancels the create dialog", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve([]),
		});
		renderExercises();
		await waitFor(() => expect(screen.getByText("Create")).toBeInTheDocument());
		fireEvent.click(screen.getByText("Create"));

		expect(screen.getByText("New Exercise")).toBeInTheDocument();

		fireEvent.click(screen.getByTestId("mock-dialog-close"));

		expect(screen.queryByText("New Exercise")).not.toBeInTheDocument();
	});

	it("shows an error when API save fails", async () => {
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
		await waitFor(() => expect(screen.getByText("Create")).toBeInTheDocument());
		fireEvent.click(screen.getByText("Create"));

		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "Diamond Push-up" },
		});
		fireEvent.submit(screen.getByText("Save Exercise").closest("form"));

		await waitFor(() => {
			expect(screen.getByText("Name must be unique")).toBeInTheDocument();
		});
	});

	it("shows an alert when delete fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: false });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_EXERCISES),
			});
		});

		renderExercises();
		await waitFor(() =>
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getAllByText("Delete")[0]);

		await waitFor(() => {
			expect(window.alert).toHaveBeenCalledWith("Failed to delete exercise");
		});
	});

	it("opens update dialog with pre-filled data", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_EXERCISES),
		});
		renderExercises();
		await waitFor(() =>
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument(),
		);

		const exRow = screen.getByText("Diamond Push-up").closest("section");
		fireEvent.click(within(exRow).getAllByText("Update")[0]);

		expect(screen.getByText("Update Exercise")).toBeInTheDocument();
		expect(screen.getByDisplayValue("Diamond Push-up")).toBeInTheDocument();
		expect(screen.getByDisplayValue("Push-up")).toBeInTheDocument();
	});

	it("updates an exercise", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "PUT") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({
								...MOCK_EXERCISES[0],
								name: "One-arm Push-up",
							}),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISES),
				});
			});

		renderExercises();
		await waitFor(() =>
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getAllByText("Update")[0]);
		fireEvent.input(screen.getByLabelText("Name"), {
			target: { value: "One-arm Push-up" },
		});

		fireEvent.submit(screen.getByText("Save Exercise").closest("form"));

		await waitFor(() => {
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({
						name: "One-arm Push-up",
						progression: "Push-up",
						description: "Triceps and chest",
						is_public: true,
					}),
				}),
			);
		});
	});

	it("deletes an exercise after confirmation", async () => {
		const fetchSpy = vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_EXERCISES),
		});

		renderExercises();
		await waitFor(() =>
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument(),
		);

		fireEvent.click(screen.getAllByText("Delete")[0]);

		expect(window.confirm).toHaveBeenCalled();
		await waitFor(() => {
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises/1",
				expect.objectContaining({
					method: "DELETE",
				}),
			);
		});
	});
});
