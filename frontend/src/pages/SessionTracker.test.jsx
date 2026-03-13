// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
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
	Combobox: () => null,
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

async function startSession() {
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
