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
	ActivityPicker: () => null,
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
	vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
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

// ─── Tests ───────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
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
				// 100 kg → 220.46 lb
				expect(screen.getByText("220.46 lb")).toBeInTheDocument(),
			);
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
