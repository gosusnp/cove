// Copyright (c) 2026 Jimmy Ma
// SPDX-License-Identifier: Elastic-2.0

import { fireEvent, screen, waitFor } from "@testing-library/preact";
import { afterEach, describe, expect, it, vi } from "vitest";
import { withProviders } from "../test-utils.jsx";
import { SessionDetail } from "./SessionDetail.jsx";

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
						} catch (_) {
							// noop — let the component handle errors
						}
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

// ─── Fixtures ────────────────────────────────────────────────────────────────

const MOCK_SESSION = {
	id: 42,
	activity: "Run",
	session_notes: "Felt strong",
};

const MOCK_USER = { email: "user@example.com", name: "Test User" };

const renderDetail = (props = {}) =>
	withProviders(
		<SessionDetail sessionId={42} onDelete={vi.fn()} {...props} />,
		{ user: MOCK_USER },
	);

// ─── Tests ───────────────────────────────────────────────────────────────────

afterEach(() => {
	vi.restoreAllMocks();
});

describe("SessionDetail — rendering", () => {
	it("shows loading state initially", () => {
		vi.spyOn(global, "fetch").mockReturnValue(new Promise(() => {}));
		renderDetail();
		expect(screen.getByText("Loading…")).toBeInTheDocument();
	});

	it("renders session data after fetch", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_SESSION),
		});
		renderDetail();
		await waitFor(() =>
			expect(screen.getByRole("heading", { name: "Run" })).toBeInTheDocument(),
		);
		expect(screen.getByText("Felt strong")).toBeInTheDocument();
	});

	it("shows error when fetch fails", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({ ok: false });
		renderDetail();
		await waitFor(() =>
			expect(screen.getByText("Failed to fetch session")).toBeInTheDocument(),
		);
	});
});

describe("SessionDetail — delete", () => {
	it("opens confirm dialog when delete button is clicked", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_SESSION),
		});
		renderDetail();
		await waitFor(() => screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByLabelText("Delete session"));
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
		expect(screen.getByText("Delete Session")).toBeInTheDocument();
	});

	it("calls DELETE /api/sessions/:id on confirm", async () => {
		const fetchSpy = vi
			.spyOn(global, "fetch")
			.mockImplementation((_url, opts) => {
				if (opts?.method === "DELETE") {
					return Promise.resolve({ ok: true });
				}
				return Promise.resolve({
					ok: true,
					json: () => Promise.resolve(MOCK_SESSION),
				});
			});

		renderDetail();
		await waitFor(() => screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/sessions/42",
				expect.objectContaining({ method: "DELETE" }),
			),
		);
	});

	it("calls onDelete with the session id after successful delete", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: true });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_SESSION),
			});
		});

		const onDelete = vi.fn();
		renderDetail({ onDelete });
		await waitFor(() => screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() => expect(onDelete).toHaveBeenCalledWith(42));
	});

	it("keeps dialog open and does not call onDelete when delete API fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "DELETE") {
				return Promise.resolve({ ok: false });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_SESSION),
			});
		});

		const onDelete = vi.fn();
		renderDetail({ onDelete });
		await waitFor(() => screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByText("Confirm"));

		await waitFor(() =>
			expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument(),
		);
		expect(onDelete).not.toHaveBeenCalled();
	});

	it("dismisses confirm dialog on cancel", async () => {
		vi.spyOn(global, "fetch").mockResolvedValue({
			ok: true,
			json: () => Promise.resolve(MOCK_SESSION),
		});
		renderDetail();
		await waitFor(() => screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByLabelText("Delete session"));
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
		fireEvent.click(screen.getByText("Cancel"));
		expect(screen.queryByTestId("mock-confirm-dialog")).not.toBeInTheDocument();
	});
});
