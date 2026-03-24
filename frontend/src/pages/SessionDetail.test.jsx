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

// ActivityPicker fetches /api/activities with a module-level cache that persists
// across tests, so mock the component to avoid cross-test contamination.
vi.mock("../components/shared/ActivityPicker.jsx", () => ({
	ActivityPicker: ({ value, onChange }) => (
		<select
			aria-label="Activity"
			value={value ?? ""}
			onChange={(e) => onChange(e.target.value)}
		>
			<option value="">None</option>
			<option value="Run">Run</option>
		</select>
	),
}));

// ─── Fixtures ────────────────────────────────────────────────────────────────

const MOCK_SESSION = {
	id: 42,
	activity: "Run",
	session_notes: "Felt strong",
	duration_s: 2700,
	program_name: "Strength A",
	program_structure: "## Squats\n3x5",
};

const MOCK_USER = { email: "user@example.com", name: "Test User" };

function makeFetch(session = MOCK_SESSION) {
	return vi.spyOn(global, "fetch").mockImplementation((url, opts) => {
		if (opts?.method === "DELETE") {
			return Promise.resolve({ ok: true });
		}
		if (opts?.method === "PATCH") {
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve({ ...session, ...JSON.parse(opts.body) }),
			});
		}
		return Promise.resolve({
			ok: true,
			json: () => Promise.resolve(session),
		});
	});
}

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
		makeFetch();
		renderDetail();
		await waitFor(() =>
			expect(
				screen.getByRole("heading", { name: "Strength A" }),
			).toBeInTheDocument(),
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

	it("renders formatted duration", async () => {
		makeFetch();
		renderDetail();
		await waitFor(() => screen.getByText("45m"));
	});

	it("renders '—' when duration is not set", async () => {
		makeFetch({ ...MOCK_SESSION, duration_s: null });
		renderDetail();
		await waitFor(() =>
			expect(screen.getAllByText("—").length).toBeGreaterThan(0),
		);
	});
});

describe("SessionDetail — duration editing", () => {
	it("shows input when duration is clicked", async () => {
		makeFetch();
		renderDetail();
		await waitFor(() => screen.getByText("45m"));
		fireEvent.click(screen.getByText("45m"));
		expect(screen.getByPlaceholderText("e.g. 1h 30m")).toBeInTheDocument();
	});

	it("pre-fills input with current formatted duration", async () => {
		makeFetch();
		renderDetail();
		await waitFor(() => screen.getByText("45m"));
		fireEvent.click(screen.getByText("45m"));
		expect(screen.getByPlaceholderText("e.g. 1h 30m").value).toBe("45m");
	});

	it("PATCHes session with parsed duration on blur", async () => {
		const fetchSpy = makeFetch();
		renderDetail();
		await waitFor(() => screen.getByText("45m"));
		fireEvent.click(screen.getByText("45m"));
		const input = screen.getByPlaceholderText("e.g. 1h 30m");
		fireEvent.input(input, { target: { value: "1h 30m" } });
		fireEvent.blur(input);
		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/sessions/42",
				expect.objectContaining({
					method: "PATCH",
					body: JSON.stringify({ duration_s: 5400 }),
				}),
			),
		);
	});

	it("cancels edit on Escape without saving", async () => {
		const fetchSpy = makeFetch();
		renderDetail();
		await waitFor(() => screen.getByText("45m"));
		fireEvent.click(screen.getByText("45m"));
		const input = screen.getByPlaceholderText("e.g. 1h 30m");
		fireEvent.input(input, { target: { value: "1h" } });
		fireEvent.keyDown(input, { key: "Escape" });
		expect(
			screen.queryByPlaceholderText("e.g. 1h 30m"),
		).not.toBeInTheDocument();
		const patchCalls = fetchSpy.mock.calls.filter(
			([, opts]) => opts?.method === "PATCH",
		);
		expect(patchCalls).toHaveLength(0);
	});

	it("shows error message when duration PATCH fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "PATCH") {
				return Promise.resolve({ ok: false });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_SESSION),
			});
		});
		renderDetail();
		await waitFor(() => screen.getByText("45m"));
		fireEvent.click(screen.getByText("45m"));
		const input = screen.getByPlaceholderText("e.g. 1h 30m");
		fireEvent.input(input, { target: { value: "2h" } });
		fireEvent.blur(input);
		await waitFor(() =>
			expect(screen.getByText("Save failed")).toBeInTheDocument(),
		);
	});

	it("does not PATCH when duration is unchanged", async () => {
		const fetchSpy = makeFetch();
		renderDetail();
		await waitFor(() => screen.getByText("45m"));
		fireEvent.click(screen.getByText("45m"));
		const input = screen.getByPlaceholderText("e.g. 1h 30m");
		fireEvent.blur(input);
		await waitFor(() =>
			expect(
				screen.queryByPlaceholderText("e.g. 1h 30m"),
			).not.toBeInTheDocument(),
		);
		const patchCalls = fetchSpy.mock.calls.filter(
			([, opts]) => opts?.method === "PATCH",
		);
		expect(patchCalls).toHaveLength(0);
	});
});

describe("SessionDetail — program name editing", () => {
	it("shows program name as a clickable button", async () => {
		makeFetch();
		renderDetail();
		await waitFor(() => screen.getByLabelText("Edit planned program name"));
		expect(
			screen.getByLabelText("Edit planned program name"),
		).toHaveTextContent("Strength A");
	});

	it("shows input when program name is clicked", async () => {
		makeFetch();
		renderDetail();
		await waitFor(() => screen.getByLabelText("Edit planned program name"));
		fireEvent.click(screen.getByLabelText("Edit planned program name"));
		expect(screen.getByPlaceholderText("Program name")).toBeInTheDocument();
	});

	it("PATCHes session with new program name on blur", async () => {
		const fetchSpy = makeFetch();
		renderDetail();
		await waitFor(() => screen.getByLabelText("Edit planned program name"));
		fireEvent.click(screen.getByLabelText("Edit planned program name"));
		const input = screen.getByPlaceholderText("Program name");
		fireEvent.input(input, { target: { value: "Strength B" } });
		fireEvent.blur(input);
		await waitFor(() =>
			expect(fetchSpy).toHaveBeenCalledWith(
				"/api/sessions/42",
				expect.objectContaining({
					method: "PATCH",
					body: JSON.stringify({ program_name: "Strength B" }),
				}),
			),
		);
	});

	it("shows error message when program name PATCH fails", async () => {
		vi.spyOn(global, "fetch").mockImplementation((_url, opts) => {
			if (opts?.method === "PATCH") {
				return Promise.resolve({ ok: false });
			}
			return Promise.resolve({
				ok: true,
				json: () => Promise.resolve(MOCK_SESSION),
			});
		});
		renderDetail();
		await waitFor(() => screen.getByLabelText("Edit planned program name"));
		fireEvent.click(screen.getByLabelText("Edit planned program name"));
		const input = screen.getByPlaceholderText("Program name");
		fireEvent.input(input, { target: { value: "New Name" } });
		fireEvent.blur(input);
		await waitFor(() =>
			expect(screen.getByText("Save failed")).toBeInTheDocument(),
		);
	});

	it("cancels program name edit on Escape without saving", async () => {
		const fetchSpy = makeFetch();
		renderDetail();
		await waitFor(() => screen.getByLabelText("Edit planned program name"));
		fireEvent.click(screen.getByLabelText("Edit planned program name"));
		const input = screen.getByPlaceholderText("Program name");
		fireEvent.input(input, { target: { value: "Other" } });
		fireEvent.keyDown(input, { key: "Escape" });
		expect(
			screen.queryByPlaceholderText("Program name"),
		).not.toBeInTheDocument();
		const patchCalls = fetchSpy.mock.calls.filter(
			([, opts]) => opts?.method === "PATCH",
		);
		expect(patchCalls).toHaveLength(0);
	});
});

describe("SessionDetail — delete", () => {
	it("opens confirm dialog when delete button is clicked", async () => {
		makeFetch();
		renderDetail();
		await waitFor(() => screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByLabelText("Delete session"));
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
		expect(screen.getByText("Delete Session")).toBeInTheDocument();
	});

	it("calls DELETE /api/sessions/:id on confirm", async () => {
		const fetchSpy = makeFetch();
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
		makeFetch();
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
		makeFetch();
		renderDetail();
		await waitFor(() => screen.getByLabelText("Delete session"));
		fireEvent.click(screen.getByLabelText("Delete session"));
		expect(screen.getByTestId("mock-confirm-dialog")).toBeInTheDocument();
		fireEvent.click(screen.getByText("Cancel"));
		expect(screen.queryByTestId("mock-confirm-dialog")).not.toBeInTheDocument();
	});
});
