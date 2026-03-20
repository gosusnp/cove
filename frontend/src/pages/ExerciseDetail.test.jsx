// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { ExerciseDetail } from "./ExerciseDetail.jsx";

// ─── Mocks ───────────────────────────────────────────────────────────────────

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

vi.mock("../components/ui/Switch.jsx", () => ({
	Switch: ({ checkedSignal, id, onCheckedChange }) => (
		<input
			type="checkbox"
			id={id}
			checked={checkedSignal?.value ?? false}
			onChange={(e) => {
				if (checkedSignal) checkedSignal.value = e.target.checked;
				onCheckedChange?.(e.target.checked);
			}}
		/>
	),
}));

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const MOCK_USER = { email: "user@example.com", name: "Test User" };

const MOCK_EXERCISE = {
	id: 1,
	name: "Diamond Push-up",
	progression: "Push-up",
	description: "Triceps focus",
	is_public: false,
};

const renderDetail = (exerciseId = 1, props = {}) =>
	withProviders(<ExerciseDetail exerciseId={exerciseId} {...props} />, {
		user: MOCK_USER,
	});

function mockDefaultFetch(exercise = MOCK_EXERCISE) {
	vi.spyOn(global, "fetch").mockResolvedValue({
		ok: true,
		json: () => Promise.resolve(exercise),
	});
}

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

// ── Loading & errors ──────────────────────────────────────────────────────────

describe("ExerciseDetail — loading & errors", () => {
	it("shows loading state initially", () => {
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
			expect(screen.getByText("Failed to load exercise")).toBeInTheDocument(),
		);
	});

	it("renders exercise fields after successful load", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Diamond Push-up")).toBeInTheDocument(),
		);
		expect(screen.getByText("Push-up")).toBeInTheDocument();
		expect(screen.getByText("Triceps focus")).toBeInTheDocument();
	});
});

// ── Inline name edit ──────────────────────────────────────────────────────────

describe("ExerciseDetail — inline name edit", () => {
	it("clicking Edit exercise name shows input pre-filled with current name", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit exercise name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit exercise name" }));
		expect(screen.getByDisplayValue("Diamond Push-up")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
	});

	it("cancel restores the original name display", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit exercise name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit exercise name" }));
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(screen.getByText("Diamond Push-up")).toBeInTheDocument();
		expect(
			screen.queryByDisplayValue("Diamond Push-up"),
		).not.toBeInTheDocument();
	});

	it("shows validation error when name is cleared", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit exercise name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit exercise name" }));
		fireEvent.input(screen.getByDisplayValue("Diamond Push-up"), {
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
						json: () => Promise.resolve({}),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISE),
				});
			});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit exercise name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit exercise name" }));
		fireEvent.input(screen.getByDisplayValue("Diamond Push-up"), {
			target: { value: "One-arm Push-up" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({
						name: "One-arm Push-up",
						progression: "Push-up",
						description: "Triceps focus",
						is_public: false,
					}),
				}),
			),
		);
		await waitFor(() =>
			expect(screen.getByText("One-arm Push-up")).toBeInTheDocument(),
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
				json: () => Promise.resolve(MOCK_EXERCISE),
			});
		});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit exercise name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit exercise name" }));
		fireEvent.input(screen.getByDisplayValue("Diamond Push-up"), {
			target: { value: "Taken Name" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(screen.getByText("Name already taken")).toBeInTheDocument(),
		);
	});

	it("calls onExerciseUpdated after successful rename", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "PUT") {
				return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_EXERCISE),
			});
		});

		const onExerciseUpdated = vi.fn();
		withProviders(
			<ExerciseDetail exerciseId={1} onExerciseUpdated={onExerciseUpdated} />,
			{ user: MOCK_USER },
		);

		await waitFor(() =>
			screen.getByRole("button", { name: "Edit exercise name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit exercise name" }));
		fireEvent.input(screen.getByDisplayValue("Diamond Push-up"), {
			target: { value: "Renamed" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() => expect(onExerciseUpdated).toHaveBeenCalledOnce());
	});
});

// ── Inline progression edit ───────────────────────────────────────────────────

describe("ExerciseDetail — inline progression edit", () => {
	it("clicking Edit progression shows input pre-filled", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit progression" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit progression" }));
		expect(screen.getByDisplayValue("Push-up")).toBeInTheDocument();
	});

	it("saves updated progression via PUT", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "PUT") {
					return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISE),
				});
			});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit progression" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit progression" }));
		fireEvent.input(screen.getByDisplayValue("Push-up"), {
			target: { value: "Standard Push-up" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({
						name: "Diamond Push-up",
						progression: "Standard Push-up",
						description: "Triceps focus",
						is_public: false,
					}),
				}),
			),
		);
	});
});

// ── Description ───────────────────────────────────────────────────────────────

describe("ExerciseDetail — description", () => {
	it("shows placeholder when no description", async () => {
		mockDefaultFetch({ ...MOCK_EXERCISE, description: null });
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Add a description…")).toBeInTheDocument(),
		);
	});

	it("saves description via PUT", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "PUT") {
					return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve({ ...MOCK_EXERCISE, description: null }),
				});
			});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit exercise description" }),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Edit exercise description" }),
		);
		fireEvent.input(screen.getByRole("textbox"), {
			target: { value: "Great triceps exercise" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({
						name: "Diamond Push-up",
						progression: "Push-up",
						description: "Great triceps exercise",
						is_public: false,
					}),
				}),
			),
		);
	});
});

// ── Public toggle ─────────────────────────────────────────────────────────────

describe("ExerciseDetail — public toggle", () => {
	it("saves is_public=true via PUT when toggled on", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "PUT") {
					return Promise.resolve({ ok: true, json: () => Promise.resolve({}) });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISE),
				});
			});

		renderDetail();
		await waitFor(() =>
			expect(screen.getByRole("checkbox")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("checkbox"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({
						name: "Diamond Push-up",
						progression: "Push-up",
						description: "Triceps focus",
						is_public: true,
					}),
				}),
			),
		);
	});
});

// ── Delete ────────────────────────────────────────────────────────────────────

describe("ExerciseDetail — delete", () => {
	it("shows Delete exercise button", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Delete exercise" }),
			).toBeInTheDocument(),
		);
	});

	it("opens confirm dialog on Delete exercise click", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Delete exercise" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete exercise" }));
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
		expect(screen.getByText("Delete Exercise")).toBeInTheDocument();
	});

	it("calls DELETE /api/exercises/1 on confirm", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "DELETE") {
					return Promise.resolve({ ok: true });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISE),
				});
			});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Delete exercise" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete exercise" }));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/exercises/1",
				expect.objectContaining({ method: "DELETE" }),
			),
		);
	});

	it("shows error message when delete fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: false });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_EXERCISE),
			});
		});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Delete exercise" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete exercise" }));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(
				screen.getByText("Failed to delete exercise."),
			).toBeInTheDocument(),
		);
	});

	it("redirects to /login when DELETE returns 401", async () => {
		const assignSpy = vi.fn();
		vi.stubGlobal("location", { ...window.location, assign: assignSpy });
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ status: 401, ok: false });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_EXERCISE),
			});
		});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Delete exercise" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete exercise" }));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() => expect(assignSpy).toHaveBeenCalledWith("/login"));
	});

	it("shows error message when DELETE returns 403", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ status: 403, ok: false });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_EXERCISE),
			});
		});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Delete exercise" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete exercise" }));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(
				screen.getByText("Failed to delete exercise."),
			).toBeInTheDocument(),
		);
	});

	it("calls onExerciseDeleted after successful delete", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: true });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_EXERCISE),
			});
		});

		const onExerciseDeleted = vi.fn();
		withProviders(
			<ExerciseDetail exerciseId={1} onExerciseDeleted={onExerciseDeleted} />,
			{ user: MOCK_USER },
		);

		await waitFor(() =>
			screen.getByRole("button", { name: "Delete exercise" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Delete exercise" }));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() => expect(onExerciseDeleted).toHaveBeenCalledOnce());
	});
});
