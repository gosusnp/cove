// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor, within } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { ProgramDetail } from "./ProgramDetail.jsx";

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
	ConfirmDialog: ({ openSignal, title, description, onConfirm }) =>
		openSignal.value ? (
			<div data-testid="mock-confirm-dialog">
				<p>{title}</p>
				<p>{description}</p>
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

vi.mock("../components/ui/Accordion.jsx", () => ({
	Accordion: ({ children }) => <div>{children}</div>,
	AccordionItem: ({ children }) => <div>{children}</div>,
	AccordionTrigger: ({ children }) => <div>{children}</div>,
	AccordionContent: ({ children }) => <div>{children}</div>,
	AccordionDragHandle: () => null,
}));

vi.mock("../components/ui/Combobox.jsx", () => ({
	Combobox: ({ label, value, onChange, options, disabled }) => (
		<div>
			{label && <label htmlFor="mock-combobox">{label}</label>}
			<select
				id="mock-combobox"
				value={value ?? ""}
				disabled={disabled}
				onChange={(e) => onChange(Number(e.target.value) || e.target.value)}
				data-testid="mock-combobox"
			>
				<option value="">-- select --</option>
				{options.map((o) => (
					<option key={o.value} value={o.value}>
						{o.label}
					</option>
				))}
			</select>
		</div>
	),
}));

vi.mock("../components/ui/ToggleGroup.jsx", () => ({
	ToggleGroup: ({ label, value, onChange, options }) => (
		<div>
			{label && <span>{label}</span>}
			{options.map((o) => (
				<button
					key={o.value}
					type="button"
					data-selected={value === o.value}
					onClick={() => onChange(value === o.value ? null : o.value)}
				>
					{o.label}
				</button>
			))}
		</div>
	),
}));

// ─── Fixtures ─────────────────────────────────────────────────────────────────

const MOCK_USER = { email: "user@example.com", name: "Test User" };

const MOCK_PROGRAM = {
	id: 1,
	name: "Strength A/B",
	sets: [
		{
			id: 10,
			name: "Push",
			rounds: 4,
			rest_s: 120,
			sort_order: 0,
			exercises: [
				{
					id: 100,
					exercise_id: 5,
					name: "Bench Press",
					laterality: "bilateral",
					reps: 8,
					duration_s: null,
					weight_kg: 60,
					sort_order: 0,
				},
			],
		},
		{
			id: 11,
			name: "Pull",
			rounds: 3,
			rest_s: 90,
			sort_order: 1,
			exercises: [],
		},
	],
};

const MOCK_EXERCISES = [
	{ id: 5, name: "Bench Press" },
	{ id: 6, name: "Squat" },
];

// ─── Helpers ──────────────────────────────────────────────────────────────────

const renderDetail = (programId = 1) =>
	withProviders(<ProgramDetail programId={programId} />, { user: MOCK_USER });

function mockDefaultFetch() {
	vi.spyOn(global, "fetch").mockImplementation((url) => {
		if (String(url).includes("/api/exercises")) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_EXERCISES),
			});
		}
		return Promise.resolve({
			ok: true,
			json: () => Promise.resolve(MOCK_PROGRAM),
		});
	});
}

// ─── Tests ────────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

// ── Loading & errors ──────────────────────────────────────────────────────────

describe("ProgramDetail — loading & errors", () => {
	it("shows loading state initially", () => {
		vi.spyOn(global, "fetch").mockReturnValue(new Promise(() => {}));
		renderDetail();
		expect(screen.getByText("Loading…")).toBeInTheDocument();
	});

	it("shows error when program fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: false,
			json: () => Promise.resolve({}),
		});
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Failed to load program")).toBeInTheDocument(),
		);
	});

	it("renders program name and sets after successful load", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		expect(screen.getByText("Push")).toBeInTheDocument();
		expect(screen.getByText("Pull")).toBeInTheDocument();
	});
});

// ── Set display ───────────────────────────────────────────────────────────────

describe("ProgramDetail — set display", () => {
	it("shows set name, rounds and rest for each set", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => expect(screen.getByText("Push")).toBeInTheDocument());
		expect(screen.getByText("4× · 120s rest")).toBeInTheDocument();
		expect(screen.getByText("3× · 90s rest")).toBeInTheDocument();
	});

	it('shows "Unnamed Set" when set name is null or empty', async () => {
		const prog = {
			...MOCK_PROGRAM,
			sets: [
				{ ...MOCK_PROGRAM.sets[0], name: null, exercises: [] },
				{ ...MOCK_PROGRAM.sets[1], name: "", exercises: [] },
			],
		};
		vi.spyOn(global, "fetch").mockImplementation((url) => {
			if (String(url).includes("/api/exercises")) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISES),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(prog),
			});
		});
		renderDetail();
		await waitFor(() =>
			expect(screen.getAllByText("Unnamed Set")).toHaveLength(2),
		);
	});

	it('shows "No sets yet." when sets array is empty', async () => {
		const prog = { ...MOCK_PROGRAM, sets: [] };
		vi.spyOn(global, "fetch").mockImplementation((url) => {
			if (String(url).includes("/api/exercises")) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISES),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(prog),
			});
		});
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByText(
					"No sets yet. Add a set to start building this program.",
				),
			).toBeInTheDocument(),
		);
	});

	it("shows exercise name and reps inside set", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Bench Press")).toBeInTheDocument(),
		);
		expect(screen.getByText("8 reps")).toBeInTheDocument();
	});

	it('shows "+60kg" for weight_kg=60', async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => expect(screen.getByText("+60kg")).toBeInTheDocument());
	});

	it('shows "bodyweight" for weight_kg=0', async () => {
		const prog = {
			...MOCK_PROGRAM,
			sets: [
				{
					...MOCK_PROGRAM.sets[0],
					exercises: [{ ...MOCK_PROGRAM.sets[0].exercises[0], weight_kg: 0 }],
				},
			],
		};
		vi.spyOn(global, "fetch").mockImplementation((url) => {
			if (String(url).includes("/api/exercises")) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISES),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(prog),
			});
		});
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("bodyweight")).toBeInTheDocument(),
		);
	});

	it("shows bodyweight for weight_kg=null", async () => {
		const prog = {
			...MOCK_PROGRAM,
			sets: [
				{
					...MOCK_PROGRAM.sets[0],
					exercises: [
						{ ...MOCK_PROGRAM.sets[0].exercises[0], weight_kg: null },
					],
				},
			],
		};
		vi.spyOn(global, "fetch").mockImplementation((url) => {
			if (String(url).includes("/api/exercises")) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISES),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(prog),
			});
		});
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("bodyweight")).toBeInTheDocument(),
		);
	});
});

// ── Inline edit ───────────────────────────────────────────────────────────────

describe("ProgramDetail — inline name edit", () => {
	it("clicking name shows input pre-filled with current name", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Strength A/B")).toBeInTheDocument(),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit program name" }));
		expect(screen.getByDisplayValue("Strength A/B")).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Save" })).toBeInTheDocument();
		expect(screen.getByRole("button", { name: "Cancel" })).toBeInTheDocument();
	});

	it("cancel restores the original name display", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit program name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit program name" }));
		fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
		expect(screen.getByText("Strength A/B")).toBeInTheDocument();
		expect(screen.queryByDisplayValue("Strength A/B")).not.toBeInTheDocument();
	});

	it("shows validation error when name is empty", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit program name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit program name" }));
		fireEvent.input(screen.getByDisplayValue("Strength A/B"), {
			target: { value: "" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));
		expect(screen.getByText("Name is required.")).toBeInTheDocument();
	});

	it("saves new name via PUT and updates display", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (opts?.method === "PUT" && String(url) === "/api/programs/1") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 1, name: "Renamed Program" }),
					});
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit program name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit program name" }));
		fireEvent.input(screen.getByDisplayValue("Strength A/B"), {
			target: { value: "Renamed Program" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({ name: "Renamed Program" }),
				}),
			),
		);
		await waitFor(() =>
			expect(screen.getByText("Renamed Program")).toBeInTheDocument(),
		);
	});

	it("shows API error on save failure", async () => {
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (opts?.method === "PUT") {
				return Promise.resolve({
					ok: false,
					json: () => Promise.resolve({ error: "Name already taken" }),
				});
			}
			if (String(url).includes("/api/exercises")) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISES),
				});
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_PROGRAM),
			});
		});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit program name" }),
		);
		fireEvent.click(screen.getByRole("button", { name: "Edit program name" }));
		fireEvent.input(screen.getByDisplayValue("Strength A/B"), {
			target: { value: "Taken Name" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(screen.getByText("Name already taken")).toBeInTheDocument(),
		);
	});
});

describe("ProgramDetail — inline description edit", () => {
	it("shows placeholder when no description and clicking opens textarea", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("button", { name: "Edit program description" }),
			).toBeInTheDocument(),
		);
		expect(screen.getByText("Add a description…")).toBeInTheDocument();
		fireEvent.click(
			screen.getByRole("button", { name: "Edit program description" }),
		);
		expect(screen.getByRole("textbox")).toBeInTheDocument();
	});

	it("saves description via PUT and updates display", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (opts?.method === "PUT" && String(url) === "/api/programs/1") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 1, name: "Strength A/B" }),
					});
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() =>
			screen.getByRole("button", { name: "Edit program description" }),
		);
		fireEvent.click(
			screen.getByRole("button", { name: "Edit program description" }),
		);
		fireEvent.input(screen.getByRole("textbox"), {
			target: { value: "A great strength program" },
		});
		fireEvent.click(screen.getByRole("button", { name: "Save" }));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1",
				expect.objectContaining({
					method: "PUT",
					body: JSON.stringify({
						name: "Strength A/B",
						description: "A great strength program",
					}),
				}),
			),
		);
		await waitFor(() =>
			expect(screen.getByText("A great strength program")).toBeInTheDocument(),
		);
	});
});

// ── Add Set dialog ────────────────────────────────────────────────────────────

describe("ProgramDetail — add set dialog", () => {
	it('opens add set dialog with "Add Set" title and default values', async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => screen.getByText("+ Add Set"));
		fireEvent.click(screen.getByText("+ Add Set"));
		expect(screen.getByText("Add Set")).toBeInTheDocument();
		expect(screen.getByDisplayValue("3")).toBeInTheDocument();
	});

	it("shows validation error for invalid rounds", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => screen.getByText("+ Add Set"));
		fireEvent.click(screen.getByText("+ Add Set"));
		const form = screen.getByText("Add Set").closest("form");
		const roundsInput = within(form).getByDisplayValue("3");
		fireEvent.input(roundsInput, { target: { value: "0" } });
		fireEvent.submit(form);
		expect(screen.getByText("Rounds must be at least 1.")).toBeInTheDocument();
	});

	it("creates set via POST /api/programs/1/sets", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({
								id: 20,
								name: null,
								rounds: 3,
								rest_s: 90,
								sort_order: 2,
								exercises: [],
							}),
					});
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() => screen.getByText("+ Add Set"));
		fireEvent.click(screen.getByText("+ Add Set"));
		const form = screen.getByText("Add Set").closest("form");
		fireEvent.submit(form);

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1/sets",
				expect.objectContaining({
					method: "POST",
				}),
			),
		);
	});
});

// ── Edit Set dialog ───────────────────────────────────────────────────────────

describe("ProgramDetail — edit set dialog", () => {
	it("opens edit set dialog pre-filled with set data", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => screen.getByText("Push"));
		fireEvent.click(screen.getAllByRole("button", { name: "Edit set" })[0]);
		expect(screen.getByText("Edit Set")).toBeInTheDocument();
		expect(screen.getByDisplayValue("Push")).toBeInTheDocument();
		expect(screen.getByDisplayValue("4")).toBeInTheDocument();
		expect(screen.getByDisplayValue("120")).toBeInTheDocument();
	});

	it("saves edited set via PUT /api/programs/1/sets/10", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (
					opts?.method === "PUT" &&
					String(url) === "/api/programs/1/sets/10"
				) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 10 }),
					});
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() => screen.getByText("Push"));
		fireEvent.click(screen.getAllByRole("button", { name: "Edit set" })[0]);
		const form = screen.getByText("Edit Set").closest("form");
		fireEvent.submit(form);

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1/sets/10",
				expect.objectContaining({ method: "PUT" }),
			),
		);
	});
});

// ── Delete Set ────────────────────────────────────────────────────────────────

describe("ProgramDetail — delete set", () => {
	it("opens confirm dialog on Delete click", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => screen.getByText("Push"));
		fireEvent.click(screen.getAllByRole("button", { name: "Delete set" })[0]);
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
	});

	it("deletes set via DELETE /api/programs/1/sets/10", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (opts?.method === "DELETE") {
					return Promise.resolve({ ok: true });
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() => screen.getByText("Push"));
		fireEvent.click(screen.getAllByRole("button", { name: "Delete set" })[0]);
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1/sets/10",
				expect.objectContaining({ method: "DELETE" }),
			),
		);
	});
});

// ── Add Exercise dialog ───────────────────────────────────────────────────────

describe("ProgramDetail — add exercise dialog", () => {
	it('opens add exercise dialog with "Add Exercise" title; combobox not disabled', async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getAllByRole("button", { name: "Add exercise" }).length,
			).toBeGreaterThan(0),
		);
		fireEvent.click(screen.getAllByRole("button", { name: "Add exercise" })[0]);
		expect(
			screen.getByRole("heading", { name: "Add Exercise" }),
		).toBeInTheDocument();
		const combobox = screen.getByTestId("mock-combobox");
		expect(combobox).not.toBeDisabled();
	});

	it("shows validation error when no exercise selected", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getAllByRole("button", { name: "Add exercise" }).length,
			).toBeGreaterThan(0),
		);
		fireEvent.click(screen.getAllByRole("button", { name: "Add exercise" })[0]);
		const form = screen
			.getByRole("heading", { name: "Add Exercise" })
			.closest("form");
		fireEvent.submit(form);
		expect(screen.getByText("Exercise is required.")).toBeInTheDocument();
	});

	it("creates exercise via POST /api/programs/1/sets/10/exercises", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 101 }),
					});
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() =>
			expect(
				screen.getAllByRole("button", { name: "Add exercise" }).length,
			).toBeGreaterThan(0),
		);
		fireEvent.click(screen.getAllByRole("button", { name: "Add exercise" })[0]);

		await waitFor(() =>
			expect(
				screen.getByTestId("mock-combobox").options.length,
			).toBeGreaterThan(1),
		);

		const combobox = screen.getByTestId("mock-combobox");
		fireEvent.change(combobox, { target: { value: "5" } });
		const form = screen
			.getByRole("heading", { name: "Add Exercise" })
			.closest("form");
		fireEvent.submit(form);

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1/sets/10/exercises",
				expect.objectContaining({ method: "POST" }),
			),
		);
	});
});

// ── Edit Exercise dialog ──────────────────────────────────────────────────────

describe("ProgramDetail — edit exercise dialog", () => {
	it("opens edit exercise dialog pre-filled; combobox disabled", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => screen.getByText("Bench Press"));
		fireEvent.click(
			screen.getAllByRole("button", { name: "Edit exercise" })[0],
		);
		expect(screen.getByText("Edit Exercise")).toBeInTheDocument();
		const combobox = screen.getByTestId("mock-combobox");
		expect(combobox).toBeDisabled();
	});

	it("saves edited exercise via PUT /api/programs/1/sets/10/exercises/100", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (
					opts?.method === "PUT" &&
					String(url) === "/api/programs/1/sets/10/exercises/100"
				) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 100 }),
					});
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() => screen.getByText("Bench Press"));
		fireEvent.click(
			screen.getAllByRole("button", { name: "Edit exercise" })[0],
		);
		const form = screen.getByText("Edit Exercise").closest("form");
		fireEvent.submit(form);

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1/sets/10/exercises/100",
				expect.objectContaining({ method: "PUT" }),
			),
		);
	});
});

// ── Remove Exercise ───────────────────────────────────────────────────────────

describe("ProgramDetail — remove exercise", () => {
	it("opens confirm dialog showing exercise name", async () => {
		mockDefaultFetch();
		renderDetail();
		await waitFor(() => screen.getByText("Bench Press"));
		fireEvent.click(
			screen.getAllByRole("button", { name: "Remove exercise" })[0],
		);
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
		expect(
			screen.getByText(/Remove "Bench Press" from this set\?/),
		).toBeInTheDocument();
	});

	it("removes exercise via DELETE /api/programs/1/sets/10/exercises/100", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((url, opts) => {
				if (opts?.method === "DELETE") {
					return Promise.resolve({ ok: true });
				}
				if (String(url).includes("/api/exercises")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve(MOCK_EXERCISES),
					});
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_PROGRAM),
				});
			});

		renderDetail();
		await waitFor(() => screen.getByText("Bench Press"));
		fireEvent.click(
			screen.getAllByRole("button", { name: "Remove exercise" })[0],
		);
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/programs/1/sets/10/exercises/100",
				expect.objectContaining({ method: "DELETE" }),
			),
		);
	});
});

// ── Stale data after CRUD ─────────────────────────────────────────────────────

describe("ProgramDetail — refresh after CRUD", () => {
	it("shows updated set list after a set is deleted", async () => {
		const programAfterDelete = {
			...MOCK_PROGRAM,
			sets: [MOCK_PROGRAM.sets[1]], // only "Pull" remains
		};

		let fetchCount = 0;
		vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: true });
			}
			if (String(url).includes("/api/exercises")) {
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_EXERCISES),
				});
			}
			// First program fetch returns full data; subsequent returns updated data.
			fetchCount += 1;
			return Promise.resolve({
				ok: true,
				json: () =>
					Promise.resolve(fetchCount === 1 ? MOCK_PROGRAM : programAfterDelete),
			});
		});

		renderDetail();
		await waitFor(() => screen.getByText("Push"));

		// Delete the "Push" set
		fireEvent.click(screen.getAllByRole("button", { name: "Delete set" })[0]);
		fireEvent.click(screen.getByText("Confirm"));

		// After refresh, "Push" should be gone and only "Pull" remains
		await waitFor(() => {
			expect(screen.queryByText("Push")).not.toBeInTheDocument();
			expect(screen.getByText("Pull")).toBeInTheDocument();
		});
	});
});
