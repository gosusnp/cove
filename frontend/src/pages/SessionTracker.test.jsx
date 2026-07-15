// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { SessionTracker } from "./SessionTracker.jsx";

// ─── Mocks ──────────────────────────────────────────────────────────────────

// Capture dialog callbacks so tests can trigger them directly.
let capturedOnSave = null;
let capturedOnCancel = null;

vi.mock("./SessionSummaryDialog.jsx", () => ({
	SessionSummaryDialog: ({ openSignal, onCancel, onSave, saveError }) => {
		capturedOnSave = onSave;
		capturedOnCancel = onCancel;
		if (!openSignal.value) return null;
		return (
			<div data-testid="mock-summary-dialog">
				{saveError && <p data-testid="dialog-error">{saveError}</p>}
				<button type="button" onClick={onCancel}>
					Cancel
				</button>
				<button type="button" onClick={onSave}>
					Save Session
				</button>
			</div>
		);
	},
}));

vi.mock("../components/ui/Combobox.jsx", () => ({
	// Render a clickable button so tests can trigger program selection.
	Combobox: ({ label, onChange }) => (
		<button type="button" onClick={() => onChange("1")}>
			{label}
		</button>
	),
}));

vi.mock("../components/shared/ActivityPicker.jsx", () => ({
	ActivityPicker: ({ value, onChange }) => (
		<>
			<button type="button" onClick={() => onChange("Climbing")}>
				Activity: {value || "none"}
			</button>
			<button type="button" onClick={() => onChange("bouldering")}>
				Set Bouldering
			</button>
		</>
	),
}));

// ─── Helpers ─────────────────────────────────────────────────────────────────

const MOCK_USER = { email: "user@example.com", name: "Test" };

function renderTracker() {
	return withProviders(<SessionTracker />, {
		user: MOCK_USER,
		path: "/workout",
	});
}

function mockFetch({ sessionId = 1, patchOk = true } = {}) {
	vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		if (url.includes("/api/programs")) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve([]),
			});
		}
		if (opts?.method === "POST") {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve({ id: sessionId }),
			});
		}
		if (opts?.method === "PATCH") {
			return Promise.resolve({ ok: patchOk });
		}
		return Promise.reject(new Error(`Unexpected fetch: ${url}`));
	});
}

function mockFetchWithProgram(exercises) {
	return vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		if (url.includes("/api/programs/")) {
			return Promise.resolve({
				ok: true,
				json: () =>
					Promise.resolve({
						id: 1,
						name: "Test Program",
						sets: [{ id: 1, name: "Set A", rounds: 1, exercises }],
					}),
			});
		}
		if (url.includes("/api/programs")) {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve([{ id: 1, name: "Test Program" }]),
			});
		}
		if (opts?.method === "POST") {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve({ id: 1 }),
			});
		}
		if (opts?.method === "PATCH") {
			return Promise.resolve({ ok: true });
		}
		return Promise.reject(new Error(`Unexpected fetch: ${url}`));
	});
}

async function startSession() {
	fireEvent.click(screen.getByRole("button", { name: "Start" }));
	await waitFor(() =>
		expect(
			screen.getByRole("button", { name: "End Session" }),
		).toBeInTheDocument(),
	);
}

async function selectProgramAndStart() {
	fireEvent.click(screen.getByRole("button", { name: "Program (optional)" }));
	fireEvent.click(screen.getByRole("button", { name: "Start" }));
	await waitFor(() =>
		expect(
			screen.getByRole("button", { name: "End Session" }),
		).toBeInTheDocument(),
	);
}

async function selectActivityAndStart() {
	fireEvent.click(screen.getByRole("button", { name: "Activity: none" }));
	fireEvent.click(screen.getByRole("button", { name: "Start" }));
	await waitFor(() =>
		expect(
			screen.getByRole("button", { name: "End Session" }),
		).toBeInTheDocument(),
	);
}

// ─── Tests ───────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
	localStorage.clear();
	capturedOnSave = null;
	capturedOnCancel = null;
});

describe("SessionTracker", () => {
	it("shows End Session button after session is started", async () => {
		mockFetch();
		renderTracker();
		await startSession();
		expect(
			screen.getByRole("button", { name: "End Session" }),
		).toBeInTheDocument();
	});

	it("opens summary dialog when End Session is clicked", async () => {
		mockFetch();
		renderTracker();
		await startSession();
		fireEvent.click(screen.getByRole("button", { name: "End Session" }));
		expect(screen.getByTestId("mock-summary-dialog")).toBeInTheDocument();
	});

	describe("timer", () => {
		// Only fake setInterval and Date — leave setTimeout real so waitFor keeps working.
		beforeEach(() => {
			vi.useFakeTimers({ toFake: ["setInterval", "clearInterval", "Date"] });
		});

		afterEach(() => {
			vi.useRealTimers();
		});

		it("counts up based on elapsed wall-clock time since start", async () => {
			mockFetch();
			renderTracker();
			await startSession();

			vi.advanceTimersByTime(5000);

			await waitFor(() =>
				expect(screen.getByText("00:00:05")).toBeInTheDocument(),
			);
		});

		it("pause freezes elapsed and resume accumulates correctly", async () => {
			mockFetch();
			renderTracker();
			await startSession();

			vi.advanceTimersByTime(10000);
			await waitFor(() =>
				expect(screen.getByText("00:00:10")).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Pause" }));
			// Time passing while paused must not be counted.
			vi.advanceTimersByTime(5000);
			await waitFor(() =>
				expect(screen.getByText("00:00:10")).toBeInTheDocument(),
			);

			// Resume and wait for the effect to set up the new interval before advancing.
			fireEvent.click(screen.getByRole("button", { name: "Resume" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Pause" }),
				).toBeInTheDocument(),
			);

			vi.advanceTimersByTime(3000);

			await waitFor(() =>
				expect(screen.getByText("00:00:13")).toBeInTheDocument(),
			);
		});

		it("stop while paused uses accumulated elapsed without corrupting it", async () => {
			mockFetch();
			renderTracker();
			await startSession();

			vi.advanceTimersByTime(10000);
			await waitFor(() =>
				expect(screen.getByText("00:00:10")).toBeInTheDocument(),
			);

			fireEvent.click(screen.getByRole("button", { name: "Pause" }));
			await waitFor(() =>
				expect(screen.getByText("00:00:10")).toBeInTheDocument(),
			);

			// End Session while paused — segmentStartRef is null at this point.
			fireEvent.click(screen.getByRole("button", { name: "End Session" }));

			expect(screen.getByTestId("mock-summary-dialog")).toBeInTheDocument();
			// summaryElapsed must equal the 10s accumulated before pause, not a corrupted value.
			expect(screen.getByText("00:00:10")).toBeInTheDocument();
		});

		it("snaps elapsed immediately on visibilitychange without waiting for the next tick", async () => {
			mockFetch();
			renderTracker();
			await startSession();

			// Advance only the clock (no interval ticks) to simulate a background gap.
			vi.setSystemTime(Date.now() + 7000);

			Object.defineProperty(document, "hidden", {
				value: false,
				configurable: true,
			});
			fireEvent(document, new Event("visibilitychange"));

			await waitFor(() =>
				expect(screen.getByText("00:00:07")).toBeInTheDocument(),
			);
		});
	});

	describe("exercise label formatting", () => {
		it("prefixes reps with x before the exercise name", async () => {
			mockFetchWithProgram([{ name: "Push-up", reps: 10 }]);
			renderTracker();
			await selectProgramAndStart();
			await waitFor(() =>
				expect(screen.getByText("10x Push-up")).toBeInTheDocument(),
			);
		});

		it("prefixes duration in seconds before the exercise name", async () => {
			mockFetchWithProgram([{ name: "Plank", duration_s: 30 }]);
			renderTracker();
			await selectProgramAndStart();
			await waitFor(() =>
				expect(screen.getByText("30s Plank")).toBeInTheDocument(),
			);
		});

		it("shows reps and duration combined before the name", async () => {
			mockFetchWithProgram([
				{ name: "Crimp block hold", reps: 3, duration_s: 10 },
			]);
			renderTracker();
			await selectProgramAndStart();
			await waitFor(() =>
				expect(screen.getByText("3x 10s Crimp block hold")).toBeInTheDocument(),
			);
		});

		it("shows just the name when no reps or duration are set", async () => {
			mockFetchWithProgram([{ name: "Squat" }]);
			renderTracker();
			await selectProgramAndStart();
			await waitFor(() =>
				expect(screen.getByText("Squat")).toBeInTheDocument(),
			);
		});

		it("shows weight in the subtitle", async () => {
			mockFetchWithProgram([
				{ name: "Back Squat", reps: 5, weight: 80, weight_unit: "kg" },
			]);
			renderTracker();
			await selectProgramAndStart();
			await waitFor(() =>
				expect(screen.getByText("80 kg")).toBeInTheDocument(),
			);
		});

		it("shows laterality in the subtitle", async () => {
			mockFetchWithProgram([
				{ name: "Lunge", reps: 12, laterality: "alternating" },
			]);
			renderTracker();
			await selectProgramAndStart();
			await waitFor(() =>
				expect(screen.getByText("alternating")).toBeInTheDocument(),
			);
		});

		it("shows weight and laterality together in the subtitle", async () => {
			mockFetchWithProgram([
				{
					name: "Single-leg press",
					reps: 8,
					weight: 40,
					weight_unit: "kg",
					laterality: "bilateral",
				},
			]);
			renderTracker();
			await selectProgramAndStart();
			await waitFor(() =>
				expect(screen.getByText("40 kg · bilateral")).toBeInTheDocument(),
			);
		});

		it("converts weight to the user's preferred unit", async () => {
			mockFetchWithProgram([
				{ name: "Back Squat", reps: 5, weight: 100, weight_unit: "kg" },
			]);
			withProviders(<SessionTracker />, {
				user: { email: "user@example.com", fitness_unit_system: "imperial" },
				path: "/workout",
			});
			await selectProgramAndStart();
			await waitFor(() =>
				// 100 kg → 220.46 lb → quantized to nearest 0.5 lb = 220.5 lb
				expect(screen.getByText("220.5 lb")).toBeInTheDocument(),
			);
		});
	});

	describe("Add Program", () => {
		it("shows + Add Program button when session is running", async () => {
			mockFetch();
			renderTracker();
			await startSession();
			expect(
				screen.getByRole("button", { name: "+ Add Program" }),
			).toBeInTheDocument();
		});

		it("hides + Add Program button before session starts", () => {
			mockFetch();
			renderTracker();
			expect(
				screen.queryByRole("button", { name: "+ Add Program" }),
			).not.toBeInTheDocument();
		});

		it("appends exercises from the selected program", async () => {
			mockFetchWithProgram([{ name: "Ring row", reps: 8 }]);
			renderTracker();
			await startSession();

			fireEvent.click(screen.getByRole("button", { name: "+ Add Program" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Program" }),
				).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Program" }));
			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				expect(screen.getByText("8x Ring row")).toBeInTheDocument(),
			);
		});

		it("shows program name labels when multiple programs are tracked", async () => {
			mockFetchWithProgram([{ name: "Pull-up", reps: 5 }]);
			renderTracker();
			await selectProgramAndStart();

			fireEvent.click(screen.getByRole("button", { name: "+ Add Program" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Program" }),
				).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Program" }));
			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				// Two programs → program name label appears above each CheckList
				expect(
					screen.getAllByText("Test Program").length,
				).toBeGreaterThanOrEqual(2),
			);
		});

		it("PATCHes the session with the accumulated program_name", async () => {
			const fetchSpy = mockFetchWithProgram([{ name: "Pull-up", reps: 5 }]);
			renderTracker();
			await startSession();

			fireEvent.click(screen.getByRole("button", { name: "+ Add Program" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Program" }),
				).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Program" }));
			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				expect(screen.getByText("5x Pull-up")).toBeInTheDocument(),
			);

			const patchCalls = fetchSpy.mock.calls.filter(
				([url, opts]) =>
					opts?.method === "PATCH" && url.includes("/api/sessions/"),
			);
			expect(patchCalls).toHaveLength(1);
			const body = JSON.parse(patchCalls[0][1].body);
			expect(body.program_name).toBe("Test Program");
		});

		it("accumulates program names when adding a second program", async () => {
			const fetchSpy = mockFetchWithProgram([{ name: "Pull-up", reps: 5 }]);
			renderTracker();
			await selectProgramAndStart();

			fireEvent.click(screen.getByRole("button", { name: "+ Add Program" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Program" }),
				).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Program" }));
			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				expect(screen.getAllByText("5x Pull-up").length).toBeGreaterThanOrEqual(
					2,
				),
			);

			const patchCalls = fetchSpy.mock.calls.filter(
				([url, opts]) =>
					opts?.method === "PATCH" && url.includes("/api/sessions/"),
			);
			expect(patchCalls).toHaveLength(1);
			const body = JSON.parse(patchCalls[0][1].body);
			expect(body.program_name).toBe("Test Program, Test Program");
		});

		it("shows error message when PATCH fails", async () => {
			vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
				if (url.includes("/api/programs/")) {
					return Promise.resolve({
						ok: true,
						json: () =>
							Promise.resolve({
								id: 1,
								name: "Test Program",
								sets: [],
							}),
					});
				}
				if (url.includes("/api/programs")) {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve([{ id: 1, name: "Test Program" }]),
					});
				}
				if (opts?.method === "POST") {
					return Promise.resolve({
						ok: true,
						json: () => Promise.resolve({ id: 1 }),
					});
				}
				if (opts?.method === "PATCH") {
					return Promise.resolve({ ok: false });
				}
				return Promise.reject(new Error(`Unexpected fetch: ${url}`));
			});
			renderTracker();
			await startSession();

			fireEvent.click(screen.getByRole("button", { name: "+ Add Program" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Program" }),
				).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "Program" }));
			fireEvent.click(screen.getByRole("button", { name: "Add" }));

			await waitFor(() =>
				expect(
					screen.getByText("Failed to update session"),
				).toBeInTheDocument(),
			);
		});
	});

	describe("free-form session", () => {
		it("shows a checklist with the activity name when no program is selected", async () => {
			mockFetch();
			renderTracker();
			await selectActivityAndStart();

			await waitFor(() =>
				expect(screen.getAllByText("Climbing").length).toBeGreaterThanOrEqual(
					1,
				),
			);
		});
	});

	describe("bouldering tracker", () => {
		async function startBouldSession() {
			fireEvent.click(screen.getByRole("button", { name: "Set Bouldering" }));
			fireEvent.click(screen.getByRole("button", { name: "Start" }));
			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "End Session" }),
				).toBeInTheDocument(),
			);
		}

		it("shows the bouldering tracker when activity is bouldering", async () => {
			mockFetch();
			renderTracker();
			await startBouldSession();
			expect(screen.getByText("Bouldering")).toBeInTheDocument();
		});

		it("does not show the bouldering tracker for other activities", async () => {
			mockFetch();
			renderTracker();
			await selectActivityAndStart();
			expect(screen.queryByText("Bouldering")).not.toBeInTheDocument();
		});

		it("prefills notes with serialized entries when End Session is tapped", async () => {
			mockFetch();
			renderTracker();
			await startBouldSession();

			fireEvent.click(screen.getByRole("button", { name: "Log Send" }));
			fireEvent.click(screen.getByRole("button", { name: "Log Attempt" }));
			fireEvent.click(screen.getByRole("button", { name: "End Session" }));

			await waitFor(() =>
				expect(screen.getByTestId("mock-summary-dialog")).toBeInTheDocument(),
			);

			// The mock dialog exposes notesSignal via its rendered textarea.
			// Instead, check the PATCH body includes the serialized summary.
			const fetchSpy = global.fetch;
			await capturedOnSave();
			const patchCall = fetchSpy.mock.calls.find(
				([url, opts]) =>
					opts?.method === "PATCH" && url.includes("/api/sessions/"),
			);
			const body = JSON.parse(patchCall[1].body);
			expect(body.session_notes).toContain("V5");
			expect(body.session_notes).toContain("Attempt");
			expect(body.session_notes).toContain("Send");
		});

		it("appends serialized summary after pre-existing notes", async () => {
			mockFetch();
			renderTracker();
			await startBouldSession();

			fireEvent.click(screen.getByRole("button", { name: "Log Send" }));

			// Type notes before ending the session
			const notesField = screen.getByRole("textbox", { name: /notes/i });
			fireEvent.input(notesField, { target: { value: "Felt strong today" } });

			fireEvent.click(screen.getByRole("button", { name: "End Session" }));
			await waitFor(() =>
				expect(screen.getByTestId("mock-summary-dialog")).toBeInTheDocument(),
			);
			await capturedOnSave();

			const patchCall = global.fetch.mock.calls.find(
				([url, opts]) =>
					opts?.method === "PATCH" && url.includes("/api/sessions/"),
			);
			const body = JSON.parse(patchCall[1].body);
			expect(body.session_notes).toMatch(/^Felt strong today/); // user notes first
			expect(body.session_notes).toContain("- V5"); // summary after
		});

		it("does not set session_notes when no bouldering entries are logged", async () => {
			mockFetch();
			renderTracker();
			await startBouldSession();

			fireEvent.click(screen.getByRole("button", { name: "End Session" }));
			await waitFor(() =>
				expect(screen.getByTestId("mock-summary-dialog")).toBeInTheDocument(),
			);
			await capturedOnSave();

			const patchCall = global.fetch.mock.calls.find(
				([url, opts]) =>
					opts?.method === "PATCH" && url.includes("/api/sessions/"),
			);
			const body = JSON.parse(patchCall[1].body);
			expect(body.session_notes).toBeNull();
		});
	});

	describe("session restore", () => {
		function seedStorage(overrides = {}) {
			localStorage.setItem(
				"cove_active_workout_session",
				JSON.stringify({
					sessionId: 42,
					startedAt: "2026-07-13T10:00:00Z",
					activity: "Climbing",
					sessionPrograms: [],
					checkedItems: {},
					bouldEntries: [],
					notes: "",
					accumulatedS: 0,
					segmentStart: Date.now(),
					running: true,
					...overrides,
				}),
			);
		}

		it("skips Start screen when an active session is in localStorage", async () => {
			mockFetch();
			seedStorage();
			renderTracker();

			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "End Session" }),
				).toBeInTheDocument(),
			);
			expect(
				screen.queryByRole("button", { name: "Start" }),
			).not.toBeInTheDocument();
		});

		it("restores elapsed accounting for time away", async () => {
			vi.useFakeTimers({ toFake: ["setInterval", "clearInterval", "Date"] });
			try {
				const now = Date.now();
				mockFetch();
				seedStorage({ accumulatedS: 300, segmentStart: now - 120_000 });
				renderTracker();

				// 300s accumulated + 120s since segment started = 420s = 7 minutes
				await waitFor(() =>
					expect(screen.getByText("00:07:00")).toBeInTheDocument(),
				);
			} finally {
				vi.useRealTimers();
			}
		});

		it("restores notes", async () => {
			mockFetch();
			seedStorage({ notes: "Felt strong today" });
			renderTracker();

			await waitFor(() =>
				expect(screen.getByRole("textbox", { name: /notes/i })).toHaveValue(
					"Felt strong today",
				),
			);
		});

		it("restores bouldering tracker with logged entries", async () => {
			mockFetch();
			seedStorage({
				activity: "bouldering",
				bouldEntries: [{ id: 1, grade: "V5", labels: [], type: "send" }],
			});
			renderTracker();

			await waitFor(() =>
				expect(screen.getByText("Bouldering")).toBeInTheDocument(),
			);
			// Entry appears in the logged list (grade selector also shows V5, so use getAllByText)
			expect(screen.getAllByText("V5").length).toBeGreaterThanOrEqual(2);
			expect(screen.getByText("Send")).toBeInTheDocument();
		});

		it("restores program checklist exercises", async () => {
			mockFetch();
			seedStorage({
				sessionPrograms: [
					{
						id: 1,
						name: "Test Program",
						sets: [
							{
								id: 1,
								name: "Set A",
								rounds: 1,
								exercises: [{ name: "Pull-up", reps: 5 }],
							},
						],
					},
				],
			});
			renderTracker();

			await waitFor(() =>
				expect(screen.getByText("5x Pull-up")).toBeInTheDocument(),
			);
		});

		it("restores a paused session", async () => {
			mockFetch();
			seedStorage({ accumulatedS: 180, segmentStart: null, running: false });
			renderTracker();

			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "Resume" }),
				).toBeInTheDocument(),
			);
			expect(screen.getByText("00:03:00")).toBeInTheDocument();
		});

		it("clears localStorage on successful save", async () => {
			mockFetch();
			seedStorage();
			renderTracker();

			await waitFor(() =>
				expect(
					screen.getByRole("button", { name: "End Session" }),
				).toBeInTheDocument(),
			);
			fireEvent.click(screen.getByRole("button", { name: "End Session" }));
			await waitFor(() =>
				expect(screen.getByTestId("mock-summary-dialog")).toBeInTheDocument(),
			);
			await capturedOnSave();

			expect(localStorage.getItem("cove_active_workout_session")).toBeNull();
		});
	});

	it("clears saveError when summary dialog is cancelled after a failed save", async () => {
		mockFetch({ patchOk: false });
		renderTracker();
		await startSession();

		// Open the summary dialog.
		fireEvent.click(screen.getByRole("button", { name: "End Session" }));
		await waitFor(() =>
			expect(screen.getByTestId("mock-summary-dialog")).toBeInTheDocument(),
		);

		// Trigger a save attempt that will fail.
		await capturedOnSave();

		// Error should appear inside the open dialog.
		await waitFor(() =>
			expect(screen.getByTestId("dialog-error")).toBeInTheDocument(),
		);

		// Cancel — this should close the dialog AND clear the error.
		capturedOnCancel();

		// The dialog is closed and the error must not appear in the tracker header.
		await waitFor(() => {
			expect(
				screen.queryByTestId("mock-summary-dialog"),
			).not.toBeInTheDocument();
			expect(
				screen.queryByText("Failed to save session"),
			).not.toBeInTheDocument();
		});
	});
});
